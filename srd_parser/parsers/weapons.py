from __future__ import annotations

from typing import Dict, List

from .items_common import (
    collect_labeled_fields,
    slugify,
    split_items,
)


IT_KEYS = {
    "Costo": "costo",
    "Peso": "peso",
    "Danno": "danno",
    "Categoria": "categoria",
    "Proprietà": "proprieta",
    "Maestria": "maestria",
    "Gittata": "gittata_normale",
    "Gittata lunga": "gittata_lunga",
}

EN_KEYS = {
    "Cost": "cost",
    "Weight": "weight",
    "Damage": "damage",
    "Category": "category",
    "Properties": "properties",
    "Mastery": "mastery",
    "Range": "range",
    "Long Range": "long_range",
}


def _map_fields(fields: Dict[str, str], lang: str) -> Dict:
    m = IT_KEYS if lang == "it" else EN_KEYS
    out: Dict = {}
    for k, v in fields.items():
        if k in m:
            out[m[k]] = v
    # split properties/proprietà
    prop_key = "proprieta" if lang == "it" else "properties"
    if prop_key in out and out[prop_key]:
        out[prop_key] = [p.strip() for p in out[prop_key].split(",") if p.strip()]
    # group range
    if lang == "it":
        r1, r2 = out.pop("gittata_normale", None), out.pop("gittata_lunga", None)
        if r1 or r2:
            out["gittata"] = {"normale": r1, "lunga": r2}
    else:
        r1, r2 = out.pop("range", None), out.pop("long_range", None)
        if r1 or r2:
            out["range"] = {"normal": r1, "long": r2}
    return out


def parse_weapons_en(md_lines: List[str]) -> List[Dict]:
    items = split_items(md_lines, level="h2")
    docs: List[Dict] = []
    for idx, (title, block) in enumerate(items, start=1):
        if not block:
            continue
        name = title.strip()
        fields = collect_labeled_fields(block)
        mapped = _map_fields(fields, "en")
        content = (f"## {title}\n" + "\n".join(block)).strip() + "\n"
        docs.append(
            {
                "slug": slugify(name),
                "name": name,
                **mapped,
                "content": content,
            }
        )
    return docs


def parse_weapons_it(md_lines: List[str]) -> List[Dict]:
    items = split_items(md_lines, level="h2")
    docs: List[Dict] = []
    for idx, (title, block) in enumerate(items, start=1):
        if not block:
            continue
        name = title.strip()
        fields = collect_labeled_fields(block)
        mapped = _map_fields(fields, "it")
        content = (f"## {title}\n" + "\n".join(block)).strip() + "\n"
        docs.append(
            {
                "slug": slugify(name),
                "nome": name,
                **mapped,
                "content": content,
            }
        )
    return docs

