"""Pre-merge table detection and markdown formatting.

Operates on list[RawBlock] **before** blocks_to_paragraphs().  Detects table
regions using bbox coordinates (still available at this stage), reconstructs
column/row grids, and replaces table blocks with synthetic RawBlock(s)
containing pre-formatted markdown table text.

Two table types in class pages:
  - Traits box: 2-column key-value pairs (TABLE_HEADER_SMALL labels + SIDEBAR values)
  - Level progression table: N-column grid (TABLE_HEADER_SMALL headers + SIDEBAR data)

Detection trigger: TABLE_HEADER span (10.5pt GillSans-SemiBold #231f20) followed
by TABLE_HEADER_SMALL and SIDEBAR blocks.  Spells/backgrounds use TABLE_HEADER_SMALL
after headings (not TABLE_HEADER), so they won't false-positive.
"""

from __future__ import annotations

from .classify import SpanRole, classify_span
from .extract import RawBlock, RawLine, RawSpan
from .profiles import FontProfile


# Roles that form table content (headers + data cells)
_TABLE_CONTENT_ROLES = {
    SpanRole.TABLE_HEADER_SMALL,
    SpanRole.SIDEBAR,
    SpanRole.SIDEBAR_BOLD,
    SpanRole.SIDEBAR_ITALIC,
    SpanRole.SIDEBAR_BOLD_ITALIC,
    SpanRole.TABLE_BODY,
}

# Roles that stop a table region
_TABLE_STOP_ROLES = {
    SpanRole.H1, SpanRole.H2, SpanRole.H3,
    SpanRole.H4, SpanRole.H5, SpanRole.H6,
    SpanRole.BODY, SpanRole.BODY_BOLD, SpanRole.BODY_ITALIC,
    SpanRole.BODY_BOLD_ITALIC, SpanRole.LINK,
}


def _block_role(block: RawBlock, profile: FontProfile | None = None) -> SpanRole:
    """Dominant role of the first span in a block."""
    for line in block.lines:
        for span in line.spans:
            return classify_span(span, profile)
    return SpanRole.UNKNOWN


def _block_text(block: RawBlock) -> str:
    """All text from a block joined together."""
    parts: list[str] = []
    for line in block.lines:
        for span in line.spans:
            parts.append(span.text)
    return " ".join(parts).strip()


def _block_page(block: RawBlock) -> int:
    """Page number from first span."""
    for line in block.lines:
        for span in line.spans:
            return span.page_num
    return 0


def _cluster_values(values: list[float], tolerance: float) -> list[float]:
    """Cluster nearby float values and return sorted cluster centers."""
    if not values:
        return []
    sorted_vals = sorted(values)
    clusters: list[list[float]] = [[sorted_vals[0]]]
    for v in sorted_vals[1:]:
        if v - clusters[-1][-1] <= tolerance:
            clusters[-1].append(v)
        else:
            clusters.append([v])
    return [sum(c) / len(c) for c in clusters]


def _assign_to_cluster(value: float, centers: list[float]) -> int:
    """Find the cluster index closest to value."""
    best_idx = 0
    best_dist = abs(value - centers[0])
    for i, c in enumerate(centers[1:], 1):
        d = abs(value - c)
        if d < best_dist:
            best_dist = d
            best_idx = i
    return best_idx


class _Cell:
    """A single table cell with position and text."""
    __slots__ = ("x0", "y0", "x1", "y1", "text", "is_header", "page_num")

    def __init__(
        self,
        x0: float, y0: float, x1: float, y1: float,
        text: str, is_header: bool, page_num: int,
    ):
        self.x0 = x0
        self.y0 = y0
        self.x1 = x1
        self.y1 = y1
        self.text = text
        self.is_header = is_header
        self.page_num = page_num


def _extract_cells(
    blocks: list[RawBlock],
    profile: FontProfile | None = None,
) -> list[_Cell]:
    """Extract individual cells from table content blocks.

    Each span is classified individually (not at block level) because
    a single block can contain both TABLE_HEADER_SMALL and SIDEBAR spans
    on the same line (e.g., "Dado Vita" + "D12 per ogni livello...").
    """
    cells: list[_Cell] = []
    for block in blocks:
        for line in block.lines:
            for span in line.spans:
                text = span.text.strip()
                if text:
                    role = classify_span(span, profile)
                    cells.append(_Cell(
                        x0=span.bbox[0], y0=span.bbox[1],
                        x1=span.bbox[2], y1=span.bbox[3],
                        text=text,
                        is_header=(role == SpanRole.TABLE_HEADER_SMALL),
                        page_num=span.page_num,
                    ))
    return cells


# ── Traits box (2-column key-value) ─────────────────────────────────────────


