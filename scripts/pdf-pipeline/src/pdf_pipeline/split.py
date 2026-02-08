"""Step 2: Split a converted markdown (with page markers) into per-collection files."""

import re
from pathlib import Path

from pdf_pipeline.sections import SECTIONS_5_2_1, Section

# Matches <!-- PAGE:N --> markers inserted by the convert step.
PAGE_MARKER = re.compile(r"^<!-- PAGE:(\d+) -->$", re.MULTILINE)


def split_markdown(
    md_path: Path,
    output_dir: Path,
    sections: dict[str, Section] | None = None,
) -> list[Path]:
    """Split a converted markdown file into per-collection files.

    Uses page markers (``<!-- PAGE:N -->``) to locate section boundaries
    based on the edition's section config.

    Returns list of paths to created files.
    """
    if sections is None:
        sections = SECTIONS_5_2_1

    content = md_path.read_text(encoding="utf-8")
    pages = _parse_pages(content)

    output_dir.mkdir(parents=True, exist_ok=True)
    created: list[Path] = []

    for filename, section in sections.items():
        start, end = section.pages
        section_parts: list[str] = []

        for page_num in range(start, end + 1):
            if page_num in pages:
                section_parts.append(pages[page_num])

        if not section_parts:
            print(f"  Warning: no content found for {filename} (pages {start}-{end})")
            continue

        body = "\n\n".join(section_parts)
        out_path = output_dir / filename
        out_path.write_text(f"# {section.header}\n\n{body}\n", encoding="utf-8")
        created.append(out_path)
        print(f"  Created: {out_path} (pages {start}-{end})")

    return created


def _parse_pages(content: str) -> dict[int, str]:
    """Parse page markers and return {page_number: text_content}."""
    pages: dict[int, str] = {}
    matches = list(PAGE_MARKER.finditer(content))

    for i, match in enumerate(matches):
        page_num = int(match.group(1))
        start = match.end()
        end = matches[i + 1].start() if i + 1 < len(matches) else len(content)
        text = content[start:end].strip()
        if text:
            pages[page_num] = text

    return pages
