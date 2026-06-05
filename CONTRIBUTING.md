# Contributing to LinBoard

Thank you for helping make LinBoard work great on every Linux desktop!

## Development setup

```bash
git clone https://github.com/your-org/linboard.git
cd linboard
./scripts/setup.sh   # Debian/Ubuntu deps + build
make run
```

## Code guidelines

- Go 1.26+
- Match existing package layout under `internal/`
- Test on **GNOME Wayland** and at least one of: KDE, XFCE, or X11
- Keep hotkey logic in `internal/hotkey/` — one file per desktop environment
- Use `fyne.DoFromGoroutine` for all UI updates from background threads

## Pull requests

1. Fork and create a feature branch
2. `make vet && make build`
3. Describe which desktop environment(s) you tested
4. Open a PR with a clear summary

## Reporting bugs

Include:

- Distro and version (e.g. Ubuntu 24.04)
- Desktop (GNOME, KDE Plasma, XFCE, …)
- Session type (`echo $XDG_SESSION_TYPE`)
- Log output from `./linboard`

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
