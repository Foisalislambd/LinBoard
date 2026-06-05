#!/usr/bin/env python3
"""Generate LinBoard icons from assets/source/linboard.png"""

from __future__ import annotations

import sys
from pathlib import Path

from PIL import Image

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "assets" / "source" / "linboard.png"
OUT_DIR = ROOT / "internal" / "assets" / "icons"
HICOLOR = OUT_DIR / "hicolor"

# Linux hicolor + app/tray sizes
SIZES = (16, 22, 24, 32, 48, 64, 128, 256, 512)


def remove_background(img: Image.Image) -> Image.Image:
    """Ensure transparent background; strip leftover black matte."""
    img = img.convert("RGBA")
    px = img.load()
    w, h = img.size

    def is_bg(r: int, g: int, b: int, a: int) -> bool:
        if a < 8:
            return True
        # Pure / near-black (common leftover from export)
        if r < 18 and g < 18 and b < 18:
            return True
        return False

    # Flood-fill from corners to remove black halos
    from collections import deque

    visited = [[False] * w for _ in range(h)]
    q: deque[tuple[int, int]] = deque()

    for x in range(w):
        for y in (0, h - 1):
            q.append((x, y))
    for y in range(h):
        for x in (0, w - 1):
            q.append((x, y))

    while q:
        x, y = q.popleft()
        if x < 0 or y < 0 or x >= w or y >= h or visited[y][x]:
            continue
        visited[y][x] = True
        r, g, b, a = px[x, y]
        if not is_bg(r, g, b, a):
            continue
        px[x, y] = (r, g, b, 0)
        q.append((x + 1, y))
        q.append((x - 1, y))
        q.append((x, y + 1))
        q.append((x, y - 1))

    return img


def trim_and_pad(img: Image.Image, pad_ratio: float = 0.06) -> Image.Image:
    bbox = img.getbbox()
    if not bbox:
        return img
    cropped = img.crop(bbox)
    side = max(cropped.size)
    pad = max(1, int(side * pad_ratio))
    canvas = Image.new("RGBA", (side + pad * 2, side + pad * 2), (0, 0, 0, 0))
    ox = (canvas.width - cropped.width) // 2
    oy = (canvas.height - cropped.height) // 2
    canvas.paste(cropped, (ox, oy), cropped)
    return canvas


def resize_icon(img: Image.Image, size: int) -> Image.Image:
    return img.resize((size, size), Image.Resampling.LANCZOS)


def main() -> int:
    if not SOURCE.exists():
        print(f"Missing source image: {SOURCE}", file=sys.stderr)
        return 1

    print(f"==> Source: {SOURCE}")
    base = remove_background(Image.open(SOURCE))
    base = trim_and_pad(base)
    print(f"==> Processed canvas: {base.size[0]}x{base.size[1]}")

    OUT_DIR.mkdir(parents=True, exist_ok=True)

    for size in SIZES:
        icon = resize_icon(base, size)

        # Flat exports for Go embed / systray
        flat = OUT_DIR / f"linboard-{size}.png"
        icon.save(flat, format="PNG", optimize=True)
        print(f"    wrote {flat.relative_to(ROOT)}")

        # Freedesktop hicolor theme
        app_dir = HICOLOR / f"{size}x{size}" / "apps"
        app_dir.mkdir(parents=True, exist_ok=True)
        themed = app_dir / "linboard.png"
        icon.save(themed, format="PNG", optimize=True)

    # Master symlink-style copy
    master = OUT_DIR / "linboard.png"
    resize_icon(base, 512).save(master, format="PNG", optimize=True)
    print(f"==> Done ({len(SIZES)} sizes + hicolor theme)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
