# LinBoard

**Windows-style clipboard history for Linux** — press **Super+V** and pick from everything you've copied.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Latest release](https://img.shields.io/github/v/release/Foisalislambd/LinBoard?label=download)](https://github.com/Foisalislambd/LinBoard/releases)

Works on **GNOME · KDE Plasma · XFCE · Cinnamon** — **X11 and Wayland**.

---

## Install

Copy and paste this in your terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/Foisalislambd/LinBoard/main/scripts/install-release.sh | bash
```

That's it. The script downloads the latest version, installs LinBoard, sets up autostart, and registers **Super+V**.

> **Need auto-paste?** Install `wtype` (Wayland) or `xdotool` (X11) first:
> `sudo apt install wtype xdotool`

---

## Start using LinBoard

```bash
linboard          # start (runs in the system tray)
```

| Action | How |
|--------|-----|
| Open history | **Super+V** (Win key + V) |
| Search | Type in the search box |
| Paste item | Click or press **Enter** |
| Pin item | Press **P** |
| Delete item | Press **Delete** |

LinBoard starts automatically on login after install.

---

## Features

- Clipboard history — text, links, and images
- **Super+V** global hotkey on every major Linux desktop
- Search, pin, and keyboard navigation
- Runs quietly in the system tray
- History saved locally (SQLite, your machine only)
- Auto-paste when you select an item

---

## Manual install

Prefer downloading yourself? Get the package for your CPU from **[Releases](https://github.com/Foisalislambd/LinBoard/releases)**:

| File | For |
|------|-----|
| `linboard-linux-amd64-*.tar.gz` | Most PCs (Intel / AMD) |
| `linboard-linux-arm64-*.tar.gz` | Raspberry Pi 4+, ARM devices |

```bash
mkdir -p ~/linboard-install
tar -xzf linboard-linux-amd64-*.tar.gz -C ~/linboard-install
~/linboard-install/install.sh
```

---

## After install

**Command not found?** Add LinBoard to your PATH:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc
```

**Super+V not working?**

1. Make sure LinBoard is running (`linboard`)
2. Run `linboard install-shortcut`
3. Check Settings → Keyboard → look for **LinBoard**

**No tray icon on GNOME?**  
Install the [AppIndicator extension](https://extensions.gnome.org/extension/615/appindicator-support/).

---

## Upgrade

Run the install command again — it always fetches the latest version:

```bash
curl -fsSL https://raw.githubusercontent.com/Foisalislambd/LinBoard/main/scripts/install-release.sh | bash
pkill linboard; linboard
```

---

## Uninstall

```bash
pkill linboard 2>/dev/null || true
rm -f ~/.local/bin/linboard
rm -f ~/.local/share/applications/linboard.desktop
rm -f ~/.config/autostart/linboard.desktop
rm -rf ~/.config/linboard   # optional — removes history
```

Remove the **LinBoard** entry from Settings → Keyboard if it remains.

---

## Settings

Edit `~/.config/linboard/config.json`:

```json
{
  "max_history": 200,
  "paste_on_select": true,
  "theme": "system"
}
```

| File | What's stored |
|------|-----------------|
| `~/.config/linboard/config.json` | Your settings |
| `~/.config/linboard/data/history.db` | Clipboard history |
| `~/.config/linboard/data/images/` | Copied images |

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `linboard: command not found` | Add `~/.local/bin` to PATH (see above) |
| Super+V does nothing | Start `linboard`, then run `linboard install-shortcut` |
| Tray icon missing (GNOME) | Install [AppIndicator](https://extensions.gnome.org/extension/615/appindicator-support/) |
| Paste doesn't work (Wayland) | `sudo apt install wtype` |
| Paste doesn't work (X11) | `sudo apt install xdotool` |
| Hotkey conflict | Remove the old binding in Settings → Keyboard |

---

## Contributing

Want to build from source or help improve LinBoard? See **[CONTRIBUTING.md](CONTRIBUTING.md)**.

## License

MIT — see [LICENSE](LICENSE).
