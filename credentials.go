package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Credentials represents the structure of credentials JSON
type Credentials struct {
	ClaudeAIOAuth struct {
		AccessToken string `json:"accessToken"`
	} `json:"claudeAiOauth"`
}

func getCredentialsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", ".credentials.json")
}

// readCredentialsFromFile reads credentials from ~/.claude/.credentials.json
func readCredentialsFromFile() (string, error) {
	credPath := getCredentialsPath()
	if credPath == "" {
		return "", fmt.Errorf("could not determine home directory")
	}

	data, err := os.ReadFile(credPath)
	if err != nil {
		return "", err
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", fmt.Errorf("failed to parse credentials: %w", err)
	}

	if creds.ClaudeAIOAuth.AccessToken == "" {
		return "", fmt.Errorf("no access token found in credentials")
	}

	return creds.ClaudeAIOAuth.AccessToken, nil
}
