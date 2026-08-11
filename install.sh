#!/usr/bin/env bash
set -e

INSTALL_DIR="$HOME/.local/bin"
TARGET_BINARY="$INSTALL_DIR/mac-remote-server"

VERSION=$(git describe --tags 2>/dev/null || cat VERSION 2>/dev/null || echo "v0.1.0")

echo "Building mac-remote-server ($VERSION)..."
go build -ldflags "-X main.Version=${VERSION}" -o mac-remote-server ./cmd/server

mkdir -p "$INSTALL_DIR"

echo "Installing binary to $TARGET_BINARY..."
cp mac-remote-server "$TARGET_BINARY"
chmod +x "$TARGET_BINARY"

echo "Successfully installed mac-remote-server to $TARGET_BINARY!"

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "Note: $INSTALL_DIR is not in your PATH."
    echo "Add the following to your ~/.zshrc or ~/.bash_profile:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
fi
