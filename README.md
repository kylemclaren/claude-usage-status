![image](https://github.com/user-attachments/assets/8f65d14c-59bc-4beb-a515-faf5377ad91f)


A Claude Code status line widget that displays your API usage with gradient progress bars.

## Features

- Shows 5-hour session and 7-day usage limits
- Gradient progress bars (green → lime → yellow → orange → red)
- Time until next reset
- Reads OAuth credentials from Claude Code CLI
- Fast Go binary with no dependencies

## Installation

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

### Configure Claude Code

Add to your `~/.claude/settings.json`:

```json
{
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

## Requirements

- Claude Code CLI installed and authenticated (`claude` command)
- Go 1.21+ (for building from source)

## Credits

Inspired by [claudecodeusage](https://github.com/richhickson/claudecodeusage) by Rich Hickson.

## License

MIT
