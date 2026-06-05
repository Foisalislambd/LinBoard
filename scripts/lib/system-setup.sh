#!/usr/bin/env bash
# LinBoard — automatic system detection & dependency setup
# Sourced by install-release.sh and package/install.sh

linboard_log() { echo "==> $*"; }
linboard_warn() { echo "==> [note] $*" >&2; }

linboard_detect_os() {
  LINBOARD_OS_ID=""
  LINBOARD_OS_LIKE=""
  LINBOARD_PM=""   # apt, dnf, pacman, zypper

  if [[ -f /etc/os-release ]]; then
    # shellcheck disable=SC1091
    . /etc/os-release
    LINBOARD_OS_ID="${ID:-unknown}"
    LINBOARD_OS_LIKE="${ID_LIKE:-}"
  fi

  case "$LINBOARD_OS_ID" in
    ubuntu|debian|linuxmint|pop|zorin|elementary|neon)
      LINBOARD_PM="apt" ;;
    fedora|nobara)
      LINBOARD_PM="dnf" ;;
    rhel|centos|rocky|almalinux|ol)
      LINBOARD_PM="dnf" ;;
    arch|manjaro|endeavouros|garuda)
      LINBOARD_PM="pacman" ;;
    opensuse-leap|opensuse-tumbleweed|opensuse|sles|sled)
      LINBOARD_PM="zypper" ;;
    *)
      if [[ "$LINBOARD_OS_LIKE" == *debian* ]] || [[ "$LINBOARD_OS_LIKE" == *ubuntu* ]]; then
        LINBOARD_PM="apt"
      elif [[ "$LINBOARD_OS_LIKE" == *fedora* ]] || [[ "$LINBOARD_OS_LIKE" == *rhel* ]]; then
        LINBOARD_PM="dnf"
      elif [[ "$LINBOARD_OS_LIKE" == *arch* ]]; then
        LINBOARD_PM="pacman"
      fi
      ;;
  esac
}

linboard_detect_session() {
  LINBOARD_SESSION="${XDG_SESSION_TYPE:-unknown}"
  LINBOARD_DESKTOP="${XDG_CURRENT_DESKTOP:-}"
  LINBOARD_DESKTOP_LC="$(echo "$LINBOARD_DESKTOP" | tr '[:upper:]' '[:lower:]')"
}

linboard_have() { command -v "$1" >/dev/null 2>&1; }

linboard_sudo() {
  if linboard_have sudo; then
    sudo "$@"
  elif [[ "$(id -u)" -eq 0 ]]; then
    "$@"
  else
    return 1
  fi
}

linboard_install_packages() {
  local pkgs=("$@")
  [[ ${#pkgs[@]} -eq 0 ]] && return 0

  linboard_detect_os
  linboard_log "Installing system packages: ${pkgs[*]} (${LINBOARD_PM:-unknown distro})"

  case "$LINBOARD_PM" in
    apt)
      linboard_sudo apt-get update -qq
      linboard_sudo apt-get install -y "${pkgs[@]}"
      ;;
    dnf)
      linboard_sudo dnf install -y "${pkgs[@]}"
      ;;
    pacman)
      linboard_sudo pacman -Sy --noconfirm "${pkgs[@]}"
      ;;
    zypper)
      linboard_sudo zypper --non-interactive install "${pkgs[@]}"
      ;;
    *)
      linboard_warn "Could not auto-install packages on this distro."
      linboard_warn "Please install manually: ${pkgs[*]}"
      return 1
      ;;
  esac
}

