from __future__ import annotations

import re
from dataclasses import asdict, dataclass, field
from typing import Dict, List, Optional, Tuple, Union

from ..utils import (
    BOLD_FIELD_RE, ITALIC_LINE_RE, split_sections,
    source_label, 
)

SPELL_HDR_RE = re.compile(r"^####\s+\**(?P<name>.+?)\**\s*$")
UPCAST_LINE_RE = re.compile(
    r"^\s*\*\*\s*_?\s*(?P<label>Using a Higher-Level Spell Slot|Cantrip Upgrade)\.?_?\s*\*\*\s*(?P<rest>.*)$",
    re.IGNORECASE,
)

def _parse_level_school_classes(meta: str) -> Tuple[int, str, List[str]]:
    meta = meta.strip()
    m = re.match(r"^Level\s+(?P<lvl>\d+)\s+(?P<school>[A-Za-z]+)\s*\((?P<classes>[^)]+)\)\s*$", meta)
    if m:
        return int(m.group("lvl")), m.group("school").title(), [c.strip() for c in m.group("classes").split(",")]
    m = re.match(r"^(?P<school>[A-Za-z]+)\s+Cantrip\s*\((?P<classes>[^)]+)\)\s*$", meta)
    if m:
        return 0, m.group("school").title(), [c.strip() for c in m.group("classes").split(",")]
    classes: List[str] = []
    m_paren = re.search(r"\((?P<classes>[^)]+)\)", meta)
    meta_wo = meta
    if m_paren:
        classes = [c.strip() for c in m_paren.group("classes").split(",")]
        meta_wo = meta[: m_paren.start()].strip()
    if "cantrip" in meta_wo.lower():
        parts = re.sub(r"\b[Cc]antrip\b", "", meta_wo).strip().split()
        return 0, (parts[0].title() if parts else "Unknown"), classes
    m2 = re.match(r"^Level\s+(?P<lvl>\d+)\s+(?P<school>[A-Za-z]+)", meta_wo)
    if m2:
        return int(m2.group("lvl")), m2.group("school").title(), classes
    return -1, "Unknown", classes

def _parse_components(value: str) -> Tuple[Optional[str], Optional[str], Optional[str]]:
    raw = value.strip()
    if not raw:
        return None, None, None
    mat = None
    m = re.search(r"\bM\s*\((?P<mat>.+?)\)\s*$", raw)
    if m:
        mat = m.group("mat").strip()
    letters = re.findall(r"\b(V|S|M)\b", raw)
    vsm = ",".join(dict.fromkeys([x for x in letters])) if letters else None
    return raw, vsm, mat

def _parse_duration(value: str) -> Tuple[Optional[str], bool]:
    v = value.strip()
    if not v:
        return None, False
    if v.lower().startswith("concentration"):
        v2 = re.sub(r"^Concentration\s*,\s*", "", v, flags=re.IGNORECASE)
        return v2, True
    return v, False

def _parse_casting_time(value: str) -> Tuple[Optional[str], bool]:
    v = value.strip()
    if not v:
        return None, False
    ritual = False
    if re.search(r"\bor\s+Ritual\b", v, flags=re.IGNORECASE):
        ritual = True
        v = re.sub(r"\s*or\s*Ritual\b", "", v, flags=re.IGNORECASE).strip()
    return v, ritual

@dataclass
class SpellDoc:
    name: str
    level: int
    school: str
    classes: List[str] = field(default_factory=list)
    casting_time: Optional[str] = None
    ritual: bool = False
    range: Optional[str] = None
    components: Dict[str, Optional[str]] = field(default_factory=dict)
    duration: Dict[str, Union[str, bool, None]] = field(default_factory=dict)
    description: Optional[str] = None
    higher_level: Optional[str] = None
    source: str = field(default_factory=source_label)

def parse_spells(md: List[str]) -> List[Dict]:
    spells: List[Dict] = []
    n = len(md)
    i = 0
    while i < n:
        line = md[i].rstrip("\n")
        m_hdr = SPELL_HDR_RE.match(line)
        if not m_hdr:
            i += 1
            continue
        name = re.sub(r"^\*\*|\*\*$", "", m_hdr.group("name").strip())
        # meta
        i += 1
        while i < n and not md[i].strip():
            i += 1
        if i >= n or not ITALIC_LINE_RE.match(md[i].strip()):
            i += 1
            continue
        level, school, classes = _parse_level_school_classes(ITALIC_LINE_RE.match(md[i].strip()).group("meta"))
        i += 1
        # fields
        fields = {"Casting Time": None, "Range": None, "Components": None, "Duration": None}
        start_desc = None
        while i < n:
            s = md[i].rstrip("\n")
            if not s.strip():
                i += 1
                continue
            if SPELL_HDR_RE.match(s):
                break
            m_field = BOLD_FIELD_RE.match(s.strip())
            if m_field and m_field.group("field").strip() in fields:
                fields[m_field.group("field").strip()] = m_field.group("value").strip()
                i += 1
                continue
            start_desc = i
            break
        # description block
        block: List[str] = []
        if start_desc is not None:
            j = start_desc
            while j < n and not SPELL_HDR_RE.match(md[j].rstrip("\n")):
                block.append(md[j].rstrip("\n"))
                j += 1
            i = j
        # split upcast
        desc_lines: List[str] = []
        upcast_lines: List[str] = []
        in_up = False
        for ln in block:
            m_up = UPCAST_LINE_RE.match(ln.strip())
            if m_up:
                in_up = True
                rest = m_up.group("rest").strip()
                if rest:
                    upcast_lines.append(rest)
                continue
            if in_up:
                upcast_lines.append(ln)
            else:
                desc_lines.append(ln)
        description = ("\n".join(desc_lines).strip() or None) if desc_lines else None
        higher = ("\n".join(upcast_lines).strip() or None) if upcast_lines else None
        # pack
        ct, ritual = _parse_casting_time(fields["Casting Time"] or "")
        rng = (fields["Range"] or "").strip() or None
        comp_raw, comp_vsm, material = _parse_components(fields["Components"] or "")
        dur_text, concentration = _parse_duration(fields["Duration"] or "")
        doc = SpellDoc(
            name=name,
            level=level,
            school=school,
            classes=classes,
            casting_time=ct,
            ritual=bool(ritual),
            range=rng,
            components={"raw": comp_raw, "vsm": comp_vsm, "material": material},
            duration={"text": dur_text, "concentration": bool(concentration)},
            description=description,
            higher_level=higher,
        )
        spells.append(asdict(doc))
    return spells
