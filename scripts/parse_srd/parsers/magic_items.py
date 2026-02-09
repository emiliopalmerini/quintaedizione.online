"""Magic Items (Oggetti Magici) parser.

Structure in PDF:
  H1: Oggetti magici
    H2: Oggetti magici A–Z (marker — items start here)
      H5: Item name
        BODY_ITALIC: "{Tipo}, {rarità} (richiede sintonia [da un ...])"
        BODY paragraphs: description
"""

from __future__ import annotations

import re
import warnings

from ..classify import SpanRole
from ..heading_tree import HeadingNode
from ..markdown_gen import paragraph_to_markdown, paragraphs_to_markdown
from ..merge import Paragraph
from ..schemas import MagicItem
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

_RARITY_WORDS = {
    "comune", "non comune", "raro", "rara", "molto raro", "molto rara",
    "leggendario", "leggendaria", "manufatto",
}

# Pattern: "{Type}, {rarity}" or "{Type}, {rarity} (richiede sintonia ...)"
_SUBTITLE_RE = re.compile(
    r"^(.+?),\s*(comune|non comune|rar[oa]|molto rar[oa]|leggendari[oa]|manufatto)"
    r"(?:\s*\(richiede sintonia(?:\s+(.+?))?\))?\s*",
    re.IGNORECASE,
)


# Manual overrides for items where automatic subtitle parsing fails.
# Key: lowercase item name. Value: (type, rarity, attunement, attunement_details).
_OVERRIDES: dict[str, tuple[str, str, bool, str]] = {
    "cintura della forza dei giganti": (
        "Oggetto meraviglioso", "varia", True, "richiede sintonia",
    ),
}


def _parse_subtitle(text: str) -> tuple[str, str, bool, str]:
    """Parse magic item subtitle.

    Returns (type, rarity, requires_attunement, attunement_details).
    """
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
    """Try to extract type and rarity from description text.

    Scans the first few lines for a known rarity word preceded by a type.
    Returns (type, rarity) or ("", "").
    """
    first_lines = description[:500].lower()
    for rarity_word in ("molto raro", "molto rara", "non comune",
                        "comune", "raro", "rara",
                        "leggendario", "leggendaria", "manufatto"):
        idx = first_lines.find(rarity_word)
        if idx > 0:
            before = first_lines[:idx].strip().rstrip(",").strip()
            # Take the last line fragment as the type
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
    """Parse magic items from the Oggetti Magici section."""
    items: list[MagicItem] = []

    # Find "Oggetti magici A–Z" or similar heading
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

        # Parse subtitle from first BODY_ITALIC paragraph
        item_type = ""
        rarity = ""
        attunement = False
        attunement_details = ""
        body_paras: list[Paragraph] = []

        for para in content:
            if not item_type and para.role == SpanRole.BODY_ITALIC:
                # The subtitle and description may be merged in one paragraph
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

                # Remaining spans are description
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

        # Apply overrides for known edge cases
        override = _OVERRIDES.get(name.lower())
        if override:
            item_type, rarity, attunement, attunement_details = override

        # Fallback: if type/rarity still empty, try to extract from description
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
