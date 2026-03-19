"""Line assembly, hyphen joining, paragraph grouping, and column handling.

Takes classified raw blocks and produces a flat list of paragraphs with
their semantic roles, handling two-column layout and text continuation.
"""

from __future__ import annotations

from dataclasses import dataclass, field

from .classify import SpanRole, classify_span
from .extract import RawBlock, RawLine, RawSpan
from .profiles import FontProfile


@dataclass(slots=True)
class ClassifiedSpan:
    text: str
    role: SpanRole


@dataclass(slots=True)
class Paragraph:
    """A logical paragraph: one or more lines merged together."""

    spans: list[ClassifiedSpan] = field(default_factory=list)
    role: SpanRole = SpanRole.UNKNOWN
    page_num: int = 0

    @property
    def text(self) -> str:
        return "".join(s.text for s in self.spans)

    def is_heading(self) -> bool:
        return self.role in (
            SpanRole.H1,
            SpanRole.H2,
            SpanRole.H3,
            SpanRole.H4,
            SpanRole.H5,
            SpanRole.H6,
        )


def _dominant_role(spans: list[ClassifiedSpan]) -> SpanRole:
    """Pick the dominant role from a list of classified spans.

    Headings take priority. Among body/sidebar text, pick the first.
    """
    for s in spans:
        if s.role in (
            SpanRole.H1, SpanRole.H2, SpanRole.H3,
            SpanRole.H4, SpanRole.H5, SpanRole.H6,
        ):
            return s.role
    # For stat-related, return the first one
    if spans:
        return spans[0].role
    return SpanRole.UNKNOWN


def _classify_line(
    line: RawLine,
    page_num: int,
    profile: FontProfile | None = None,
) -> list[ClassifiedSpan]:
    """Classify all spans in a line."""
    result: list[ClassifiedSpan] = []
    for span in line.spans:
        role = classify_span(span, profile)
        result.append(ClassifiedSpan(text=span.text, role=role))
    return result


def _is_footer(spans: list[ClassifiedSpan]) -> bool:
    """Check if a line is a page footer."""
    return all(s.role == SpanRole.FOOTER for s in spans)


def _is_toc(spans: list[ClassifiedSpan]) -> bool:
    return all(s.role == SpanRole.TOC for s in spans)


def _is_drop_cap(spans: list[ClassifiedSpan]) -> bool:
    return any(s.role == SpanRole.DROP_CAP for s in spans)


def _should_join_hyphen(prev_text: str, next_text: str) -> bool:
    """Decide whether to join two consecutive text fragments at a hyphen.

    Join when previous line ends with hyphen/soft-hyphen and next starts lowercase.
    """
    if not prev_text or not next_text:
        return False
    if prev_text.rstrip().endswith(("-", "\u00ad")):
        first_char = next_text.lstrip()[:1]
        if first_char and first_char.islower():
            return True
    return False


def _page_midpoint(page_num: int) -> float:
    """Return the x midpoint for A4 pages (595pt width)."""
    return 297.5


def _block_page(block: RawBlock) -> int:
    """Get the page number from the first span of a block."""
    for line in block.lines:
        for span in line.spans:
            return span.page_num
    return 0


def sort_blocks_reading_order(blocks: list[RawBlock]) -> list[RawBlock]:
    """Sort blocks in reading order: per-page left column then right column.

    For each page, reads left column top-to-bottom, then right column
    top-to-bottom, before moving to the next page.
    """
    midpoint = 297.5

    # Group by page
    by_page: dict[int, list[RawBlock]] = {}
    for block in blocks:
        pg = _block_page(block)
        by_page.setdefault(pg, []).append(block)

    result: list[RawBlock] = []
    for pg in sorted(by_page.keys()):
        page_blocks = by_page[pg]
        left = []
        right = []
        for block in page_blocks:
            if block.bbox[0] < midpoint:
                left.append(block)
            else:
                right.append(block)
        left.sort(key=lambda b: b.bbox[1])
        right.sort(key=lambda b: b.bbox[1])
        result.extend(left)
        result.extend(right)

    return result


def _same_paragraph_group(role_a: SpanRole, role_b: SpanRole) -> bool:
    """Check if two roles belong to the same paragraph group."""
    body_roles = {
        SpanRole.BODY, SpanRole.BODY_BOLD, SpanRole.BODY_ITALIC,
        SpanRole.BODY_BOLD_ITALIC, SpanRole.LINK,
    }
    sidebar_roles = {
        SpanRole.SIDEBAR, SpanRole.SIDEBAR_BOLD,
        SpanRole.SIDEBAR_ITALIC, SpanRole.SIDEBAR_BOLD_ITALIC,
    }
    stat_body_roles = {
        SpanRole.STAT_BODY, SpanRole.STAT_ITALIC,
        SpanRole.STAT_BOLD_ITALIC,
    }
    stat_header_roles = {
        SpanRole.STAT_LABEL, SpanRole.STAT_VALUE,
    }
    for group in (body_roles, sidebar_roles, stat_body_roles, stat_header_roles):
        if role_a in group and role_b in group:
            return True
    return False


