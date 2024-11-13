package utils

import (
	"errors"
	"fmt"
	"io/fs"

	log "github.com/sirupsen/logrus"
)

func CheckError(err error) {
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "CheckError",
		}).Fatalf(err.Error())
	}
}

func GetCurrentNamespace(fileSystem fs.FS) (string, error) {
	namespace, err := fs.ReadFile(fileSystem, "var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not get Namespace: %s", err.Error()))
	}
	return string(namespace), nil
}
