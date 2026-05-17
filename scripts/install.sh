#!/usr/bin/env sh
# install.sh — installs mcp-postman from GitHub Releases
# Usage: curl -sSfL https://raw.githubusercontent.com/jinkp/postman-go-mcp/master/scripts/install.sh | sh
set -e

REPO="jinkp/postman-go-mcp"
BINARY="mcp-postman"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# ── helpers ────────────────────────────────────────────────────────────────────

say()  { printf "  \033[32m✔\033[0m %s\n" "$*"; }
warn() { printf "  \033[33m⚠\033[0m %s\n" "$*"; }
die()  { printf "  \033[31m✗\033[0m %s\n" "$*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || die "Required tool not found: $1"
}

# ── detect OS / arch ───────────────────────────────────────────────────────────

detect_platform() {
  OS="$(uname -s)"
  ARCH="$(uname -m)"

  case "$OS" in
    Linux*)  OS="linux"  ;;
    Darwin*) OS="darwin" ;;
    *)       die "Unsupported OS: $OS" ;;
  esac

  case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) die "Unsupported architecture: $ARCH" ;;
  esac
}

# ── fetch latest tag ───────────────────────────────────────────────────────────

fetch_latest_version() {
  need curl
  VERSION="$(curl -sSfL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed 's/.*"tag_name": *"\(.*\)".*/\1/')"
  [ -n "$VERSION" ] || die "Could not determine latest version"
}

# ── download and install ───────────────────────────────────────────────────────

install_binary() {
  ASSET="${BINARY}_${OS}_${ARCH}"
  EXT="tar.gz"
  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}.${EXT}"
  CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

  TMP_DIR="$(mktemp -d)"
  trap 'rm -rf "$TMP_DIR"' EXIT

  printf "  Downloading %s %s (%s_%s)...\n" "$BINARY" "$VERSION" "$OS" "$ARCH"
  curl -sSfL "$DOWNLOAD_URL" -o "${TMP_DIR}/${ASSET}.${EXT}" \
    || die "Download failed: $DOWNLOAD_URL"

  # Verify checksum
  curl -sSfL "$CHECKSUM_URL" -o "${TMP_DIR}/checksums.txt" 2>/dev/null && {
    cd "$TMP_DIR"
    grep "${ASSET}.${EXT}" checksums.txt | sha256sum --check --status 2>/dev/null \
      && say "Checksum verified" \
      || warn "Could not verify checksum — proceeding anyway"
    cd - >/dev/null
  }

  # Extract
  tar -xzf "${TMP_DIR}/${ASSET}.${EXT}" -C "$TMP_DIR"

  # Install
  if [ -w "$INSTALL_DIR" ]; then
    cp "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    chmod +x "${INSTALL_DIR}/${BINARY}"
  else
    sudo cp "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    sudo chmod +x "${INSTALL_DIR}/${BINARY}"
  fi
}

# ── verify install ─────────────────────────────────────────────────────────────

verify() {
  INSTALLED_VERSION="$("${INSTALL_DIR}/${BINARY}" --version 2>/dev/null || true)"
  say "Installed: ${INSTALL_DIR}/${BINARY}"
  say "Version:   ${INSTALLED_VERSION}"
}

# ── main ───────────────────────────────────────────────────────────────────────

main() {
  printf "\n  \033[1mpostman-go-mcp installer\033[0m\n\n"

  detect_platform
  fetch_latest_version
  install_binary
  verify

  printf "\n  \033[1mDone!\033[0m Run the setup wizard to configure your AI assistant:\n\n"
  printf "    mcp-postman setup\n\n"
}

main
