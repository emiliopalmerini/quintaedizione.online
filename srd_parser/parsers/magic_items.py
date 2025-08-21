from __future__ import annotations

from typing import Dict, List
from ..utils import ITALIC_LINE_RE, SECTION_H3_RE, split_sections, source_label

RARITIES = {"common", "uncommon", "rare", "very rare", "legendary", "artifact"}

def parse_magic_items(md: List[str]) -> List[Dict]:
    items: List[Dict] = []
    sections = split_sections(md, SECTION_H3_RE)
    for title, block in sections:
        low = title.strip().lower()
        if low.endswith("a–z") or low == "magic items a–z":
            continue
        i = 0
        while i < len(block) and not block[i].strip():
            i += 1
        item: Dict[str, object] = {"name": title.strip(), "source": source_label()}
        if i < len(block) and ITALIC_LINE_RE.match(block[i].strip()):
            meta = ITALIC_LINE_RE.match(block[i].strip()).group("meta")
            parts = [p.strip() for p in meta.split(",")]
            if parts:
                item["type_line"] = parts[0]
            rarity = None
            requires = False
            for p in parts[1:]:
                lowp = p.lower()
                for r in RARITIES:
                    if r in lowp:
                        rarity = r.title() if r != "very rare" else "Very Rare"
                if "requires attunement" in lowp:
                    requires = True
            if rarity:
                item["rarity"] = rarity
            if requires:
                item["requires_attunement"] = True
            i += 1
        desc = "\n".join(block[i:]).strip()
        if desc:
            item["description_md"] = desc
        items.append(item)
    return items
