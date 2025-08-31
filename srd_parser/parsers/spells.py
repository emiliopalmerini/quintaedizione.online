from __future__ import annotations

import re
from typing import Dict, List

from .items_common import (
    collect_labeled_fields,
    first_italic_line,
    shared_id_for,
    slugify,
    split_items,
)
from ..utils import SECTION_H3_RE


META_RE = re.compile(
    r"^\s*"  # start
    r"(?:(?:Level\s+(?P<lvl_en>\d{1,2})\s+(?P<school_en>[^()]+))"  # EN: Level X School
    r"|(?P<cantrip_school_en>[^()]+)\s+Cantrip"  # EN: School Cantrip
    r"|(?:Livello\s+(?P<lvl_it>\d{1,2})\s+(?P<school_it>[^()]+))"  # IT: Livello X Scuola
    r"|(?:(?:Trucchetto\s+di\s+(?P<cantrip_school_it1>[^()]+))|(?P<cantrip_school_it2>[^()]+)\s+Trucchetto))"  # IT: Trucchetto di Scuola / Scuola Trucchetto
    r"\s*\((?P<classes>[^)]+)\)\s*$",
    re.IGNORECASE,
)

# Canonical Italian mapping for schools and classes
SCHOOL_EN_TO_IT = {
    "Abjuration": "Abiurazione",
    "Conjuration": "Evocazione",
    "Divination": "Divinazione",
    "Enchantment": "Ammaliamento",
    "Evocation": "Invocazione",
    "Illusion": "Illusione",
    "Necromancy": "Necromanzia",
    "Transmutation": "Trasmutazione",
}

SCHOOL_IT_NORMALIZE = {v.lower(): v for v in SCHOOL_EN_TO_IT.values()}

CLASS_EN_TO_IT = {
    "Barbarian": "Barbaro",
    "Bard": "Bardo",
    "Cleric": "Chierico",
    "Druid": "Druido",
    "Fighter": "Guerriero",
    "Monk": "Monaco",
    "Paladin": "Paladino",
    "Ranger": "Ranger",
    "Rogue": "Ladro",
    "Sorcerer": "Stregone",
    "Warlock": "Warlock",
    "Wizard": "Mago",
}
CLASS_IT_NORMALIZE = {v.lower(): v for v in CLASS_EN_TO_IT.values()}


def _norm_school_to_it(name: str) -> str:
    if not name:
        return ""
    name = name.strip()
    # Try English mapping first
    if name in SCHOOL_EN_TO_IT:
        return SCHOOL_EN_TO_IT[name]
    # Normalize Italian capitalization
    low = name.lower()
    if low in SCHOOL_IT_NORMALIZE:
        return SCHOOL_IT_NORMALIZE[low]
    return name


def _norm_classes_to_it(classes: List[str]) -> List[str]:
    out: List[str] = []
    for c in classes:
        k = c.strip()
        if not k:
            continue
        if k in CLASS_EN_TO_IT:
            out.append(CLASS_EN_TO_IT[k])
            continue
        low = k.lower()
        if low in CLASS_IT_NORMALIZE:
            out.append(CLASS_IT_NORMALIZE[low])
        else:
            out.append(k)
    return out


def _parse_meta(line: str) -> Dict:
    out: Dict = {}
    m = META_RE.match(line.strip())
    if not m:
        return out
    # Determine level and school from EN or IT groups
    lvl = m.group("lvl_en") or m.group("lvl_it")
    cantrip_school = (
        m.group("cantrip_school_en")
        or m.group("cantrip_school_it1")
        or m.group("cantrip_school_it2")
    )
    school = m.group("school_en") or m.group("school_it") or cantrip_school or ""
    if lvl:
        try:
            out["livello"] = int(lvl)
        except Exception:
            pass
        out["scuola"] = _norm_school_to_it(school)
    else:
        out["livello"] = 0
        out["scuola"] = _norm_school_to_it(school)
    classes = [c.strip() for c in (m.group("classes") or "").split(",") if c.strip()]
    if classes:
        out["classi"] = _norm_classes_to_it(classes)
    return out


def parse_spells(md_lines: List[str]) -> List[Dict]:
    # Spells are H4 sections under a H3 letter heading; parse all H4
    items = split_items(md_lines, level="h4")
    docs: List[Dict] = []
    for idx, (title, block) in enumerate(items, start=1):
        if not block:
            continue
        name = title.strip().strip("*")
        slug = slugify(name)
        meta = _parse_meta(first_italic_line(block) or "")
        fields = collect_labeled_fields(block)
        # Italian/English labeled keys are identical for spells content in this dataset
        lancio = {
            "tempo": fields.get("Casting Time") or fields.get("Tempo di Lancio"),
            "gittata": fields.get("Range") or fields.get("Gittata"),
            "componenti": fields.get("Components") or fields.get("Componenti"),
            "durata": fields.get("Duration") or fields.get("Durata"),
        }
        lancio = {k: v for k, v in lancio.items() if v}
        content = (f"#### {title}\n" + "\n".join(block)).strip() + "\n"
        doc: Dict = {
            "shared_id": shared_id_for("spell", idx),
            "slug": slug,
            "nome": name,
            "content": content,
        }
        doc.update(meta)
        if lancio:
            doc["lancio"] = lancio
        docs.append(doc)
    return docs
