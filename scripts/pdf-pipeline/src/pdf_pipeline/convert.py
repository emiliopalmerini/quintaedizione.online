"""Step 1: PDF → Markdown using pdftext (no ML models needed)."""

from pathlib import Path

from pdftext.extraction import plain_text_output


def convert_pdf(pdf_path: Path, output_dir: Path) -> Path:
    """Convert a PDF to a single markdown file with page markers.

    Uses pdftext to extract embedded text directly — no OCR or layout
    detection needed for born-digital PDFs.  Each page is delimited by
    a ``<!-- PAGE:N -->`` comment so the split step can locate sections
    by page number.

    Returns the path to the generated markdown file.
    """
    output_dir.mkdir(parents=True, exist_ok=True)

    total_pages = _count_pages(pdf_path)
    lines: list[str] = []

    for page_idx in range(total_pages):
        text = plain_text_output(
            str(pdf_path),
            sort=True,
            hyphens=False,
            page_range=[page_idx],
        )
        # 1-indexed page number for human readability
        lines.append(f"<!-- PAGE:{page_idx + 1} -->")
        lines.append(text.strip())
        lines.append("")  # blank line between pages

    out_path = output_dir / f"{pdf_path.stem}.md"
    out_path.write_text("\n".join(lines), encoding="utf-8")
    return out_path


def _count_pages(pdf_path: Path) -> int:
    import pypdfium2 as pdfium

    doc = pdfium.PdfDocument(str(pdf_path))
    count = len(doc)
    doc.close()
    return count
