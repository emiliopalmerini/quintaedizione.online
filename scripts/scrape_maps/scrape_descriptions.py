#!/usr/bin/env python3
"""Scrape map descriptions from Dyson Logos blog posts.

Reads dyson_raw.json for source URLs, scrapes 1-2 sentences of flavor text
from each blog post, and outputs descriptions keyed by source URL.

Usage:
    python scripts/scrape_maps/scrape_descriptions.py [--delay N] [--limit N]
"""

import argparse
import json
import re
import sys
import time

import requests
from bs4 import BeautifulSoup

USER_AGENT = "QuintaEdizione-MapScraper/1.0 (D&D 5e Italian fansite)"
DEFAULT_DELAY = 1.5


def fetch(url: str, session: requests.Session) -> BeautifulSoup:
    resp = session.get(url, timeout=30)
    resp.raise_for_status()
    return BeautifulSoup(resp.text, "html.parser")


def extract_description(soup: BeautifulSoup) -> str:
    """Extract the first meaningful paragraph(s) of flavor text from a blog post."""
    # Find the post content area
    content = soup.select_one(".entry-content, .post-entry, .post-content")
    if not content:
        return ""

    paragraphs = []
    for p in content.find_all("p"):
        text = p.get_text(strip=True)

        # Skip tag lists (comma-separated WordPress tags)
        if text.count(",") > 3 and len(text.split(",")[0].split()) <= 3:
            continue

        # Skip empty, very short, or boilerplate
        if len(text) < 30:
            continue
        # Skip paragraphs that are just links/credits
        if text.startswith("This map") and ("Patreon" in text or "patron" in text.lower()):
            continue
        if "released under" in text.lower() or "commercial license" in text.lower():
            continue
        if "free for" in text.lower() and "use" in text.lower():
            continue
        if text.startswith("The ") and "dpi" in text and "square" in text:
            continue
        if "1200 dpi" in text or "300 dpi" in text:
            # Grid dimension descriptions
            if "squares" in text and len(text) < 100:
                continue
        # Skip meta-commentary about map-making process
        if "hesitant to post" in text.lower() or "first attempt" in text.lower():
            continue
        if text.lower().startswith("i'm ") or text.lower().startswith("i am "):
            continue
        if "patreon" in text.lower() or "ko-fi" in text.lower():
            continue

        paragraphs.append(text)

        # Take at most 2 good paragraphs
        if len(paragraphs) >= 2:
            break

    return " ".join(paragraphs)


def main():
    parser = argparse.ArgumentParser(description="Scrape map descriptions")
    parser.add_argument("--input", "-i", default="dyson_raw.json")
    parser.add_argument("--output", "-o", default="descriptions_raw.json")
    parser.add_argument("--delay", "-d", type=float, default=DEFAULT_DELAY)
    parser.add_argument("--limit", "-l", type=int, default=0)
    args = parser.parse_args()

    with open(args.input, encoding="utf-8") as f:
        maps = json.load(f)

    if args.limit > 0:
        maps = maps[:args.limit]

    session = requests.Session()
    session.headers["User-Agent"] = USER_AGENT

    results = {}
    for i, m in enumerate(maps, 1):
        url = m["source_url"]
        name = m["name"]
        print(f"[{i}/{len(maps)}] {name}", file=sys.stderr)

        try:
            soup = fetch(url, session)
            desc = extract_description(soup)
            if desc:
                results[url] = {"name": name, "description": desc}
            else:
                print(f"  WARN: no description found", file=sys.stderr)
                results[url] = {"name": name, "description": ""}
        except requests.RequestException as e:
            print(f"  ERROR: {e}", file=sys.stderr)
            results[url] = {"name": name, "description": ""}

        time.sleep(args.delay)

    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(results, f, indent=2, ensure_ascii=False)
        f.write("\n")

    found = sum(1 for v in results.values() if v["description"])
    print(f"\nDone: {found}/{len(results)} descriptions found", file=sys.stderr)


if __name__ == "__main__":
    main()