def blocks_to_paragraphs(
    blocks: list[RawBlock],
    profile: FontProfile | None = None,
) -> list[Paragraph]:
    """Convert sorted raw blocks into a list of paragraphs.

    Handles:
    - Footer filtering
    - Drop cap merging
    - Hyphen joining across lines
    - Paragraph grouping by role continuity
    """
    sorted_blocks = sort_blocks_reading_order(blocks)
    paragraphs: list[Paragraph] = []
    current: Paragraph | None = None

    for block in sorted_blocks:
        for line in block.lines:
            if not line.spans:
                continue

            page_num = line.spans[0].page_num
            classified = _classify_line(line, page_num, profile)

            # Skip footers and TOC
            if _is_footer(classified) or _is_toc(classified):
                continue

            # Handle drop caps — merge into next span
            if _is_drop_cap(classified):
                # Drop cap lines have a large first letter + small continuation
                # Merge all spans text together
                merged_text = "".join(s.text for s in classified)
                classified = [ClassifiedSpan(text=merged_text, role=SpanRole.SIDEBAR)]

            line_role = _dominant_role(classified)

            # Headings always start a new paragraph
            if line_role in (
                SpanRole.H1, SpanRole.H2, SpanRole.H3,
                SpanRole.H4, SpanRole.H5, SpanRole.H6,
            ):
                if current:
                    paragraphs.append(current)
                current = Paragraph(spans=classified, role=line_role, page_num=page_num)
                paragraphs.append(current)
                current = None
                continue

            # Stat block subtitle — always its own paragraph
            if line_role == SpanRole.STAT_SUBTITLE:
                if current:
                    paragraphs.append(current)
                current = Paragraph(spans=classified, role=line_role, page_num=page_num)
                paragraphs.append(current)
                current = None
                continue

            # Score header — always its own paragraph
            if line_role == SpanRole.STAT_SCORE_HEADER:
                if current:
                    paragraphs.append(current)
                current = Paragraph(spans=classified, role=line_role, page_num=page_num)
                paragraphs.append(current)
                current = None
                continue

            # Score values/labels — group together
            if line_role in (SpanRole.STAT_SCORE_LABEL, SpanRole.STAT_SCORE_VALUE):
                if current and current.role in (SpanRole.STAT_SCORE_LABEL, SpanRole.STAT_SCORE_VALUE):
                    current.spans.extend(classified)
                    continue
                if current:
                    paragraphs.append(current)
                current = Paragraph(spans=classified, role=line_role, page_num=page_num)
                continue

            # Stat label/value lines — always own paragraph
            if line_role in (SpanRole.STAT_LABEL, SpanRole.STAT_VALUE):
                if current and current.role in (SpanRole.STAT_LABEL, SpanRole.STAT_VALUE):
                    # Continue the stat line if it wraps
                    current.spans.extend(classified)
                    continue
                if current:
                    paragraphs.append(current)
                current = Paragraph(spans=classified, role=line_role, page_num=page_num)
                continue

            # Body/sidebar/stat text — try to group into paragraphs
            if current and _same_paragraph_group(current.role, line_role):
                # Check for hyphen joining
                prev_text = current.spans[-1].text if current.spans else ""
                next_text = classified[0].text if classified else ""
                if _should_join_hyphen(prev_text, next_text):
                    # Remove trailing hyphen from previous
                    last = current.spans[-1]
                    stripped = last.text.rstrip()
                    if stripped.endswith("\u00ad"):
                        current.spans[-1] = ClassifiedSpan(
                            text=stripped[:-1], role=last.role
                        )
                    elif stripped.endswith("-"):
                        current.spans[-1] = ClassifiedSpan(
                            text=stripped[:-1], role=last.role
                        )
                    current.spans.extend(classified)
                else:
                    # Add a space between lines
                    current.spans.append(ClassifiedSpan(text=" ", role=line_role))
                    current.spans.extend(classified)
                continue

            # Role change — new paragraph
            if current:
                paragraphs.append(current)
            current = Paragraph(spans=classified, role=line_role, page_num=page_num)

    if current:
        paragraphs.append(current)

    return paragraphs
