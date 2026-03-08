#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
GO_DIR="$ROOT_DIR/go"
BIN_DIR="$ROOT_DIR/bin"

mkdir -p "$BIN_DIR"

generate_wrapper() {
  # プラットフォームに応じて適切なバイナリを実行するラッパースクリプトを生成する
  # npm の bin エントリはこのラッパーを参照する
  local WRAPPER="$BIN_DIR/claude-commit-msg-gen"
  cat > "$WRAPPER" << 'EOF'
#!/bin/sh
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  arm64|aarch64) ARCH="arm64" ;;
  x86_64|amd64)  ARCH="amd64" ;;
esac
DIR="$(cd "$(dirname "$0")" && pwd)"
BINARY="$DIR/claude-commit-msg-gen-${OS}-${ARCH}"
if [ ! -x "$BINARY" ]; then
  echo "[claude-commit-msg-gen] Binary not found: $BINARY" >&2
  exit 0
fi
exec "$BINARY" "$@"
EOF
  chmod +x "$WRAPPER"
  chmod +x "$BIN_DIR"/claude-commit-msg-gen-* 2>/dev/null || true
  echo "Wrapper generated: $WRAPPER"
}

# --wrapper-only: CI でバイナリ収集後にラッパーだけ生成する用途
if [[ "${1:-}" == "--wrapper-only" ]]; then
  generate_wrapper
  exit 0
fi

PLATFORMS=(
  "darwin/arm64"
  "darwin/amd64"
  "linux/amd64"
  "linux/arm64"
)

VERSION="v$(grep '"version"' "$ROOT_DIR/package.json" | sed 's/.*"version": *"\([^"]*\)".*/\1/')"

for PLATFORM in "${PLATFORMS[@]}"; do
  OS="${PLATFORM%/*}"
  ARCH="${PLATFORM#*/}"
  OUTPUT="$BIN_DIR/claude-commit-msg-gen-${OS}-${ARCH}"
  echo "Building ${OS}/${ARCH} -> $(basename "$OUTPUT") (${VERSION})"
  cd "$GO_DIR"
  GOOS="$OS" GOARCH="$ARCH" go build -ldflags "-X main.version=${VERSION}" -o "$OUTPUT" .
done

generate_wrapper

echo ""
echo "Build complete:"
ls -lh "$BIN_DIR/"
