## LinBoard — Windows-style clipboard history for Linux

Press **Super+V** to open your clipboard history. Click an item to paste it into the previous window — just like Windows Win+V.

### Supported desktops

- GNOME (Ubuntu, Fedora, etc.)
- KDE Plasma
- XFCE, Cinnamon, MATE
- X11 and Wayland

### Highlights

- Clipboard history for text, URLs, and images
- **Super+V** global hotkey
- **CopyQ-style auto-paste** — built-in uinput, no extra tools required
- Search, pin, keyboard navigation
- System tray integration
- Local JSON storage — your data stays on your machine

### Install

```bash
curl -fsSL https://raw.githubusercontent.com/Foisalislambd/LinBoard/main/scripts/install-release.sh | bash
```

After install, run `linboard setup-paste` and log out/in if prompted.

### Paste support

| Platform | Backend |
|----------|---------|
| All Linux (Wayland + X11) | Built-in uinput |
| X11 fallback | xdotool (optional) |
