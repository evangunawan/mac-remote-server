#!/usr/bin/env bash
set -e

INSTALL_DIR="$HOME/.local/bin"
TARGET_BINARY="$INSTALL_DIR/mac-remote-server"

VERSION=$(git describe --tags 2>/dev/null || cat VERSION 2>/dev/null || echo "v0.1.0")

echo "Building mac-remote-server ($VERSION)..."
go build -ldflags "-X main.Version=${VERSION}" -o mac-remote-server ./cmd/server

mkdir -p "$INSTALL_DIR"

WAS_RUNNING=false
if [ -f "$TARGET_BINARY" ]; then
    if "$TARGET_BINARY" stop >/dev/null 2>&1; then
        WAS_RUNNING=true
        echo "Stopped running background server for update..."
    fi
fi

echo "Installing binary to $TARGET_BINARY..."
rm -f "$TARGET_BINARY"
cp mac-remote-server "$TARGET_BINARY"
chmod +x "$TARGET_BINARY"

echo "Successfully installed mac-remote-server ($VERSION) to $TARGET_BINARY!"

if [ "$WAS_RUNNING" = true ]; then
    echo "Restarting background server..."
    "$TARGET_BINARY" start -d
fi

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "Note: $INSTALL_DIR is not in your PATH."
    echo "Add the following to your ~/.zshrc or ~/.bash_profile:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
fi
