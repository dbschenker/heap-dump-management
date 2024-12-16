package requests

import (
	"fmt"
	"net/http"

	"github.com/dbschenker/heap-dump-management/heap-dump-service/internal/config"
	"github.com/dbschenker/heap-dump-management/heap-dump-service/internal/rest-api/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Health(c *gin.Context) {
	cfg := c.MustGet("cfg").(*config.AppConfig)
	client, err := utils.GenerateVaultClient(cfg.Vault.VaultRole, cfg.Vault.VaultAuthMountPath, cfg.ServiceAccount.JWTokenMountPoint)
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "Health",
		}).Fatalf(fmt.Sprintf("unable to initialize Vanilla Vault Client : %s", err.Error()))
	}

	err = utils.CheckAWSAccess()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	err = utils.CheckVaultAccess(client)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
