//go:build darwin

package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// readCredentialsFromKeychain reads credentials from macOS Keychain using security command
func readCredentialsFromKeychain() (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", "Claude Code-credentials",
		"-w")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to read from Keychain: %w", err)
	}

	jsonStr := strings.TrimSpace(string(output))

	var creds Credentials
	if err := json.Unmarshal([]byte(jsonStr), &creds); err != nil {
		return "", fmt.Errorf("failed to parse Keychain credentials: %w", err)
	}

	if creds.ClaudeAIOAuth.AccessToken == "" {
		return "", fmt.Errorf("no access token found in Keychain credentials")
	}

	return creds.ClaudeAIOAuth.AccessToken, nil
}

// readCredentials tries file first, then Keychain on macOS
func readCredentials() (string, error) {
	// Try file-based credentials first
	token, err := readCredentialsFromFile()
	if err == nil {
		return token, nil
	}

	// Try Keychain
	token, err = readCredentialsFromKeychain()
	if err == nil {
		return token, nil
	}

	return "", fmt.Errorf("credentials not found in file or Keychain: %w", err)
}
