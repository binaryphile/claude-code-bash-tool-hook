#!/bin/bash

# Installation script for claude-code-bash-tool-hook
# Builds the binary and symlinks it to ~/.claude/hooks/

set -e

BINARY_NAME="claude-code-bash-tool-hook"
BIN_DIR="bin"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="$HOME/.claude/hooks"
CONFIG_DIR="$HOME/.claude"
CONFIG_FILE="$CONFIG_DIR/bash-hook-config.json"
SETTINGS_FILE="$CONFIG_DIR/settings.json"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "========================================="
echo "claude-code-bash-tool-hook Installer"
echo "========================================="
echo ""

# Check for Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}ERROR: Go is not installed${NC}"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

echo "Go version: $(go version)"
echo ""

# Build the binary
echo "Building binary..."
cd "$SCRIPT_DIR"
if [ -f "Makefile" ]; then
    make build
else
    mkdir -p "$BIN_DIR"
    go build -o "$BIN_DIR/$BINARY_NAME" -ldflags="-s -w" .
fi

if [ ! -f "$BIN_DIR/$BINARY_NAME" ]; then
    echo -e "${RED}ERROR: Build failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Build successful${NC}"
echo ""

# Create installation directory
echo "Creating installation directory..."
mkdir -p "$INSTALL_DIR"
echo -e "${GREEN}✓ Created $INSTALL_DIR${NC}"
echo ""

# Install symlink
echo "Installing symlink to $INSTALL_DIR/$BINARY_NAME..."
if [ -L "$INSTALL_DIR/$BINARY_NAME" ]; then
    rm "$INSTALL_DIR/$BINARY_NAME"
fi
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo -e "${YELLOW}⚠ Existing binary found, backing up...${NC}"
    mv "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME.bak"
fi
ln -s "$SCRIPT_DIR/$BIN_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
echo -e "${GREEN}✓ Symlinked $BIN_DIR/$BINARY_NAME -> $INSTALL_DIR/$BINARY_NAME${NC}"
echo ""

# Create config file if it doesn't exist
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Creating default config file..."
    mkdir -p "$CONFIG_DIR"
    cat > "$CONFIG_FILE" <<EOF
{
  "enabled": true,
  "debug_log": false,
  "log_file": "",
  "additional_safe_patterns": [],
  "additional_escape_markers": [],
  "force_wrap_patterns": []
}
EOF
    chmod 600 "$CONFIG_FILE"
    echo -e "${GREEN}✓ Created $CONFIG_FILE${NC}"
else
    echo -e "${YELLOW}⚠ Config file already exists: $CONFIG_FILE${NC}"
fi
echo ""

# Check settings.json
echo "========================================="
echo "Next Steps"
echo "========================================="
echo ""

if [ -f "$SETTINGS_FILE" ]; then
    if grep -q "claude-code-bash-tool-hook" "$SETTINGS_FILE" 2>/dev/null; then
        echo -e "${GREEN}✓ Hook already configured in settings.json${NC}"
    else
        echo -e "${YELLOW}⚠ You need to add the hook to settings.json${NC}"
        echo ""
        echo "Add this to $SETTINGS_FILE:"
        echo ""
        cat <<'EOF'
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "hooks": [{
        "type": "command",
        "command": "~/.claude/hooks/claude-code-bash-tool-hook",
        "timeout": 5
      }]
    }]
  }
}
EOF
    fi
else
    echo -e "${YELLOW}⚠ settings.json not found${NC}"
    echo ""
    echo "Create $SETTINGS_FILE with:"
    echo ""
    cat <<'EOF'
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "hooks": [{
        "type": "command",
        "command": "~/.claude/hooks/claude-code-bash-tool-hook",
        "timeout": 5
      }]
    }]
  }
}
EOF
fi

echo ""
echo "========================================="
echo "Installation Complete"
echo "========================================="
echo ""
echo "Binary:  $SCRIPT_DIR/$BIN_DIR/$BINARY_NAME"
echo "Symlink: $INSTALL_DIR/$BINARY_NAME"
echo "Config:  $CONFIG_FILE"
echo ""
echo "To test the hook:"
echo "  $INSTALL_DIR/$BINARY_NAME --test 'ls | grep foo'"
echo ""
echo "To enable debug logging, edit $CONFIG_FILE and set:"
echo '  "debug_log": true'
echo ""
echo "See README.md for configuration options and full documentation."
echo ""
