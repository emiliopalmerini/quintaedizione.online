from __future__ import annotations

from typing import Dict, List
from ..utils import SECTION_H4_RE, split_sections, source_label

def parse_rules_glossary(md: List[str]) -> List[Dict]:
    entries: List[Dict] = []
    for title, block in split_sections(md, SECTION_H4_RE):
        body = "\n".join(block).strip()
        entries.append({"term": title.strip(), "body_md": body, "source": source_label()})
    return entries
