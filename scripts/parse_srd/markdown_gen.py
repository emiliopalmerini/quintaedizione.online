"""Classified spans → markdown text.

Converts paragraphs with classified spans into markdown format,
applying appropriate inline formatting (bold, italic) and heading markers.
"""

from __future__ import annotations

from .classify import SpanRole
from .merge import ClassifiedSpan, Paragraph


def _span_to_markdown(span: ClassifiedSpan) -> str:
    """Convert a single classified span to markdown text."""
    text = span.text
    role = span.role

    if role in (SpanRole.BODY_BOLD, SpanRole.SIDEBAR_BOLD):
        return f"**{text.strip()}**" if text.strip() else text
    if role in (SpanRole.BODY_ITALIC, SpanRole.SIDEBAR_ITALIC, SpanRole.STAT_ITALIC):
        return f"*{text.strip()}*" if text.strip() else text
    if role in (SpanRole.BODY_BOLD_ITALIC, SpanRole.SIDEBAR_BOLD_ITALIC, SpanRole.STAT_BOLD_ITALIC):
        return f"_**{text.strip()}**_" if text.strip() else text
    if role == SpanRole.STAT_LABEL:
        return f"**{text.strip()}**" if text.strip() else text

    return text


def paragraph_to_markdown(para: Paragraph, heading_offset: int = 0) -> str:
    """Convert a paragraph to markdown.

    Args:
        para: The paragraph to convert.
        heading_offset: Offset to add to heading levels (for nesting).

    Returns:
        Markdown string.
    """
    role = para.role

    # Headings
    level_map = {
        SpanRole.H1: 1,
        SpanRole.H2: 2,
        SpanRole.H3: 3,
        SpanRole.H4: 4,
        SpanRole.H5: 5,
        SpanRole.H6: 6,
    }

    if role in level_map:
        level = min(level_map[role] + heading_offset, 6)
        title = para.text.strip()
        return f"{'#' * level} {title}"

    # Stat subtitle
    if role == SpanRole.STAT_SUBTITLE:
        return f"*{para.text.strip()}*"

    # Score header
    if role == SpanRole.STAT_SCORE_HEADER:
        return ""  # Skip MOD SALV headers in markdown

    # Body/sidebar/stat text
    parts: list[str] = []
    for span in para.spans:
        parts.append(_span_to_markdown(span))
    return "".join(parts).strip()


def paragraphs_to_markdown(paragraphs: list[Paragraph], heading_offset: int = 0) -> str:
    """Convert a list of paragraphs to a full markdown document."""
    lines: list[str] = []
    for para in paragraphs:
        md = paragraph_to_markdown(para, heading_offset)
        if md:
            lines.append(md)
            lines.append("")  # blank line between paragraphs

    return "\n".join(lines).strip()
