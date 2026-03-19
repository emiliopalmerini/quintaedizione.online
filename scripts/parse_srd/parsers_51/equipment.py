"""Equipment parser for SRD 5.1.

Tables are sequential TABLE_HEADER_SMALL + TABLE_BODY paragraphs
(one cell per paragraph) rather than pipe-table markdown.
"""

from __future__ import annotations

import re

from ..classify import SpanRole
from ..heading_tree import HeadingNode
from ..markdown_gen import paragraphs_to_markdown
from ..merge import Paragraph
from ..schemas import EquipmentItem
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

_H2_CATEGORY: dict[str, str] = {
    "Armi": "weapons",
    "Armature": "armor",
    "Strumenti": "tools",
    "Equipaggiamento d'avventura": "gear",
    "Cavalcature e veicoli": "mounts",
    "Spese dello stile di vita": "services",
}


def _collect_all_paragraphs(node: HeadingNode) -> list[Paragraph]:
    result = list(node.content)
    for child in node.children:
        result.extend(_collect_all_paragraphs(child))
    return result


def _parse_sequential_table(node: HeadingNode, category: str) -> list[EquipmentItem]:
    """Parse equipment from sequential TABLE_HEADER_SMALL/TABLE_BODY paragraphs."""
    all_paras = _collect_all_paragraphs(node)
    items: list[EquipmentItem] = []
    seen_ids: set[str] = set()

    headers: list[str] = []
    data_start = -1
    for i, p in enumerate(all_paras):
        if p.role == SpanRole.TABLE_HEADER_SMALL:
            headers.append(p.text.strip())
        elif headers:
            data_start = i
            break

    if not headers or data_start < 0:
        return items

    num_cols = len(headers)
    header_lower = [h.lower() for h in headers]

    name_col = -1
    col_map: dict[str, int] = {}
    for ci, h in enumerate(header_lower):
        if h in ("armatura", "nome", "oggetto", "nave"):
            name_col = ci
        elif "costo" in h:
            col_map["cost"] = ci
        elif "classe armatura" in h:
            col_map["classe_armatura"] = ci
        elif h == "danni":
            col_map["danni"] = ci
        elif h == "peso":
            col_map["peso"] = ci
        elif "forza" in h:
            col_map["forza"] = ci
        elif "furtività" in h:
            col_map["furtività"] = ci
        elif "proprietà" in h:
            col_map["proprietà"] = ci
        elif "velocità" in h:
            col_map["velocità"] = ci
        elif "capacità" in h:
            col_map["capacità"] = ci

    if name_col < 0:
        name_col = 0

    subcategory = ""
    row: list[str] = []

    for p in all_paras[data_start:]:
        if p.role == SpanRole.UNKNOWN:
            subcategory = p.text.strip()
            continue
        if p.role == SpanRole.TABLE_HEADER_SMALL:
            continue
        if p.role != SpanRole.TABLE_BODY:
            continue

        row.append(p.text.strip())
        if len(row) < num_cols:
            continue

        name = row[name_col] if name_col < len(row) else ""
        if not name or name == "—":
            row = []
            continue

        item_id = slugify(name)
        if item_id in seen_ids:
            row = []
            continue
        seen_ids.add(item_id)

        props: dict[str, str] = {}
        for key, ci in col_map.items():
            if ci < len(row):
                val = row[ci].strip()
                if val and val != "—":
                    props[key] = val

        desc_parts: list[str] = []
        label_map = {
            "cost": "Costo", "classe_armatura": "Classe Armatura (CA)",
            "danni": "Danni", "peso": "Peso", "forza": "Forza",
            "furtività": "Furtività", "proprietà": "Proprietà",
            "velocità": "Velocità", "capacità": "Capacità",
        }
        for key, label in label_map.items():
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
        row = []

    return items


def _has_table_paragraphs(node: HeadingNode) -> bool:
    for p in _collect_all_paragraphs(node):
        if p.role == SpanRole.TABLE_HEADER_SMALL:
            return True
    return False


@register("equipment")
def parse_equipment(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[EquipmentItem]:
    """Parse equipment from 5.1 SRD using sequential table paragraphs."""
    items: list[EquipmentItem] = []

    def _walk(nodes: list[HeadingNode]) -> None:
        for node in nodes:
            if node.level == 1:
                _walk(node.children)
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

                if _has_table_paragraphs(node):
                    items.extend(_parse_sequential_table(node, new_category))
                continue

    _walk(tree)
    return items
