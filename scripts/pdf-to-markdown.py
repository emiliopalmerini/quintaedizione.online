# /// script
# requires-python = ">=3.12"
# dependencies = ["pymupdf4llm>=0.0.17"]
# ///
"""Convert the Italian SRD PDF into per-collection markdown files.

Usage:
    uv run scripts/pdf-to-markdown.py data/ita/IT_SRD_CC_v5.2.1.pdf --output-dir data/ita/lists/
"""

import argparse
import re
import sys
from dataclasses import dataclass
from pathlib import Path

import pymupdf4llm


@dataclass(frozen=True)
class Section:
    header: str
    pages: tuple[int, int]  # 1-indexed, inclusive


# SRD v5.2.1 Italian — 405 pages total.
# Page numbers taken from the PDF table of contents.
SECTIONS: dict[str, Section] = {
    "classi.md": Section("Classi", (32, 92)),
    "backgrounds.md": Section("Backgrounds", (93, 97)),
    "talenti.md": Section("Talenti", (98, 100)),
    "armi.md": Section("Armi", (101, 103)),
    "armature.md": Section("Armature", (104, 105)),
    "strumenti.md": Section("Strumenti", (106, 107)),
    "equipaggiamenti.md": Section("Equipaggiamenti", (108, 112)),
    "cavalcature_veicoli_items.md": Section("Cavalcature e Veicoli", (113, 114)),
    "servizi.md": Section("Servizi", (115, 117)),
    "incantesimi.md": Section("Incantesimi", (118, 201)),
    "regole.md": Section("Regole", (202, 231)),
    "oggetti_magici.md": Section("Oggetti Magici", (232, 288)),
    "mostri.md": Section("Mostri", (289, 384)),
    "animali.md": Section("Animali", (385, 405)),
}


def _strip_bold(text: str) -> str:
    """Remove bold markdown markers from text."""
    return text.replace("**", "")


def remove_page_numbers(md: str) -> str:
    """Remove PDF page footer/header lines like '**108** System Reference Document 5.2.1'."""
    return re.sub(
        r"^\*\*\d+\*\*\s+System Reference Document.*$",
        "",
        md,
        flags=re.MULTILINE,
    )


def promote_bold_to_h2(md: str) -> str:
    """Promote standalone bold lines to H2 headers.

    A line that is entirely a single bold span (e.g. ``**Longsword**``)
    becomes ``## Longsword``.

    Exceptions:
    - Lines containing ``:**`` (key-value patterns like ``**Caratteristica:** Forza``)
    - Lines starting with ``_**`` (italic+bold inline text)
    """
    lines = md.split("\n")
    result = []
    for line in lines:
        stripped = line.strip()
        if (
            stripped.startswith("**")
            and not stripped.startswith("_**")
            and stripped.endswith("**")
            and ":**" not in stripped
            # Ensure it's a single bold span: only two ** markers (open and close)
            and stripped.count("**") == 2
        ):
            title = stripped[2:-2]
            result.append(f"## {title}")
        else:
            result.append(line)
    return "\n".join(result)


def deduplicate_consecutive_h2(md: str) -> str:
    """Remove duplicate consecutive H2 headers (ignoring blank lines between them)."""
    lines = md.split("\n")
    result = []
    last_h2 = None
    for line in lines:
        stripped = line.strip()
        if stripped.startswith("## ") and not stripped.startswith("### "):
            if stripped == last_h2:
                continue  # Skip duplicate
            last_h2 = stripped
        elif stripped:  # Non-blank, non-H2 line resets tracking
            last_h2 = None
        result.append(line)
    return "\n".join(result)


def strip_intro_text(md: str) -> str:
    """Remove all text before the first H2 header, keeping only the H1 and entities."""
    lines = md.split("\n")
    h1_line = None
    first_h2_idx = None

    for i, line in enumerate(lines):
        stripped = line.strip()
        if stripped.startswith("# ") and not stripped.startswith("## "):
            h1_line = line
        if stripped.startswith("## ") and not stripped.startswith("### "):
            first_h2_idx = i
            break

    if first_h2_idx is None:
        return md  # No H2 found, return as-is

    if h1_line is not None:
        return h1_line + "\n\n" + "\n".join(lines[first_h2_idx:])

    return "\n".join(lines[first_h2_idx:])


def normalize_headers(md: str, section_title: str) -> str:
    """Normalize markdown headers to match Go parser expectations.

    The Go parsers expect:
      - A single H1 with the section title (e.g. ``# Animali``)
      - H2 headers for each entity (e.g. ``## Alce``)
      - H3/H4 for sub-sections within entities

    pymupdf4llm produces header levels based on font size which don't
    match this convention. This function:
      1. Replaces all existing H1 headers with a single canonical one
      2. Promotes H3 (``###``) to H2 (``##``) — entity names
      3. Promotes H4 (``####``) to H3 (``###``) — sub-sections
      4. Strips bold markers from all headers
    """

    def replace_header(match: re.Match) -> str:
        hashes = match.group(1)
        title = _strip_bold(match.group(2)).strip()
        level = len(hashes)
        if level == 1:
            return ""  # Remove PDF H1s; we prepend our own
        # Promote: ## stays ##, ### becomes ##, #### becomes ###
        new_level = max(level - 1, 2)
        return f"{'#' * new_level} {title}"

    result = re.sub(r"^(#{1,6})\s+(.+)$", replace_header, md, flags=re.MULTILINE)

    # Clean up multiple consecutive blank lines left by removed H1s
    result = re.sub(r"\n{3,}", "\n\n", result)

    return f"# {section_title}\n\n{result.strip()}\n"


def extract_and_split(pdf_path: Path, output_dir: Path) -> None:
    """Extract markdown from PDF and split into per-collection files."""
    if not pdf_path.exists():
        print(f"Error: PDF not found: {pdf_path}", file=sys.stderr)
        sys.exit(1)

    output_dir.mkdir(parents=True, exist_ok=True)

    print(f"Extracting markdown from {pdf_path} ...")
    chunks = pymupdf4llm.to_markdown(str(pdf_path), page_chunks=True)
    print(f"  Got {len(chunks)} page chunks")

    for filename, section in SECTIONS.items():
        start, end = section.pages
        # page_chunks returns 0-indexed pages; section pages are 1-indexed
        section_chunks = chunks[start - 1 : end]
        md = "\n\n".join(chunk["text"] for chunk in section_chunks)

        md = remove_page_numbers(md)
        md = promote_bold_to_h2(md)
        content = normalize_headers(md, section.header)
        content = strip_intro_text(content)
        content = deduplicate_consecutive_h2(content)

        out_path = output_dir / filename
        out_path.write_text(content, encoding="utf-8")
        print(f"  Wrote {out_path} (pages {start}-{end})")


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Convert SRD PDF to per-collection markdown files"
    )
    parser.add_argument("pdf", type=Path, help="Path to the SRD PDF file")
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path("data/ita/lists"),
        help="Output directory for markdown files (default: data/ita/lists)",
    )
    args = parser.parse_args()

    extract_and_split(args.pdf, args.output_dir)
    print("Done!")


if __name__ == "__main__":
    main()
