package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialsParsing(t *testing.T) {
	// Create a temporary credentials file
	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, ".credentials.json")

	creds := map[string]interface{}{
		"claudeAiOauth": map[string]interface{}{
			"accessToken": "test-token-12345",
		},
	}

	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("failed to marshal test credentials: %v", err)
	}

	if err := os.WriteFile(credPath, data, 0600); err != nil {
		t.Fatalf("failed to write test credentials: %v", err)
	}

	// Read and parse
	fileData, err := os.ReadFile(credPath)
	if err != nil {
		t.Fatalf("failed to read credentials: %v", err)
	}

	var parsedCreds Credentials
	if err := json.Unmarshal(fileData, &parsedCreds); err != nil {
		t.Fatalf("failed to parse credentials: %v", err)
	}

	if parsedCreds.ClaudeAIOAuth.AccessToken != "test-token-12345" {
		t.Errorf("expected token 'test-token-12345', got '%s'", parsedCreds.ClaudeAIOAuth.AccessToken)
	}
}

func TestEmptyToken(t *testing.T) {
	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, ".credentials.json")

	// Credentials with empty token
	creds := map[string]interface{}{
		"claudeAiOauth": map[string]interface{}{
			"accessToken": "",
		},
	}

	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("failed to marshal test credentials: %v", err)
	}

	if err := os.WriteFile(credPath, data, 0600); err != nil {
		t.Fatalf("failed to write test credentials: %v", err)
	}

	fileData, err := os.ReadFile(credPath)
	if err != nil {
		t.Fatalf("failed to read credentials: %v", err)
	}

	var parsedCreds Credentials
	if err := json.Unmarshal(fileData, &parsedCreds); err != nil {
		t.Fatalf("failed to parse credentials: %v", err)
	}

	if parsedCreds.ClaudeAIOAuth.AccessToken != "" {
		t.Errorf("expected empty token, got '%s'", parsedCreds.ClaudeAIOAuth.AccessToken)
	}
}

func TestFetchUsage(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		// Verify headers
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept: application/json, got %s", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Authorization: Bearer test-token, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("anthropic-beta") != "oauth-2025-04-20" {
			t.Errorf("expected anthropic-beta: oauth-2025-04-20, got %s", r.Header.Get("anthropic-beta"))
		}

		// Return mock response
		response := UsageResponse{
			FiveHour: UsageBucket{
				Utilization: 0.45,
				ResetsAt:    "2026-01-09T15:00:00Z",
			},
			SevenDay: UsageBucket{
				Utilization: 0.78,
				ResetsAt:    "2026-01-16T00:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override the URL for testing
	originalURL := usageAPIURL
	defer func() { _ = originalURL }() // Keep reference to original

	// We need to test with the mock server URL
	// Since usageAPIURL is a const, we'll test the parsing separately
	// and use the mock server to test the full flow

	// Test that we can parse the response format
	testJSON := `{"five_hour":{"utilization":0.45,"resets_at":"2026-01-09T15:00:00Z"},"seven_day":{"utilization":0.78,"resets_at":"2026-01-16T00:00:00Z"}}`
	var usage UsageResponse
	if err := json.Unmarshal([]byte(testJSON), &usage); err != nil {
		t.Fatalf("failed to parse usage response: %v", err)
	}

	if usage.FiveHour.Utilization != 0.45 {
		t.Errorf("expected FiveHour.Utilization 0.45, got %f", usage.FiveHour.Utilization)
	}
	if usage.SevenDay.Utilization != 0.78 {
		t.Errorf("expected SevenDay.Utilization 0.78, got %f", usage.SevenDay.Utilization)
	}
	if usage.FiveHour.ResetsAt != "2026-01-09T15:00:00Z" {
		t.Errorf("expected FiveHour.ResetsAt '2026-01-09T15:00:00Z', got '%s'", usage.FiveHour.ResetsAt)
	}
	if usage.SevenDay.ResetsAt != "2026-01-16T00:00:00Z" {
		t.Errorf("expected SevenDay.ResetsAt '2026-01-16T00:00:00Z', got '%s'", usage.SevenDay.ResetsAt)
	}
}

func TestFetchUsageWithMockServer(t *testing.T) {
	// Create a mock server that returns usage data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept: application/json, got %s", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token-123" {
			t.Errorf("expected Authorization: Bearer test-token-123, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("anthropic-beta") != "oauth-2025-04-20" {
			t.Errorf("expected anthropic-beta: oauth-2025-04-20, got %s", r.Header.Get("anthropic-beta"))
		}

		response := `{"five_hour":{"utilization":0.5,"resets_at":"2026-01-09T15:00:00Z"},"seven_day":{"utilization":0.75,"resets_at":"2026-01-16T00:00:00Z"}}`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Test fetchUsageFromURL helper (we'll add this function)
	usage, err := fetchUsageFromURL(server.URL, "test-token-123")
	if err != nil {
		t.Fatalf("fetchUsageFromURL failed: %v", err)
	}

	if usage.FiveHour.Utilization != 0.5 {
		t.Errorf("expected FiveHour.Utilization 0.5, got %f", usage.FiveHour.Utilization)
	}
	if usage.SevenDay.Utilization != 0.75 {
		t.Errorf("expected SevenDay.Utilization 0.75, got %f", usage.SevenDay.Utilization)
	}
}

func TestFetchUsageAPIError(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid token"}`))
	}))
	defer server.Close()

	_, err := fetchUsageFromURL(server.URL, "bad-token")
	if err == nil {
		t.Error("expected error for unauthorized request")
	}
}
