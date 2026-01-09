package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

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

// Color definitions
var (
	greenColor  = lipgloss.Color("#00ff00")
	yellowColor = lipgloss.Color("#ffff00")
	redColor    = lipgloss.Color("#ff0000")
)

// getGradientEndColor returns an end color that smoothly interpolates from green to red
// based on the fill percentage, creating a "scaled gradient" effect
func getGradientEndColor(pct float64) string {
	t := pct / 100
	r := int(255 * t)
	g := int(255 * (1 - t))
	return fmt.Sprintf("#%02x%02x00", r, g)
}

// createProgressBar creates a gradient progress bar with smooth color scaling
func createProgressBar(percent float64, width int) string {
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}

	endColor := getGradientEndColor(percent)
	prog := progress.New(
		progress.WithColorProfile(termenv.TrueColor),
		progress.WithGradient("#00ff00", endColor),
		progress.WithWidth(width),
		progress.WithoutPercentage(),
	)

	return prog.ViewAs(percent / 100)
}

// getLabelStyle returns a lipgloss style for the percentage label based on usage
func getLabelStyle(percent float64) lipgloss.Style {
	var color lipgloss.Color
	if percent < 70 {
		color = greenColor
	} else if percent < 90 {
		color = yellowColor
	} else {
		color = redColor
	}
	return lipgloss.NewStyle().Foreground(color)
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
func formatStatusLine(usage *UsageResponse, now time.Time) string {
	fiveHourPct := usage.FiveHour.Utilization
	sevenDayPct := usage.SevenDay.Utilization

	fiveHourBar := createProgressBar(fiveHourPct, 10)
	sevenDayBar := createProgressBar(sevenDayPct, 10)

	fiveHourLabel := getLabelStyle(fiveHourPct).Render(fmt.Sprintf("%d%%", int(fiveHourPct)))
	sevenDayLabel := getLabelStyle(sevenDayPct).Render(fmt.Sprintf("%d%%", int(sevenDayPct)))

	fiveHourReset := formatTimeUntilReset(usage.FiveHour.ResetsAt, now)

	return fmt.Sprintf("5h %s %s │ 7d %s %s │ ⏱ %s",
		fiveHourBar, fiveHourLabel,
		sevenDayBar, sevenDayLabel,
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
