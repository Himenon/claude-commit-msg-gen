#!/bin/sh
set -eu

REPO="Himenon/claude-commit-msg-gen"
BINARY_NAME="claude-commit-msg-gen"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect ARCH
ARCH="$(uname -m)"
case "$ARCH" in
  arm64|aarch64) ARCH="arm64" ;;
  x86_64|amd64)  ARCH="amd64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Resolve version
VERSION="${VERSION:-}"
if [ -z "$VERSION" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
fi

ASSET_NAME="${BINARY_NAME}-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET_NAME}"

mkdir -p "$INSTALL_DIR"
echo "Downloading ${ASSET_NAME} ${VERSION} ..."
curl -fsSL "$DOWNLOAD_URL" -o "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
echo "Installed: ${INSTALL_DIR}/${BINARY_NAME}"
echo ""

case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo "Add ${INSTALL_DIR} to your PATH:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    echo ""
    ;;
esac

echo "Setup:"
echo "  1. Set your Anthropic API key:"
echo "       export ANTHROPIC_API_KEY=\"sk-ant-...\""
echo "  2. Run in your repository:"
echo "       lefthook install"
echo "  3. Stage files and commit — the message is generated automatically:"
echo "       git add <files> && git commit"
echo ""
echo "Version: ${INSTALL_DIR}/${BINARY_NAME} --version"
