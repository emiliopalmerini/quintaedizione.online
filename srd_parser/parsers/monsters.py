from __future__ import annotations

import re
from typing import Dict, List, Optional, Tuple

from ..utils import (
    ITALIC_LINE_RE, BOLD_FIELD_RE, DASH_FIELD_RE,
    SECTION_H2_RE, SECTION_H3_RE, split_sections, source_label, 
)

ITALIC_META_CRE = re.compile(r"^\*(?P<size>[A-Za-z]+)\s+(?P<type>[^,]+),\s*(?P<alignment>[^*]+)\*\s*$")

def _is_h3(line: str) -> Optional[str]:
    m = SECTION_H3_RE.match(line.rstrip("\n"))
    if not m:
        return None
    return m.group("title").strip().lower()

def _find_h3(block: List[str], start: int, title_text: str) -> Optional[int]:
    tgt = title_text.strip().lower()
    for k in range(start, len(block)):
        t = _is_h3(block[k])
        if t == tgt:
            return k + 1
    return None

def _pull_named_sections(block: List[str], start_idx: int) -> Tuple[List[Dict[str, str]], int]:
    out_list: List[Dict[str, str]] = []
    j = start_idx
    name_re = re.compile(r"^\*{3}(?P<name>.+?)\.\*{3}\s*(?P<text>.*)$")
    current: Optional[Dict[str, str]] = None
    acc: List[str] = []

    while j < len(block):
        if _is_h3(block[j]):
            break
        line = block[j].rstrip("\n")
        mname = name_re.match(line.strip())
        if mname:
            if current is not None:
                current["text"] = "\n".join(x for x in acc if x.strip()).strip()
                out_list.append(current)
            inline = mname.group("text").strip()
            current = {"name": mname.group("name").strip(), "text": ""}
            acc = [inline] if inline else []
        else:
            if current is not None:
                acc.append(line)
        j += 1

    if current is not None:
        current["text"] = "\n".join(x for x in acc if x.strip()).strip()
        out_list.append(current)

    return out_list, j

def _parse_ability_table(block: List[str], start_idx: int) -> Tuple[Dict[str, Dict[str, str]], int]:
    abilities: Dict[str, Dict[str, str]] = {}
    j = start_idx
    while j < len(block) and not block[j].lstrip().startswith("|STAT"):
        j += 1
    if j >= len(block):
        return abilities, start_idx
    j += 2
    while j < len(block) and block[j].strip().startswith("|"):
        row = [c.strip() for c in block[j].strip().strip("|").split("|")]
        if len(row) >= 4 and len(row[0]) == 3:
            abbr = row[0]
            abilities[abbr] = {"score": row[1], "mod": row[2], "save": row[3]}
        else:
            break
        j += 1
    return abilities, j

def parse_monster_block(title: str, block: List[str]) -> Dict:
    doc: Dict[str, object] = {"name": title.strip(), "source": source_label()}
    i = 0
    while i < len(block) and not block[i].strip():
        i += 1
    if i < len(block) and ITALIC_LINE_RE.match(block[i].strip()):
        m = ITALIC_META_CRE.match(block[i].strip())
        if m:
            doc["size"] = m.group("size").title()
            doc["type"] = m.group("type").strip()
            doc["alignment"] = m.group("alignment").strip()
        i += 1
    # simple dash fields
    stats: Dict[str, str] = {}
    k = i
    while k < len(block):
        s = block[k].rstrip("\n")
        m = DASH_FIELD_RE.match(s)
        if not m:
            break
        stats[m.group("field").strip()] = m.group("value").strip()
        k += 1
    for fk, fv in stats.items():
        key = fk.strip().lower().replace(" ", "_")
        doc[key] = fv.strip()
    i = k
    abilities, i_after = _parse_ability_table(block, i)
    if abilities:
        doc["abilities"] = abilities
        i = i_after
    for section_name, key in (
        ("Traits", "traits"),
        ("Actions", "actions"),
        ("Bonus Actions", "bonus_actions"),
        ("Reactions", "reactions"),
        ("Legendary Actions", "legendary_actions"),
        ("Lair Actions", "lair_actions"),
    ):
        start_at = _find_h3(block, i, section_name)
        if start_at is not None:
            entries, _ = _pull_named_sections(block, start_at)
            if entries:
                doc[key] = entries
    doc["block_md"] = "\n".join(block).strip()
    return doc

def parse_monsters(md: List[str]) -> List[Dict]:
    out: List[Dict] = []
    for title, block in split_sections(md, SECTION_H2_RE):
        if not title or title.startswith("# "):
            continue
        out.append(parse_monster_block(title, block))
    return out
