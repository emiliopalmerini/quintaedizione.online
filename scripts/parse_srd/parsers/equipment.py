"""Equipment parser — weapons, armor, tools, gear, mounts, services.

Handles two types of equipment entries:
- **Table-based** (weapons, armor): data is in markdown pipe tables
  extracted from the PDF. Uses pattern matching for robust extraction
  since pymupdf table columns can shift or merge.
- **Heading-based** (tools, gear, services): each item is an H4/H5
  heading with a price in parentheses and descriptive body content.

Category is determined by the H2 heading in the tree, not by the
section page range, to avoid cross-boundary misclassification.
"""

from __future__ import annotations

import re

from ..heading_tree import HeadingNode
from ..markdown_gen import paragraphs_to_markdown
from ..merge import Paragraph
from ..schemas import EquipmentItem
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

# ── H2 heading → category mapping ───────────────────────────────────────────

_H2_CATEGORY: dict[str, str] = {
    "Armi": "weapons",
    "Armature": "armor",
    "Strumenti": "tools",
    "Equipaggiamento d'avventura": "gear",
    "Cavalcature e veicoli": "mounts",
    "Spese dello stile di vita": "services",
}

# Regex to detect a price in parentheses — indicates an actual equipment item.
_PRICE_RE = re.compile(
    r"\((?:.*?(?:\d+[\.\d]*\s*(?:mo|ma|mr)|gratis|variabil[ei]).*?)\)",
    re.IGNORECASE,
)

# Patterns for identifying cell roles in weapon/armor tables
_DAMAGE_RE = re.compile(r"^\d+d\d+\s+\w+|^1\s+\w+ante")
_WEIGHT_RE = re.compile(r"^\d+[,.]?\d*\s*kg$")
_COST_RE = re.compile(r"^\d+[,.]?\d*\s*m[oar]$|^gratis$", re.IGNORECASE)
_WEAPON_SUBCAT_RE = re.compile(r"Armi\s+(da\s+mischia|a\s+distanza)")
_ARMOR_SUBCAT_RE = re.compile(r"Armatura\s+(leggera|media|pesante)|Scudo\s*\(")

# Known mastery properties for pattern detection
_MASTERIES = {
    "Colpo di striscio", "Colpo di Striscio",
    "Doppio fendente", "Fiaccare", "Graffio",
    "Lentezza", "Rovesciamento", "Spinta", "Vessazione",
}


def _is_item_title(title: str) -> bool:
    """Return True if the title looks like an equipment item (has price)."""
    return bool(_PRICE_RE.search(title))


# ── Markdown table parsing ──────────────────────────────────────────────────


def _collect_markdown_tables(node: HeadingNode) -> list[list[list[str]]]:
    """Extract all markdown pipe tables from a node's full subtree."""
    lines: list[str] = []

    def _gather(n: HeadingNode) -> None:
        md = paragraphs_to_markdown(n.content)
        lines.extend(md.splitlines())
        for child in n.children:
            _gather(child)

    _gather(node)

    tables: list[list[list[str]]] = []
    current_table: list[list[str]] = []

    for line in lines:
        stripped = line.strip()
        if stripped.startswith("|"):
            if not stripped.endswith("|"):
                stripped += "|"  # fix truncated rows (e.g., Scudo)
            cells = [c.strip() for c in stripped.split("|")]
            cells = cells[1:-1]  # trim empty first/last from split
            if cells and all(re.match(r"^-+$", c) or c == "" for c in cells):
                continue
            current_table.append(cells)
        else:
            if current_table:
                tables.append(current_table)
                current_table = []
    if current_table:
        tables.append(current_table)

    return tables


def _clean_name(raw: str) -> str:
    """Strip bold/italic markdown from a name string."""
    name = raw.strip()
    name = re.sub(r"\*\*(.+?)\*\*", r"\1", name)
    name = re.sub(r"\*(.+?)\*", r"\1", name)
    name = re.sub(r"_(.+?)_", r"\1", name)
    return name.strip()


# ── Weapon table parsing (pattern-based) ─────────────────────────────────────


