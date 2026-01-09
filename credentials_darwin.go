//go:build darwin

package main

import (
	"encoding/json"
	"fmt"

	"github.com/keybase/go-keychain"
)

// readCredentialsFromKeychain reads credentials from macOS Keychain
func readCredentialsFromKeychain() (string, error) {
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService("Claude Code-credentials")
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)

	results, err := keychain.QueryItem(query)
	if err != nil {
		return "", fmt.Errorf("failed to query Keychain: %w", err)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no credentials found in Keychain")
	}

	jsonStr := string(results[0].Data)

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
