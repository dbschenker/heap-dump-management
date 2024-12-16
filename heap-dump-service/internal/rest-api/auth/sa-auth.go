package auth

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dbschenker/heap-dump-management/heap-dump-service/internal/config"
	"github.com/shaj13/go-guardian/v2/auth"
	"github.com/shaj13/go-guardian/v2/auth/strategies/kubernetes"
	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/fifo"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func generateHttpClient() *http.Client {
	var tlsConfig tls.Config

	tlsConfig.InsecureSkipVerify = true

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = &tlsConfig
	t.MaxIdleConnsPerHost = 100
	t.TLSHandshakeTimeout = 10 * time.Second

	httpClient := &http.Client{
		Transport: t,
	}
	return httpClient
}

func readTokenFromFile(filepath string) (string, error) {
	jwt, err := os.ReadFile(filepath)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Unable to read file containing service account token: %s", err.Error()))
	}
	return string(jwt), nil
}

func setupGoGuardian(token string) auth.Strategy {
	cacheObj := libcache.FIFO.NewUnsafe(10)
	cacheObj.SetTTL(time.Minute * 5)
	kubAuthAddr := kubernetes.SetAddress("https://kubernetes.default.svc")
	httpClient := generateHttpClient()
	kubeClientConfig := kubernetes.SetHTTPClient(httpClient)
	authToken := kubernetes.SetServiceAccountToken(token)
	return kubernetes.New(cacheObj, kubAuthAddr, kubeClientConfig, authToken)
}

func SaAuth(c *gin.Context) {
	log.WithFields(log.Fields{
		"caller": "SaAuth",
	}).Info("Handling request")

	cfg := c.MustGet("cfg").(*config.AppConfig)

	token, err := readTokenFromFile(cfg.ServiceAccount.JWTokenMountPoint)
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "SaAuth",
		}).Errorf("Authentication Failure: %s", err.Error())
		c.JSON(http.StatusForbidden, gin.H{"error": "Authentication Failure"})
		c.Abort()
		c.Writer.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
		return
	}

	strategy := setupGoGuardian(token)

	_, err = strategy.Authenticate(c, c.Request)
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "SaAuth",
		}).Errorf("Authentication Failure: %s", err.Error())
		c.JSON(http.StatusForbidden, gin.H{"error": "Authentication Failure"})
		c.Abort()
		c.Writer.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
		return
	}
}