def _split_name_damage(text: str) -> tuple[str, str, str]:
    """Split a merged name+damage+properties cell.

    E.g., "Martello da guerra 1d8 contundenti Versatile (1d10)"
    → ("Martello da guerra", "1d8 contundenti", "Versatile (1d10)")
    """
    # Find where damage dice pattern starts
    m = re.search(r"\b(\d+d\d+\s+\w+)", text)
    if not m:
        m = re.search(r"\b(1\s+\w+ante)", text)  # "1 perforante"
    if m:
        name = text[:m.start()].strip()
        rest = text[m.start():].strip()
        # Split rest into damage and properties
        # Damage type words: taglienti, contundenti, perforanti
        dm = re.match(
            r"(\d+d\d+\s+(?:taglienti|contundenti|perforanti)|1\s+perforante)\s*(.*)",
            rest,
        )
        if dm:
            return name, dm.group(1), dm.group(2)
        return name, rest, ""
    return text, "", ""


def _classify_cells(cells: list[str]) -> dict[str, str]:
    """Classify non-empty cells by pattern into semantic roles.

    Returns dict with keys: name, damage, properties, mastery, weight, cost.
    """
    result: dict[str, str] = {}
    remaining: list[str] = []

    for cell in cells:
        cell = cell.strip()
        if not cell:
            continue

        if _WEIGHT_RE.match(cell):
            result.setdefault("weight", cell)
        elif _COST_RE.match(cell):
            result.setdefault("cost", cell)
        elif cell in _MASTERIES:
            result.setdefault("mastery", cell)
        elif _DAMAGE_RE.match(cell):
            # Split merged damage+properties like "1d8 contundenti Due mani"
            dm = re.match(
                r"(\d+d\d+\s+(?:taglienti|contundenti|perforanti)"
                r"|1\s+perforante)\s*(.*)",
                cell,
            )
            if dm:
                result.setdefault("damage", dm.group(1))
                if dm.group(2):
                    remaining.append(dm.group(2))
            else:
                result.setdefault("damage", cell)
        else:
            remaining.append(cell)

    return result, remaining


def _parse_weapon_tables(node: HeadingNode) -> list[EquipmentItem]:
    """Parse weapon items from markdown tables under the Armi H2 node."""
    tables = _collect_markdown_tables(node)
    items: list[EquipmentItem] = []
    seen_ids: set[str] = set()

    subcategory = ""

    for table in tables:
        if not table:
            continue

        # Detect weapon table
        flat = " ".join(" ".join(row) for row in table)
        if not re.search(r"\d+d\d+|perforant|taglien|contunden", flat):
            continue

        # Skip header row(s)
        data_start = 0
        for i, row in enumerate(table):
            if any("Nome" in c or "Danni" in c for c in row):
                data_start = i + 1
                break

        for row in table[data_start:]:
            non_empty = [c.strip() for c in row if c.strip()]
            if not non_empty:
                continue

            # Subcategory in col0 applies to NEXT rows, not the item
            # on the same row (e.g., "Armi a distanza da guerra" row has
            # "Tridente" as name — still belongs to previous subcategory).
            col0 = row[0].strip() if row else ""
            new_weapon_subcat = ""
            if col0 and _WEAPON_SUBCAT_RE.search(col0):
                new_weapon_subcat = _clean_name(col0)

            # Get the name: column 1 if it exists and col0 is empty/subcategory
            name = ""
            name_col_idx = -1
            if len(row) > 1 and row[1].strip():
                name = _clean_name(row[1])
                name_col_idx = 1
            elif col0 and not _WEAPON_SUBCAT_RE.search(col0):
                name = _clean_name(col0)
                name_col_idx = 0

            if not name:
                # No item on this row; apply subcategory immediately
                if new_weapon_subcat:
                    subcategory = new_weapon_subcat
                continue

            # Check if name has merged damage data
            if re.search(r"\d+d\d+", name) or re.search(r"\b1\s+perforante", name):
                split_name, split_damage, split_props = _split_name_damage(name)
                if split_name:
                    name = split_name

            # Classify remaining cells (after name)
            other_cells = [
                c.strip() for i, c in enumerate(row)
                if i != name_col_idx and i != 0 and c.strip()
            ]

            classified, remaining = _classify_cells(other_cells)

            damage = classified.get("damage", "")
            mastery = classified.get("mastery", "")
            weight = classified.get("weight", "")
            cost = classified.get("cost", "")

            # Properties: remaining cells that aren't already classified
            # Filter out subcategory text and known non-property values
            properties = ""
            prop_candidates = [
                r for r in remaining
                if not _WEAPON_SUBCAT_RE.search(r)
                and r != name
                and r != "—"
                and not _DAMAGE_RE.match(r)
            ]
            if prop_candidates:
                properties = ", ".join(prop_candidates)

            # If name had merged data, use split values
            if "split_damage" in dir() and split_damage and not damage:
                damage = split_damage
            if "split_props" in dir() and split_props and not properties:
                properties = split_props

            item_id = slugify(name)
            if item_id in seen_ids:
                # Still apply deferred subcategory even for duplicates
                if new_weapon_subcat:
                    subcategory = new_weapon_subcat
                continue
            seen_ids.add(item_id)

            desc_parts = []
            if damage:
                desc_parts.append(f"**Danni:** {damage}")
            if properties and properties != "—":
                desc_parts.append(f"**Proprietà:** {properties}")
            if mastery and mastery != "—":
                desc_parts.append(f"**Padronanza:** {mastery}")
            if weight and weight != "—":
                desc_parts.append(f"**Peso:** {weight}")
            if cost and cost != "—":
                desc_parts.append(f"**Costo:** {cost}")

            items.append(EquipmentItem(
                id=item_id,
                name=name,
                category="weapons",
                subcategory=subcategory,
                properties={
                    k: v for k, v in [
                        ("danni", damage),
                        ("proprietà", properties),
                        ("padronanza", mastery),
                        ("peso", weight),
                        ("costo", cost),
                    ] if v and v != "—"
                },
                description="\n\n".join(desc_parts),
            ))

            # Apply deferred subcategory after appending the item
            if new_weapon_subcat:
                subcategory = new_weapon_subcat

    return items


