#!/bin/bash

set -e

echo "==================================="
echo "  Screen Recorder Installation"
echo "==================================="

# Detect package manager
detect_package_manager() {
    if command -v pacman &> /dev/null; then
        echo "Detected: Arch Linux"
        echo "Package manager: pacman"
        PACKAGE_MANAGER="pacman"
    elif command -v apt &> /dev/null; then
        echo "Detected: Debian/Ubuntu"
        echo "Package manager: apt"
        PACKAGE_MANAGER="apt"
    elif command -v dnf &> /dev/null; then
        echo "Detected: Fedora"
        echo "Package manager: dnf"
        PACKAGE_MANAGER="dnf"
    elif command -v zypper &> /dev/null; then
        echo "Detected: openSUSE"
        echo "Package manager: zypper"
        PACKAGE_MANAGER="zypper"
    else
        echo "Error: Could not detect package manager"
        echo "Please install dependencies manually"
        exit 1
    fi
}

# Install FFmpeg
install_ffmpeg() {
    echo ""
    echo "Installing FFmpeg..."

    case $PACKAGE_MANAGER in
        pacman)
            sudo pacman -S --noconfirm ffmpeg
            ;;
        apt)
            sudo apt update
            sudo apt install -y ffmpeg
            ;;
        dnf)
            sudo dnf install -y ffmpeg
            ;;
        zypper)
            sudo zypper install -y ffmpeg
            ;;
    esac

    echo "FFmpeg installed successfully"
}

# Check if Go is installed
check_go() {
    echo ""
    echo "Checking Go installation..."

    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
        echo "Go version: $GO_VERSION"
    else
        echo "Error: Go is not installed"
        echo "Please install Go from: https://go.dev/doc/install"
        exit 1
    fi
}

# Build the application
build_app() {
    echo ""
    echo "Building Screen Recorder..."

    cd "$(dirname "$0")"

    go build -o screenrecord ./cmd/screenrecorder

    if [ $? -eq 0 ]; then
        echo "Build successful!"
    else
        echo "Build failed!"
        exit 1
    fi
}

# Install the binary
install_binary() {
    echo ""
    echo "Installing binary..."

    BIN_DIR="/usr/local/bin"

    if [ -w "$BIN_DIR" ]; then
        cp screenrecord "$BIN_DIR/"
        echo "Installed to: $BIN_DIR/screenrecord"
    else
        echo "Note: Cannot write to $BIN_DIR"
        echo "Running without installation..."
        echo "To install, run: sudo cp screenrecord $BIN_DIR/"
    fi
}

# Main installation process
main() {
    detect_package_manager
    install_ffmpeg
    check_go
    build_app
    install_binary

    echo ""
    echo "==================================="
    echo "  Installation Complete!"
    echo "==================================="
    echo ""
    echo "Run with: screenrecord"
    echo ""
}

main "$@"
