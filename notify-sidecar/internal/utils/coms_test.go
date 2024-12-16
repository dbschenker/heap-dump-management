package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/dbschenker/heap-dump-management/notify-sidecar/internal/config"
	"github.com/dbschenker/heap-dump-management/notify-sidecar/internal/models"
)

var TestResponse = models.SigningResponse{
	URL:                "http://test-url.org/upload",
	EncryptedAesKey:    "cryptedTest",
	EncryptedAesKeyURL: "http://test-url.org/upload2",
	AesKey:             "test",
}

var ValidFs = fstest.MapFS{
	"var/run/secrets/kubernetes.io/serviceaccount/token":     {Data: []byte("test_token")},
	"var/run/secrets/kubernetes.io/serviceaccount/namespace": {Data: []byte("platform")},
}

var InvalidFs = fstest.MapFS{
	"invalid": {Data: []byte("none")},
}

func TestMain(m *testing.M) {
	httpServer := http.Server{
		Addr: ":21337",
	}
	setup(&httpServer)
	time.Sleep(1000 * time.Millisecond)
	m.Run()
	shutdown(&httpServer)
}

func setup(serverPointer *http.Server) {
	returnGoodJson, _ := json.Marshal(TestResponse)
	http.HandleFunc("/request", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(returnGoodJson))
	})
	go func() {
		if err := serverPointer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe Error: %v", err)
		}
	}()
}

func shutdown(serverPointer *http.Server) {
	if err := serverPointer.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP Server Shutdown Error: %v", err)
	}
}

func TestConstructBearerAuth(t *testing.T) {
	want := "Bearer test_token"
	got, err := constructBearerAuth(ValidFs, "var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		t.Errorf("Failed to construct authHeader: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestConstructRequestBody(t *testing.T) {
	testSystem := "devops"
	testComponent := "platform"
	testFileName := "heapDump"
	testPodName := "TestingPod"
	now := time.Now()

	os.Setenv("POD_NAME", testPodName)

	testData := models.Payload{
		Tenant:    testSystem,
		Namespace: testComponent,
		FileName:  fmt.Sprintf("%s-%s-%s.hprof.crypted", testPodName, testFileName, now.Format("2006-01-02-15-04-05")),
	}

	testBytes, _ := json.Marshal(testData)
	want := bytes.NewReader(testBytes)
	got, err := constructRequestBody(testFileName, testSystem, testComponent, testPodName)
	if err != nil {
		t.Errorf("Failed to construct request body: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestRequestUploadConfig(t *testing.T) {
	testConfig := config.AppConfig{
		Metrics: struct {
			Port int
			Path string
		}{
			Port: 8081,
			Path: "/metrics",
		},
		WatchPath: struct{ Path string }{
			Path: "/test",
		},
		Middleware: struct{ Endpoint string }{
			Endpoint: "http://localhost:21337/request",
		},
		ServiceOwner: struct {
			Tenant string
		}{
			Tenant: "testTenant",
		},
	}

	got := new(models.SigningResponse)

	/*go func() {
		mockMiddleware()
	}()*/

	err := RequestUploadConfig(ValidFs, testConfig, "test", got)

	if err != nil {
		t.Errorf("Error requesting upload config %v", err)
	}
	if !reflect.DeepEqual(*got, TestResponse) {
		t.Errorf("got %+v, want %+v", got, TestResponse)
	}
}

func TestBadRequestUploadConfig(t *testing.T) {
	badTestConfig := config.AppConfig{
		Metrics: struct {
			Port int
			Path string
		}{
			Port: 8081,
			Path: "/metrics",
		},
		WatchPath: struct{ Path string }{
			Path: "/test",
		},
		Middleware: struct{ Endpoint string }{
			Endpoint: "http://localhost:1337/request",
		},
		ServiceOwner: struct {
			Tenant string
		}{
			Tenant: "testTenant",
		},
	}
	testResponseModel := new(models.SigningResponse)

	got := RequestUploadConfig(InvalidFs, badTestConfig, "does not matter", testResponseModel)
	wantNoToken := errors.New(fmt.Sprintf("Error reading SA Token: %s", "open var/run/secrets/kubernetes.io/serviceaccount/token: file does not exist"))

	if got == nil {
		t.Errorf("Request should not be constructed without a valid token!")
	}
	if !reflect.DeepEqual(got.Error(), wantNoToken.Error()) {
		t.Errorf("got %+v, want %+v", got.Error(), wantNoToken.Error())
	}

	got = RequestUploadConfig(ValidFs, badTestConfig, "does not matter", testResponseModel)
	wantNoNetwork := "Error sending request to middleware"

	if got == nil {
		t.Errorf("Network Failure should probagate error!")
	}
	if !(strings.Contains(got.Error(), wantNoNetwork)) {
		t.Errorf("got wrong error %+v, want %+v", got.Error(), wantNoNetwork)
	}

}
