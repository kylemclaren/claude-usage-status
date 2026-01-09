package main

import (
	"encoding/json"
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
