package main

import (
	"io"
	"net/http"
	"testing"
	"time"
)

func TestHealthEndpoint(t *testing.T) {
	// Wait a bit for services
	time.Sleep(5 * time.Second)

	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		t.Fatalf("Failed to get health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	t.Logf("Health endpoint response: %s", string(body))
	t.Log("✅ Health endpoint test passed!")
}
