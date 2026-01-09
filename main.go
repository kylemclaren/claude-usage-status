package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Credentials represents the structure of ~/.claude/.credentials.json
type Credentials struct {
	ClaudeAIOAuth struct {
		AccessToken string `json:"accessToken"`
	} `json:"claudeAiOauth"`
}

// UsageBucket represents a usage time bucket (five_hour or seven_day)
type UsageBucket struct {
	Utilization float64 `json:"utilization"`
	ResetsAt    string  `json:"resets_at"`
}

// UsageResponse represents the API response from /api/oauth/usage
type UsageResponse struct {
	FiveHour UsageBucket `json:"five_hour"`
	SevenDay UsageBucket `json:"seven_day"`
}

const usageAPIURL = "https://api.anthropic.com/api/oauth/usage"

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

// fetchUsageFromURL fetches usage data from a specified URL (for testing)
func fetchUsageFromURL(url, token string) (*UsageResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch usage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var usage UsageResponse
	if err := json.Unmarshal(body, &usage); err != nil {
		return nil, fmt.Errorf("failed to parse usage response: %w", err)
	}

	return &usage, nil
}

// fetchUsage fetches usage data from the Anthropic API
func fetchUsage(token string) (*UsageResponse, error) {
	return fetchUsageFromURL(usageAPIURL, token)
}

func main() {
	token, err := readCredentials()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	usage, err := fetchUsage(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("5-hour utilization: %.1f%%\n", usage.FiveHour.Utilization*100)
	fmt.Printf("7-day utilization: %.1f%%\n", usage.SevenDay.Utilization*100)
	fmt.Printf("5-hour resets at: %s\n", usage.FiveHour.ResetsAt)
	fmt.Printf("7-day resets at: %s\n", usage.SevenDay.ResetsAt)
}
