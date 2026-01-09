#!/bin/bash
set -e

# claude-usage-status installer
# Usage: curl -fsSL https://raw.githubusercontent.com/kylemclaren/claude-usage-status/main/install.sh | bash

REPO="kylemclaren/claude-usage-status"
INSTALL_DIR="$HOME/.claude"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            ;;
        *)
            error "Unsupported operating system: $OS"
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    info "Detected platform: $PLATFORM"
}

# Get latest release version
get_latest_version() {
    LATEST_VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$LATEST_VERSION" ]; then
        error "Failed to get latest version"
    fi
    info "Latest version: $LATEST_VERSION"
}

# Download and install
install() {
    BINARY_NAME="claude-usage-status-${PLATFORM}"
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${BINARY_NAME}"
    SCRIPT_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/statusline.sh"

    info "Creating install directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"

    info "Downloading binary..."
    curl -fsSL "$DOWNLOAD_URL" -o "$INSTALL_DIR/claude-usage-status"
    chmod +x "$INSTALL_DIR/claude-usage-status"

    info "Downloading statusline script..."
    curl -fsSL "$SCRIPT_URL" -o "$INSTALL_DIR/statusline.sh"
    chmod +x "$INSTALL_DIR/statusline.sh"

    info "Installation complete!"
}

# Configure Claude Code settings
configure_settings() {
    SETTINGS_FILE="$INSTALL_DIR/settings.json"

    if [ -f "$SETTINGS_FILE" ]; then
        # Check if statusLine is already configured
        if grep -q '"statusLine"' "$SETTINGS_FILE"; then
            warn "statusLine already configured in settings.json"
            warn "Add manually if needed:"
            echo '  "statusLine": {'
            echo '    "type": "command",'
            echo '    "command": "~/.claude/statusline.sh"'
            echo '  }'
            return
        fi

        # Backup existing settings
        cp "$SETTINGS_FILE" "${SETTINGS_FILE}.backup"
        info "Backed up existing settings to ${SETTINGS_FILE}.backup"

        # Add statusLine config (insert before last closing brace)
        if command -v jq &> /dev/null; then
            jq '. + {"statusLine": {"type": "command", "command": "~/.claude/statusline.sh"}}' "$SETTINGS_FILE" > "${SETTINGS_FILE}.tmp"
            mv "${SETTINGS_FILE}.tmp" "$SETTINGS_FILE"
            info "Updated settings.json with statusLine config"
        else
            warn "jq not found. Please add statusLine config manually to $SETTINGS_FILE:"
            echo '  "statusLine": {'
            echo '    "type": "command",'
            echo '    "command": "~/.claude/statusline.sh"'
            echo '  }'
        fi
    else
        # Create new settings file
        cat > "$SETTINGS_FILE" << 'EOF'
{
  "statusLine": {
    "type": "command",
    "command": "~/.claude/statusline.sh"
  }
}
EOF
        info "Created settings.json with statusLine config"
    fi
}

# Verify installation
verify() {
    if [ -x "$INSTALL_DIR/claude-usage-status" ]; then
        info "Binary installed: $INSTALL_DIR/claude-usage-status"
    else
        error "Binary installation failed"
    fi

    if [ -x "$INSTALL_DIR/statusline.sh" ]; then
        info "Script installed: $INSTALL_DIR/statusline.sh"
    else
        error "Script installation failed"
    fi

    echo ""
    echo -e "${GREEN}Installation successful!${NC}"
    echo ""
    echo "To test, run:"
    echo "  ~/.claude/claude-usage-status"
    echo ""
    echo "Restart Claude Code to see the status line."
}

main() {
    echo ""
    echo "╔═══════════════════════════════════════════╗"
    echo "║   claude-usage-status installer           ║"
    echo "╚═══════════════════════════════════════════╝"
    echo ""

    detect_platform
    get_latest_version
    install
    configure_settings
    verify
}

main
