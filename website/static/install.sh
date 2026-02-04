#!/bin/sh
# OpenBoundary CLI Installer
# Usage: curl -fsSL https://openboundary.org/install.sh | sh
#
# This script detects your OS and architecture, downloads the appropriate
# bound binary, and installs it to /usr/local/bin (or ~/.local/bin if no sudo).

set -e

GITHUB_REPO="openboundary/openboundary"
BINARY_NAME="bound"
INSTALL_DIR="/usr/local/bin"

# Colors (if terminal supports them)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

info() {
    printf "${BLUE}→${NC} %s\n" "$1"
}

success() {
    printf "${GREEN}✓${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}!${NC} %s\n" "$1"
}

error() {
    printf "${RED}✗${NC} %s\n" "$1" >&2
    exit 1
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) error "Unsupported operating system: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        armv7l) echo "arm" ;;
        i386|i686) echo "386" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest release version from GitHub
get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        curl -sL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
}

# Download file
download() {
    url="$1"
    output="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "$output"
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
}

# Check if we can write to install directory
can_write_to() {
    dir="$1"
    if [ -w "$dir" ]; then
        return 0
    fi
    return 1
}

# Main installation
main() {
    echo ""
    echo "  ${BLUE}OpenBoundary CLI Installer${NC}"
    echo "  ─────────────────────────"
    echo ""

    OS=$(detect_os)
    ARCH=$(detect_arch)

    info "Detected OS: $OS"
    info "Detected architecture: $ARCH"

    # Get latest version
    info "Fetching latest version..."
    VERSION=$(get_latest_version)

    if [ -z "$VERSION" ]; then
        error "Could not determine latest version. Please check your internet connection."
    fi

    success "Latest version: $VERSION"

    # Construct download URL
    FILENAME="${BINARY_NAME}_${VERSION#v}_${OS}_${ARCH}"
    if [ "$OS" = "windows" ]; then
        FILENAME="${FILENAME}.zip"
    else
        FILENAME="${FILENAME}.tar.gz"
    fi

    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${FILENAME}"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TMP_DIR"' EXIT

    info "Downloading ${BINARY_NAME}..."
    download "$DOWNLOAD_URL" "$TMP_DIR/$FILENAME" || error "Download failed. Release may not exist for your platform ($OS/$ARCH)."
    success "Downloaded successfully"

    # Extract
    info "Extracting..."
    cd "$TMP_DIR"
    if [ "$OS" = "windows" ]; then
        unzip -q "$FILENAME"
    else
        tar -xzf "$FILENAME"
    fi

    # Find the binary
    if [ -f "$BINARY_NAME" ]; then
        BINARY_PATH="$TMP_DIR/$BINARY_NAME"
    elif [ -f "${BINARY_NAME}.exe" ]; then
        BINARY_PATH="$TMP_DIR/${BINARY_NAME}.exe"
    else
        error "Binary not found in archive"
    fi

    # Determine install location
    if ! can_write_to "$INSTALL_DIR"; then
        if [ -n "$HOME" ]; then
            INSTALL_DIR="$HOME/.local/bin"
            mkdir -p "$INSTALL_DIR"
            warn "No write access to /usr/local/bin, installing to $INSTALL_DIR"
        else
            error "Cannot write to /usr/local/bin and HOME is not set"
        fi
    fi

    # Install
    info "Installing to $INSTALL_DIR..."
    if can_write_to "$INSTALL_DIR"; then
        cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    else
        sudo cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
        sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi

    success "Installed ${BINARY_NAME} to $INSTALL_DIR/$BINARY_NAME"

    # Verify installation
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        success "$($BINARY_NAME --version)"
    else
        echo ""
        warn "Installation complete, but $BINARY_NAME is not in your PATH."
        echo ""
        echo "  Add this to your shell profile (.bashrc, .zshrc, etc.):"
        echo ""
        echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
    fi

    echo ""
    echo "  ${GREEN}Installation complete!${NC}"
    echo ""
    echo "  Get started:"
    echo "    ${BLUE}bound init basic -o my-project${NC}"
    echo "    ${BLUE}bound compile my-project/spec.yaml${NC}"
    echo ""
    echo "  Documentation: https://openboundary.org/docs/"
    echo ""
}

# Alternative: Go install
go_install_hint() {
    echo ""
    echo "  Alternatively, if you have Go installed:"
    echo ""
    echo "    go install github.com/${GITHUB_REPO}/cmd/bound@latest"
    echo ""
}

main "$@"
