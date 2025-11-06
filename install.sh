#!/usr/bin/env bash

# Swarm Browser Installation Script
# https://github.com/Mendes11/swarm-browser
#
# This script installs the swarm-browser binary from GitHub releases.
# - If run with sudo: installs to /usr/local/bin
# - If run without sudo: installs to ~/.local/bin
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Mendes11/swarm-browser/main/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/Mendes11/swarm-browser/main/install.sh | sudo bash

set -e

# Configuration
REPO_OWNER="Mendes11"
REPO_NAME="swarm-browser"
BINARY_NAME="swarm-browser"
GITHUB_BASE_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored messages
print_error() {
    echo -e "${RED}Error: $1${NC}" >&2
}

print_success() {
    echo -e "${GREEN}$1${NC}"
}

print_info() {
    echo -e "${BLUE}$1${NC}"
}

print_warning() {
    echo -e "${YELLOW}$1${NC}"
}

# Detect OS
detect_os() {
    local os
    case "$(uname -s)" in
        Linux*)     os="Linux";;
        Darwin*)    os="Darwin";;
        CYGWIN*|MINGW*|MSYS*)
            print_error "Windows is not supported. Please use WSL or a Linux VM."
            exit 1
            ;;
        *)
            print_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    echo "$os"
}

# Detect architecture
detect_arch() {
    local arch
    case "$(uname -m)" in
        x86_64|amd64)   arch="x86_64";;
        aarch64|arm64)  arch="arm64";;
        i386|i686)      arch="i386";;
        armv7l)
            # ARM v7 is not directly supported, but we can try arm64
            print_warning "ARM v7 detected. Attempting to use arm64 binary..."
            arch="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    echo "$arch"
}

# Determine installation directory based on privileges
determine_install_dir() {
    if [ "$EUID" -eq 0 ] || [ -n "$SUDO_USER" ]; then
        # Running as root or with sudo
        echo "/usr/local/bin"
    else
        # Running as regular user
        echo "${HOME}/.local/bin"
    fi
}

# Check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Download file with curl or wget
download_file() {
    local url="$1"
    local output="$2"

    if command_exists curl; then
        curl -fsSL "$url" -o "$output"
    elif command_exists wget; then
        wget -q "$url" -O "$output"
    else
        print_error "Neither curl nor wget found. Please install curl or wget."
        exit 1
    fi
}

# Verify checksum
verify_checksum() {
    local file="$1"
    local checksums_file="$2"
    local expected_checksum

    # Extract the checksum for our file
    expected_checksum=$(grep "$(basename "$file")" "$checksums_file" | cut -d' ' -f1)

    if [ -z "$expected_checksum" ]; then
        print_warning "Checksum not found for $(basename "$file"). Skipping verification."
        return 0
    fi

    local actual_checksum
    if command_exists sha256sum; then
        actual_checksum=$(sha256sum "$file" | cut -d' ' -f1)
    elif command_exists shasum; then
        actual_checksum=$(shasum -a 256 "$file" | cut -d' ' -f1)
    else
        print_warning "No checksum tool found. Skipping verification."
        return 0
    fi

    if [ "$expected_checksum" != "$actual_checksum" ]; then
        print_error "Checksum verification failed!"
        print_error "Expected: $expected_checksum"
        print_error "Got: $actual_checksum"
        return 1
    fi

    print_success "✓ Checksum verified"
    return 0
}

# Check if binary is already installed and get version
get_installed_version() {
    local binary_path="$1"
    if [ -f "$binary_path" ] && [ -x "$binary_path" ]; then
        # Try to get version (assuming the binary supports --version flag)
        "$binary_path" --version 2>/dev/null || echo "unknown"
    else
        echo "none"
    fi
}

