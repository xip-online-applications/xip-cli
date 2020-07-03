#!/usr/bin/env sh

## Installation for MacOS

# Install dir
INSTALL_PATH="/usr/local/bin/x-ip"

# Download link
DOWNLOAD_LINK="https://github.com/xip-online-applications/xip-cli/releases/latest/download/x-ip_macos_amd64"
if [ $(uname -m) = "arm64" ]; then
  DOWNLOAD_LINK="https://github.com/xip-online-applications/xip-cli/releases/latest/download/x-ip_macos_arm64"
fi

# Install the executable
curl -s -o "$INSTALL_PATH" "$DOWNLOAD_LINK"
chmod +x "$INSTALL_PATH"

echo "The command has been installed to this path: $INSTALL_PATH"
