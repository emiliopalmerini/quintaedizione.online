# srd_parser/parsers/classes.py
from __future__ import annotations

import re
from dataclasses import asdict, dataclass, field
from typing import Dict, List, Optional, Tuple

from ..utils import SECTION_H2_RE, split_sections, source_label

LEVEL_HEADER_RE = re.compile(r"^(?:###|####)\s+Level\s+(?P<lvl>\d+)\s*:\s*(?P<name>.+?)\s*$")
SUBCLASS_HEADER_RE = re.compile(r"^###\s+.*?\bSubclass:\s*(?P<name>.+?)\s*$")
TABLE_CAPTION_RE = re.compile(r"^Table:\s*(?P<title>.+?)\s*:?\s*$", re.IGNORECASE)
SPELL_LIST_H3_RE = re.compile(r"^###\s+.*\bSpell List\b", re.IGNORECASE)

def _is_table_row(s: str) -> bool:
    s = s.strip()
    return s.startswith("|") and s.endswith("|")

def _parse_md_table(lines: List[str], start: int) -> Tuple[Dict[str, object], int]:
    i = start
    buf: List[str] = []
    while i < len(lines) and _is_table_row(lines[i]):
        buf.append(lines[i].rstrip("\n"))
        i += 1
    if len(buf) < 2:
        return {"headers": [], "rows": [], "md": ""}, i
    raw_headers = [c.strip() for c in buf[0].strip().strip("|").split("|")]
    if all(h == "" for h in raw_headers):
        headers = ["Key", "Value"] if len(raw_headers) == 2 else [f"col{j+1}" for j in range(len(raw_headers))]
    else:
        headers = raw_headers
    rows: List[Dict[str, str]] = []
    for r in buf[2:]:
        cells = [c.strip() for c in r.strip().strip("|").split("|")]
        if len(cells) < len(headers):
            cells += [""] * (len(headers) - len(cells))
        elif len(cells) > len(headers):
            cells = cells[: len(headers)]
        if headers == ["Key", "Value"] and len(cells) > 2:
            cells = [cells[0], " | ".join(cells[1:])]
        rows.append({headers[j]: cells[j] for j in range(len(headers))})
    md_str = "\n".join(buf)
    return {"headers": headers, "rows": rows, "md": md_str}, i

def _scan_captioned_tables(block: List[str]) -> List[Dict[str, object]]:
    out: List[Dict[str, object]] = []
    i = 0
    while i < len(block):
        m = TABLE_CAPTION_RE.match(block[i].strip())
        if not m:
            i += 1
            continue
        title = m.group("title").strip()
        i += 1
        while i < len(block) and not block[i].strip():
            i += 1
        table, nxt = _parse_md_table(block, i)
        if table["md"]:
            out.append({"title": title, **table})
        i = nxt
    return out

def _tokenize_choices(s: str) -> List[str]:
    s = s.replace(";", ",")
    parts = re.split(r",|\bor\b|\band\b", s, flags=re.IGNORECASE)
    return [p.strip(" .") for p in parts if p.strip(" .")]

def _parse_starting_equipment(text: str) -> Dict[str, object]:
    raw = text.strip()
    options = []
    for m in re.finditer(r"\((?P<label>[A-Z])\)\s*(?P<body>[^;]+)", raw):
        label = m.group("label").strip()
        items = _tokenize_choices(m.group("body"))
        options.append({"id": label, "items": items})
    choose_labels = [o["id"] for o in options]
    return {"choose": choose_labels or [], "options": options or [], "raw": raw}

def _norm_core_key(k: str) -> str:
    s = k.strip().lower()
    s = re.sub(r"\s+", " ", s)
    if "primary ability" in s: return "primary_ability"
    if "hit point die" in s: return "hit_point_die"
    if "saving throw proficiencie" in s: return "saving_throw_proficiencies"
    if "skill proficiencie" in s: return "skill_proficiencies"
    if "weapon proficiencie" in s: return "weapon_proficiencies"
    if "tool proficiencie" in s: return "tool_proficiencies"
    if "armor training" in s: return "armor_training"
    if "starting equipment" in s: return "starting_equipment"
    return re.sub(r"[^a-z0-9]+", "_", s).strip("_")

