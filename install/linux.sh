#!/usr/bin/env sh

## Installation for Linux

# Install dir
INSTALL_PATH="/usr/local/bin/x-ip"

# Install the executable
curl -s -o "$INSTALL_PATH" https://github.com/xip-online-applications/xip-cli/releases/latest/download/x-ip_linux_amd64
chmod +x "$INSTALL_PATH"

echo "The command has been installed to this path: $INSTALL_PATH"
