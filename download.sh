#!/bin/bash

# GitHub repository details
REPO="mreider/krems"

# Fetch the latest release tag
LATEST_RELEASE=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep "tag_name" | cut -d '"' -f 4)

# Construct the download URL
URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/krems-darwin-amd64"

# Download the latest binary
echo "Downloading latest version: $LATEST_RELEASE..."
curl -L -o krems "$URL"

# Make the file executable
chmod +x krems

# Move to /usr/local/bin (requires sudo)
echo "Moving krems to /usr/local/bin..."
sudo mv krems /usr/local/bin/

echo "krems has been installed successfully!"
