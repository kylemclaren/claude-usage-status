package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
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

// getColorEmoji returns a colored circle emoji based on utilization percentage
// Green (<70%), Yellow (70-90%), Red (>90%)
func getColorEmoji(utilization float64) string {
	pct := utilization * 100
	if pct < 70 {
		return "🟢"
	} else if pct <= 90 {
		return "🟡"
	}
	return "🔴"
}

// formatDuration formats a duration into human-readable format like "2h15m"
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "now"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 && minutes > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return "now"
}

// formatTimeUntilReset calculates and formats time until the reset timestamp
func formatTimeUntilReset(resetAt string, now time.Time) string {
	resetTime, err := time.Parse(time.RFC3339, resetAt)
	if err != nil {
		return "unknown"
	}

	duration := resetTime.Sub(now)
	return formatDuration(duration)
}

// formatStatusLine formats the usage data as a single status line
func formatStatusLine(usage *UsageResponse, now time.Time) string {
	fiveHourPct := int(usage.FiveHour.Utilization * 100)
	sevenDayPct := int(usage.SevenDay.Utilization * 100)

	fiveHourEmoji := getColorEmoji(usage.FiveHour.Utilization)
	sevenDayEmoji := getColorEmoji(usage.SevenDay.Utilization)

	// Use the soonest reset time
	fiveHourReset := formatTimeUntilReset(usage.FiveHour.ResetsAt, now)

	return fmt.Sprintf("%s 5h:%d%% | %s 7d:%d%% | resets %s",
		fiveHourEmoji, fiveHourPct,
		sevenDayEmoji, sevenDayPct,
		fiveHourReset)
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

	fmt.Println(formatStatusLine(usage, time.Now()))
}
