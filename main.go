package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Credentials represents the structure of ~/.claude/.credentials.json
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

func readCredentials() (string, error) {
	credPath := getCredentialsPath()
	if credPath == "" {
		return "", fmt.Errorf("could not determine home directory")
	}

	data, err := os.ReadFile(credPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("credentials not found at %s", credPath)
		}
		return "", fmt.Errorf("failed to read credentials: %w", err)
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

func main() {
	token, err := readCredentials()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// For now, just confirm we read the token (don't print it for security)
	fmt.Printf("Successfully read access token (%d chars)\n", len(token))
}