def _group_by_y_adjacency(
    cells: list[_Cell], threshold: float = 14.0,
) -> list[list[_Cell]]:
    """Group cells that are vertically adjacent (within threshold pt)."""
    if not cells:
        return []
    sorted_cells = sorted(cells, key=lambda c: c.y0)
    groups: list[list[_Cell]] = [[sorted_cells[0]]]
    for cell in sorted_cells[1:]:
        prev = groups[-1][-1]
        if cell.y0 - prev.y0 <= threshold:
            groups[-1].append(cell)
        else:
            groups.append([cell])
    return groups


def _build_traits_table(cells: list[_Cell]) -> tuple[list[str], list[list[str]]]:
    """Build a 2-column key-value table from traits box cells.

    Labels are TABLE_HEADER_SMALL cells (left column).
    Values are SIDEBAR cells (right column).
    Multi-line labels/values are merged by y-adjacency.
    """
    labels = [c for c in cells if c.is_header]
    values = [c for c in cells if not c.is_header]

    label_groups = _group_by_y_adjacency(labels, threshold=14.0)
    value_groups = _group_by_y_adjacency(values, threshold=14.0)

    rows: list[list[str]] = []
    used_vg: set[int] = set()

    for lg in label_groups:
        label_text = " ".join(c.text for c in lg)
        lg_y_min = min(c.y0 for c in lg)
        lg_y_max = max(c.y1 for c in lg)

        # Find the value group whose y-range overlaps with the label
        best_idx = -1
        best_overlap = -1.0
        for vi, vg in enumerate(value_groups):
            if vi in used_vg:
                continue
            vg_y_min = min(c.y0 for c in vg)
            vg_y_max = max(c.y1 for c in vg)
            # Overlap: positive means ranges intersect
            overlap = min(lg_y_max, vg_y_max) - max(lg_y_min, vg_y_min)
            # Also allow close proximity (value starts within 5pt of label)
            if overlap < 0 and abs(vg_y_min - lg_y_min) <= 5.0:
                overlap = 0.1
            if overlap > best_overlap:
                best_overlap = overlap
                best_idx = vi

        value_text = ""
        if best_idx >= 0 and best_overlap >= 0:
            used_vg.add(best_idx)
            value_text = " ".join(c.text for c in value_groups[best_idx])

        rows.append([label_text, value_text])

    # No header row — the table title serves as context
    # Use empty header for markdown formatting
    header = ["", ""]
    return header, rows


# ── Level table (N-column grid) ──────────────────────────────────────────────


def _build_level_table(cells: list[_Cell]) -> tuple[list[str], list[list[str]]]:
    """Build an N-column grid table from level progression cells.

    Column headers are TABLE_HEADER_SMALL, data cells are SIDEBAR.
    Each data row is a single line (no multi-line data cells).
    Column headers may span 2 lines and their x-positions may not
    exactly match data columns, so we derive columns from data cells
    and assign headers to the nearest column.
    """
    data_cells = [c for c in cells if not c.is_header]

    if not data_cells:
        return [], []

    # Filter header cells: exclude decorative spanning titles like
    # "—Slot incantesimo per livello di incantesimo—" that appear above
    # the real column headers.  Two filters:
    # 1. Proximity: only keep headers within 28pt of the first data row
    # 2. Content: exclude em-dash decorated text (starts/ends with —)
    min_data_y = min(c.y0 for c in data_cells)
    header_cells = [
        c for c in cells
        if c.is_header
        and min_data_y - c.y0 < 28.0
        and not c.text.startswith("\u2014")
        and not c.text.endswith("\u2014")
    ]

    # Derive column positions from data cells (consistent x-values)
    col_centers = _cluster_values([c.x0 for c in data_cells], tolerance=5.0)
    if not col_centers:
        return [], []

    n_cols = len(col_centers)

    # Assign header cells to nearest data column
    header = [""] * n_cols
    for cell in header_cells:
        col_idx = _assign_to_cluster(cell.x0, col_centers)
        if header[col_idx]:
            header[col_idx] += " " + cell.text
        else:
            header[col_idx] = cell.text

    # Build data rows by y-clustering (tolerance 3pt)
    row_centers = _cluster_values([c.y0 for c in data_cells], tolerance=3.0)
    if not row_centers:
        return header, []

    n_rows = len(row_centers)
    grid: list[list[str]] = [[""] * n_cols for _ in range(n_rows)]

    for cell in data_cells:
        col_idx = _assign_to_cluster(cell.x0, col_centers)
        row_idx = _assign_to_cluster(cell.y0, row_centers)
        existing = grid[row_idx][col_idx]
        if existing:
            grid[row_idx][col_idx] = existing + " " + cell.text
        else:
            grid[row_idx][col_idx] = cell.text

    # Merge continuation rows: rows where only 1-2 columns have content
    # are likely wrapped text from the previous row's long cell.
    merged: list[list[str]] = []
    for row in grid:
        non_empty = sum(1 for c in row if c.strip())
        if merged and non_empty <= 2 and non_empty < n_cols // 2:
            # Merge into previous row
            for ci in range(n_cols):
                if row[ci].strip():
                    if merged[-1][ci]:
                        merged[-1][ci] += " " + row[ci]
                    else:
                        merged[-1][ci] = row[ci]
        else:
            merged.append(list(row))

    return header, merged


