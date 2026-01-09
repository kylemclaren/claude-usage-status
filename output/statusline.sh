#!/bin/bash

# statusline.sh - Wrapper script for claude-usage-status
# Called by Claude Code to display usage stats in the status line
#
# Claude Code passes JSON context via stdin which we read and discard

# Read and discard stdin (required by Claude Code)
cat > /dev/null

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Path to the Go binary (same directory as this script)
BINARY="${SCRIPT_DIR}/claude-usage-status"

# Run the binary and capture output
if [ -x "$BINARY" ]; then
    output=$("$BINARY" 2>&1)
    exit_code=$?

    if [ $exit_code -eq 0 ]; then
        echo "$output"
    else
        echo "⚠️ Usage unavailable"
    fi
else
    echo "⚠️ Usage unavailable"
fi
