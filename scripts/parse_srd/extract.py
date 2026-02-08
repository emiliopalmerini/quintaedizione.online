"""Raw span extraction from pymupdf."""

from __future__ import annotations

from dataclasses import dataclass

import fitz


@dataclass(frozen=True, slots=True)
class RawSpan:
    text: str
    font_name: str
    font_size: float
    color: int
    bbox: tuple[float, float, float, float]  # x0, y0, x1, y1
    page_num: int  # 1-indexed


@dataclass(frozen=True, slots=True)
class RawLine:
    spans: list[RawSpan]
    bbox: tuple[float, float, float, float]


@dataclass(frozen=True, slots=True)
class RawBlock:
    lines: list[RawLine]
    bbox: tuple[float, float, float, float]


def extract_page(page: fitz.Page, page_num: int) -> list[RawBlock]:
    """Extract all text blocks from a single page.

    Args:
        page: A pymupdf Page object.
        page_num: 1-indexed page number.

    Returns:
        List of RawBlock with lines and spans preserving font metadata.
    """
    d = page.get_text("dict")
    blocks: list[RawBlock] = []
    for block in d["blocks"]:
        if block["type"] != 0:  # skip image blocks
            continue
        raw_lines: list[RawLine] = []
        for line in block["lines"]:
            spans: list[RawSpan] = []
            for span in line["spans"]:
                text = span["text"]
                if not text.strip():
                    continue
                spans.append(
                    RawSpan(
                        text=text,
                        font_name=span["font"],
                        font_size=round(span["size"], 1),
                        color=span["color"],
                        bbox=(
                            span["bbox"][0],
                            span["bbox"][1],
                            span["bbox"][2],
                            span["bbox"][3],
                        ),
                        page_num=page_num,
                    )
                )
            if spans:
                raw_lines.append(
                    RawLine(
                        spans=spans,
                        bbox=(
                            line["bbox"][0],
                            line["bbox"][1],
                            line["bbox"][2],
                            line["bbox"][3],
                        ),
                    )
                )
        if raw_lines:
            blocks.append(
                RawBlock(
                    lines=raw_lines,
                    bbox=(
                        block["bbox"][0],
                        block["bbox"][1],
                        block["bbox"][2],
                        block["bbox"][3],
                    ),
                )
            )
    return blocks


def extract_pages(doc: fitz.Document, start: int, end: int) -> list[RawBlock]:
    """Extract blocks from a range of pages (1-indexed, inclusive)."""
    all_blocks: list[RawBlock] = []
    for page_num in range(start, end + 1):
        page = doc[page_num - 1]  # 0-indexed
        all_blocks.extend(extract_page(page, page_num))
    return all_blocks
