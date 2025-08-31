from __future__ import annotations

import re
from typing import Dict, List

from .items_common import first_italic_line, shared_id_for, slugify, split_items


TYPE_LINE_RE = re.compile(
    r"^(?P<type>[^,]+),\s*(?P<rarity>[^()]+?)(?:\s*\((?P<paren>[^)]+)\))?\s*$",
    re.IGNORECASE,
)


def _parse_type_line(s: str) -> Dict:
    out: Dict = {}
    m = TYPE_LINE_RE.match(s.strip())
    if not m:
        return out
    out["tipo"] = m.group("type").strip()
    out["rarita"] = m.group("rarity").strip()
    par = (m.group("paren") or "").lower()
    if "requires attunement" in par or "richiede sintonizzazione" in par:
        out["sintonizzazione"] = True
    else:
        out["sintonizzazione"] = False
    return out


def parse_magic_items(md_lines: List[str]) -> List[Dict]:
    # Magic items entries are H3 under a H2 A–Z section
    items = split_items(md_lines, level="h3")
    docs: List[Dict] = []
    for idx, (title, block) in enumerate(items, start=1):
        if not block:
            continue
        name = title.strip()
        type_line = first_italic_line(block) or ""
        content = (f"### {title}\n" + "\n".join(block)).strip() + "\n"
        meta = _parse_type_line(type_line) if type_line else {}
        doc: Dict = {
            "shared_id": shared_id_for("magicitem", idx),
            "slug": slugify(name),
            "nome": name,
            "content": content,
        }
        doc.update(meta)
        docs.append(doc)
    return docs