# ── Armor table parsing ─────────────────────────────────────────────────────


def _parse_armor_tables(node: HeadingNode) -> list[EquipmentItem]:
    """Parse armor items from markdown tables under the Armature H2 node."""
    tables = _collect_markdown_tables(node)
    items: list[EquipmentItem] = []
    seen_ids: set[str] = set()

    for table in tables:
        if not table:
            continue

        flat = " ".join(" ".join(row) for row in table)
        if "Classe Armatura" not in flat:
            continue

        # Find header row and column mapping
        header_idx = -1
        col_map: dict[str, int] = {}
        for i, row in enumerate(table):
            for ci, cell in enumerate(row):
                if "Classe Armatura" in cell:
                    header_idx = i
                    break
            if header_idx >= 0:
                # Map known headers
                for ci, cell in enumerate(table[header_idx]):
                    cl = cell.strip()
                    if "Classe Armatura" in cl:
                        col_map["ac"] = ci
                    elif cl == "Forza":
                        col_map["strength"] = ci
                    elif "Furtività" in cl:
                        col_map["stealth"] = ci
                    elif cl == "Peso":
                        col_map["weight"] = ci
                    elif cl == "Costo":
                        col_map["cost"] = ci
                break

        if header_idx < 0:
            continue

        name_col = 1  # Name is always in column 1
        subcategory = ""
        don_doff = ""

        for row in table[header_idx + 1:]:
            n = len(row)
            col0 = row[0].strip() if n > 0 else ""
            name = _clean_name(row[name_col]) if n > name_col else ""

            # Subcategory in col0 applies to NEXT rows, not the item
            # on the same row (e.g., "Scudo (...)" row has "Armatura a
            # piastre" as name — still belongs to previous subcategory).
            new_subcat = ""
            if col0 and _ARMOR_SUBCAT_RE.search(col0):
                cleaned = _clean_name(col0)
                paren_match = re.search(r"\((.+)\)", cleaned)
                new_subcat = re.sub(r"\s*\(.+\)", "", cleaned)
                don_doff = paren_match.group(1) if paren_match else ""
                if not name:
                    subcategory = new_subcat
                    continue

            if not name:
                continue

            item_id = slugify(name)
            if item_id in seen_ids:
                continue
            seen_ids.add(item_id)

            def _get(key: str) -> str:
                ci = col_map.get(key)
                if ci is not None and ci < n:
                    return row[ci].strip()
                return ""

            ac = _get("ac")
            strength = _get("strength")
            stealth = _get("stealth")
            weight = _get("weight")
            cost = _get("cost")

            # Handle merged weight+cost (e.g., "32,5 kg 1.500 mo")
            if weight and not cost and re.search(r"kg.*mo", weight):
                m = re.match(r"(.+?kg)\s+(.+?mo)", weight)
                if m:
                    weight = m.group(1).strip()
                    cost = m.group(2).strip()

            # Fallback: scan all cells for weight/cost by pattern
            if not weight or not cost:
                for cell in row:
                    cell = cell.strip()
                    if not cell:
                        continue
                    if not weight and _WEIGHT_RE.match(cell):
                        weight = cell
                    elif not cost and _COST_RE.match(cell) and cell != "—":
                        cost = cell

            desc_parts = []
            if ac:
                desc_parts.append(f"**Classe Armatura (CA):** {ac}")
            if strength and strength != "—":
                desc_parts.append(f"**Forza:** {strength}")
            if stealth and stealth != "—":
                desc_parts.append(f"**Furtività:** {stealth}")
            if weight and weight != "—":
                desc_parts.append(f"**Peso:** {weight}")
            if cost and cost != "—":
                desc_parts.append(f"**Costo:** {cost}")

            items.append(EquipmentItem(
                id=item_id,
                name=name,
                category="armor",
                subcategory=subcategory,
                properties={
                    k: v for k, v in [
                        ("classe_armatura", ac),
                        ("forza", strength),
                        ("furtività", stealth),
                        ("vestizione", don_doff),
                        ("peso", weight),
                        ("costo", cost),
                    ] if v and v != "—"
                },
                description="\n\n".join(desc_parts),
            ))

            # Apply deferred subcategory after appending the item
            if new_subcat:
                subcategory = new_subcat

    return items


