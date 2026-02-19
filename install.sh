#!/bin/bash
set -e

VERSION="1.0.0"

echo "Installing AutoFix ${VERSION}..."

detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ -f /etc/os-release ]]; then
        . /etc/os-release
        if [[ "$ID" == "ubuntu" || "$ID" == "debian" ]]; then
            echo "ubuntu"
        elif [[ "$ID" == "fedora" ]]; then
            echo "fedora"
        elif [[ "$ID" == "arch" ]]; then
            echo "arch"
        else
            echo "linux"
        fi
    else
        echo "linux"
    fi
}

detect_arch() {
    local arch=$(uname -m)
    case $arch in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

OS=$(detect_os)
ARCH=$(detect_arch)

if [[ "$ARCH" == "unknown" ]]; then
    echo "Error: Unsupported architecture"
    exit 1
fi

BINARY_NAME="autofix-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/steliosot/autofix/releases/download/v${VERSION}/${BINARY_NAME}"

echo "Downloading binary for ${OS}/${ARCH}..."

if command -v curl &> /dev/null; then
    curl -fsSL -o /tmp/autofix "$DOWNLOAD_URL"
elif command -v wget &> /dev/null; then
    wget -q -O /tmp/autofix "$DOWNLOAD_URL"
else
    echo "Error: Neither curl nor wget found"
    exit 1
fi

chmod +x /tmp/autofix

echo "Installing to /usr/local/bin/autofix..."

if [[ -w /usr/local/bin ]] || sudo -n true 2>/dev/null; then
    if [[ -w /usr/local/bin ]]; then
        mv /tmp/autofix /usr/local/bin/autofix
    else
        sudo mv /tmp/autofix /usr/local/bin/autofix
    fi
else
    echo "Error: Cannot write to /usr/local/bin. Please run with sudo or install manually."
    exit 1
fi

cat << 'EOF'
AutoFix installed successfully!

To get started:
  autofix setup           # Run interactive setup
  autofix version         # Verify installation

Example usage:
  autofix run 'npm install'
  autofix run 'pip install requests'

For more information, visit: https://github.com/autofix/cli
EOF