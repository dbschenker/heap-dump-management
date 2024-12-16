package main

import (
	"container/list"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dbschenker/heap-dump-management/notify-sidecar/internal/config"
	"github.com/dbschenker/heap-dump-management/notify-sidecar/internal/logging"
	"github.com/dbschenker/heap-dump-management/notify-sidecar/internal/metrics"
	"github.com/dbschenker/heap-dump-management/notify-sidecar/internal/models"
	"github.com/dbschenker/heap-dump-management/notify-sidecar/internal/utils"
)

func handleNewHeapDump(fileSystem fs.FS, cfg config.AppConfig, file string) error {
	response := new(models.SigningResponse)
	err := utils.RequestUploadConfig(fileSystem, cfg, file, response)
	if err != nil {
		return errors.New(fmt.Sprintf("Error requesting upload URL: %s", err.Error()))
	}
	key, err := base64.StdEncoding.DecodeString(response.AesKey)
	if err != nil {
		return errors.New(fmt.Sprintf("Error decoding aes key: %s", err.Error()))
	}
	entryptedFileLocation, err := utils.EncryptDump(fileSystem, strings.TrimPrefix(file, "/"), key)
	if err != nil {
		return errors.New(fmt.Sprintf("Error encrypting dump: %s", err.Error()))
	}
	encryptDumpFileHandler, err := os.Open(entryptedFileLocation)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading %s: %s", entryptedFileLocation, err.Error()))
	}
	encryptedKeyFile, err := os.CreateTemp("/tmp", "key")
	if err != nil {
		return errors.New(fmt.Sprintf("Error creating tmp file %s: %s", encryptedKeyFile.Name(), err.Error()))
	}
	_, err = encryptedKeyFile.WriteString(response.EncryptedAesKey)

	if err != nil {
		return errors.New(fmt.Sprintf("Error writing encrypted AesKey to file %s: %s", encryptedKeyFile.Name(), err.Error()))
	}
	err = utils.UploadToS3(response.URL, encryptDumpFileHandler)
	if err != nil {
		return err
	}

	encryptedKeyFileHandler, err := os.Open(encryptedKeyFile.Name())

	defer encryptedKeyFileHandler.Close()
	defer os.Remove(encryptedKeyFile.Name())
	defer os.Remove(entryptedFileLocation)
	defer os.Remove(file)

	if err != nil {
		return errors.New(fmt.Sprintf("Error Creating FileHandler for %s: %s", encryptedKeyFile.Name(), err.Error()))
	}

	err = utils.UploadToS3(response.EncryptedAesKeyURL, encryptedKeyFileHandler)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"caller": "handleNewHeapDump",
	}).Info(fmt.Sprintf("Uploaded encrypted Heap dump for %s successfully", cfg.ServiceOwner.Tenant))

	metrics.HeapDumpHandled.WithLabelValues(cfg.ServiceOwner.Tenant).Inc()

	return nil
}

func cleanupStaleFiles(basePath string, staleFiles *list.List, appConfig config.AppConfig) {
	for e := staleFiles.Front(); e != nil; e = e.Next() {
		f := e.Value.(fs.FileInfo)
		if time.Now().Sub(f.ModTime()) > time.Minute*5 {
			log.WithFields(log.Fields{
				"caller": "pollChanges",
			}).Info(fmt.Sprintf("Deleting %s", f.Name()))
			os.Remove(path.Join(basePath, f.Name()))
			staleFiles.Remove(e)
			metrics.FailedDumps.WithLabelValues(appConfig.ServiceOwner.Tenant).Inc()
		}
	}
}

func pollChanges(appConfig config.AppConfig) {
	c := time.Tick(10 * time.Second)
	var fzise int64 = 0
	staleFiles := list.New()
	for range c {
		log.WithFields(log.Fields{
			"caller": "pollChanges",
		}).Info(fmt.Sprintf("Checking for new files at %s", time.Now()))
		files, err := ioutil.ReadDir(appConfig.WatchPath.Path)
		if err != nil {
			log.WithFields(log.Fields{
				"caller": "pollChanges",
			}).Fatal(fmt.Sprintf("Could not read files in dir: %s", err.Error()))
		}

		for _, file := range files {
			//defer os.Remove(path.Join(appConfig.WatchPath.Path, file.Name()))
			log.WithFields(log.Fields{
				"caller": "pollChanges",
			}).Debug(fmt.Sprintf("Checking Heap Dump: %s modified at %v, with size %d", file.Name(), file.ModTime(), file.Size()))
			// Flag file for deletion
			if time.Now().Sub(file.ModTime()) > time.Minute*1 {
				log.WithFields(log.Fields{
					"caller": "pollChanges",
				}).Info(fmt.Sprintf("Flagging %s for deletion", file.Name()))
				staleFiles.PushBack(file)
			}
			// Ignore files smaller than 16 MB
			if file.Size() < 16777216 {
				log.WithFields(log.Fields{
					"caller": "pollChanges",
				}).Debug(fmt.Sprintf("%s is too small for a Heap Dump: %d", file.Name(), file.Size()))
			} else {
				if time.Now().Sub(file.ModTime()) > time.Second*15 && fzise == file.Size() {
					log.WithFields(log.Fields{
						"caller": "pollChanges",
					}).Info(fmt.Sprintf("Processing file: %s modified at %v, with size %d", file.Name(), file.ModTime(), file.Size()))
					err := handleNewHeapDump(os.DirFS("/"), appConfig, fmt.Sprintf("%s/%s", appConfig.WatchPath.Path, file.Name()))
					utils.CheckError(err)
				}
			}
			cleanupStaleFiles(appConfig.WatchPath.Path, staleFiles, appConfig)
			log.WithFields(log.Fields{
				"caller": "pollChanges",
			}).Debug(fmt.Sprintf("Write operation still in progress for Heap Dump %s", file.Name()))
			fzise = file.Size()
			time.Sleep(time.Second * 2)
		}
	}
}

func main() {
	logging.SetupLogging()

	appConfig, err := config.LoadConfigFromEnvironment("APP_CONFIG_FILE")
	utils.CheckError(err)

	go pollChanges(appConfig)

	metrics.StartMetricServer(appConfig.Metrics.Port, appConfig.Metrics.Path)
}
