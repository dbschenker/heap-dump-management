package utils

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"testing"
	"testing/fstest"

	"github.com/docker/go-connections/nat"
	"github.com/hashicorp/vault/api"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	vaultTransit "github.com/mittwald/vaultgo"
)

var Token = "test"

var ValidFs = fstest.MapFS{
	"var/var/run/secrets/kubernetes.io/serviceaccount/token": {Data: []byte(Token)},
}

type VaultContainer struct {
	container  testcontainers.Container
	mappedPort nat.Port
	hostIP     string
	token      string
}

var Vault *VaultContainer

func (v *VaultContainer) URI() string {
	return fmt.Sprintf("http://%s:%s/", v.HostIP(), v.Port())
}

func (v *VaultContainer) Port() string {
	return v.mappedPort.Port()
}

func (v *VaultContainer) HostIP() string {
	return v.hostIP
}

func (v *VaultContainer) Token() string {
	return v.token
}

func (v *VaultContainer) Terminate(ctx context.Context) error {
	return v.container.Terminate(ctx)
}

func InitVaultContainer(ctx context.Context, version string) (*VaultContainer, error) {
	port := nat.Port("8200/tcp")

	req := testcontainers.GenericContainerRequest{
		ProviderType: testcontainers.ProviderDocker,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "vault:" + version,
			ExposedPorts: []string{string(port)},
			WaitingFor:   wait.ForListeningPort(port),
			SkipReaper:   true,
			Env: map[string]string{
				"VAULT_ADDR":              fmt.Sprintf("http://0.0.0.0:%s", port.Port()),
				"VAULT_DEV_ROOT_TOKEN_ID": Token,
				"VAULT_TOKEN":             Token,
				"VAULT_LOG_LEVEL":         "trace",
			},
			Cmd: []string{
				"server",
				"-dev",
			},
			Privileged: true,
		},
	}

	v, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req.ContainerRequest,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	vc := &VaultContainer{
		container:  v,
		mappedPort: "",
		hostIP:     "",
		token:      Token,
	}

	vc.hostIP, err = v.Host(ctx)
	if err != nil {
		return nil, err
	}

	vc.mappedPort, err = v.MappedPort(ctx, port)
	if err != nil {
		return nil, err
	}

	_, _, err = vc.container.Exec(ctx, []string{
		"vault",
		"secrets",
		"enable",
		"transit",
	})
	if err != nil {
		return nil, err
	}

	return vc, nil
}

func TestMain(m *testing.M) {
	var err error
	Vault, err = InitVaultContainer(context.Background(), "1.11.4")
	if err != nil {
		fmt.Printf("Could not start test container for vault: %s", err.Error())
	}
	m.Run()
}

func TestVaultEncryptString(t *testing.T) {
	os.Setenv("VAULT_ADDR", Vault.URI())
	defer os.Unsetenv("VAULT_ADDR")
	os.Setenv("VAULT_TOKEN", Vault.Token())
	defer os.Unsetenv("VAULT_TOKEN")
	rnd, err := GenerateRandomBytes(32)
	if err != nil {
		t.Errorf(err.Error())
	}

	testClient, err := vaultTransit.NewClient(Vault.URI(), vaultTransit.WithCaPath(""), vaultTransit.WithAuthToken(Vault.token))
	if err != nil {
		t.Errorf("Error creating test client: %s", err.Error())
	}

	ret, err := TransitEncryptString(testClient, "transit", "test-topic", base64.URLEncoding.EncodeToString(rnd))

	if err != nil {
		t.Errorf("Error encrypting string: %s", err.Error())
	}

	transit := testClient.TransitWithMountPoint("transit")
	plainTestResponse, err := transit.Decrypt("test-topic", &vaultTransit.TransitDecryptOptions{
		Ciphertext: string(ret),
	})

	if plainTestResponse.Data.Plaintext != base64.URLEncoding.EncodeToString(rnd) {
		t.Errorf("Decryption failed! want %v, got %v", string(rnd), plainTestResponse.Data.Plaintext)
	}

}

func TestCheckVaultAccess(t *testing.T) {
	os.Setenv("VAULT_ADDR", Vault.URI())
	defer os.Unsetenv("VAULT_ADDR")
	os.Setenv("VAULT_TOKEN", Vault.Token())
	defer os.Unsetenv("VAULT_TOKEN")
	config := api.DefaultConfig()
	testClient, err := api.NewClient(config)
	if err != nil {
		t.Errorf("Could not initialize vault client: %s", err.Error())
	}
	err = CheckVaultAccess(testClient)
	if err != nil {
		t.Errorf("Failed Health Check: %s", err.Error())
	}
}
