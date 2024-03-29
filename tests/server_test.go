package tests

import (
	"encoding/json"
	"github.com/rovergulf/busybox/handler"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"testing"
	"time"
)

func init() {
	viper.SetDefault("listen_addr", ":8081")
	_ = runTestServer()
}

func runTestServer() *handler.Handler {
	viper.AutomaticEnv()
	h := new(handler.Handler)
	go func() {
		if err := h.Run(); err != nil {
			{
				log.Fatalf("Unable to run server: %s", err)
			}
		}
	}()
	return h
}

func TestServerHealth(t *testing.T) {
	// wait until server goroutine is completed to run
	time.Sleep(1 * time.Second)

	res, err := http.Get("http://127.0.0.1:8081/health")
	if err != nil {
		t.Fatalf("Failed to complete request: %s", err)
	}

	var result map[string]any
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&result); err != nil {
		t.Errorf("Unable to unmarshal request response")
	}

	if healthy, ok := result["healthy"].(bool); !ok || !healthy {
		t.Errorf("invalid server health result")
	}

	if _, ok := result["timestamp"].(string); !ok {
		t.Errorf("Invalid server timestamp result")
	}
}

func TestServerDebugRequest(t *testing.T) {
	res, err := http.Get("http://127.0.0.1:8081/debug")
	if err != nil {
		t.Fatalf("Failed to complete request: %s", err)
	}

	var result map[string]any
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&result); err != nil {
		t.Errorf("Unable to unmarshal request response")
	}
}
