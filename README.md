# LinBoard

**LinBoard** is a human-friendly clipboard manager for Linux, inspired by Windows clipboard history (Win+V). Built with Go.

## Features


- **Clipboard history** — Automatically saves text, URLs, and images
- **Global hotkey** — `Super+V` (Win+V) opens history — same as Windows
- **Search** — Filter history instantly
- **Pin items** — Keep important clips permanently (press `P`)
- **Keyboard navigation** — ↑↓ navigate, Enter paste, Del remove, Esc close
- **System tray** — Runs quietly in the background
- **Persistent storage** — History survives reboots (SQLite)
- **Auto-paste** — Selected item is pasted into the active window

## Stack (latest as of June 2026)

| Component | Version |
|-----------|---------|
| Go | **1.26.4** |
| Fyne | **v2.7.4** |
| golang.design/x/clipboard | **v0.7.1** |
| golang.design/x/hotkey | **v0.4.1** |
| modernc.org/sqlite | **v1.51.0** |
| getlantern/systray | **v1.2.2** |

## Quick setup

```bash
./scripts/setup.sh
# or
make setup
```

This installs system dependencies, verifies Go 1.26.4, and builds the binary.

## Build & dev commands

```bash
make build        # build ./linboard
make run          # build + run
make rebuild      # clean + build
make clean        # remove binary
make clean-all    # remove binary + Go cache
make deps         # go mod tidy + download
make test         # run tests
make vet          # static analysis
make install      # install to ~/.local/bin
make stop         # stop running LinBoard
make help         # all commands
```

Same via script: `./scripts/dev.sh <command>`

## Requirements

- Linux with **X11** (recommended) or Wayland
- `xdotool` for auto-paste (optional but recommended):
  ```bash
  sudo apt install xdotool   # Debian/Ubuntu
  sudo dnf install xdotool   # Fedora
  ```
- System tray support (for GNOME: install AppIndicator extension)

### Build dependencies

```bash
# Debian/Ubuntu
sudo apt install gcc libc6-dev libx11-dev libxcursor-dev libxrandr-dev \
  libxinerama-dev libxi-dev libgl1-mesa-dev libxxf86vm-dev

# Fedora
sudo dnf install gcc libX11-devel libXcursor-devel libXrandr-devel \
  libXinerama-devel libXi-devel mesa-libGL-devel libXxf86vm-devel
```

## Install & Run

```bash
git clone <repo-url> linboard
cd linboard
go build -o linboard ./cmd/linboard
./linboard
```

## Usage

| Action | Shortcut |
|--------|----------|
| Open history | `Super+V` (Win key + V) |
| Navigate | `↑` / `↓` |
| Paste selected | `Enter` |
| Pin / Unpin | `P` |
| Delete item | `Delete` |
| Close | `Esc` |

Right-click the tray icon for menu options.

## Configuration

Config file: `~/.config/linboard/config.json`

```json
{
  "max_history": 200,
  "start_minimized": true,
  "paste_on_select": true,
  "theme": "system"
}
```

| Option | Description |
|--------|-------------|
| `max_history` | Max unpinned items to keep |
| `paste_on_select` | Auto-paste when selecting an item |
| `theme` | `light`, `dark`, or `system` |

History hotkey is fixed at **Super+V** (Win key + V), like Windows.

## Data location

- Config: `~/.config/linboard/config.json`
- Database: `~/.config/linboard/data/history.db`
- Images: `~/.config/linboard/data/images/`

## Autostart

Add to `~/.config/autostart/linboard.desktop`:

```ini
[Desktop Entry]
Type=Application
Name=LinBoard
Comment=Clipboard Manager
Exec=/path/to/linboard
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
```

## License

MIT
