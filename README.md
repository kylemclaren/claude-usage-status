![image](https://github.com/user-attachments/assets/8f65d14c-59bc-4beb-a515-faf5377ad91f)


A Claude Code status line widget that displays your API usage with gradient progress bars.

## Features

- Shows 5-hour session and 7-day usage limits
- Gradient progress bars (green → lime → yellow → orange → red)
- Time until next reset
- Reads OAuth credentials from Claude Code CLI
- Fast Go binary with no dependencies

## Installation

### Quick Install (Recommended)

Run the install script - works on **Linux**, **macOS**, and **Windows** (WSL/Git Bash):

```bash
curl -fsSL https://raw.githubusercontent.com/kylemclaren/claude-usage-status/main/install.sh | bash
```

The script will:
- Detect your OS and architecture
- Download the latest release
- Install to `~/.claude/`
- Configure your Claude Code settings

### Manual Download

Download the binary for your platform from [Releases](https://github.com/kylemclaren/claude-usage-status/releases):

| Platform | Binary |
|----------|--------|
| Linux (x64) | `claude-usage-status-linux-amd64` |
| Linux (ARM64) | `claude-usage-status-linux-arm64` |
| macOS (Intel) | `claude-usage-status-darwin-amd64` |
| macOS (Apple Silicon) | `claude-usage-status-darwin-arm64` |
| Windows (x64) | `claude-usage-status-windows-amd64.exe` |

Then install manually:

```bash
# Download binary (example for Linux x64)
curl -fsSL https://github.com/kylemclaren/claude-usage-status/releases/latest/download/claude-usage-status-linux-amd64 -o ~/.claude/claude-usage-status

# Download statusline script
curl -fsSL https://github.com/kylemclaren/claude-usage-status/releases/latest/download/statusline.sh -o ~/.claude/statusline.sh

# Make executable
chmod +x ~/.claude/claude-usage-status ~/.claude/statusline.sh
```

### From Source

```bash
# Clone the repo
git clone https://github.com/kylemclaren/claude-usage-status.git
cd claude-usage-status

# Build
go build -o claude-usage-status

# Install to ~/.claude
cp claude-usage-status ~/.claude/
cp output/statusline.sh ~/.claude/
chmod +x ~/.claude/statusline.sh ~/.claude/claude-usage-status
```

### Homebrew (macOS/Linux)

Coming soon!

## Configuration

Add to your `~/.claude/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "~/.claude/statusline.sh"
  }
}
```

Or merge with existing settings:

```json
{
  "permissions": {
    "defaultMode": "bypassPermissions"
  },
  "statusLine": {
    "type": "command",
    "command": "~/.claude/statusline.sh"
  }
}
```

## Output

The status line displays:

```
5h ██████░░░░ 69% │ 7d ██████░░░░ 62% │ ⏱ 2h39m
```

- **5h**: 5-hour rolling session usage
- **7d**: 7-day rolling usage
- **⏱**: Time until 5-hour limit resets

### Color Thresholds

| Usage | Bar Color | Label Color |
|-------|-----------|-------------|
| < 70% | Green/Lime | Green |
| 70-90% | Yellow/Orange | Yellow |
| > 90% | Red | Red |

## How It Works

This tool uses the undocumented Anthropic API endpoint at `api.anthropic.com/api/oauth/usage` to fetch usage data. It reads your OAuth credentials from `~/.claude/.credentials.json` (created when you log in with `claude` CLI).

> **Note**: This uses an undocumented API that could change at any time.

## Troubleshooting

### "Usage unavailable" in status line

1. Make sure you're logged in to Claude Code: `claude`
2. Check credentials exist: `cat ~/.claude/.credentials.json`
3. Test the binary directly: `~/.claude/claude-usage-status`

### Binary not found

Make sure the binary is in `~/.claude/` and the path in `statusline.sh` is correct.

### Permission denied

```bash
chmod +x ~/.claude/claude-usage-status ~/.claude/statusline.sh
```

## Requirements

- Claude Code CLI installed and authenticated (`claude` command)
- Go 1.21+ (for building from source only)

## Credits

Inspired by [claudecodeusage](https://github.com/richhickson/claudecodeusage) by Rich Hickson.

## License

MIT
