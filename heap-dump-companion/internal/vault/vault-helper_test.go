package vault

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	vaultTransit "github.com/mittwald/vaultgo"
)

var Token = "test"

var TestTransitKey = "eyJwb2xpY3kiOnsibmFtZSI6InRlc3Qta2V5Iiwia2V5cyI6eyIxIjp7ImtleSI6InEyRU9ra0JhaTFpYkNtSE1RR3RocGQ4TEJodWs1Z3ZWZVRoRzY0ZW84YkE9IiwiaG1hY19rZXkiOiJrQk9maTlkNjl2L1QvbjYvMGg1WU9hUlRUVUZtejE4Q3ZCMFg0UlJiQ2hJPSIsInRpbWUiOiIyMDIzLTAyLTEzVDE1OjQyOjQ2Ljk3NDU4OTM2MVoiLCJlY194IjpudWxsLCJlY195IjpudWxsLCJlY19kIjpudWxsLCJyc2Ffa2V5IjpudWxsLCJwdWJsaWNfa2V5IjoiIiwiY29udmVyZ2VudF92ZXJzaW9uIjowLCJjcmVhdGlvbl90aW1lIjoxNjc2MzAyOTY2fX0sImRlcml2ZWQiOmZhbHNlLCJrZGYiOjAsImNvbnZlcmdlbnRfZW5jcnlwdGlvbiI6ZmFsc2UsImV4cG9ydGFibGUiOnRydWUsIm1pbl9kZWNyeXB0aW9uX3ZlcnNpb24iOjEsIm1pbl9lbmNyeXB0aW9uX3ZlcnNpb24iOjAsImxhdGVzdF92ZXJzaW9uIjoxLCJhcmNoaXZlX3ZlcnNpb24iOjEsImFyY2hpdmVfbWluX3ZlcnNpb24iOjAsIm1pbl9hdmFpbGFibGVfdmVyc2lvbiI6MCwiZGVsZXRpb25fYWxsb3dlZCI6ZmFsc2UsImNvbnZlcmdlbnRfdmVyc2lvbiI6MCwidHlwZSI6MCwiYmFja3VwX2luZm8iOnsidGltZSI6IjIwMjMtMDItMTNUMTU6NTA6MjcuNzYwODE0ODcyWiIsInZlcnNpb24iOjF9LCJyZXN0b3JlX2luZm8iOm51bGwsImFsbG93X3BsYWludGV4dF9iYWNrdXAiOnRydWUsInZlcnNpb25fdGVtcGxhdGUiOiIiLCJzdG9yYWdlX3ByZWZpeCI6IiIsImF1dG9fcm90YXRlX3BlcmlvZCI6MCwiSW1wb3J0ZWQiOmZhbHNlLCJBbGxvd0ltcG9ydGVkS2V5Um90YXRpb24iOmZhbHNlfSwiYXJjaGl2ZWRfa2V5cyI6eyJrZXlzIjpbeyJrZXkiOm51bGwsImhtYWNfa2V5IjpudWxsLCJ0aW1lIjoiMDAwMS0wMS0wMVQwMDowMDowMFoiLCJlY194IjpudWxsLCJlY195IjpudWxsLCJlY19kIjpudWxsLCJyc2Ffa2V5IjpudWxsLCJwdWJsaWNfa2V5IjoiIiwiY29udmVyZ2VudF92ZXJzaW9uIjowLCJjcmVhdGlvbl90aW1lIjowfSx7ImtleSI6InEyRU9ra0JhaTFpYkNtSE1RR3RocGQ4TEJodWs1Z3ZWZVRoRzY0ZW84YkE9IiwiaG1hY19rZXkiOiJrQk9maTlkNjl2L1QvbjYvMGg1WU9hUlRUVUZtejE4Q3ZCMFg0UlJiQ2hJPSIsInRpbWUiOiIyMDIzLTAyLTEzVDE1OjQyOjQ2Ljk3NDU4OTM2MVoiLCJlY194IjpudWxsLCJlY195IjpudWxsLCJlY19kIjpudWxsLCJyc2Ffa2V5IjpudWxsLCJwdWJsaWNfa2V5IjoiIiwiY29udmVyZ2VudF92ZXJzaW9uIjowLCJjcmVhdGlvbl90aW1lIjoxNjc2MzAyOTY2fV19fQo="

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

	_, _, err = vc.container.Exec(ctx, []string{
		"vault",
		"write",
		"/transit/restore/test-topic",
		fmt.Sprintf("backup=%s", TestTransitKey),
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

func TestGenerateTransitVaultClient(t *testing.T) {
	os.Setenv("VAULT_ADDR", Vault.URI())
	defer os.Unsetenv("VAULT_ADDR")
	os.Setenv("VAULT_TOKEN", Vault.Token())
	defer os.Unsetenv("VAULT_TOKEN")
	client, err := GenerateTransitVaultClient()
	if err != nil {
		t.Errorf("Error creating test client: %s", err.Error())
	}

	if client.Client.Address() != Vault.URI() {
		t.Errorf("Generated Client missconfigured")
	}

	if client.Client.Token() != Vault.Token() {
		t.Errorf("Generated Client missconfigured")
	}
}

func TestGenerateTransitVaultClientNoEnv(t *testing.T) {
	_, err := GenerateTransitVaultClient()
	if err == nil {
		t.Errorf("No VAULT_ADDR or VAULT_TOKEN should generate an error!")
	}
	os.Setenv("VAULT_ADDR", Vault.URI())
	defer os.Unsetenv("VAULT_ADDR")
	_, err = GenerateTransitVaultClient()
	if err == nil {
		t.Errorf("No VAULT_TOKEN should generate an error!")
	}
}

func TestVaultDecryptString(t *testing.T) {
	os.Setenv("VAULT_ADDR", Vault.URI())
	defer os.Unsetenv("VAULT_ADDR")
	os.Setenv("VAULT_TOKEN", Vault.Token())
	defer os.Unsetenv("VAULT_TOKEN")

	testClient, err := vaultTransit.NewClient(Vault.URI(), vaultTransit.WithCaPath(""), vaultTransit.WithAuthToken(Vault.token))
	if err != nil {
		t.Errorf("Error creating test client: %s", err.Error())
	}

	ret, err := TransitDecryptString(testClient, "transit", "test-topic", "vault:v1:QPRzx8HS54xZB2v/7KpBnIojMOulGuudYz12Z2x08rg=")

	if err != nil {
		t.Errorf("Error decrypting string: %s", err.Error())
	}

	if ret != "test" {
		t.Errorf("Decryption failed! want %v, got %v", "test", ret)
	}
}
