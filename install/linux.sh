#!/usr/bin/env sh

## Installation for Linux

# Install dir
INSTALL_PATH="/usr/bin/x-ip"

# Install the executable
curl -s -o "$INSTALL_PATH" https://raw.githubusercontent.com/xip-online-applications/xip-cli/master/builds/xip_linux_amd64_linux
chmod +x "$INSTALL_PATH"

echo "The command has been installed to this path: $INSTALL_PATH"
