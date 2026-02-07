"""Step 2: Split a single markdown file into per-collection files."""

import re
from pathlib import Path

# Section mapping for SRD v5.2.1.
# Maps H1 header text (case-insensitive) to output filename.
# Extend or add new mappings for future editions.
SECTION_MAP_5_2_1: dict[str, str] = {
    "incantesimi": "incantesimi.md",
    "mostri": "mostri.md",
    "classi": "classi.md",
    "armature": "armature.md",
    "armi": "armi.md",
    "backgrounds": "backgrounds.md",
    "talenti": "talenti.md",
    "equipaggiamento": "equipaggiamenti.md",
    "oggetti magici": "oggetti_magici.md",
    "regole": "regole.md",
    "animali": "animali.md",
    "cavalcature, veicoli e servizi": "cavalcature_veicoli_items.md",
    "servizi": "servizi.md",
    "strumenti": "strumenti.md",
}

# H1 header pattern: lines starting with exactly "# " (not "## ")
H1_PATTERN = re.compile(r"^# (.+)", re.MULTILINE)


def split_markdown(
    md_path: Path,
    output_dir: Path,
    section_map: dict[str, str] | None = None,
) -> list[Path]:
    """Split a markdown file into per-collection files based on H1 headers.

    Returns list of paths to created files.
    """
    if section_map is None:
        section_map = SECTION_MAP_5_2_1

    # Normalize keys to lowercase for case-insensitive matching
    normalized_map = {k.lower(): v for k, v in section_map.items()}

    content = md_path.read_text(encoding="utf-8")
    sections = _extract_sections(content)

    output_dir.mkdir(parents=True, exist_ok=True)
    created: list[Path] = []

    for header, body in sections:
        key = header.strip().rstrip(".").lower()
        filename = normalized_map.get(key)

        if filename is None:
            print(f"  Skipping unmapped section: '{header}'")
            continue

        out_path = output_dir / filename
        # Write section with its H1 header
        out_path.write_text(f"# {header}\n\n{body}", encoding="utf-8")
        created.append(out_path)
        print(f"  Created: {out_path}")

    return created


def _extract_sections(content: str) -> list[tuple[str, str]]:
    """Extract (header, body) pairs from markdown, split on H1 headers."""
    sections: list[tuple[str, str]] = []
    matches = list(H1_PATTERN.finditer(content))

    for i, match in enumerate(matches):
        header = match.group(1).strip()
        start = match.end()
        end = matches[i + 1].start() if i + 1 < len(matches) else len(content)
        body = content[start:end].strip()
        sections.append((header, body))

    return sections
