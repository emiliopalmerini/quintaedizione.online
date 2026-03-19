"""Magic items parser for SRD 5.1.

Mostly identical to 5.2.1 — the subtitle regex and overrides are extended
for variable rarity and Avatar della morte.
"""

from __future__ import annotations

import re
import warnings

from ..classify import SpanRole
from ..heading_tree import HeadingNode
from ..markdown_gen import paragraphs_to_markdown
from ..merge import Paragraph
from ..schemas import MagicItem
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

_SUBTITLE_RE = re.compile(
    r"^(.+?),\s*(comune|non comune|rar[oa]|molto rar[oa]|leggendari[oa]|manufatto|rarità variabile|varia)"
    r"(?:\s*\(richiede sintonia(?:\s+(.+?))?\))?\s*",
    re.IGNORECASE,
)

_OVERRIDES: dict[str, tuple[str, str, bool, str]] = {
    "cintura della forza dei giganti": (
        "Oggetto meraviglioso", "varia", True, "richiede sintonia",
    ),
    "avatar della morte": (
        "Creatura", "leggendario", False, "",
    ),
}


def _parse_subtitle(text: str) -> tuple[str, str, bool, str]:
    text = " ".join(text.split())
    m = _SUBTITLE_RE.match(text)
    if m:
        item_type = m.group(1).strip()
        rarity = m.group(2).strip()
        attunement = "(richiede sintonia" in text
        attunement_details = m.group(3).strip() if m.group(3) else ""
        return item_type, rarity, attunement, attunement_details
    return "", "", False, ""


def _fallback_rarity(description: str) -> tuple[str, str]:
    first_lines = description[:500].lower()
    for rarity_word in ("molto raro", "molto rara", "non comune",
                        "comune", "raro", "rara",
                        "leggendario", "leggendaria", "manufatto"):
        idx = first_lines.find(rarity_word)
        if idx > 0:
            before = first_lines[:idx].strip().rstrip(",").strip()
            last_line = before.rsplit("\n", 1)[-1].strip()
            if last_line:
                return last_line.title(), rarity_word
    return "", ""


@register("magic_items")
def parse_magic_items(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[MagicItem]:
    """Parse magic items from 5.1 SRD."""
    items: list[MagicItem] = []

    item_nodes: list[HeadingNode] = []

    def _find_items(nodes: list[HeadingNode]) -> None:
        for node in nodes:
            title_lower = node.title.lower().strip()
            if "oggetti magici" in title_lower and ("a-z" in title_lower or "a–z" in title_lower):
                item_nodes.extend(node.children)
                return
            _find_items(node.children)

    _find_items(tree)

    for node in item_nodes:
        if node.level != 5:
            continue

        name = node.title.strip()
        content = node.content

        item_type = ""
        rarity = ""
        attunement = False
        attunement_details = ""
        body_paras: list[Paragraph] = []

        for para in content:
            if not item_type and para.role == SpanRole.BODY_ITALIC:
                subtitle_text = ""
                body_start_idx = 0
                for i, span in enumerate(para.spans):
                    if span.role == SpanRole.BODY_ITALIC:
                        subtitle_text += span.text
                    elif span.text.strip() == "":
                        continue
                    else:
                        body_start_idx = i
                        break
                else:
                    body_start_idx = len(para.spans)

                subtitle_text = " ".join(subtitle_text.split())
                item_type, rarity, attunement, attunement_details = _parse_subtitle(subtitle_text)

                remaining = para.spans[body_start_idx:]
                while remaining and remaining[0].text.strip() == "":
                    remaining = remaining[1:]
                if remaining:
                    body_paras.append(Paragraph(
                        spans=remaining, role=SpanRole.BODY, page_num=para.page_num,
                    ))
                continue

            body_paras.append(para)

        description = paragraphs_to_markdown(body_paras)

        override = _OVERRIDES.get(name.lower())
        if override:
            item_type, rarity, attunement, attunement_details = override

        if not item_type or not rarity:
            fb_type, fb_rarity = _fallback_rarity(description)
            if not item_type and fb_type:
                item_type = fb_type
            if not rarity and fb_rarity:
                rarity = fb_rarity

        if not item_type:
            warnings.warn(f"Magic item '{name}': empty type")
        if not rarity:
            warnings.warn(f"Magic item '{name}': empty rarity")

        items.append(MagicItem(
            id=slugify(name),
            name=name,
            type=item_type,
            rarity=rarity,
            attunement=attunement,
            attunement_details=attunement_details,
            description=description,
        ))

    return items
