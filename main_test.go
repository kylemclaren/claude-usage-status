package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestGetColorEmoji(t *testing.T) {
	tests := []struct {
		name        string
		utilization float64
		want        string
	}{
		{"zero usage", 0.0, "🟢"},
		{"low usage", 0.45, "🟢"},
		{"below threshold", 0.69, "🟢"},
		{"at yellow threshold", 0.70, "🟡"},
		{"medium usage", 0.78, "🟡"},
		{"at red threshold", 0.90, "🟡"},
		{"high usage", 0.91, "🔴"},
		{"max usage", 1.0, "🔴"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getColorEmoji(tt.utilization)
			if got != tt.want {
				t.Errorf("getColorEmoji(%v) = %v, want %v", tt.utilization, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"negative", -1 * time.Hour, "now"},
		{"zero", 0, "now"},
		{"only minutes", 15 * time.Minute, "15m"},
		{"only hours", 2 * time.Hour, "2h"},
		{"hours and minutes", 2*time.Hour + 15*time.Minute, "2h15m"},
		{"many hours", 12*time.Hour + 30*time.Minute, "12h30m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.d)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %v, want %v", tt.d, got, tt.want)
			}
		})
	}
}

func TestFormatTimeUntilReset(t *testing.T) {
	now := time.Date(2026, 1, 9, 12, 45, 0, 0, time.UTC)

	tests := []struct {
		name    string
		resetAt string
		want    string
	}{
		{"2 hours 15 minutes ahead", "2026-01-09T15:00:00Z", "2h15m"},
		{"past time", "2026-01-09T10:00:00Z", "now"},
		{"invalid format", "invalid", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTimeUntilReset(tt.resetAt, now)
			if got != tt.want {
				t.Errorf("formatTimeUntilReset(%v) = %v, want %v", tt.resetAt, got, tt.want)
			}
		})
	}
}

func TestFormatStatusLine(t *testing.T) {
	now := time.Date(2026, 1, 9, 12, 45, 0, 0, time.UTC)

	tests := []struct {
		name  string
		usage *UsageResponse
		want  string
	}{
		{
			name: "low usage both",
			usage: &UsageResponse{
				FiveHour: UsageBucket{Utilization: 0.45, ResetsAt: "2026-01-09T15:00:00Z"},
				SevenDay: UsageBucket{Utilization: 0.30, ResetsAt: "2026-01-16T00:00:00Z"},
			},
			want: "🟢 5h:45% | 🟢 7d:30% | resets 2h15m",
		},
		{
			name: "mixed usage",
			usage: &UsageResponse{
				FiveHour: UsageBucket{Utilization: 0.45, ResetsAt: "2026-01-09T15:00:00Z"},
				SevenDay: UsageBucket{Utilization: 0.78, ResetsAt: "2026-01-16T00:00:00Z"},
			},
			want: "🟢 5h:45% | 🟡 7d:78% | resets 2h15m",
		},
		{
			name: "high usage both",
			usage: &UsageResponse{
				FiveHour: UsageBucket{Utilization: 0.95, ResetsAt: "2026-01-09T15:00:00Z"},
				SevenDay: UsageBucket{Utilization: 0.92, ResetsAt: "2026-01-16T00:00:00Z"},
			},
			want: "🔴 5h:95% | 🔴 7d:92% | resets 2h15m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatStatusLine(tt.usage, now)
			if got != tt.want {
				t.Errorf("formatStatusLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