def _parse_core_traits_table(table: Dict[str, object]) -> Dict[str, object]:
    rows: List[Dict[str, str]] = table.get("rows", [])  # type: ignore
    headers: List[str] = table.get("headers", [])       # type: ignore
    out: Dict[str, object] = {}
    if not rows:
        return out
    for r in rows:
        if headers == ["Key", "Value"] or ("Key" in r and "Value" in r):
            key_cell = r.get("Key", "")
            val_cell = r.get("Value", "")
        else:
            vals = list(r.values())
            key_cell = vals[0] if vals else ""
            val_cell = vals[1] if len(vals) > 1 else ""
        key = _norm_core_key(key_cell)
        val = val_cell.strip() or None
        if key:
            out[key] = val
    if isinstance(out.get("saving_throw_proficiencies"), str):
        out["saving_throw_proficiencies_list"] = _tokenize_choices(out["saving_throw_proficiencies"])  # type: ignore
    if isinstance(out.get("skill_proficiencies"), str):
        m = re.match(r"^choose\s+(?P<n>\d+)\s*:\s*(?P<rest>.+)$", str(out["skill_proficiencies"]), flags=re.IGNORECASE)
        if m:
            out["skill_proficiencies_parsed"] = {
                "choose": int(m.group("n")),
                "options": _tokenize_choices(m.group("rest")),
                "raw": out["skill_proficiencies"],
            }
    if isinstance(out.get("starting_equipment"), str):
        out["starting_equipment_parsed"] = _parse_starting_equipment(out["starting_equipment"])  # type: ignore
    return out

@dataclass
class FeatureEntry:
    name: str
    text: str = ""

@dataclass
class LevelFeatures:
    level: int
    features: List[FeatureEntry] = field(default_factory=list)

@dataclass
class SubclassDoc:
    name: str
    features_by_level: List[LevelFeatures] = field(default_factory=list)

@dataclass
class ClassDoc:
    name: str
    features_by_level: List[LevelFeatures] = field(default_factory=list)
    subclasses: List[SubclassDoc] = field(default_factory=list)
    core_traits: Dict[str, object] = field(default_factory=dict)
    core_traits_md: str = ""
    features_table_md: str = ""
    spellcasting_progression: Dict[str, object] = field(default_factory=dict)
    spell_lists_by_level: List[Dict[str, object]] = field(default_factory=list)
    spell_list_tables_md: str = ""
    source: str = field(default_factory=source_label)

def _extract_level_features(lines: List[str]) -> List[LevelFeatures]:
    buckets: Dict[int, List[FeatureEntry]] = {}
    current_level: Optional[int] = None
    current_feat: Optional[FeatureEntry] = None
    acc: List[str] = []

    def flush():
        nonlocal current_level, current_feat, acc
        if current_level is not None and current_feat is not None:
            current_feat.text = "\n".join(x for x in acc if x.strip()).strip()
            buckets.setdefault(current_level, []).append(current_feat)
        current_level, current_feat, acc = None, None, []

    for ln in lines:
        s = ln.rstrip("\n")
        if SUBCLASS_HEADER_RE.match(s):
            flush()
            break
        m = LEVEL_HEADER_RE.match(s.strip())
        if m:
            flush()
            current_level = int(m.group("lvl"))
            current_feat = FeatureEntry(name=m.group("name").strip())
            acc = []
            continue
        # stop collecting on any other H3 to avoid swallowing "Spell List" or other sections
        if s.strip().startswith("### ") and not LEVEL_HEADER_RE.match(s.strip()):
            flush()
            break
        if TABLE_CAPTION_RE.match(s.strip()):
            flush()
            continue
        if current_feat is not None:
            acc.append(s)
    flush()
    out: List[LevelFeatures] = []
    for lvl in sorted(buckets.keys()):
        out.append(LevelFeatures(level=lvl, features=buckets[lvl]))
    return out

