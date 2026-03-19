# /// script
# requires-python = ">=3.12"
# dependencies = ["pymupdf>=1.24.0"]
# ///
"""CLI entry point and orchestrator for SRD PDF parsing.

Usage:
    uv run scripts/parse_srd <pdf> --output-dir ./output
    uv run scripts/parse_srd <pdf> --profile 5.1 --output-dir ./output
    uv run scripts/parse_srd <pdf> --debug-page 297
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

import fitz

from ._cli import resolve_profile
from .classify import SpanRole, classify_span
from .extract import extract_page, extract_pages
from .heading_tree import build_heading_tree, walk_tree
from .merge import blocks_to_paragraphs
from .markdown_gen import paragraphs_to_markdown
from .profiles import FontProfile
from .section_split import SectionDef
from .tables import process_tables


def debug_page(doc: fitz.Document, page_num: int, profile: FontProfile) -> None:
    """Dump classified spans for a single page for calibration."""
    blocks = extract_page(doc[page_num - 1], page_num)
    blocks = process_tables(blocks, profile)
    paragraphs = blocks_to_paragraphs(blocks, profile=profile)

    print(f"\n{'=' * 80}")
    print(f"Page {page_num} — {len(blocks)} blocks, {len(paragraphs)} paragraphs")
    print(f"{'=' * 80}\n")

    # Raw spans with classification
    for block in blocks:
        for line in block.lines:
            for span in line.spans:
                role = classify_span(span, profile)
                text = span.text.strip()[:70]
                print(
                    f"  {span.font_size:5.1f}  {span.font_name:<40s}  "
                    f"#{span.color:06x}  [{role.name:<20s}]  {text}"
                )
        print()

    # Merged paragraphs
    print(f"\n{'─' * 80}")
    print("PARAGRAPHS:")
    print(f"{'─' * 80}\n")
    for i, para in enumerate(paragraphs):
        text = para.text[:100]
        print(f"  [{para.role.name:<20s}]  {text}")

    # Heading tree
    print(f"\n{'─' * 80}")
    print("HEADING TREE:")
    print(f"{'─' * 80}\n")
    tree = build_heading_tree(paragraphs)
    walk_tree(tree)

    # Markdown output
    print(f"\n{'─' * 80}")
    print("MARKDOWN:")
    print(f"{'─' * 80}\n")
    md = paragraphs_to_markdown(paragraphs)
    print(md[:3000])


def debug_section(
    doc: fitz.Document,
    section_name: str,
    sections: list[SectionDef],
    profile: FontProfile,
) -> None:
    """Dump heading tree for a section."""
    section = None
    for s in sections:
        if s.name.lower() == section_name.lower() or s.output_file == section_name:
            section = s
            break
    if not section:
        print(f"Unknown section: {section_name}")
        print(f"Available: {', '.join(s.name for s in sections)}")
        sys.exit(1)

    blocks = extract_pages(doc, section.pages[0], section.pages[1])
    blocks = process_tables(blocks, profile)
    paragraphs = blocks_to_paragraphs(blocks, profile=profile)

    print(f"\n{'=' * 80}")
    print(f"Section: {section.name} (pages {section.pages[0]}-{section.pages[1]})")
    print(f"Blocks: {len(blocks)}, Paragraphs: {len(paragraphs)}")
    print(f"{'=' * 80}\n")

    tree = build_heading_tree(paragraphs)
    walk_tree(tree)


def run_parsers(
    doc: fitz.Document,
    output_dir: Path,
    sections: list[SectionDef],
    profile: FontProfile,
) -> None:
    """Run all section parsers and write JSON output."""
    from .profiles import PROFILE_51
    if profile is PROFILE_51:
        from .parsers_51 import get_parser
    else:
        from .parsers import get_parser

    output_dir.mkdir(parents=True, exist_ok=True)

    # Group sections by output file to merge results
    outputs: dict[str, list] = {}

    for section in sections:
        parser = get_parser(section.parser_name)
        if parser is None:
            print(f"  SKIP {section.name} — no parser '{section.parser_name}'")
            continue

        print(f"  Parsing {section.name} (pages {section.pages[0]}-{section.pages[1]})...")

        blocks = extract_pages(doc, section.pages[0], section.pages[1])
        blocks = process_tables(blocks, profile)
        paragraphs = blocks_to_paragraphs(blocks, profile=profile)
        tree = build_heading_tree(paragraphs)

        result = parser(section, paragraphs, tree)

        if section.output_file not in outputs:
            outputs[section.output_file] = []

        if isinstance(result, list):
            outputs[section.output_file].extend(result)
        elif isinstance(result, dict):
            outputs[section.output_file].append(result)

    for filename, data in outputs.items():
        out_path = output_dir / filename
        out_path.write_text(
            json.dumps(data, ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )
        print(f"  Wrote {out_path} ({len(data)} entries)")


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Parse Italian SRD PDF into structured JSON"
    )
    parser.add_argument("pdf", type=Path, help="Path to the SRD PDF file")
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path("output"),
        help="Output directory for JSON files (default: output)",
    )
    parser.add_argument(
        "--profile",
        type=str,
        choices=["5.1", "5.2.1"],
        default=None,
        help="SRD version profile (default: 5.2.1)",
    )
    parser.add_argument(
        "--debug-page",
        type=int,
        metavar="N",
        help="Dump classified spans for page N and exit",
    )
    parser.add_argument(
        "--debug-section",
        type=str,
        metavar="NAME",
        help="Dump heading tree for a section and exit",
    )

    args = parser.parse_args()

    if not args.pdf.exists():
        print(f"Error: PDF not found: {args.pdf}", file=sys.stderr)
        sys.exit(1)

    font_profile, sections = resolve_profile(args.profile)
    print(f"Using profile: {font_profile.name}")

    doc = fitz.open(str(args.pdf))
    print(f"Opened {args.pdf} ({len(doc)} pages)")

    if args.debug_page:
        debug_page(doc, args.debug_page, font_profile)
        return

    if args.debug_section:
        debug_section(doc, args.debug_section, sections, font_profile)
        return

    print("Running parsers...")
    run_parsers(doc, args.output_dir, sections, font_profile)

    from .quality import validate_output
    validate_output(args.output_dir)
    print("\nDone!")


if __name__ == "__main__":
    main()
