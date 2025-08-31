from __future__ import annotations

import hashlib
import re
from typing import Dict, List, Optional, Tuple

from ..utils import BOLD_FIELD_RE, DASH_FIELD_RE, SECTION_H2_RE, SECTION_H3_RE, SECTION_H4_RE


def slugify(name: str) -> str:
    x = name.strip().lower()
    x = (
        x.replace(" ", "-")
        .replace("_", "-")
        .replace("'", "")
        .replace("’", "")
        .replace("/", "-")
    )
    x = re.sub(r"-+", "-", x)
    return x


def split_items(md: List[str], level: str = "h2") -> List[Tuple[str, List[str]]]:
    re_map = {"h2": SECTION_H2_RE, "h3": SECTION_H3_RE, "h4": SECTION_H4_RE}
    header_re = re_map[level]
    out: List[Tuple[str, List[str]]] = []
    i = 0
    n = len(md)
    current_title: Optional[str] = None
    current_block: List[str] = []
    while i < n:
        line = md[i].rstrip("\n")
        m = header_re.match(line)
        if m:
            title = m.group("title").strip()
            if current_title is not None:
                out.append((current_title, current_block))
            current_title = title
            current_block = []
        else:
            if current_title is not None:
                current_block.append(line)
        i += 1
    if current_title is not None:
        out.append((current_title, current_block))
    return out


def collect_labeled_fields(block: List[str]) -> Dict[str, str]:
    fields: Dict[str, str] = {}
    for ln in block:
        ln_s = ln.strip()
        m = BOLD_FIELD_RE.match(ln_s) or DASH_FIELD_RE.match(ln_s)
        if not m:
            continue
        key = m.group("field").strip()
        val = m.group("value").strip()
        if key and val:
            fields[key] = val
    return fields


def shared_id_for(namespace: str, index1: int) -> str:
    return f"{namespace}:{index1:04d}"


def first_italic_line(block: List[str]) -> Optional[str]:
    for ln in block:
        s = ln.strip()
        if s.startswith("*") and s.endswith("*") and len(s) > 2:
            return s.strip("*").strip()
    return None

