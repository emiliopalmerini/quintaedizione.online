# /// script
# requires-python = ">=3.12"
# dependencies = ["pymupdf4llm>=0.0.17"]
# ///
"""Convert the Italian SRD PDF into per-collection markdown files.

Usage:
    uv run scripts/pdf-to-markdown.py data/ita/IT_SRD_CC_v5.2.1.pdf --output-dir data/ita/lists/
"""

import argparse
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

        # Prepend the section header
        content = f"# {section.header}\n\n{md}\n"

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