# ── Simple table parsing (mounts, vehicles, etc.) ─────────────────────────────

# Regex matching a bold-only line like **Cavalcature e altri animali**
_BOLD_TITLE_RE = re.compile(r"^\*\*(.+?)\*\*$")


def _collect_titled_tables(
    node: HeadingNode,
) -> list[tuple[str, list[list[str]]]]:
    """Extract markdown tables with their preceding bold title.

    Returns a list of (title, rows) pairs. The title is the text from
    the last ``**bold line**`` before the table starts, or "" if none.
    """
    lines: list[str] = []

    def _gather(n: HeadingNode) -> None:
        md = paragraphs_to_markdown(n.content)
        lines.extend(md.splitlines())
        for child in n.children:
            _gather(child)

    _gather(node)

    tables: list[tuple[str, list[list[str]]]] = []
    current_table: list[list[str]] = []
    last_title = ""
    table_title = ""

    for line in lines:
        stripped = line.strip()
        if stripped.startswith("|"):
            if not stripped.endswith("|"):
                stripped += "|"
            cells = [c.strip() for c in stripped.split("|")]
            cells = cells[1:-1]
            if cells and all(re.match(r"^-+$", c) or c == "" for c in cells):
                continue
            if not current_table:
                table_title = last_title
            current_table.append(cells)
        else:
            if current_table:
                tables.append((table_title, current_table))
                current_table = []
                table_title = ""
            m = _BOLD_TITLE_RE.match(stripped)
            if m:
                last_title = m.group(1)

    if current_table:
        tables.append((table_title, current_table))

    return tables


