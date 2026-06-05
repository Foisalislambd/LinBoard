#!/usr/bin/env bash
# LinBoard — install from release package (no Go required)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${LINBOARD_BIN_DIR:-$HOME/.local/bin}"
DESKTOP_DIR="$HOME/.local/share/applications"
AUTOSTART_DIR="$HOME/.config/autostart"

echo "==> Installing LinBoard..."
mkdir -p "$BIN_DIR" "$DESKTOP_DIR" "$AUTOSTART_DIR"

install -m 755 "$ROOT/linboard" "$BIN_DIR/linboard"

# Desktop launcher
sed "s|@EXEC@|$BIN_DIR/linboard|g" "$ROOT/linboard.desktop" > "$DESKTOP_DIR/linboard.desktop"
chmod 644 "$DESKTOP_DIR/linboard.desktop"

# Autostart
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

# Runtime tools (optional but recommended)
MISSING=()
for cmd in wtype xdotool; do
  command -v "$cmd" >/dev/null 2>&1 || MISSING+=("$cmd")
done
if ((${#MISSING[@]} > 0)); then
  echo "==> Note: install for auto-paste support:"
  echo "    Debian/Ubuntu: sudo apt install ${MISSING[*]}"
  echo "    Fedora:        sudo dnf install ${MISSING[*]}"
  echo "    Arch:          sudo pacman -S ${MISSING[*]}"
fi

# Super+V shortcut (GNOME/KDE/XFCE/Cinnamon)
if "$BIN_DIR/linboard" install-shortcut; then
  echo "==> Super+V shortcut registered"
else
  echo "==> Shortcut: add manually in Settings → Keyboard"
  echo "    Command: $BIN_DIR/linboard toggle"
fi

echo ""
echo "Installed: $BIN_DIR/linboard"
echo "Start now: linboard"
echo "Hotkey:    Super+V (Win+V)"
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
  echo "Add to PATH: export PATH=\"$BIN_DIR:\$PATH\""
fi
