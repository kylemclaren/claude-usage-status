package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

// readCredentialsFromFile reads credentials from ~/.claude/.credentials.json (Linux)
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

// readCredentialsFromKeychain reads credentials from macOS Keychain
func readCredentialsFromKeychain() (string, error) {
	// Use security command to read from Keychain
	cmd := exec.Command("security", "find-generic-password",
		"-s", "Claude Code-credentials",
		"-w")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to read from Keychain: %w", err)
	}

	// The output is the password (JSON string), parse it
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

// readCredentials tries to read credentials from file first, then Keychain (macOS)
func readCredentials() (string, error) {
	// Try file-based credentials first (Linux and some configurations)
	token, err := readCredentialsFromFile()
	if err == nil {
		return token, nil
	}

	// On macOS, try Keychain
	if runtime.GOOS == "darwin" {
		token, err := readCredentialsFromKeychain()
		if err == nil {
			return token, nil
		}
		return "", fmt.Errorf("credentials not found in file or Keychain: %w", err)
	}

	// Return the original file error for non-macOS
	credPath := getCredentialsPath()
	return "", fmt.Errorf("credentials not found at %s", credPath)
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

// ANSI color codes for gradient progress bar
const (
	reset   = "\033[0m"
	// Gradient colors from green to yellow to red
	green   = "\033[38;5;46m"  // bright green
	lime    = "\033[38;5;118m" // lime
	yellow  = "\033[38;5;226m" // yellow
	orange  = "\033[38;5;208m" // orange
	red     = "\033[38;5;196m" // red
	dimGray = "\033[38;5;240m" // dim gray for empty bar
)

// getGradientColor returns ANSI color based on position in bar (0.0-1.0)
func getGradientColor(position float64) string {
	if position < 0.5 {
		return green
	} else if position < 0.7 {
		return lime
	} else if position < 0.85 {
		return yellow
	} else if position < 0.95 {
		return orange
	}
	return red
}

// renderProgressBar creates a gradient progress bar with ANSI colors
// width is the total character width of the bar
func renderProgressBar(percent float64, width int) string {
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}

	filled := int(float64(width) * percent / 100)
	empty := width - filled

	var bar string

	// Filled portion with gradient
	for i := 0; i < filled; i++ {
		position := float64(i) / float64(width)
		color := getGradientColor(position)
		bar += color + "█"
	}

	// Empty portion
	if empty > 0 {
		bar += dimGray
		for i := 0; i < empty; i++ {
			bar += "░"
		}
	}

	bar += reset
	return bar
}

// getLabelColor returns color for the percentage label based on usage
func getLabelColor(percent float64) string {
	if percent < 70 {
		return green
	} else if percent < 90 {
		return yellow
	}
	return red
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

// formatStatusLine formats the usage data as a single status line with gradient progress bars
// Note: API returns utilization as percentage (0-100), not decimal (0-1)
func formatStatusLine(usage *UsageResponse, now time.Time) string {
	fiveHourPct := usage.FiveHour.Utilization
	sevenDayPct := usage.SevenDay.Utilization

	// Create gradient progress bars (10 chars wide each)
	fiveHourBar := renderProgressBar(fiveHourPct, 10)
	sevenDayBar := renderProgressBar(sevenDayPct, 10)

	// Color the percentage labels
	fiveHourColor := getLabelColor(fiveHourPct)
	sevenDayColor := getLabelColor(sevenDayPct)

	// Use the soonest reset time
	fiveHourReset := formatTimeUntilReset(usage.FiveHour.ResetsAt, now)

	return fmt.Sprintf("5h %s %s%d%%%s │ 7d %s %s%d%%%s │ ⏱ %s",
		fiveHourBar, fiveHourColor, int(fiveHourPct), reset,
		sevenDayBar, sevenDayColor, int(sevenDayPct), reset,
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