# ── Table type detection ─────────────────────────────────────────────────────


def _detect_and_build(cells: list[_Cell]) -> tuple[list[str], list[list[str]]]:
    """Detect table type and build appropriate grid.

    Traits box: TABLE_HEADER_SMALL cells cluster in one x-region (labels).
    Level table: TABLE_HEADER_SMALL cells span multiple x-regions (column headers).
    """
    header_cells = [c for c in cells if c.is_header]
    if not header_cells:
        return [], []

    header_x = _cluster_values([c.x0 for c in header_cells], tolerance=5.0)

    if len(header_x) <= 2:
        # Traits box: labels are in 1-2 x-clusters (one column, maybe wrapping)
        return _build_traits_table(cells)
    else:
        # Level table: headers span many columns
        return _build_level_table(cells)


# ── Markdown formatting ──────────────────────────────────────────────────────


def _format_markdown_grid(header: list[str], rows: list[list[str]]) -> str:
    """Format a markdown pipe table grid (without title)."""
    if not header and not rows:
        return ""

    n_cols = len(header) if header else (max(len(r) for r in rows) if rows else 0)
    if n_cols == 0:
        return ""

    lines: list[str] = []

    # Header row
    h = header if header else [""] * n_cols
    lines.append("| " + " | ".join(h[:n_cols]) + " |")
    # Separator
    lines.append("| " + " | ".join("---" for _ in range(n_cols)) + " |")
    # Data rows
    for row in rows:
        padded = row + [""] * (n_cols - len(row))
        lines.append("| " + " | ".join(padded[:n_cols]) + " |")

    return "\n".join(lines)


def _make_synthetic_blocks(
    title: str,
    grid_text: str,
    page_num: int,
    bbox: tuple[float, float, float, float],
) -> list[RawBlock]:
    """Create synthetic RawBlocks for a table.

    Returns two blocks:
    1. Title block using TABLE_HEADER font (GillSans-SemiBold 10.5pt)
       so it breaks paragraph grouping and doesn't merge with adjacent BODY.
    2. Grid block using BODY font (Cambria 10pt) for the table rows.

    Both preserve the original bbox for correct reading-order sort.
    """
    blocks: list[RawBlock] = []

    if title:
        title_span = RawSpan(
            text=f"**{title}**",
            font_name="GillSans-SemiBold",
            font_size=10.5,
            color=0x231f20,
            bbox=bbox,
            page_num=page_num,
        )
        title_line = RawLine(spans=[title_span], bbox=bbox)
        blocks.append(RawBlock(lines=[title_line], bbox=bbox))

    if grid_text:
        # Offset y slightly so grid sorts after title
        grid_bbox = (bbox[0], bbox[1] + 0.1, bbox[2], bbox[3])
        grid_span = RawSpan(
            text=grid_text,
            font_name="Cambria",
            font_size=10.0,
            color=0x231f20,
            bbox=grid_bbox,
            page_num=page_num,
        )
        grid_line = RawLine(spans=[grid_span], bbox=grid_bbox)
        blocks.append(RawBlock(lines=[grid_line], bbox=grid_bbox))

    return blocks


# ── Table region detection ───────────────────────────────────────────────────


def _count_header_small_cells(
    blocks: list[RawBlock],
    start: int,
    profile: FontProfile | None = None,
) -> int:
    """Count TABLE_HEADER_SMALL cells in consecutive blocks from start."""
    count = 0
    for i in range(start, len(blocks)):
        role = _block_role(blocks[i], profile)
        if role != SpanRole.TABLE_HEADER_SMALL:
            break
        for line in blocks[i].lines:
            for span in line.spans:
                if classify_span(span, profile) == SpanRole.TABLE_HEADER_SMALL:
                    count += 1
    return count