# Main installation function
main() {
    print_info "======================================"
    print_info " Swarm Browser Installation Script"
    print_info "======================================"
    echo

    # Detect OS and architecture
    local os=$(detect_os)
    local arch=$(detect_arch)
    print_info "Detected OS: $os"
    print_info "Detected Architecture: $arch"

    # Determine installation directory
    local install_dir=$(determine_install_dir)
    print_info "Installation directory: $install_dir"

    # Check if installation directory exists and create if needed
    if [ ! -d "$install_dir" ]; then
        print_info "Creating installation directory: $install_dir"
        mkdir -p "$install_dir"
    fi

    # Check if binary already exists
    local binary_path="${install_dir}/${BINARY_NAME}"
    local current_version=$(get_installed_version "$binary_path")
    if [ "$current_version" != "none" ]; then
        print_warning "Swarm Browser is already installed (version: $current_version)"
        read -p "Do you want to reinstall/update? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Installation cancelled."
            exit 0
        fi
    fi

    # Create temporary directory
    local temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT

    print_info "Downloading Swarm Browser..."

    # Get the latest version tag
    local version_tag
    if command_exists curl; then
        version_tag=$(curl -s "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
    elif command_exists wget; then
        version_tag=$(wget -qO- "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
    fi

    if [ -z "$version_tag" ]; then
        print_warning "Could not determine latest version. Continuing without version info."
    else
        print_info "Latest version: $version_tag"
    fi

    # Construct download URLs
    local archive_name="${BINARY_NAME}_${os}_${arch}.tar.gz"
    local archive_url="${GITHUB_BASE_URL}/releases/latest/download/${archive_name}"

    # Checksums file includes version in the name
    local checksums_url
    if [ -n "$version_tag" ]; then
        checksums_url="${GITHUB_BASE_URL}/releases/latest/download/${BINARY_NAME}_${version_tag}_checksums.txt"
    fi

    # Download archive
    print_info "Downloading from: $archive_url"
    download_file "$archive_url" "${temp_dir}/${archive_name}" || {
        print_error "Failed to download binary archive"
        exit 1
    }

    # Download and verify checksums
    if [ -n "$checksums_url" ]; then
        print_info "Downloading checksums..."
        local checksums_file="${temp_dir}/checksums.txt"
        download_file "$checksums_url" "$checksums_file" 2>/dev/null || {
            print_warning "Could not download checksums file. Skipping verification."
        }
    else
        local checksums_file=""
        print_warning "Skipping checksum verification (version could not be determined)."
    fi

    if [ -f "$checksums_file" ]; then
        # The checksum file contains entries for all artifacts, we need to filter
        grep "${archive_name}" "$checksums_file" > "${temp_dir}/checksum_single.txt" 2>/dev/null || true
        if [ -s "${temp_dir}/checksum_single.txt" ]; then
            verify_checksum "${temp_dir}/${archive_name}" "${temp_dir}/checksum_single.txt" || exit 1
        fi
    fi

    # Extract archive
    print_info "Extracting archive..."
    tar -xzf "${temp_dir}/${archive_name}" -C "$temp_dir" || {
        print_error "Failed to extract archive"
        exit 1
    }

    # Find the binary (it should be named swarm-browser after extraction)
    if [ ! -f "${temp_dir}/${BINARY_NAME}" ]; then
        print_error "Binary not found in archive"
        exit 1
    fi

    # Install the binary
    print_info "Installing binary to ${binary_path}..."
    if [ "$EUID" -eq 0 ] || [ -n "$SUDO_USER" ]; then
        # Running as root/sudo
        mv "${temp_dir}/${BINARY_NAME}" "$binary_path"
        chmod 755 "$binary_path"
    else
        # Running as regular user
        mv "${temp_dir}/${BINARY_NAME}" "$binary_path"
        chmod 755 "$binary_path"
    fi

    print_success "✓ Swarm Browser installed successfully!"

    # Check if install directory is in PATH
    if [[ ":$PATH:" != *":${install_dir}:"* ]]; then
        print_warning "Warning: ${install_dir} is not in your PATH."
        echo
        echo "To add it to your PATH, run one of the following commands:"
        echo
        echo "  For bash:"
        echo "    echo 'export PATH=\"${install_dir}:\$PATH\"' >> ~/.bashrc && source ~/.bashrc"
        echo
        echo "  For zsh:"
        echo "    echo 'export PATH=\"${install_dir}:\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
        echo
        echo "  For fish:"
        echo "    fish_add_path ${install_dir}"
        echo
        echo "After updating your PATH, you can run: ${BINARY_NAME}"
        echo "Or run directly: ${binary_path}"
    else
        echo
        echo "You can now run: ${BINARY_NAME}"
    fi

    # Try to show version of installed binary
    echo
    if [ -x "$binary_path" ]; then
        print_info "Installed version:"
        "$binary_path" --version 2>/dev/null || print_info "Version information not available"
    fi
}

# Run main function
main "$@"