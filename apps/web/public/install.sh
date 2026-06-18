#!/bin/sh
# Tene CLI installer
# Usage: curl -sSfL https://tene.sh/install.sh | sh
#
# Installs the latest tene binary to /usr/local/bin (or ~/.local/bin as fallback).
# Supports macOS (arm64, amd64) and Linux (arm64, amd64).

set -eu

RELEASE_BASE="https://tene-releases.s3.ap-northeast-2.amazonaws.com"
INSTALL_DIR="/usr/local/bin"
BINARY="tene"

info() {
  printf '\033[1;34m%s\033[0m\n' "$1"
}

error() {
  printf '\033[1;31mError: %s\033[0m\n' "$1" >&2
  exit 1
}

detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux)  echo "linux" ;;
    *)      error "Unsupported OS: $(uname -s). Use WSL on Windows." ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)             error "Unsupported architecture: $(uname -m)" ;;
  esac
}

get_latest_version() {
  if command -v curl > /dev/null 2>&1; then
    curl -sSfL "${RELEASE_BASE}/LATEST_VERSION"
  elif command -v wget > /dev/null 2>&1; then
    wget -qO- "${RELEASE_BASE}/LATEST_VERSION"
  else
    error "curl or wget is required"
  fi
}

download() {
  url="$1"
  output="$2"
  if command -v curl > /dev/null 2>&1; then
    curl -sSfL -o "$output" "$url"
  elif command -v wget > /dev/null 2>&1; then
    wget -qO "$output" "$url"
  fi
}

main() {
  info "Installing tene..."

  os="$(detect_os)"
  arch="$(detect_arch)"
  version="$(get_latest_version)"

  if [ -z "$version" ]; then
    error "Failed to fetch latest version"
  fi

  info "  Version: v${version}"
  info "  Platform: ${os}/${arch}"

  filename="tene_${version}_${os}_${arch}.tar.gz"
  url="${RELEASE_BASE}/v${version}/${filename}"
  checksum_url="${RELEASE_BASE}/v${version}/checksums.txt"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  info "  Downloading ${filename}..."
  download "$url" "${tmpdir}/${filename}"
  download "$checksum_url" "${tmpdir}/checksums.txt"

  # SHA-256 checksum verification
  if command -v sha256sum > /dev/null 2>&1; then
    expected=$(grep "$filename" "${tmpdir}/checksums.txt" | awk '{print $1}')
    actual=$(sha256sum "${tmpdir}/${filename}" | awk '{print $1}')
  elif command -v shasum > /dev/null 2>&1; then
    expected=$(grep "$filename" "${tmpdir}/checksums.txt" | awk '{print $1}')
    actual=$(shasum -a 256 "${tmpdir}/${filename}" | awk '{print $1}')
  fi

  if [ -n "$expected" ] && [ "$expected" != "$actual" ]; then
    error "Checksum verification failed (expected: ${expected}, got: ${actual})"
  fi

  tar xzf "${tmpdir}/${filename}" -C "$tmpdir"

  if [ ! -f "${tmpdir}/${BINARY}" ]; then
    error "Binary not found in archive"
  fi

  # Try /usr/local/bin first, fall back to ~/.local/bin
  if [ -w "$INSTALL_DIR" ] || [ "$(id -u)" = "0" ]; then
    mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    chmod +x "${INSTALL_DIR}/${BINARY}"
    info "  Installed to ${INSTALL_DIR}/${BINARY}"
  else
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "$INSTALL_DIR"
    mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    chmod +x "${INSTALL_DIR}/${BINARY}"
    info "  Installed to ${INSTALL_DIR}/${BINARY}"
    case ":$PATH:" in
      *":${INSTALL_DIR}:"*) ;;
      *) printf '\033[1;33m%s\033[0m\n' "  Add to PATH: export PATH=\"${INSTALL_DIR}:\$PATH\"" ;;
    esac
  fi

  info ""
  info "  tene v${version} installed successfully!"
  info ""
  info "  Next step: tene init"
  info ""
  info "  Documentation:"
  info "    README:    https://github.com/tene-ai/tene#readme"
  info "    CLI ref:   https://tene.sh/cli"
  info "    AI index:  https://tene.sh/llms.txt"
  info "    Issues:    https://github.com/tene-ai/tene/issues"
}

main