def _find_table_regions(
    blocks: list[RawBlock],
    profile: FontProfile | None = None,
) -> list[tuple[int, int, str]]:
    """Find (start, end, title) of table regions in the block list.

    A table region starts with a TABLE_HEADER block and extends through
    consecutive TABLE_HEADER_SMALL and SIDEBAR/TABLE_BODY blocks.

    Also detects orphaned tables: TABLE_HEADER_SMALL blocks with 3+ column
    headers followed by SIDEBAR data, without a preceding TABLE_HEADER.
    This handles tables that span page boundaries (e.g. weapon table
    continuation).
    """
    regions: list[tuple[int, int, str]] = []
    i = 0
    n = len(blocks)

    while i < n:
        role = _block_role(blocks[i], profile)

        if role == SpanRole.TABLE_HEADER:
            title = _block_text(blocks[i])
            start = i
            j = i + 1

            while j < n:
                r = _block_role(blocks[j], profile)
                if r in _TABLE_CONTENT_ROLES:
                    j += 1
                elif r == SpanRole.TABLE_HEADER:
                    break
                elif r in _TABLE_STOP_ROLES:
                    break
                else:
                    if j + 1 < n and _block_role(blocks[j + 1], profile) in _TABLE_CONTENT_ROLES:
                        j += 1
                    else:
                        break

            if j > start + 1:
                regions.append((start, j, title))
                i = j
            else:
                i += 1

        elif role == SpanRole.TABLE_HEADER_SMALL:
            header_cells = _count_header_small_cells(blocks, i, profile)
            if header_cells >= 3:
                start = i
                j = i + 1
                while j < n and _block_role(blocks[j], profile) == SpanRole.TABLE_HEADER_SMALL:
                    j += 1
                while j < n:
                    r = _block_role(blocks[j], profile)
                    if r in _TABLE_CONTENT_ROLES:
                        j += 1
                    elif r in _TABLE_STOP_ROLES or r == SpanRole.TABLE_HEADER:
                        break
                    else:
                        if j + 1 < n and _block_role(blocks[j + 1], profile) in _TABLE_CONTENT_ROLES:
                            j += 1
                        else:
                            break
                if j > start + 1:
                    regions.append((start, j, ""))
                    i = j
                else:
                    i += 1
            else:
                i += 1
        else:
            i += 1

    return regions


# ── Main entry point ─────────────────────────────────────────────────────────


def process_tables(
    blocks: list[RawBlock],
    profile: FontProfile | None = None,
) -> list[RawBlock]:
    """Replace table regions with pre-formatted markdown table blocks.

    Scans for TABLE_HEADER → TABLE_HEADER_SMALL/SIDEBAR sequences,
    reconstructs the column/row grid from bbox coordinates, and emits
    synthetic BODY blocks containing markdown pipe tables.

    For class pages, "Privilegi" (level) tables may be on a different page
    from their "Tratti" (traits) table.  To keep them together after
    reading-order sort, the level table is placed right after its
    matching traits table (same page/column, y offset).
    """
    regions = _find_table_regions(blocks, profile)
    if not regions:
        return blocks

    # Build a map of "Tratti" regions so "Privilegi" tables can anchor to them
    traits_bboxes: dict[int, tuple[int, tuple[float, float, float, float]]] = {}
    for idx, (start, _end, title) in enumerate(regions):
        if "Tratti" in title:
            traits_bboxes[idx] = (_block_page(blocks[start]), blocks[start].bbox)

    result: list[RawBlock] = []
    prev_end = 0

    for idx, (start, end, title) in enumerate(regions):
        # Copy blocks before this table
        result.extend(blocks[prev_end:start])

        # Extract cells from content blocks.
        # For titled tables, skip the title block (start + 1).
        # For orphan tables (no title), include all blocks from start.
        content_start = start + 1 if title else start
        content_blocks = [
            b for b in blocks[content_start:end]
            if _block_role(b, profile) in _TABLE_CONTENT_ROLES
        ]

        cells = _extract_cells(content_blocks, profile)
        if not cells:
            result.extend(blocks[start:end])
            prev_end = end
            continue

        header, data_rows = _detect_and_build(cells)
        if not header and not data_rows:
            result.extend(blocks[start:end])
            prev_end = end
            continue

        grid_text = _format_markdown_grid(header, data_rows)

        # Determine placement: "Privilegi" tables anchor after their "Tratti"
        page_num = _block_page(blocks[start])
        region_bbox = blocks[start].bbox

        if "Privilegi" in title:
            # Find the nearest preceding "Tratti" region
            for j in range(idx - 1, -1, -1):
                if j in traits_bboxes:
                    anchor_page, anchor_bbox = traits_bboxes[j]
                    page_num = anchor_page
                    # Place just below the traits table (y + 1pt offset)
                    region_bbox = (
                        anchor_bbox[0],
                        anchor_bbox[1] + 1.0,
                        anchor_bbox[2],
                        anchor_bbox[3],
                    )
                    break

        result.extend(
            _make_synthetic_blocks(title, grid_text, page_num, region_bbox)
        )

        prev_end = end

    # Append remaining blocks after last table
    result.extend(blocks[prev_end:])
    return result
