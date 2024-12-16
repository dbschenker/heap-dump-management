package main

import (
	"context"
	b64 "encoding/base64"
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

	cfg "github.com/dbschenker/heap-dump-management/notify-sidecar/internal/config"
	"github.com/dbschenker/heap-dump-management/notify-sidecar/internal/models"
)

var staticTestKey = b64.StdEncoding.EncodeToString([]byte{52, 74, 93, 7, 97, 74, 50, 186, 172, 14, 125, 208, 130, 218, 177, 215, 219, 219, 247, 163, 81, 86, 105, 60, 22, 162, 54, 81, 19, 37, 212, 49})

var staticBadTestKey = b64.StdEncoding.EncodeToString([]byte{52, 74})

var BadTestResponse = models.SigningResponse{
	URL:                "http://test-url.org/upload",
	EncryptedAesKey:    "cryptedTest",
	EncryptedAesKeyURL: "http://test-url.org/upload2",
	AesKey:             staticBadTestKey,
}

var GoodTestResponse = models.SigningResponse{
	URL:                "http://test-url.org/upload",
	EncryptedAesKey:    "cryptedTest",
	EncryptedAesKeyURL: "http://test-url.org/upload2",
	AesKey:             staticTestKey,
}

func cleanup(target string) {
	os.Remove(target)
}

func TestMain(m *testing.M) {
	httpServer := http.Server{
		Addr: ":21347",
	}
	setup(&httpServer)
	time.Sleep(1000 * time.Millisecond)
	m.Run()
	shutdown(&httpServer)
}

func setup(serverPointer *http.Server) {
	returnGoodJson, _ := json.Marshal(GoodTestResponse)
	returnBadJson, _ := json.Marshal(BadTestResponse)
	http.HandleFunc("/request/good", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(returnGoodJson))
	})
	http.HandleFunc("/request/bad", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(returnBadJson))
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

func TestNoToken(t *testing.T) {
	goodConfig := cfg.AppConfig{
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
			Endpoint: "http://localhost:21347/request/good",
		},
		ServiceOwner: struct {
			Tenant string
		}{
			Tenant: "testTenant",
		},
	}
	fs := fstest.MapFS{
		"test_heap_dump": {Data: []byte("asdfasdfasdf")},
		"var/run/secrets/kubernetes.io/serviceaccount/namespace": {Data: []byte("platform")},
	}
	want := errors.New(fmt.Sprintf("Error requesting upload URL: %s", "Error reading SA Token: open var/run/secrets/kubernetes.io/serviceaccount/token: file does not exist"))
	got := handleNewHeapDump(fs, goodConfig, "test_heap_dump")
	if got == nil {
		t.Errorf("This should produce an Error")
	}
	if !reflect.DeepEqual(got.Error(), want.Error()) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestBadAESKey(t *testing.T) {
	badConfig := cfg.AppConfig{
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
			Endpoint: "http://localhost:21347/request/bad",
		},
		ServiceOwner: struct {
			Tenant string
		}{
			Tenant: "testTenant",
		},
	}
	fs := fstest.MapFS{
		"var/run/secrets/kubernetes.io/serviceaccount/token":     {Data: []byte("test_token")},
		"var/run/secrets/kubernetes.io/serviceaccount/namespace": {Data: []byte("platform")},
		"test_heap_dump": {Data: []byte("dummy")},
	}
	want := errors.New(fmt.Sprintf("Error encrypting dump: %s", "Error initializing ARE Cipher: crypto/aes: invalid key size 2"))
	got := handleNewHeapDump(fs, badConfig, "test_heap_dump")
	if got == nil {
		t.Errorf("This should produce an Error")
	}
	if !reflect.DeepEqual(got.Error(), want.Error()) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestHandleNewHeapDump(t *testing.T) {
	config := cfg.AppConfig{
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
			Endpoint: "http://localhost:21347/request/good",
		},
		ServiceOwner: struct {
			Tenant string
		}{
			Tenant: "testTenant",
		},
	}
	fs := fstest.MapFS{
		"var/run/secrets/kubernetes.io/serviceaccount/token":     {Data: []byte("test_token")},
		"var/run/secrets/kubernetes.io/serviceaccount/namespace": {Data: []byte("platform")},
		"test_heap_dump": {Data: []byte("dummy")},
	}
	err := handleNewHeapDump(fs, config, "test_heap_dump")
	if err == nil {
		t.Errorf("This should fail!")
	}
	if !(strings.Contains(err.Error(), "Error making request")) {
		t.Errorf("Got error: %s , want: %s", err.Error(), "Error making request: Put \"http://test-url.org/upload\": dial tcp: lookup test-url.org on 10.227.160.2:53: no such host")
	}
	cleanup("test_heap_dump.crypted")
}
