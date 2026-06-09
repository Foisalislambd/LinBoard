# LinBoard

**Windows-style clipboard history for Linux** — press **Super+V** and pick from everything you've copied.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Latest release](https://img.shields.io/github/v/release/Foisalislambd/LinBoard?label=download)](https://github.com/Foisalislambd/LinBoard/releases)

Works on **GNOME · KDE Plasma · XFCE · Cinnamon · MATE** — **X11 and Wayland**.

---

## Install

Copy and paste this in your terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/Foisalislambd/LinBoard/main/scripts/install-release.sh | bash
```

That's it — fully automatic. The installer will:

- Detect your Linux distro and CPU (apt, dnf, pacman, zypper)
- Install dependencies (tray support, `xclip` for images, auto-paste setup)
- Configure **uinput** for click-to-paste on all desktops
- Install LinBoard, autostart, and **Super+V** shortcut
- Add `~/.local/bin` to your PATH and start LinBoard

After install, **log out and back in once** if `linboard setup-paste` asks — this activates the `input` group for auto-paste.

---

## Start using LinBoard

```bash
linboard          # start (runs in the system tray)
```

| Action | How |
|--------|-----|
| Open history | **Super+V** (Win key + V) |
| Search | Type in the search box |
| Paste item | Click or press **Enter** (auto-pastes to previous window) |
| Pin item | Press **P** |
| Delete item | Press **Delete** |

LinBoard starts automatically on login after install.

---

## Features

- Clipboard history — text, links, and images
- **Super+V** global hotkey on every major Linux desktop
- **CopyQ-style auto-paste** — built-in uinput (no ydotool/wtype needed)
- Search, pin, and keyboard navigation
- Runs quietly in the system tray
- History saved locally (SQLite, your machine only)

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

1. Make sure LinBoard is running (`linboard` or `linboard-start`)
2. Run `linboard install-shortcut`
3. Check Settings → Keyboard → look for **LinBoard**

**Paste not working?**

```bash
linboard setup-paste   # shows status + fix steps
linboard doctor        # shortcut + paste check
```

Then **log out and log back in** if prompted.

**No tray icon on GNOME?**  
Install the [AppIndicator extension](https://extensions.gnome.org/extension/615/appindicator-support/).

---

## Upgrade

Run the install command again — it always fetches the latest version and restarts LinBoard in the background:

```bash
curl -fsSL https://raw.githubusercontent.com/Foisalislambd/LinBoard/main/scripts/install-release.sh | bash
```

To restart manually after an upgrade: `linboard-start` (runs in the background; survives terminal close). LinBoard also starts on login via `~/.config/autostart/` — there is no systemd service.

---

## Uninstall

```bash
pkill linboard 2>/dev/null || true
rm -f ~/.local/bin/linboard ~/.local/bin/linboard-start
rm -f ~/.local/share/applications/linboard.desktop
rm -f ~/.config/autostart/linboard.desktop
rm -rf ~/.config/linboard   # optional — removes history
sudo rm -f /etc/udev/rules.d/99-linboard-uinput.rules  # optional
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
| Paste doesn't work | Run `linboard setup-paste`, then log out/in |
| Paste works after install but not from autostart | Use `linboard-start` (installed automatically) |
| Stops when terminal closes after upgrade | Re-run the install command, or `linboard-start` — do not run bare `linboard` in a terminal |
| Hotkey conflict | Remove the old binding in Settings → Keyboard |

### Auto-paste by desktop

| Environment | How paste works |
|-------------|-----------------|
| **GNOME / KDE / XFCE Wayland** | Built-in uinput (`input` group) |
| **X11** | Built-in uinput, or `xdotool` fallback |
| **All** | CopyQ-style: remembers window → copy → focus back → Ctrl+V |

---

## Contributing

Want to build from source or help improve LinBoard? See **[CONTRIBUTING.md](CONTRIBUTING.md)**.

## License

MIT — see [LICENSE](LICENSE).