def _slice_subclasses(block: List[str]) -> List[Tuple[str, List[str]]]:
    out: List[Tuple[str, List[str]]] = []
    n = len(block)
    i = 0
    headers: List[Tuple[int, str]] = []
    while i < n:
        m = SUBCLASS_HEADER_RE.match(block[i].rstrip("\n"))
        if m:
            headers.append((i, m.group("name").strip()))
        i += 1
    if not headers:
        return out
    for idx, (start, name) in enumerate(headers):
        end = headers[idx + 1][0] if idx + 1 < len(headers) else n
        out.append((name, block[start + 1 : end]))
    return out

def _merge_features_from_table(class_name: str, rows: List[Dict[str, str]], features_by_level: List[Dict]) -> Tuple[List[Dict], Dict[str, object]]:
    by_level_extra: Dict[int, List[str]] = {}
    spell_prog: List[Dict[str, object]] = []
    for r in rows:
        lvl_str = r.get("Level") or r.get("LEVEL") or r.get("level")
        if not (lvl_str and lvl_str.strip().isdigit()):
            continue
        lvl = int(lvl_str.strip())
        feat_cell = r.get("Class Features") or r.get("Features") or r.get("Class Features ".rstrip())
        if feat_cell:
            names = [x.strip(" .") for x in re.split(r",|\band\b", feat_cell) if x.strip()]
            names = [n for n in names if n and n not in {"—", "-"}]
            if names:
                by_level_extra.setdefault(lvl, []).extend(names)
        has_can = any(k.lower().startswith("cantrip") for k in r.keys())
        has_prep = any(k.lower().startswith("prepared") for k in r.keys())
        slot_cols = [k for k in r.keys() if re.fullmatch(r"[1-9]", k.strip())]
        if has_can or has_prep or slot_cols:
            entry: Dict[str, object] = {"level": lvl}
            can_key = next((k for k in r if k.lower().startswith("cantrip")), None)
            if can_key:
                entry["cantrips"] = int(re.sub(r"[^\d]", "", r[can_key])) if re.search(r"\d", r[can_key]) else 0
            prep_key = next((k for k in r if k.lower().startswith("prepared")), None)
            if prep_key:
                entry["prepared_spells"] = int(re.sub(r"[^\d]", "", r[prep_key])) if re.search(r"\d", r[prep_key]) else 0
            slots: Dict[str, int] = {}
            for c in slot_cols:
                val = r.get(c, "").strip()
                slots[c] = int(re.sub(r"[^\d]", "", val)) if re.search(r"\d", val) else 0
            if slots:
                entry["slots"] = slots
            spell_prog.append(entry)
    by_level_map: Dict[int, Dict] = {e["level"]: e for e in features_by_level}
    for lvl, names in by_level_extra.items():
        bucket = by_level_map.setdefault(lvl, {"level": lvl, "features": []})
        existing = {f["name"] for f in bucket["features"]}
        for nm in names:
            if nm not in existing:
                bucket["features"].append({"name": nm, "text": ""})
    merged = sorted(by_level_map.values(), key=lambda x: x["level"])
    return merged, {"by_level": sorted(spell_prog, key=lambda x: x["level"])} if spell_prog else {}

