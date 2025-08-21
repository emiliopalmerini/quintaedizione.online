from __future__ import annotations

import re
from typing import Dict, List

from ..utils import BOLD_FIELD_RE, split_sections, SECTION_H2_RE, norm_key, clean_value, source_label

def parse_equipment(md: List[str], extra_meta_from_title: bool = True) -> List[Dict]:
    out: List[Dict] = []
    sections = split_sections(md, SECTION_H2_RE)
    for title, block in sections:
        name = title
        embedded_cost = None
        m_cost = re.search(r"\((?P<cost>\d+[^)]*)\)$", title)
        if m_cost and extra_meta_from_title:
            embedded_cost = m_cost.group("cost").strip()
            name = title[: m_cost.start()].strip()
        fields: Dict[str, str] = {}
        desc_lines: List[str] = []
        for ln in block:
            m = BOLD_FIELD_RE.match(ln.strip())
            if m:
                fields[m.group("field").strip()] = m.group("value").strip()
            else:
                desc_lines.append(ln)
        doc = {norm_key(k): clean_value(v) for k, v in fields.items()}
        doc["name"] = name
        if embedded_cost and not doc.get("cost"):
            doc["cost"] = embedded_cost
        desc = "\n".join(desc_lines).strip()
        if desc:
            doc["description_md"] = desc
        doc["source"] = source_label()
        out.append(doc)
    return out
