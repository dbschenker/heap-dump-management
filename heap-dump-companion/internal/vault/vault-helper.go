package vault

import (
	"errors"
	"fmt"
	"os"

	vaultTransit "github.com/mittwald/vaultgo"
	log "github.com/sirupsen/logrus"
)

func GenerateTransitVaultClient() (*vaultTransit.Client, error) {

	vaultURL, found := os.LookupEnv("VAULT_ADDR")
	if !found {
		log.WithFields(log.Fields{
			"caller": "GenerateTransitVaultClient",
		}).Error(fmt.Sprintf("Could not find valid vault URL on env: %s", "VAULT_ADDR"))
		return nil, errors.New(fmt.Sprintf("Could not find valid vault URL on env: %s", "VAULT_ADDR"))
	}

	vaultToken, found := os.LookupEnv("VAULT_TOKEN")
	if !found {
		log.WithFields(log.Fields{
			"caller": "GenerateTransitVaultClient",
		}).Error(fmt.Sprintf("Could not find valid vault token on env: %s", "VAULT_TOKEN"))
		return nil, errors.New(fmt.Sprintf("Could not find valid vault token on env: %s", "VAULT_TOKEN"))
	}

	c, err := vaultTransit.NewClient(
		vaultURL,
		vaultTransit.WithCaPath(""),
		vaultTransit.WithAuthToken(vaultToken),
	)
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "GenerateTransitVaultClient",
		}).Error(fmt.Sprintf("Error creating Transit Vault Client: %s", err.Error()))
		return nil, err
	}
	return c, nil
}

func TransitDecryptString(client *vaultTransit.Client, mountPoint string, topicKey string, cypherText string) (string, error) {

	transit := client.TransitWithMountPoint(mountPoint)
	plainTestResponse, err := transit.Decrypt(topicKey, &vaultTransit.TransitDecryptOptions{
		Ciphertext: cypherText,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "TransitDecryptString",
		}).Error(fmt.Sprintf("Could not decrypt AES Key: %s", err.Error()))
		return "", errors.New(fmt.Sprintf("Could not decrypt AES Key: %s", err.Error()))

	}
	return plainTestResponse.Data.Plaintext, nil
}
