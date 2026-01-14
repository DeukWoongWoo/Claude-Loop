#!/bin/bash
#
# Claude Loop Installer
# https://github.com/DeukWoongWoo/claude-loop
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/DeukWoongWoo/claude-loop/main/install.sh | bash
#
# Environment variables:
#   INSTALL_DIR - Installation directory (default: ~/.local/bin)
#   VERSION     - Specific version to install (default: latest)

set -euo pipefail

REPO="DeukWoongWoo/claude-loop"
BINARY_NAME="claude-loop"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Colors (disabled if not a terminal)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    RED='' GREEN='' YELLOW='' BLUE='' NC=''
fi

info()    { printf "${BLUE}INFO:${NC} %s\n" "$1"; }
success() { printf "${GREEN}SUCCESS:${NC} %s\n" "$1"; }
warn()    { printf "${YELLOW}WARNING:${NC} %s\n" "$1" >&2; }
error()   { printf "${RED}ERROR:${NC} %s\n" "$1" >&2; exit 1; }

detect_platform() {
    local os arch

    case "$(uname -s)" in
        Darwin)             os="darwin" ;;
        Linux)              os="linux" ;;
        MINGW*|MSYS*|CYGWIN*) os="windows" ;;
        *) error "Unsupported operating system: $(uname -s)" ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)  arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac

    [ "$os" = "windows" ] && [ "$arch" = "arm64" ] && error "Windows ARM64 is not supported"

    echo "${os}_${arch}"
}

get_latest_version() {
    local version=""

    if command -v gh &>/dev/null; then
        version=$(gh release view --repo "$REPO" --json tagName -q '.tagName' 2>/dev/null) || true
    fi

    if [ -z "$version" ]; then
        version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | \
            grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/') || true
    fi

    [ -z "$version" ] && error "Failed to get latest version. Check your internet connection."

    echo "$version"
}

download_binary() {
    local version="$1" platform="$2"
    local tmp_dir ext os_name base_url

    tmp_dir=$(mktemp -d)
    trap 'rm -rf "$tmp_dir"' EXIT

    os_name="${platform%_*}"
    ext="tar.gz"
    [ "$os_name" = "windows" ] && ext="zip"

    local archive="${BINARY_NAME}_${platform}.${ext}"
    base_url="https://github.com/${REPO}/releases/download/${version}"

    info "Downloading ${BINARY_NAME} ${version} for ${platform}..."
    curl -fsSL "${base_url}/${archive}" -o "${tmp_dir}/${archive}" || error "Failed to download ${archive}"

    # Verify checksum
    info "Verifying checksum..."
    if curl -fsSL "${base_url}/checksums.txt" -o "${tmp_dir}/checksums.txt" 2>/dev/null; then
        local expected actual
        expected=$(grep "$archive" "${tmp_dir}/checksums.txt" | awk '{print $1}')

        if [ -n "$expected" ]; then
            if command -v sha256sum &>/dev/null; then
                actual=$(sha256sum "${tmp_dir}/${archive}" | awk '{print $1}')
            elif command -v shasum &>/dev/null; then
                actual=$(shasum -a 256 "${tmp_dir}/${archive}" | awk '{print $1}')
            fi

            if [ -n "$actual" ] && [ "$expected" != "$actual" ]; then
                error "Checksum verification failed!\nExpected: ${expected}\nActual: ${actual}"
            fi
            success "Checksum verified"
        else
            warn "Could not find checksum for ${archive}"
        fi
    else
        warn "Could not download checksums file, skipping verification"
    fi

    # Extract
    info "Extracting binary..."
    if [ "$ext" = "tar.gz" ]; then
        tar -xzf "${tmp_dir}/${archive}" -C "$tmp_dir"
    else
        unzip -q "${tmp_dir}/${archive}" -d "$tmp_dir"
    fi

    # Install
    local binary_path="${tmp_dir}/${BINARY_NAME}"
    local install_path="${INSTALL_DIR}/${BINARY_NAME}"
    [ "$os_name" = "windows" ] && binary_path="${binary_path}.exe" && install_path="${install_path}.exe"

    [ ! -f "$binary_path" ] && error "Binary not found in archive"

    mkdir -p "$INSTALL_DIR"
    mv "$binary_path" "$install_path"
    chmod +x "$install_path"

    success "Installed ${BINARY_NAME} to ${install_path}"
}

check_path() {
    [[ ":$PATH:" == *":$INSTALL_DIR:"* ]] && return

    echo ""
    warn "$INSTALL_DIR is not in your PATH"
    echo ""
    echo "Add it to your shell profile:"
    echo ""

    case "$SHELL" in
        */zsh)  echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc && source ~/.zshrc" ;;
        */bash) echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc && source ~/.bashrc" ;;
        *)      echo "  export PATH=\"\$HOME/.local/bin:\$PATH\"" ;;
    esac
    echo ""
}

check_dependencies() {
    echo ""
    info "Checking dependencies..."

    local missing=()
    command -v claude &>/dev/null || missing+=("Claude Code CLI (https://claude.ai/code)")
    command -v gh &>/dev/null || missing+=("GitHub CLI (https://cli.github.com)")

    if [ ${#missing[@]} -eq 0 ]; then
        success "All dependencies installed"
        return
    fi

    warn "Missing optional dependencies:"
    printf "   - %s\n" "${missing[@]}"
    echo ""
    echo "Install them with:"
    case "$(uname -s)" in
        Darwin) echo "  brew install gh" ;;
        Linux)  echo "  # GitHub CLI: https://github.com/cli/cli#installation" ;;
    esac
    echo "  # Claude Code: https://claude.ai/code"
}

main() {
    echo ""
    echo "================================================"
    echo "  Claude Loop Installer"
    echo "================================================"
    echo ""

    local platform version
    platform=$(detect_platform)
    info "Detected platform: ${platform}"

    version="${VERSION:-$(get_latest_version)}"
    info "Version: ${version}"

    download_binary "$version" "$platform"
    check_path
    check_dependencies

    echo ""
    echo "================================================"
    success "Installation complete!"
    echo "================================================"
    echo ""
    echo "Get started:"
    echo "  ${BINARY_NAME} --help"
    echo ""
    echo "Example:"
    echo "  ${BINARY_NAME} --prompt \"your task\" --max-runs 5"
    echo ""
    echo "Documentation: https://github.com/${REPO}"
    echo ""
}

main "$@"
