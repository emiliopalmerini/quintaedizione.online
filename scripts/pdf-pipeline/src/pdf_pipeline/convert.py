"""Step 1: PDF → Markdown using marker-pdf."""

import subprocess
from pathlib import Path


def convert_pdf(pdf_path: Path, output_dir: Path, use_ocr: bool = False) -> Path:
    """Convert a PDF to markdown using marker-pdf.

    By default OCR and layout detection are disabled, relying on the PDF's
    embedded text.  Set use_ocr=True to enable the full vision pipeline
    (requires GPU, may crash on large documents).

    Returns the path to the generated markdown file.
    """
    output_dir.mkdir(parents=True, exist_ok=True)

    cmd = [
        "marker_single",
        str(pdf_path),
        "--output_dir", str(output_dir),
        "--output_format", "markdown",
    ]

    if not use_ocr:
        cmd += ["--disable_ocr", "--force_layout_block", "Text"]

    subprocess.run(cmd, check=True)

    # marker_single creates a subdirectory named after the PDF stem
    stem = pdf_path.stem
    result_dir = output_dir / stem
    md_files = list(result_dir.glob("*.md"))

    if not md_files:
        raise FileNotFoundError(
            f"No markdown file found in {result_dir} after conversion"
        )

    return md_files[0]
