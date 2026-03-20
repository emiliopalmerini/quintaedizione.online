#!/usr/bin/env python3
"""Scrape direct image URLs from Dyson Logos blog posts."""

import json
import re
import sys
import time

import requests
from bs4 import BeautifulSoup

DATA_FILE = "data/mappe/mappe.json"


def find_map_image_url(blog_url, nome_originale):
    """Extract the full-res map image URL from a Dyson Logos blog post."""
    r = requests.get(blog_url, timeout=15)
    r.raise_for_status()
    soup = BeautifulSoup(r.text, "html.parser")

    entry = soup.find("div", class_="entry-content") or soup

    # Collect all wp-content/uploads images that link to full-res versions
    candidates = []
    for a_tag in entry.find_all("a", href=re.compile(r"wp-content/uploads.*\.(png|jpg|jpeg)$", re.I)):
        href = a_tag.get("href", "")
        img = a_tag.find("img")
        if img and "wp-content/uploads" in img.get("src", ""):
            candidates.append(href)

    if not candidates:
        # Fallback: look for large images without parent links
        for img in entry.find_all("img", src=re.compile(r"wp-content/uploads")):
            src = img.get("src", "")
            # Strip resize params
            src = re.sub(r"\?w=\d+", "", src)
            candidates.append(src)

    if len(candidates) == 1:
        return candidates[0]

    if len(candidates) > 1:
        # Prefer the one whose filename matches the nome_originale slug
        slug = re.sub(r"[^a-z0-9]+", "-", nome_originale.lower()).strip("-")
        for c in candidates:
            if slug in c.lower():
                return c
        # Otherwise return the first one (usually the main map)
        return candidates[0]

    return None


def main():
    with open(DATA_FILE) as f:
        data = json.load(f)

    found = 0
    not_found = []

    for i, m in enumerate(data):
        url = m.get("url_originale", "")
        nome = m.get("nome_originale", "")

        if not url:
            not_found.append((m["slug"], "no url_originale"))
            continue

        try:
            img_url = find_map_image_url(url, nome)
            if img_url:
                m["url_immagine_originale"] = img_url
                found += 1
            else:
                not_found.append((m["slug"], "no image found"))
        except Exception as e:
            not_found.append((m["slug"], str(e)))

        if (i + 1) % 20 == 0:
            print(f"  {i + 1}/{len(data)} processed...", file=sys.stderr)

        time.sleep(0.3)

    with open(DATA_FILE, "w") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)

    print(f"\nDone: {found}/{len(data)} image URLs found", file=sys.stderr)
    if not_found:
        print(f"\nMissing ({len(not_found)}):", file=sys.stderr)
        for slug, reason in not_found:
            print(f"  {slug}: {reason}", file=sys.stderr)


if __name__ == "__main__":
    main()
