from __future__ import annotations

from typing import Dict, List

from .items_common import collect_labeled_fields, shared_id_for, slugify, split_items


IT_KEYS = {
    "Peso": "peso",
}

EN_KEYS = {
    "Weight": "weight",
}


def _map_fields(fields: Dict[str, str], lang: str) -> Dict:
    m = IT_KEYS if lang == "it" else EN_KEYS
    out: Dict = {}
    for k, v in fields.items():
        if k in m:
            out[m[k]] = v
    return out


def parse_gear_en(md_lines: List[str]) -> List[Dict]:
    items = split_items(md_lines, level="h2")
    docs: List[Dict] = []
    for idx, (title, block) in enumerate(items, start=1):
        if not block:
            continue
        fields = collect_labeled_fields(block)
        mapped = _map_fields(fields, "en")
        content = (f"## {title}\n" + "\n".join(block)).strip() + "\n"
        docs.append(
            {
                "shared_id": shared_id_for("gear", idx),
                "slug": slugify(title),
                "name": title.strip(),
                **mapped,
                "content": content,
            }
        )
    return docs


def parse_gear_it(md_lines: List[str]) -> List[Dict]:
    items = split_items(md_lines, level="h2")
    docs: List[Dict] = []
    for idx, (title, block) in enumerate(items, start=1):
        if not block:
            continue
        fields = collect_labeled_fields(block)
        mapped = _map_fields(fields, "it")
        content = (f"## {title}\n" + "\n".join(block)).strip() + "\n"
        docs.append(
            {
                "shared_id": shared_id_for("gear", idx),
                "slug": slugify(title),
                "nome": title.strip(),
                **mapped,
                "content": content,
            }
        )
    return docs

