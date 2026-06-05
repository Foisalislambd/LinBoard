## LinBoard @TAG@ 🎉

**Windows-style clipboard history for Linux** — press **Super+V** (Win+V) and pick from everything you've copied.

Works on **GNOME · KDE Plasma · XFCE · Cinnamon** — **X11 and Wayland**.

---

### ⚡ Quick install

```bash
curl -fsSL https://raw.githubusercontent.com/Foisalislambd/LinBoard/main/scripts/install-release.sh | bash
```

Fully automatic — detects your distro, installs dependencies, sets up **Super+V**, and starts LinBoard.

Then press **Super+V** to open your clipboard history.

---

### ✨ Features

- Clipboard history for **text, URLs, and images**
- Global hotkey **Super+V** on every major Linux desktop
- Search, pin, delete, and keyboard navigation
- System tray app with autostart on login
- Local SQLite storage — your data stays on your machine
- Auto-paste when you select an item

---

### 📦 Downloads

| Package | Platform |
|---------|----------|
| `linboard-linux-amd64-v@VERSION@.tar.gz` | Intel / AMD 64-bit |
| `linboard-linux-arm64-v@VERSION@.tar.gz` | ARM64 (Raspberry Pi 4+, ARM laptops) |

Each archive includes the binary, `install.sh`, desktop launcher, and quick start guide.

**Manual install:**

```bash
mkdir -p ~/linboard-install
tar -xzf linboard-linux-amd64-v@VERSION@.tar.gz -C ~/linboard-install
~/linboard-install/install.sh
```

---

### ⌨️ Shortcuts

| Key | Action |
|-----|--------|
| **Super+V** | Open / close history |
| **Enter** | Paste selected item |
| **P** | Pin / unpin |
| **Delete** | Remove item |
| Type | Search history |

---

### 🖥️ Notes

- LinBoard must be running in the tray for **Super+V** to work
- **GNOME users:** [AppIndicator extension](https://extensions.gnome.org/extension/615/appindicator-support/) may be needed for the tray icon
- Settings: `~/.config/linboard/config.json`

---

### 🐛 Issues & feedback

Found a bug? [Open an issue](https://github.com/Foisalislambd/LinBoard/issues)

**License:** [MIT](https://github.com/Foisalislambd/LinBoard/blob/main/LICENSE)
