package logging

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func SetupLogging() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}
