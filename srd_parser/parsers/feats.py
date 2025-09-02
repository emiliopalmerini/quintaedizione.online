from __future__ import annotations

from typing import Dict, List

from .items_common import first_italic_line, slugify, split_items


def parse_feats(md_lines: List[str], *, lang: str = "it") -> List[Dict]:
    # Feats are H4 sections within H3 category sections; parse all H4
    items = split_items(md_lines, level="h4")
    docs: List[Dict] = []
    for idx, (title, block) in enumerate(items, start=1):
        if not block:
            continue
        cat_line = first_italic_line(block) or ""
        content = (f"#### {title}\n" + "\n".join(block)).strip() + "\n"
        if lang == "en":
            doc: Dict = {
                "slug": slugify(title),
                "name": title.strip(),
                "category": cat_line.strip(),
                "content": content,
            }
        else:
            doc = {
                "slug": slugify(title),
                "nome": title.strip(),
                "categoria": cat_line.strip(),
                "content": content,
            }
        docs.append(doc)
    return docs

