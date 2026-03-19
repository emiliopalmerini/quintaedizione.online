#!/usr/bin/env python3
"""Convert map PNG images to lossless WebP format.

Usage:
    python scripts/scrape_maps/convert_webp.py [--delete-originals]

Lossless WebP typically halves the size of these line-art map PNGs.
Images exceeding WebP's 16383px limit are downscaled to fit.
"""

import argparse
import sys
from pathlib import Path

from PIL import Image

Image.MAX_IMAGE_PIXELS = None  # These are large maps, not bombs

IMG_DIR = Path(__file__).resolve().parents[2] / "web" / "static" / "img" / "mappe"
WEBP_MAX_DIM = 16383


def main():
    parser = argparse.ArgumentParser(description="Convert map PNGs to lossless WebP")
    parser.add_argument("--delete-originals", action="store_true", help="Delete PNG files after conversion")
    args = parser.parse_args()

    pngs = sorted(IMG_DIR.glob("*.png"))
    print(f"Found {len(pngs)} PNG files", file=sys.stderr)

    converted = 0
    skipped = 0
    resized = 0
    total_saved = 0

    for i, png in enumerate(pngs, 1):
        webp = png.with_suffix(".webp")
        if webp.exists():
            skipped += 1
            continue

        img = Image.open(png)
        w, h = img.size

        # Downscale if exceeding WebP dimension limit
        if w > WEBP_MAX_DIM or h > WEBP_MAX_DIM:
            scale = WEBP_MAX_DIM / max(w, h)
            new_w, new_h = int(w * scale), int(h * scale)
            img = img.resize((new_w, new_h), Image.LANCZOS)
            resized += 1
            print(f"[{i}/{len(pngs)}] {png.name} (resized {w}x{h} → {new_w}x{new_h})", file=sys.stderr)
        else:
            print(f"[{i}/{len(pngs)}] {png.name}", file=sys.stderr)

        img.save(webp, "WEBP", lossless=True)

        saved = png.stat().st_size - webp.stat().st_size
        total_saved += saved
        converted += 1

        if args.delete_originals:
            png.unlink()

    print(f"\nDone: {converted} converted, {skipped} skipped, {resized} resized", file=sys.stderr)
    print(f"Space saved: {total_saved / 1024 / 1024:.0f} MB", file=sys.stderr)


if __name__ == "__main__":
    main()
