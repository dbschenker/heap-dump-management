package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

type AppConfig struct {
	Metrics struct {
		Port int
		Path string
	}
	WatchPath struct {
		Path string
	}
	Middleware struct {
		Endpoint string
	}
	ServiceOwner struct {
		Tenant string
	}
}

func LoadConfigFromEnvironment(envVarName string) (AppConfig, error) {
	configFile, found := os.LookupEnv(envVarName)
	var appConfig AppConfig
	if !found {
		log.WithFields(log.Fields{
			"caller": "LoadConfigFromEnvironment",
		}).Warnf(fmt.Sprintf("Environment variable for config file not set: %s", envVarName))
		return appConfig, errors.New(fmt.Sprintf("Environment variable for config file not set: %s", envVarName))
	}
	return LoadConfigFromFile(configFile)
}

func LoadConfigFromFile(configFile string) (AppConfig, error) {
	jsonData, err := os.ReadFile(configFile)
	var appConfig AppConfig
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "LoadConfigFromEnvironment",
		}).Warnf(fmt.Sprintf("Failed to load config file '%v': %v", configFile, err))
		return appConfig, errors.New(fmt.Sprintf("Failed to load config file '%v': %v", configFile, err.Error()))
	}

	d := json.NewDecoder(strings.NewReader(string(jsonData)))
	d.DisallowUnknownFields()
	err = d.Decode(&appConfig)
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "LoadConfigFromEnvironment",
		}).Warnf(fmt.Sprintf("Failed to parse json data of file '%v': %v", configFile, err))
		return appConfig, errors.New(fmt.Sprintf("Failed to parse json data of file '%v': %v", configFile, err.Error()))
	}
	return appConfig, nil
}
