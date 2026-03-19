#!/usr/bin/env python3
"""Download map images from Dyson Logos using curated metadata.

Usage:
    python scripts/scrape_maps/download_images.py [--delay SECONDS] [--limit N]

Reads mappe_curated.json (which has immagine_url) and downloads images
to web/static/img/mappe/ using the slug-based filename from the curated data.
"""

import argparse
import json
import sys
import time
from pathlib import Path

import requests

USER_AGENT = "QuintaEdizione-MapScraper/1.0 (D&D 5e Italian fansite)"
DEFAULT_DELAY = 1.0
OUTPUT_DIR = Path(__file__).resolve().parents[2] / "web" / "static" / "img" / "mappe"


def main():
    parser = argparse.ArgumentParser(description="Download Dyson Logos map images")
    parser.add_argument("--input", "-i", default="mappe_curated.json", help="Input curated JSON")
    parser.add_argument("--delay", "-d", type=float, default=DEFAULT_DELAY, help="Delay between downloads")
    parser.add_argument("--limit", "-l", type=int, default=0, help="Limit number of downloads (0 = all)")
    args = parser.parse_args()

    with open(args.input, encoding="utf-8") as f:
        maps = json.load(f)

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    session = requests.Session()
    session.headers["User-Agent"] = USER_AGENT

    to_download = [m for m in maps if m.get("immagine_url")]
    if args.limit > 0:
        to_download = to_download[:args.limit]

    skipped = 0
    downloaded = 0
    errors = 0

    for i, m in enumerate(to_download, 1):
        filename = m["immagine"]
        dest = OUTPUT_DIR / filename

        if dest.exists():
            skipped += 1
            continue

        url = m["immagine_url"]
        print(f"[{i}/{len(to_download)}] {filename}", file=sys.stderr)

        try:
            resp = session.get(url, timeout=60)
            resp.raise_for_status()
            dest.write_bytes(resp.content)
            downloaded += 1
        except requests.RequestException as e:
            print(f"  ERROR: {e}", file=sys.stderr)
            errors += 1

        time.sleep(args.delay)

    print(f"\nDone: {downloaded} downloaded, {skipped} skipped (exist), {errors} errors", file=sys.stderr)


if __name__ == "__main__":
    main()
