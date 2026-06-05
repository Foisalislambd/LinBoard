# LinBoard

[![CI](https://github.com/foisal/linboard/actions/workflows/ci.yml/badge.svg)](https://github.com/foisal/linboard/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**LinBoard** is an open-source clipboard manager for Linux — Windows **Win+V** style history, built with Go.

Works on **GNOME, KDE Plasma, XFCE, Cinnamon**, **X11 and Wayland**.

## Features

- Clipboard history — text, URLs, images
- **Super+V** (Win+V) global hotkey
- Search, pin, keyboard navigation
- System tray background app
- Persistent SQLite storage
- Auto-paste on select (Wayland: `wtype` / `ydotool`, X11: `xdotool`)

## Quick install

```bash
git clone https://github.com/foisal/linboard.git
cd linboard
./scripts/setup.sh    # deps + build (Debian/Ubuntu)
./linboard install    # ~/.local/bin + autostart + Super+V
linboard              # start
```

Or manually:

```bash
make build
./linboard
```

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
| Super+V does nothing | Run `linboard install-shortcut`, ensure LinBoard is running |
| Tray icon missing (GNOME) | Install AppIndicator extension |
| Auto-paste fails (Wayland) | `sudo apt install wtype` |
| Hotkey conflict | Change binding in system keyboard settings |

## License

MIT — see [LICENSE](LICENSE).