linboard_install_paste_tools() {
  linboard_detect_session
  local need=()

  if ! linboard_have wtype; then need+=("wtype"); fi
  if ! linboard_have xdotool; then need+=("xdotool"); fi

  [[ ${#need[@]} -eq 0 ]] && return 0

  linboard_log "Setting up auto-paste tools (${LINBOARD_SESSION} session)..."
  linboard_install_packages "${need[@]}" || true
}

linboard_install_gnome_tray() {
  linboard_detect_session
  [[ "$LINBOARD_DESKTOP_LC" != *gnome* ]] && return 0

  # Already have tray support?
  if linboard_have gnome-extensions; then
    if gnome-extensions list --enabled 2>/dev/null | grep -qi appindicator; then
      return 0
    fi
  fi

  linboard_log "GNOME detected — installing tray icon support..."
  linboard_detect_os

  local pkg=""
  case "$LINBOARD_PM" in
    apt)  pkg="gnome-shell-extension-appindicator" ;;
    dnf)  pkg="gnome-shell-extension-appindicator" ;;
    pacman) pkg="gnome-shell-extension-appindicator" ;;
  esac

  if [[ -n "$pkg" ]]; then
    linboard_install_packages "$pkg" || true
  fi

  # Try to enable the extension when the CLI is available
  if linboard_have gnome-extensions; then
    local ext_uuid=""
    for uuid in \
      "appindicatorsupport@rgcjonas.gmail.com" \
      "AppIndicatorExtension@martin.zimmermann"; do
      if gnome-extensions list 2>/dev/null | grep -qF "$uuid"; then
        ext_uuid="$uuid"
        break
      fi
    done
    if [[ -n "$ext_uuid" ]]; then
      gnome-extensions enable "$ext_uuid" 2>/dev/null || true
      linboard_log "Enabled GNOME AppIndicator extension"
    fi
  fi
}

linboard_ensure_path() {
  local bin_dir="$1"
  [[ -z "$bin_dir" ]] && return 0
  [[ ":$PATH:" == *":$bin_dir:"* ]] && return 0

  linboard_log "Adding $bin_dir to PATH..."
  local line="export PATH=\"$bin_dir:\$PATH\""
  local updated=0

  for rc in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
    [[ -f "$rc" ]] || continue
    if ! grep -qF "$bin_dir" "$rc" 2>/dev/null; then
      printf '\n# LinBoard\n%s\n' "$line" >> "$rc"
      updated=1
    fi
  done

  if [[ "$updated" -eq 0 ]] && [[ ! -f "$HOME/.bashrc" ]]; then
    printf '%s\n' "$line" >> "$HOME/.profile"
  fi

  export PATH="$bin_dir:$PATH"
}

linboard_start_app() {
  local bin_dir="$1"
  local exe="$bin_dir/linboard"

  [[ -x "$exe" ]] || return 0

  pkill -x linboard 2>/dev/null || true
  sleep 0.3

  linboard_log "Starting LinBoard in the background..."
  if [[ -n "${DISPLAY:-}" ]] || [[ -n "${WAYLAND_DISPLAY:-}" ]]; then
    nohup "$exe" >/dev/null 2>&1 &
    disown 2>/dev/null || true
    sleep 0.5
    if pgrep -x linboard >/dev/null; then
      linboard_log "LinBoard is running (system tray)"
    else
      linboard_warn "Start manually: linboard"
    fi
  else
    linboard_warn "No display detected — run 'linboard' after you log in to your desktop"
  fi
}

# Run before downloading / installing LinBoard binary
linboard_preflight_setup() {
  linboard_detect_os
  linboard_detect_session

  linboard_log "Detected: ${LINBOARD_OS_ID:-Linux} | ${LINBOARD_DESKTOP:-desktop} | ${LINBOARD_SESSION:-session}"

  if ! linboard_have curl; then
    linboard_install_packages curl || true
  fi

  linboard_install_paste_tools
  linboard_install_gnome_tray
}

# Run after binary is installed to ~/.local/bin
linboard_post_install_setup() {
  local bin_dir="${1:-$HOME/.local/bin}"

  linboard_ensure_path "$bin_dir"

  if "$bin_dir/linboard" install-shortcut 2>/dev/null; then
    linboard_log "Super+V shortcut registered"
  else
    linboard_warn "Shortcut setup failed — run: linboard install-shortcut"
  fi

  linboard_start_app "$bin_dir"

  echo ""
  echo "  ┌─────────────────────────────────────────────┐"
  echo "  │  LinBoard installed successfully!           │"
  echo "  │                                             │"
  echo "  │  Press Super+V (Win+V) to open history      │"
  echo "  │  Tray icon: look at the top panel           │"
  echo "  └─────────────────────────────────────────────┘"
  echo ""
}
