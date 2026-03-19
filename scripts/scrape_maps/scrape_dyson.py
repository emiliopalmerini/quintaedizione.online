#!/usr/bin/env python3
"""Scrape map metadata from Dyson Logos commercial maps page.

Usage:
    python scripts/scrape_maps/scrape_dyson.py [--output FILE] [--delay SECONDS]

Outputs a JSON array of map metadata (English only) to stdout or a file.
Does NOT download images.
"""

import argparse
import json
import sys
import time
import requests
from bs4 import BeautifulSoup

INDEX_URL = "https://dysonlogos.blog/maps/commercial-maps/"
USER_AGENT = "QuintaEdizione-MapScraper/1.0 (D&D 5e Italian fansite)"
DEFAULT_DELAY = 1.5


def fetch(url: str, session: requests.Session) -> BeautifulSoup:
    resp = session.get(url, timeout=30)
    resp.raise_for_status()
    return BeautifulSoup(resp.text, "html.parser")


def scrape_index(session: requests.Session) -> list[dict]:
    """Scrape the commercial maps index page for map names and detail URLs."""
    soup = fetch(INDEX_URL, session)

    entries = []
    # The page lists maps as linked images/text inside the post content
    content = soup.select_one(".entry-content, .post-entry, .post-content, article")
    if not content:
        content = soup

    seen_urls = set()
    for link in content.find_all("a", href=True):
        href = link["href"]
        # Filter to blog post URLs (contain date pattern) and skip non-map links
        if "dysonlogos.blog/" not in href:
            continue
        # Match pattern like /2025/08/13/slug/
        parts = href.rstrip("/").split("/")
        if len(parts) < 5:
            continue
        # Check for date-like segments
        try:
            year, month, day = parts[-4], parts[-3], parts[-2]
            int(year)
            int(month)
            int(day)
        except (ValueError, IndexError):
            continue

        if href in seen_urls:
            continue
        seen_urls.add(href)

        # Get name from link text or image alt
        name = link.get_text(strip=True)
        if not name:
            img = link.find("img")
            if img:
                name = img.get("alt", "").strip()
        if not name:
            # Derive from slug
            name = parts[-1].replace("-", " ").title()

        entries.append({"name": name, "source_url": href})

    return entries


def scrape_detail(entry: dict, session: requests.Session) -> dict:
    """Scrape a single map detail page for image URL and tags."""
    url = entry["source_url"]
    soup = fetch(url, session)

    result = {
        "name": entry["name"],
        "image_url": "",
        "tags": [],
        "source_url": url,
    }

    # Extract full-res map image URL
    # Look for links to high-res PNGs in wp-content/uploads
    for link in soup.find_all("a", href=True):
        href = link["href"]
        if "wp-content/uploads" in href and href.endswith(".png"):
            # Prefer the version without "nogrid" or "lowres"
            lower = href.lower()
            if "nogrid" not in lower and "lowres" not in lower and "promo" not in lower:
                result["image_url"] = href
                break

    # Fallback: look for large images in the post
    if not result["image_url"]:
        for img in soup.find_all("img", src=True):
            src = img["src"]
            if "wp-content/uploads" in src and src.endswith(".png"):
                lower = src.lower()
                if "nogrid" not in lower and "lowres" not in lower:
                    result["image_url"] = src
                    break

    # Extract WordPress tags
    tag_links = soup.select('a[rel="tag"]')
    tags = []
    skip_tags = {"commercial maps", "maps", "fantasy", "d&d", "dnd", "dungeons & dragons",
                 "rpg", "osr", "old school essentials", "ose", "labyrinth lord",
                 "pathfinder", "5e", "ttrpg", "colour maps", "color maps"}
    for tag_link in tag_links:
        tag_text = tag_link.get_text(strip=True)
        if tag_text.lower() not in skip_tags:
            tags.append(tag_text)

    # Deduplicate plural/singular (keep singular form)
    normalized = {}
    for tag in tags:
        key = tag.rstrip("s").lower() if tag.lower().endswith("s") and len(tag) > 3 else tag.lower()
        if key not in normalized:
            # Keep the shorter (singular) form
            normalized[key] = tag.rstrip("s") if tag.lower().endswith("s") and len(tag) > 3 else tag
    result["tags"] = sorted(normalized.values())

    return result


def main():
    parser = argparse.ArgumentParser(description="Scrape Dyson Logos commercial maps metadata")
    parser.add_argument("--output", "-o", help="Output JSON file (default: stdout)")
    parser.add_argument("--delay", "-d", type=float, default=DEFAULT_DELAY,
                        help=f"Delay between requests in seconds (default: {DEFAULT_DELAY})")
    parser.add_argument("--limit", "-l", type=int, default=0,
                        help="Limit number of maps to scrape (0 = all)")
    args = parser.parse_args()

    session = requests.Session()
    session.headers["User-Agent"] = USER_AGENT

    print(f"Scraping index page: {INDEX_URL}", file=sys.stderr)
    entries = scrape_index(session)
    print(f"Found {len(entries)} maps on index page", file=sys.stderr)

    if args.limit > 0:
        entries = entries[:args.limit]
        print(f"Limiting to {args.limit} maps", file=sys.stderr)

    results = []
    for i, entry in enumerate(entries, 1):
        print(f"[{i}/{len(entries)}] {entry['name']}", file=sys.stderr)
        try:
            result = scrape_detail(entry, session)
            results.append(result)
        except requests.RequestException as e:
            print(f"  ERROR: {e}", file=sys.stderr)
        time.sleep(args.delay)

    output = json.dumps(results, indent=2, ensure_ascii=False)
    if args.output:
        with open(args.output, "w", encoding="utf-8") as f:
            f.write(output + "\n")
        print(f"Wrote {len(results)} maps to {args.output}", file=sys.stderr)
    else:
        print(output)


if __name__ == "__main__":
    main()
