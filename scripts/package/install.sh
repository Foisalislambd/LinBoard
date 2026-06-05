#!/usr/bin/env bash
# LinBoard — install from release package (fully automatic)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${LINBOARD_BIN_DIR:-$HOME/.local/bin}"

linboard_log() { echo "==> $*"; }
DESKTOP_DIR="$HOME/.local/share/applications"
AUTOSTART_DIR="$HOME/.config/autostart"

# shellcheck source=../lib/system-setup.sh
if [[ -f "$ROOT/lib/system-setup.sh" ]]; then
  # shellcheck disable=SC1091
  source "$ROOT/lib/system-setup.sh"
elif [[ -f "$ROOT/../lib/system-setup.sh" ]]; then
  # shellcheck disable=SC1091
  source "$ROOT/../lib/system-setup.sh"
fi

linboard_log "Installing LinBoard..."
mkdir -p "$BIN_DIR" "$DESKTOP_DIR" "$AUTOSTART_DIR"

if declare -F linboard_preflight_setup >/dev/null 2>&1; then
  linboard_preflight_setup
fi

install -m 755 "$ROOT/linboard" "$BIN_DIR/linboard"

if [[ -d "$ROOT/icons" ]]; then
  if declare -F linboard_install_theme_icons >/dev/null 2>&1; then
    linboard_install_theme_icons "$ROOT/icons"
  else
    mkdir -p "$HOME/.local/share/icons/hicolor"
    cp -r "$ROOT/icons/hicolor/." "$HOME/.local/share/icons/hicolor/" 2>/dev/null || true
  fi
fi

sed "s|@EXEC@|$BIN_DIR/linboard|g" "$ROOT/linboard.desktop" > "$DESKTOP_DIR/linboard.desktop"
chmod 644 "$DESKTOP_DIR/linboard.desktop"

cat > "$AUTOSTART_DIR/linboard.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=LinBoard
Comment=Clipboard Manager
Exec=$BIN_DIR/linboard
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
EOF

if declare -F linboard_post_install_setup >/dev/null 2>&1; then
  linboard_post_install_setup "$BIN_DIR"
else
  "$BIN_DIR/linboard" install-shortcut || true
  echo "Installed: $BIN_DIR/linboard — press Super+V"
fi