def _extract_spell_list(block: List[str]) -> Tuple[List[Dict[str, object]], str]:
    """
    Parse '### <Class> Spell List' section tables into per-level arrays.
    Returns (spell_lists_by_level, md_joined)
    spell_lists_by_level: [{'spell_level': 0..9, 'spells': [names...] }]
    """
    out: Dict[int, List[str]] = {}
    md_parts: List[str] = []
    # find the spell list header
    idx = None
    for i, ln in enumerate(block):
        if SPELL_LIST_H3_RE.match(ln.strip()):
            idx = i + 1
            break
    if idx is None:
        return [], ""
    # from idx forward, parse captioned tables
    i = idx
    while i < len(block):
        mcap = TABLE_CAPTION_RE.match(block[i].strip())
        if not mcap:
            i += 1
            continue
        title = mcap.group("title")
        # detect spell level
        level = None
        m0 = re.search(r"\bLevel\s*0\b", title, re.IGNORECASE)
        mcan = re.search(r"\bCantrips\b|\(Level\s*0\b", title, re.IGNORECASE)
        mdig = re.search(r"\bLevel\s*(\d+)\b", title, re.IGNORECASE)
        if mcan or m0:
            level = 0
        elif mdig:
            level = int(mdig.group(1))
        i += 1
        while i < len(block) and not block[i].strip():
            i += 1
        table, nxt = _parse_md_table(block, i)
        if table["md"]:
            md_parts.append(f"Table: {title}\n{table['md']}")
            # extract 'Spell' column (first column fallback)
            if level is not None:
                headers = table["headers"]  # type: ignore
                rows = table["rows"]        # type: ignore
                spell_col = next((h for h in headers if h.lower().startswith("spell")), headers[0] if headers else None)
                if spell_col:
                    names = [r.get(spell_col, "").strip() for r in rows if r.get(spell_col, "").strip()]
                    if names:
                        out.setdefault(level, []).extend(names)
        i = nxt
        # stop if we hit a new H3 section
        if i < len(block) and block[i].strip().startswith("### "):
            break
    result = [{"spell_level": k, "spells": out[k]} for k in sorted(out.keys())]
    return result, ("\n\n".join(md_parts) if md_parts else "")

def parse_classes(md: List[str]) -> List[Dict]:
    docs: List[Dict] = []
    for class_name, block in split_sections(md, SECTION_H2_RE):
        if not class_name or class_name.strip().lower() == "classes":
            continue
        pre_end = next((i for i, ln in enumerate(block) if SUBCLASS_HEADER_RE.match(ln.rstrip("\n"))), len(block))
        pre_slice = block[:pre_end]

        class_features_lvls = _extract_level_features(pre_slice)
        features_by_level = [{"level": lf.level, "features": [asdict(f) for f in lf.features]} for lf in class_features_lvls]

        tables = _scan_captioned_tables(pre_slice)
        core_traits_md = None
        core_traits = {}
        features_table_md = None
        spellcasting = {}

        for t in tables:
            title_low = t["title"].lower()
            if "core" in title_low and "traits" in title_low:
                core_traits_md = t["md"]
                core_traits = _parse_core_traits_table(t)
            if "features" in title_low:
                features_table_md = t["md"]
                features_by_level, spellcasting = _merge_features_from_table(class_name, t["rows"], features_by_level)

        # subclasses
        subclasses: List[Dict] = []
        for sub_name, sub_lines in _slice_subclasses(block):
            sub_lvls = _extract_level_features(sub_lines)
            subclasses.append({
                "name": sub_name,
                "features_by_level": [{"level": lf.level, "features": [asdict(f) for f in lf.features]} for lf in sub_lvls],
            })

        # spell list section (can be after level 20 feature)
        spell_lists_by_level, spell_list_tables_md = _extract_spell_list(block)

        doc: Dict[str, object] = {
            "name": class_name.strip(),
            "features_by_level": features_by_level,
            "subclasses": subclasses,
            "source": source_label(),
        }
        if core_traits:
            doc["core_traits"] = core_traits
        if core_traits_md:
            doc["core_traits_md"] = core_traits_md
        if features_table_md:
            doc["features_table_md"] = features_table_md
        if spellcasting:
            doc["spellcasting_progression"] = spellcasting
        if spell_lists_by_level:
            doc["spell_lists_by_level"] = spell_lists_by_level
        if spell_list_tables_md:
            doc["spell_list_tables_md"] = spell_list_tables_md

        docs.append(doc)
    return docs

