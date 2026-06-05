# LinBoard

[![CI](https://github.com/foisal/linboard/actions/workflows/ci.yml/badge.svg)](https://github.com/foisal/linboard/actions/workflows/ci.yml)
[![Release](https://github.com/foisal/linboard/actions/workflows/release.yml/badge.svg)](https://github.com/foisal/linboard/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**LinBoard** is an open-source clipboard manager for Linux — Windows **Win+V** style history, built with Go.

Works on **GNOME, KDE Plasma, XFCE, Cinnamon**, **X11 and Wayland**.

**New here?** Jump to the [Installation guide](#installation-guide).

## Features

- Clipboard history — text, URLs, images
- **Super+V** (Win+V) global hotkey
- Search, pin, keyboard navigation
- System tray background app
- Persistent SQLite storage
- Auto-paste on select (Wayland: `wtype` / `ydotool`, X11: `xdotool`)

## Installation guide

LinBoard supports **Linux x86_64 (amd64)** and **ARM64 (aarch64)**. You do **not** need Go installed if you use a [release package](https://github.com/foisal/linboard/releases).

### Before you install

| Check | Command |
|-------|---------|
| Supported CPU | `uname -m` → `x86_64` or `aarch64` |
| Session type | `echo $XDG_SESSION_TYPE` → `wayland` or `x11` |
| Desktop | `echo $XDG_CURRENT_DESKTOP` → GNOME, KDE, XFCE, Cinnamon, etc. |

**Supported desktops:** GNOME, KDE Plasma, XFCE, Cinnamon (X11 and Wayland).

**Recommended (auto-paste):** install a paste tool for your session:

```bash
# Debian / Ubuntu / Linux Mint / Pop!_OS
sudo apt update
sudo apt install wtype xdotool

# Fedora
sudo dnf install wtype xdotool

# Arch / Manjaro
sudo pacman -S wtype xdotool
```

| Session | Tool |
|---------|------|
| Wayland | `wtype` (best) or `ydotool` |
| X11 | `xdotool` |

**GNOME only:** install the [AppIndicator](https://extensions.gnome.org/extension/615/appindicator-support/) extension so the tray icon appears.

---

### Method 1 — One-line install (recommended)

Downloads the latest release, installs to `~/.local/bin`, sets autostart, and registers **Super+V**.

```bash
curl -fsSL https://raw.githubusercontent.com/foisal/linboard/main/scripts/install-release.sh | bash
```

Install a **specific version** (replace `v1.0.0` with the tag from [Releases](https://github.com/foisal/linboard/releases)):

```bash
LINBOARD_VERSION=v1.0.0 curl -fsSL https://raw.githubusercontent.com/foisal/linboard/main/scripts/install-release.sh | bash
```

**Custom install location:**

```bash
LINBOARD_BIN_DIR="$HOME/bin" curl -fsSL https://raw.githubusercontent.com/foisal/linboard/main/scripts/install-release.sh | bash
```

Then continue to [After install](#after-install).

---

### Method 2 — Manual download (offline / no curl pipe)

**Step 1 — Pick the right file** from [GitHub Releases](https://github.com/foisal/linboard/releases):

| Download | Your system |
|----------|-------------|
| `linboard-linux-amd64-v*.tar.gz` | Most PCs and laptops (Intel/AMD 64-bit) |
| `linboard-linux-arm64-v*.tar.gz` | Raspberry Pi 4+, ARM laptops, ARM VMs |

**Step 2 — Extract and install:**

```bash
mkdir -p ~/linboard-install
tar -xzf linboard-linux-amd64-v1.0.0.tar.gz -C ~/linboard-install
cd ~/linboard-install
ls
# You should see: linboard  install.sh  linboard.desktop  QUICKSTART.txt

chmod +x install.sh
./install.sh
```

**Step 3 — Verify checksum (optional):**

```bash
sha256sum -c linboard-linux-amd64-v1.0.0.tar.gz.sha256
```

Then continue to [After install](#after-install).

---

### Method 3 — Build from source

For developers or distros without a pre-built package. Requires **Go 1.26+**, **gcc**, and GUI libraries.

#### Debian / Ubuntu / Mint

```bash
git clone https://github.com/foisal/linboard.git
cd linboard
./scripts/setup.sh
```

`setup.sh` installs system dependencies, Go (if needed), and builds `./linboard`.

Then install for daily use:

```bash
./linboard install
```

#### Fedora

```bash
sudo dnf install gcc pkg-config libX11-devel libXcursor-devel libXrandr-devel \
  libXinerama-devel libXi-devel mesa-libGL-devel libXxf86vm-devel \
  libayatana-appindicator3-devel dbus-devel wtype xdotool

git clone https://github.com/foisal/linboard.git
cd linboard
go mod tidy
go build -o linboard ./cmd/linboard
./linboard install
```

#### Arch / Manjaro

```bash
sudo pacman -S gcc pkgconf libx11 libxcursor libxrandr libxinerama libxi \
  mesa libxxf86vm libayatana-appindicator3 dbus wtype xdotool go

git clone https://github.com/foisal/linboard.git
cd linboard
go mod tidy
go build -o linboard ./cmd/linboard
./linboard install
```

#### Build only (no system install)

```bash
make build
./linboard
```

---

### After install

**1. Add `~/.local/bin` to PATH** (if `linboard` is not found):

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

For **Zsh**:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**2. Start LinBoard:**

```bash
linboard
```

LinBoard runs in the **system tray**. On next login it starts automatically (autostart entry).

**3. Open clipboard history:**

Press **Super+V** (Windows key + V).

**4. Confirm the shortcut is registered:**

- **GNOME / Ubuntu:** Settings → Keyboard → Keyboard Shortcuts → Custom → look for **LinBoard**
- **KDE:** System Settings → Shortcuts → Custom Shortcuts
- **XFCE / Cinnamon:** Keyboard settings → Application shortcuts

If missing, run:

```bash
linboard install-shortcut
```

**5. Test it:**

1. Copy some text (`Ctrl+C`)
2. Press **Super+V**
3. Select an item — it should paste into the active window (needs `wtype` / `xdotool`)

**6. Check version:**

```bash
linboard version
```

---

### What gets installed

| Path | Purpose |
|------|---------|
| `~/.local/bin/linboard` | Main program |
| `~/.local/share/applications/linboard.desktop` | App menu launcher |
| `~/.config/autostart/linboard.desktop` | Start on login |
| `~/.config/linboard/` | Settings and clipboard database |

---

### Upgrade

**One-line (latest release):**

```bash
curl -fsSL https://raw.githubusercontent.com/foisal/linboard/main/scripts/install-release.sh | bash
```

**Manual:** download the new `.tar.gz`, extract, run `./install.sh` again (overwrites the binary).

**From source:**

```bash
cd linboard
git pull
make rebuild
./linboard install
```

Restart LinBoard after upgrading:

```bash
pkill linboard; linboard
```

---

### Uninstall

```bash
pkill linboard 2>/dev/null || true
rm -f ~/.local/bin/linboard
rm -f ~/.local/share/applications/linboard.desktop
rm -f ~/.config/autostart/linboard.desktop
```

Remove settings and history (optional):

```bash
rm -rf ~/.config/linboard
```

Remove the **Super+V** shortcut manually in your desktop’s keyboard settings (search for **LinBoard**).

## Commands

| Command | Description |
|---------|-------------|
| `linboard` | Start (system tray) |
| `linboard toggle` | Show/hide history |
| `linboard install` | Install binary + autostart + shortcut |
| `linboard install-shortcut` | Register Super+V only |
| `make run` | Build and run |

## Hotkey (Super+V)

LinBoard picks the best method automatically:

| Environment | Method |
|-------------|--------|
| KDE Plasma 6+ (Wayland) | xdg-desktop-portal |
| GNOME / Ubuntu | GNOME Custom Shortcut (`gsettings`) |
| KDE Plasma | KHotKeys (`khotkeysrc`) |
| XFCE | `xfconf-query` |
| Cinnamon | Cinnamon keybindings |
| X11 session | X11 key grab |

All desktop shortcuts run `linboard toggle`, which talks to the running app via IPC.

**Verify:** Settings → Keyboard → look for **LinBoard**

## Requirements

### Runtime (auto-paste)

| Session | Package |
|---------|---------|
| Wayland | `wtype` (recommended) or `ydotool` |
| X11 | `xdotool` |

```bash
# Debian/Ubuntu
sudo apt install wtype xdotool

# Fedora
sudo dnf install wtype xdotool

# Arch
sudo pacman -S wtype xdotool
```

### Build

```bash
# Debian/Ubuntu
sudo apt install gcc pkg-config libx11-dev libxcursor-dev libxrandr-dev \
  libxinerama-dev libxi-dev libgl1-mesa-dev libxxf86vm-dev \
  libayatana-appindicator3-dev libdbus-1-dev

# Fedora
sudo dnf install gcc pkg-config libX11-devel libXcursor-devel libXrandr-devel \
  libXinerama-devel libXi-devel mesa-libGL-devel libXxf86vm-devel \
  libayatana-appindicator3-devel dbus-devel

# Arch
sudo pacman -S gcc pkgconf libx11 libxcursor libxrandr libxinerama libxi \
  mesa libxxf86vm libayatana-appindicator3 dbus
```

GNOME users: install **AppIndicator** extension for tray icon.

## Configuration

`~/.config/linboard/config.json`

```json
{
  "max_history": 200,
  "start_minimized": true,
  "paste_on_select": true,
  "theme": "system"
}
```

## Data

| Path | Content |
|------|---------|
| `~/.config/linboard/config.json` | Settings |
| `~/.config/linboard/data/history.db` | History |
| `~/.config/linboard/data/images/` | Image clips |

## Development

```bash
make help
make build
make vet
make test
```

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Troubleshooting

| Problem | Fix |
|---------|-----|
| `linboard: command not found` | Add `~/.local/bin` to PATH — see [After install](#after-install) |
| Super+V does nothing | Run `linboard` first (must stay in tray), then `linboard install-shortcut` |
| Shortcut not in settings | `linboard install-shortcut` — command must be `~/.local/bin/linboard toggle` |
| Tray icon missing (GNOME) | Install [AppIndicator](https://extensions.gnome.org/extension/615/appindicator-support/) extension |
| Auto-paste fails (Wayland) | `sudo apt install wtype` (or `ydotool`) |
| Auto-paste fails (X11) | `sudo apt install xdotool` |
| Hotkey conflict | Remove duplicate binding in Settings → Keyboard |
| Install script: no release found | Publish a release on GitHub first, or set `LINBOARD_VERSION=v1.0.0` |
| Wrong architecture | Use `uname -m` — `x86_64` → amd64, `aarch64` → arm64 package |

## License

MIT — see [LICENSE](LICENSE).
