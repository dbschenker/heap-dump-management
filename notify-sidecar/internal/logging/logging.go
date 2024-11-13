package logging

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func SetupLogging() {
	logLevelString, found := os.LookupEnv("NOTIFY_SIDECAR_LOG_LEVEL")
	if !found {
		logLevelString = "WARNING"
	}
	level, err := log.ParseLevel(logLevelString)
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "SetupLogging",
		}).Error(fmt.Sprintf("error parsing %s: %v", logLevelString, err))
		panic(err)
	}
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(level)
}
