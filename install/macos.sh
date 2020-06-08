#!/usr/bin/env sh

## Installation for MacOS

# Install dir
INSTALL_PATH="/usr/local/bin/x-ip"

# Install the executable
curl -s -o "$INSTALL_PATH" https://raw.githubusercontent.com/xip-online-applications/xip-cli/master/builds/xip_darwin_amd64_darwin
chmod +x "$INSTALL_PATH"

echo "The command has been installed to this path: $INSTALL_PATH"
