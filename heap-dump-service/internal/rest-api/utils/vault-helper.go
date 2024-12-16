package utils

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/hashicorp/vault/api"
	vault "github.com/hashicorp/vault/api"
	vaultAuth "github.com/hashicorp/vault/api/auth/kubernetes"
	vaultTransit "github.com/mittwald/vaultgo"
	log "github.com/sirupsen/logrus"
)

func CheckVaultAccess(client *vault.Client) error {

	u, err := url.Parse(client.Address() + "v1/auth/token/lookup-self")
	if err != nil {
		return errors.New(fmt.Sprintf("Could not construct URL from client config: %s", err.Error()))
	}

	request := vault.Request{
		Method:      "GET",
		URL:         u,
		ClientToken: client.Token(),
	}
	_, err = client.RawRequest(&request)

	if err != nil {
		return errors.New(fmt.Sprintf("Unable to Access Vault: %s - %s", err.Error(), u.String()))
	}

	return nil
}

func GenerateVaultClient(role string, mountPath string, jwtLocation string) (*api.Client, error) {
	config := api.DefaultConfig()

	k8sAuth, err := vaultAuth.NewKubernetesAuth(
		role,
		vaultAuth.WithServiceAccountTokenPath(jwtLocation),
		vaultAuth.WithMountPath(mountPath),
	)
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "GenerateVaultClient",
		}).Fatalf(fmt.Sprintf("unable to initialize Vault Authentication : %s", err.Error()))
	}

	vanillaVaultclient, err := api.NewClient(config)
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "GenerateVaultClient",
		}).Warn(fmt.Sprintf("unable to initialize Vanilla Vault Client : %s", err.Error()))
		return nil, errors.New(fmt.Sprintf("unable to initialize Vanilla Vault Client : %s", err.Error()))
	}
	authInfo, err := vanillaVaultclient.Auth().Login(context.Background(), k8sAuth)
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "GenerateVaultClient",
		}).Warn(fmt.Sprintf("unable to log in with Kubernetes auth : %s", err.Error()))
		return nil, errors.New(fmt.Sprintf("unable to log in with Kubernetes auth: %s", err.Error()))
	}
	if authInfo == nil {
		log.WithFields(log.Fields{
			"caller": "GenerateVaultClient",
		}).Warn(fmt.Sprintf("no auth info was returned after login"))
		return nil, errors.New(fmt.Sprintf("no auth info was returned after login"))
	}
	return vanillaVaultclient, nil
}

func GenerateTransitVaultClient(role string, mountPath string, jwtLocation string) (*vaultTransit.Client, error) {

	vaultURL, found := os.LookupEnv("VAULT_ADDR")
	if !found {
		log.WithFields(log.Fields{
			"caller": "GenerateTransitVaultClient",
		}).Error(fmt.Sprintf("Could not find valid vault URL on env: %s", "VAULT_ADDR"))
		return nil, errors.New(fmt.Sprintf("Could not find valid vault URL on env: %s", "VAULT_ADDR"))
	}

	c, err := vaultTransit.NewClient(
		vaultURL,
		vaultTransit.WithCaPath(""),
		vaultTransit.WithKubernetesAuth(
			role,
			vaultTransit.WithJwtFromFile(jwtLocation),
			vaultTransit.WithMountPoint(mountPath),
		),
	)
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "GenerateTransitVaultClient",
		}).Error(fmt.Sprintf("Error creating Transit Vault Client: %s", err.Error()))
		return nil, err
	}
	return c, nil
}

func TransitEncryptString(client *vaultTransit.Client, mountPoint string, topicKey string, key string) (string, error) {

	transit := client.TransitWithMountPoint(mountPoint)

	encryptResponse, err := transit.Encrypt(topicKey, &vaultTransit.TransitEncryptOptions{
		Plaintext: key,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"caller": "TransitEncryptString",
		}).Error(fmt.Sprintf("Error occurred during encryption: %s", err.Error()))
		return "", err
	}

	return encryptResponse.Data.Ciphertext, nil
}
