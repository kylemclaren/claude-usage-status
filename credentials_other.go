//go:build !darwin

package main

import "fmt"

// readCredentials reads credentials from file (non-macOS)
func readCredentials() (string, error) {
	token, err := readCredentialsFromFile()
	if err != nil {
		credPath := getCredentialsPath()
		return "", fmt.Errorf("credentials not found at %s: %w", credPath, err)
	}
	return token, nil
}
