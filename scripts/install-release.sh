#!/usr/bin/env bash
# LinBoard — one-line installer from GitHub Releases
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Foisalislambd/LinBoard/main/scripts/install-release.sh | bash
#   LINBOARD_VERSION=v1.0.0 curl -fsSL ... | bash
set -euo pipefail

REPO="${LINBOARD_REPO:-Foisalislambd/LinBoard}"
VERSION="${LINBOARD_VERSION:-latest}"
INSTALL_DIR="${LINBOARD_INSTALL_DIR:-$HOME/.local/linboard}"
BIN_DIR="${LINBOARD_BIN_DIR:-$HOME/.local/bin}"

arch="$(uname -m)"
case "$arch" in
  x86_64|amd64)  ASSET="linboard-linux-amd64" ;;
  aarch64|arm64) ASSET="linboard-linux-arm64" ;;
  *)
    echo "Unsupported CPU: $arch (need x86_64 or aarch64)"
    exit 1
    ;;
esac

if [ "$VERSION" = "latest" ]; then
  API="https://api.github.com/repos/${REPO}/releases/latest"
  VERSION_TAG="$(curl -fsSL "$API" | grep -oP '"tag_name"\s*:\s*"\K[^"]+' | head -1)"
else
  VERSION_TAG="$VERSION"
fi

if [ -z "$VERSION_TAG" ]; then
  echo "Could not resolve release version from GitHub."
  exit 1
fi

VER="${VERSION_TAG#v}"
TARBALL="${ASSET}-v${VER}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION_TAG}/${TARBALL}"

echo "==> LinBoard ${VERSION_TAG} (${ASSET})"
echo "==> Downloading $URL"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "$TMP/$TARBALL"
tar -xzf "$TMP/$TARBALL" -C "$TMP"

export LINBOARD_BIN_DIR="$BIN_DIR"
bash "$TMP/install.sh"

echo "==> Done."
