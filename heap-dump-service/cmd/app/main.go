package main

import (
	"fmt"

	"github.com/dbschenker/heap-dump-management/heap-dump-service/internal/config"
	"github.com/dbschenker/heap-dump-management/heap-dump-service/internal/logging"
	"github.com/dbschenker/heap-dump-management/heap-dump-service/internal/metrics"
	restapi "github.com/dbschenker/heap-dump-management/heap-dump-service/internal/rest-api"
	log "github.com/sirupsen/logrus"
)

func main() {
	logging.SetupLogging()

	appConfig, err := config.LoadConfigFromEnvironment("APP_CONFIG_FILE")

	if err != nil {
		log.WithFields(log.Fields{
			"caller": "LoadConfigFromEnvironment",
		}).Fatalf(fmt.Sprintf("Failed to read Config File: %s", err.Error()))
	}

	log.Printf("Configuration loaded. Starting event handler")

	go func() {
		for {
			log.WithFields(log.Fields{
				"caller": "main",
			}).Info("Serving Metrics")
			metrics.StartMetricServer(appConfig.Metrics.Port, appConfig.Metrics.Path)
		}
	}()

	restapi.Serve(&appConfig)
}
