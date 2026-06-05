#!/usr/bin/env bash
# LinBoard — one-line installer (fully automatic)
#   curl -fsSL https://raw.githubusercontent.com/Foisalislambd/LinBoard/main/scripts/install-release.sh | bash
set -euo pipefail

REPO="${LINBOARD_REPO:-Foisalislambd/LinBoard}"
BRANCH="${LINBOARD_BRANCH:-main}"
BIN_DIR="${LINBOARD_BIN_DIR:-$HOME/.local/bin}"
RAW_BASE="https://raw.githubusercontent.com/${REPO}/${BRANCH}"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

# Load system setup helpers
curl -fsSL "$RAW_BASE/scripts/lib/system-setup.sh" -o "$TMP/system-setup.sh"
# shellcheck source=/dev/null
source "$TMP/system-setup.sh"

linboard_preflight_setup

# Use LINBOARD_RELEASE (not VERSION — /etc/os-release defines VERSION on Ubuntu)
LINBOARD_RELEASE="${LINBOARD_VERSION:-latest}"

arch="$(uname -m)"
case "$arch" in
  x86_64|amd64)  ASSET="linboard-linux-amd64" ;;
  aarch64|arm64) ASSET="linboard-linux-arm64" ;;
  *)
    echo "Unsupported CPU: $arch (need x86_64 or aarch64)"
    exit 1
    ;;
esac

if [ "$LINBOARD_RELEASE" = "latest" ]; then
  API="https://api.github.com/repos/${REPO}/releases/latest"
  VERSION_TAG="$(curl -fsSL "$API" | grep -oP '"tag_name"\s*:\s*"\K[^"]+' | head -1)"
else
  VERSION_TAG="$LINBOARD_RELEASE"
  case "$VERSION_TAG" in
    v*) ;;
    *) VERSION_TAG="v${VERSION_TAG}" ;;
  esac
fi

if [ -z "$VERSION_TAG" ]; then
  echo "No release found on GitHub yet. Try again after the first release is published."
  exit 1
fi

VER="${VERSION_TAG#v}"
TARBALL="${ASSET}-v${VER}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION_TAG}/${TARBALL}"

linboard_log "LinBoard ${VERSION_TAG} (${ASSET})"
linboard_log "Downloading..."

curl -fsSL "$URL" -o "$TMP/$TARBALL"
tar -xzf "$TMP/$TARBALL" -C "$TMP"

# Bundle system-setup for install.sh
mkdir -p "$TMP/lib"
cp "$TMP/system-setup.sh" "$TMP/lib/system-setup.sh"

export LINBOARD_BIN_DIR="$BIN_DIR"
bash "$TMP/install.sh"