def _parse_simple_tables(
    node: HeadingNode, category: str,
) -> list[EquipmentItem]:
    """Parse items from simple Oggetto/Costo tables (mounts, vehicles, etc.)."""
    titled_tables = _collect_titled_tables(node)
    items: list[EquipmentItem] = []
    seen_ids: set[str] = set()

    for subcategory, table in titled_tables:
        if not table:
            continue

        # Find header row with "Oggetto" or "Nave" column
        header_idx = -1
        name_col = -1
        col_map: dict[str, int] = {}
        for i, row in enumerate(table):
            for ci, cell in enumerate(row):
                cl = cell.strip()
                if cl in ("Oggetto", "Nave"):
                    header_idx = i
                    name_col = ci
                    break
            if header_idx >= 0:
                for ci, cell in enumerate(table[header_idx]):
                    cl = cell.strip()
                    if cl == "Costo":
                        col_map["cost"] = ci
                    elif cl == "Peso":
                        col_map["weight"] = ci
                    elif "Capacità" in cl:
                        col_map["capacity"] = ci
                    elif cl == "Velocità":
                        col_map["speed"] = ci
                    elif cl == "CA":
                        col_map["ac"] = ci
                    elif cl == "PF":
                        col_map["hp"] = ci
                break

        if header_idx < 0:
            continue

        sub_heading = ""  # e.g., "Sella" for saddle sub-items
        data_rows = table[header_idx + 1:]

        for ri, row in enumerate(data_rows):
            n = len(row)
            raw_name = row[name_col].strip() if name_col < n else ""

            # Handle sub-items with name in col1 (e.g., "" | "Da galoppo")
            if not raw_name and n > name_col + 1 and row[name_col + 1].strip():
                sub_name = _clean_name(row[name_col + 1])
                name = f"{sub_heading} {sub_name}" if sub_heading else sub_name
            else:
                name = _clean_name(raw_name)
                # Detect merged name+sub-heading: if next row has empty col0
                # and a name in col1, the last word of this name is a sub-heading.
                next_is_sub = (
                    ri + 1 < len(data_rows)
                    and not data_rows[ri + 1][name_col].strip()
                    and len(data_rows[ri + 1]) > name_col + 1
                    and data_rows[ri + 1][name_col + 1].strip()
                )
                if next_is_sub and name:
                    # Split: last word is the sub-heading
                    # e.g., "Nutrimento (al giorno) Sella" → item + "Sella"
                    parts = name.rsplit(" ", 1)
                    if len(parts) == 2:
                        name, sub_heading = parts
                    else:
                        sub_heading = name
                        name = ""
                else:
                    sub_heading = ""

            if not name:
                continue

            item_id = slugify(name)
            if item_id in seen_ids:
                continue
            seen_ids.add(item_id)

            # Collect values from mapped columns
            props: dict[str, str] = {}
            for key in col_map:
                ci = col_map[key]
                if ci < n:
                    val = row[ci].strip()
                    if val and val != "—":
                        props[key] = val

            # Fallback: scan all cells for cost by pattern
            # (columns can shift due to empty cells in the header)
            if "cost" not in props:
                for cell in row:
                    cell = cell.strip()
                    if not cell:
                        continue
                    if _COST_RE.match(cell):
                        props["cost"] = cell
                        break

            desc_parts: list[str] = []
            _LABEL_MAP = {
                "cost": "Costo", "weight": "Peso",
                "capacity": "Capacità di trasporto",
                "speed": "Velocità", "ac": "CA", "hp": "PF",
            }
            for key, label in _LABEL_MAP.items():
                if key in props:
                    desc_parts.append(f"**{label}:** {props[key]}")

            items.append(EquipmentItem(
                id=item_id,
                name=name,
                category=category,
                subcategory=subcategory,
                properties=props,
                description="\n\n".join(desc_parts),
            ))

    return items


# ── Main parser ─────────────────────────────────────────────────────────────


@register("equipment")
def parse_equipment(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[EquipmentItem]:
    """Parse the unified equipment section (pages 101-117).

    Category is determined by the H2 heading in the tree, not by the
    section name. This avoids misclassification when PDF content
    crosses page boundaries.
    """
    items: list[EquipmentItem] = []

    def _walk(nodes: list[HeadingNode], category: str = "", subcategory: str = "") -> None:
        for node in nodes:
            if node.level == 1:
                _walk(node.children, category, subcategory)
                continue

            if node.level == 2:
                h2_title = node.title.strip()
                new_category = None
                for key, val in _H2_CATEGORY.items():
                    if h2_title.lower() == key.lower():
                        new_category = val
                        break
                if new_category is None:
                    continue

                if new_category == "weapons":
                    items.extend(_parse_weapon_tables(node))
                elif new_category == "armor":
                    items.extend(_parse_armor_tables(node))
                elif new_category == "mounts":
                    items.extend(_parse_simple_tables(node, new_category))

                _walk(node.children, new_category, h2_title)
                continue

            # H3+ nodes — heading-based item extraction
            name = node.title.strip()
            content_md = paragraphs_to_markdown(node.content)

            if node.children:
                if _is_item_title(name) and content_md.strip():
                    items.append(EquipmentItem(
                        id=slugify(name),
                        name=name,
                        category=category,
                        subcategory=subcategory,
                        properties={},
                        description=content_md,
                    ))
                child_sub = name if not _is_item_title(name) else subcategory
                _walk(node.children, category, child_sub)
            else:
                if not _is_item_title(name):
                    continue
                items.append(EquipmentItem(
                    id=slugify(name),
                    name=name,
                    category=category,
                    subcategory=subcategory,
                    properties={},
                    description=content_md,
                ))

    _walk(tree)
    return items
