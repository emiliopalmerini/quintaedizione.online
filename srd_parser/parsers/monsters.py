from __future__ import annotations

import re
from typing import Dict, List

from .items_common import collect_labeled_fields, shared_id_for, slugify, split_items


TAG_RE = re.compile(r"^(?P<size>[^,]+)\s+(?P<type>[^,]+)(?:,\s*(?P<alignment>.+))?$")


def _first_meta(block: List[str]) -> str:
    for ln in block:
        s = ln.strip()
        if s.startswith("*") and s.endswith("*"):
            return s.strip("*").strip()
    return ""


def _parse_meta(s: str) -> Dict:
    out: Dict = {}
    m = TAG_RE.match(s)
    if not m:
        return out
    out["tag"] = {
        "taglia": m.group("size").strip(),
        "tipo": m.group("type").strip(),
        "allineamento": (m.group("alignment") or "").strip(),
    }
    return out


def _map_fields(fields: Dict[str, str]) -> Dict:
    # Dataset uses English keys even in IT monsters file for core stats
    key_map = {
        "Armor Class": "ac",
        "Hit Points": "hp",
        "Speed": "velocita",
        "Initiative": "iniziativa",
    }
    out: Dict = {}
    for k, v in fields.items():
        if k in key_map:
            out[key_map[k]] = v
    # Challenge Rating (CR) or GS (IT)
    cr_val = fields.get("CR") or fields.get("Gs") or fields.get("GS") or fields.get("Grado di Sfida")
    if cr_val:
        raw = str(cr_val).strip()
        # Extract first numeric or fraction like 1/2, 3, 4.5
        m = re.search(r"(\d+/(\d+))|(\d+(?:\.\d+)?)", raw)
        if m:
            frac = m.group(1)
            num = m.group(3)
            try:
                if frac:
                    a, b = frac.split("/")
                    val = float(int(a) / int(b))
                else:
                    val = float(num)
                out["cr"] = val
                out["cr_raw"] = raw
            except Exception:
                out["cr_raw"] = raw
    return out


def parse_monsters(md_lines: List[str], *, namespace: str = "monster") -> List[Dict]:
    items = split_items(md_lines, level="h2")
    docs: List[Dict] = []
    for idx, (title, block) in enumerate(items, start=1):
        if not block:
            continue
        fields = collect_labeled_fields(block)
        base = _map_fields(fields)
        meta = _parse_meta(_first_meta(block))
        content = (f"## {title}\n" + "\n".join(block)).strip() + "\n"
        doc: Dict = {
            "shared_id": shared_id_for(namespace, idx),
            "slug": slugify(title),
            "nome": title.strip(),
            **meta,
            **base,
            "content": content,
        }
        docs.append(doc)
    return docs
