from __future__ import annotations

import os
import re
from typing import List, Tuple, Optional

# Common regex
BOLD_FIELD_RE = re.compile(r"^\*\*(?P<field>[^:*]+):\*\*\s*(?P<value>.+?)\s*$")
DASH_FIELD_RE = re.compile(r"^-+\s*\*\*(?P<field>[^:*]+):\*\*\s*(?P<value>.+?)\s*$")
ITALIC_LINE_RE = re.compile(r"^\*(?P<meta>.+?)\*\s*$")
SECTION_H2_RE = re.compile(r"^##\s+(?P<title>.+?)\s*$")
SECTION_H3_RE = re.compile(r"^###\s+(?P<title>.+?)\s*$")
SECTION_H4_RE = re.compile(r"^####\s+(?P<title>.+?)\s*$")
EMDASH = "—"

def norm_key(s: str) -> str:
    return (
        s.strip()
        .lower()
        .replace(" ", "_")
        .replace("-", "_")
        .replace("’", "")
        .replace("'", "")
        .replace("/", "_")
    )

def clean_value(v: Optional[str]) -> Optional[str]:
    if v is None:
        return None
    x = v.strip()
    if not x or x == EMDASH or x.lower() == "null":
        return None
    return x

def split_sections(md: List[str], header_re: re.Pattern) -> List[Tuple[str, List[str]]]:
    out: List[Tuple[str, List[str]]] = []
    n = len(md)
    i = 0
    current_title: Optional[str] = None
    current_block: List[str] = []
    while i < n:
        line = md[i].rstrip("\n")
        m = header_re.match(line)
        if m:
            if current_title is not None:
                out.append((current_title, current_block))
            current_title = m.group("title").strip()
            current_block = []
        else:
            if current_title is not None:
                current_block.append(line)
        i += 1
    if current_title is not None:
        out.append((current_title, current_block))
    return out

def source_label() -> str:
    return os.environ.get("SOURCE_LABEL", "srd 5.2.1")
