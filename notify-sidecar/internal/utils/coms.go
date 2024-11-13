package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dbschenker/heap-dump-management/notify-sidecar/internal/config"
	"github.com/dbschenker/heap-dump-management/notify-sidecar/internal/models"
)

func constructBearerAuth(fileSystem fs.FS, tokenLocation string) (string, error) {
	sAToken, err := fs.ReadFile(fileSystem, tokenLocation)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error reading SA Token: %s", err.Error()))
	}
	return fmt.Sprintf("Bearer %s", string(sAToken)), nil
}

func constructRequestBody(fileName string, tenant string, namespace string, podName string) (*bytes.Reader, error) {
	t := time.Now()
	data := models.Payload{
		Tenant:    tenant,
		Namespace: namespace,
		FileName:  fmt.Sprintf("%s-%s-%s.hprof.crypted", podName, fileName, t.Format("2006-01-02-15-04-05")),
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error creating middleware request body: %s", err.Error()))
	}
	return bytes.NewReader(payloadBytes), nil

}

func RequestUploadConfig(fileSystem fs.FS, cfg config.AppConfig, fileName string, target interface{}) error {

	bearer, err := constructBearerAuth(fileSystem, "var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return err
	}

	ns, err := GetCurrentNamespace(fileSystem)

	CheckError(err)

	podName := os.Getenv("POD_NAME")

	body, err := constructRequestBody(filepath.Base(fileName), cfg.ServiceOwner.Tenant, ns, podName)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", cfg.Middleware.Endpoint, body)
	req.Header.Add("Authorization", bearer)
	if err != nil {
		return errors.New(fmt.Sprintf("Error creating request to middleware: %s", err.Error()))
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("Error sending request to middleware: %s", err.Error()))
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Middleware replied with error code: %d: %s", resp.StatusCode, b))
	}

	return json.NewDecoder(resp.Body).Decode(target)
}
