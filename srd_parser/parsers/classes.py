# srd_parser/parsers/classes.py
from __future__ import annotations

import re
import os
from dataclasses import asdict, dataclass, field
from typing import Dict, List, Optional, Tuple

# Local utils are not available here, so we inline minimal helpers
H2_RE = re.compile(r"^##\s+(.+?)\s*$")
H3_RE = re.compile(r"^###\s+(.+?)\s*$")
H4_RE = re.compile(r"^####\s+(.+?)\s*$")
TABLE_CAPTION_RE = re.compile(r"^Table:\s*(?P<title>.+?)\s*$", re.IGNORECASE)
SUBCLASS_HEADER_RE = re.compile(r"^###\s+.*?\bSubclass:\s*(?P<name>.+?)\s*$", re.IGNORECASE)
LEVEL_HEADER_RE = re.compile(r"^(?:###|####)\s+Level\s+(?P<lvl>\d+)\s*[:\-\u2013\u2014]\s*(?P<name>.+?)\s*$")
SPELL_LIST_H3_RE = re.compile(r"^###\s+.*\bSpell List\b", re.IGNORECASE)

def _split_sections(lines: List[str], regex: re.Pattern) -> List[Tuple[str, List[str]]]:
    headers: List[Tuple[int, str]] = []
    for i, ln in enumerate(lines):
        m = regex.match(ln)
        if m:
            headers.append((i, m.group(1).strip()))
    blocks: List[Tuple[str, List[str]]] = []
    n = len(lines)
    for idx, (start, name) in enumerate(headers):
        end = headers[idx+1][0] if idx+1 < len(headers) else n
        blocks.append((name, lines[start+1:end]))
    return blocks

def _is_table_row(s: str) -> bool:
    return "|" in s and s.strip().startswith("|") and s.strip().endswith("|")

def _consume_table(start: int, lines: List[str]) -> Tuple[int, Optional[Dict[str, object]]]:
    i = start
    # skip blank lines
    while i < len(lines) and not lines[i].strip():
        i += 1
    # Optional caption
    if i < len(lines) and TABLE_CAPTION_RE.match(lines[i].strip()):
        caption = TABLE_CAPTION_RE.match(lines[i].strip()).group("title").strip()
        i += 1
    else:
        caption = None
    # Find header row and separator
    while i < len(lines) and not _is_table_row(lines[i]):
        # allow only blank lines here; otherwise bail
        if lines[i].strip():
            return start, None
        i += 1
    if i >= len(lines):
        return start, None
    header = [c.strip() for c in lines[i].strip().strip("|").split("|")]
    i += 1
    if i >= len(lines) or not re.search(r"\|-{3,}", lines[i]):
        return start, None
    i += 1
    rows: List[Dict[str, str]] = []
    while i < len(lines) and _is_table_row(lines[i]):
        cells = [c.strip() for c in lines[i].strip().strip("|").split("|")]
        if len(cells) < len(header):
            cells += [""] * (len(header)-len(cells))
        row = {header[j].strip(): cells[j].strip() for j in range(len(header))}
        rows.append(row)
        i += 1
    return i, {"caption": caption, "header": header, "rows": rows, "md": "\n".join(lines[start:i]).strip()}

def _collect_tables(lines: List[str], start: int) -> Tuple[int, List[Dict[str, object]]]:
    """Scan forward from start and collect consecutive tables, skipping blank lines and captions.
    """
    i = start
    # advance to next caption or table header
    while i < len(lines) and not (TABLE_CAPTION_RE.match(lines[i].strip()) or _is_table_row(lines[i])):
        i += 1
    tables: List[Dict[str, object]] = []
    while i < len(lines):
        j, table = _consume_table(i, lines)
        if table:
            tables.append(table)
            i = j
            while i < len(lines) and not lines[i].strip():
                i += 1
        else:
            break
    return i, tables

def _collect_all_tables(lines: List[str]) -> List[Dict[str, object]]:
    out: List[Dict[str, object]] = []
    i = 0
    n = len(lines)
    while i < n:
        # advance to next caption or table header
        while i < n and not (TABLE_CAPTION_RE.match(lines[i].strip()) or _is_table_row(lines[i])):
            i += 1
        j, table = _consume_table(i, lines)
        if table:
            out.append(table)
            i = j
        else:
            i += 1
    return out

def _rows_from_md(table_md: str) -> List[List[str]]:
    rows: List[List[str]] = []
    for ln in table_md.splitlines():
        if _is_table_row(ln):
            cells = [c.strip() for c in ln.strip().strip("|").split("|")]
            rows.append(cells)
    return rows

@dataclass
class Feature:
    level: int
    name: str
    text_md: str

@dataclass
class Subclass:
    name: str
    text_md: str = ""
    features: List[Feature] = field(default_factory=list)
    tables: List[Dict[str, object]] = field(default_factory=list)

@dataclass
class ClassDoc:
    name: str
    core_traits: Dict[str, str] = field(default_factory=dict)
    features_table: List[Dict[str, str]] = field(default_factory=list)
    features: List[Feature] = field(default_factory=list)
    spellcasting_progression: List[Dict[str, object]] = field(default_factory=list)
    spell_lists_by_level: Dict[str, List[Dict[str, str]]] = field(default_factory=dict)
    subclasses: List[Subclass] = field(default_factory=list)
    raw_spell_tables_md: List[str] = field(default_factory=list)

def _parse_core_traits(block_lines: List[str]) -> Tuple[Dict[str, str], str]:
    i = 0
    _, tables = _collect_tables(block_lines, i)
    core: Dict[str, str] = {}
    md = ""
    if tables:
        tbl = tables[0]
        md = tbl.get("md", "")
        # Re-parse table markdown to handle empty headers
        lines = md.splitlines()
        try:
            sep_idx = next(idx for idx, ln in enumerate(lines) if re.search(r"\|-{3,}", ln))
        except StopIteration:
            sep_idx = 1
        for ln in lines[sep_idx+1:]:
            if not _is_table_row(ln):
                continue
            cells = [c.strip() for c in ln.strip().strip("|").split("|")]
            if len(cells) >= 2:
                key = cells[0].strip()
                val = cells[1].strip()
                if key:
                    core[key] = val
    return core, md

def _merge_features_from_table(rows: List[Dict[str, str]] = None, table_md: Optional[str] = None) -> Tuple[Dict[int, List[str]], List[Dict[str, object]]]:
    if rows is None:
        rows = []
    by_level: Dict[int, List[str]] = {}
    spell_prog: List[Dict[str, object]] = []
    iter_rows = []
    if table_md:
        grid = _rows_from_md(table_md)
        if len(grid) >= 2:
            header = grid[0]
            for cells in grid[2:]:  # skip header and separator
                iter_rows.append(dict(zip(header, cells)))
    if rows:
        iter_rows.extend(rows)
    for r in iter_rows:
        lvl_raw = r.get("Level") or r.get("level") or r.get("LEVEL")
        if lvl_raw and lvl_raw.strip().isdigit():
            lvl = int(lvl_raw.strip())
        else:
            continue
        cell = r.get("Class Features") or r.get("Features") or ""
        names = [x.strip(" .") for x in re.split(r",|\band\b", cell) if x.strip() and x.strip() not in {"—", "-"}]
        if names:
            by_level.setdefault(lvl, []).extend(names)
        slot_cols = [k for k in r if re.fullmatch(r"[1-9]", k.strip())]
        has_can = any(k.lower().startswith("cantrip") for k in r)
        has_prep = any(k.lower().startswith("prepared") for k in r)
        if slot_cols or has_can or has_prep:
            entry: Dict[str, object] = {"level": lvl}
            if has_can:
                k = next(k for k in r if k.lower().startswith("cantrip"))
                entry["cantrips"] = int(re.sub(r"[^\d]", "", r[k])) if re.search(r"\d", r[k]) else 0
            if has_prep:
                k = next(k for k in r if k.lower().startswith("prepared"))
                entry["prepared_spells"] = int(re.sub(r"[^\d]", "", r[k])) if re.search(r"\d", r[k]) else 0
            for sk in slot_cols:
                entry[f"slot_{sk}"] = int(re.sub(r"[^\d]", "", r[sk])) if re.search(r"\d", r[sk]) else 0
            spell_prog.append(entry)
    return by_level, sorted(spell_prog, key=lambda x: x["level"])

def _parse_features_and_tables(block_lines: List[str]) -> Tuple[List[Dict[str, str]], List[Feature], List[Dict[str, object]]]:
    tables_all = _collect_all_tables(block_lines)
    features_table_rows: List[Dict[str, str]] = []
    features_table_md_list: List[Dict[str, object]] = []
    for t in tables_all:
        cap = (t.get("caption") or "").lower()
        if "features" in cap:
            features_table_rows = t["rows"]
            features_table_md_list = [t]
            break
    # Parse level features via headers
    features: List[Feature] = []
    j = 0
    while j < len(block_lines):
        m = LEVEL_HEADER_RE.match(block_lines[j])
        if not m:
            j += 1
            continue
        lvl = int(m.group("lvl"))
        name = m.group("name").strip()
        start = j + 1
        k = start
        while k < len(block_lines):
            if LEVEL_HEADER_RE.match(block_lines[k]) or H2_RE.match(block_lines[k]) or H3_RE.match(block_lines[k]):
                break
            k += 1
        text = "\n".join(block_lines[start:k]).strip()
        features.append(Feature(level=lvl, name=name, text_md=text))
        j = k
    return features_table_rows, features, features_table_md_list

def _parse_spell_list(block_lines: List[str]) -> Tuple[Dict[str, List[Dict[str, str]]], List[str]]:
    out: Dict[str, List[Dict[str, str]]] = {}
    raw_tables: List[str] = []
    i = 0
    while i < len(block_lines):
        if SPELL_LIST_H3_RE.match(block_lines[i].strip()):
            i += 1
            break
        i += 1
    while i < len(block_lines):
        if H2_RE.match(block_lines[i]) or SUBCLASS_HEADER_RE.match(block_lines[i]):
            break
        j, table = _consume_table(i, block_lines)
        if table:
            cap = (table.get("caption") or "").strip()
            raw_tables.append(table.get("md",""))
            lvl_match = re.search(r"(Cantrips|Level\s+\d+)", cap, re.IGNORECASE)
            if lvl_match:
                key = lvl_match.group(0).title().replace(" ", "")
                key = key if key.lower() == "cantrips" else key.replace("Level", "Level")
                out.setdefault(key, []).extend(table["rows"])
            i = j
            continue
        i += 1
    return out, raw_tables

def _parse_subclasses(block_lines: List[str]) -> List[Subclass]:
    subs: List[Subclass] = []
    # Find all subclass headers
    heads: List[Tuple[int, str]] = []
    for i, ln in enumerate(block_lines):
        m = SUBCLASS_HEADER_RE.match(ln.strip())
        if m:
            heads.append((i, m.group("name").strip()))
    if not heads:
        return subs
    for idx, (start, name) in enumerate(heads):
        end = heads[idx + 1][0] if idx + 1 < len(heads) else len(block_lines)
        chunk = block_lines[start + 1 : end]
        # parse features within this subclass chunk
        feats: List[Feature] = []
        j = 0
        while j < len(chunk):
            m2 = LEVEL_HEADER_RE.match(chunk[j])
            if not m2:
                j += 1
                continue
            lvl = int(m2.group("lvl"))
            fname = m2.group("name").strip()
            s = j + 1
            k = s
            while k < len(chunk):
                if LEVEL_HEADER_RE.match(chunk[k]) or H3_RE.match(chunk[k]) or H2_RE.match(chunk[k]):
                    break
                k += 1
            text = "\n".join(chunk[s:k]).strip()
            feats.append(Feature(level=lvl, name=fname, text_md=text))
            j = k
        subs.append(Subclass(name=name, features=feats))
    return subs

def _source_label() -> str:
    return os.environ.get("SOURCE_LABEL", "srd 5.2.1")

def parse_classes(md: List[str]) -> List[Dict]:
    docs: List[Dict] = []
    # Split on H2 (## ClassName)
    for class_name, block in _split_sections(md, H2_RE):
        if not class_name or class_name.strip().lower() == "classes":
            continue
        # pre-subclass slice
        first_sub_idx = next((i for i, ln in enumerate(block) if SUBCLASS_HEADER_RE.match(ln.strip())), len(block))
        pre = block[:first_sub_idx]

        # Core traits (first captioned table)
        core_traits, core_md = _parse_core_traits(pre)

        # Features table rows and features via headers
        ft_rows, features, ft_tables = _parse_features_and_tables(pre)
        # Group captured features by level
        grouped: Dict[int, List[Dict[str, str]]] = {}
        for f in features:
            grouped.setdefault(f.level, []).append({"name": f.name, "text": f.text_md})

        # Merge with features table names and collect spellcasting progression
        by_level_names, spell_prog = _merge_features_from_table(ft_rows, ft_tables[0]["md"] if ft_tables else None)
        for lvl, names in by_level_names.items():
            bucket = grouped.setdefault(lvl, [])
            existing = {e["name"] for e in bucket}
            for nm in names:
                if nm not in existing:
                    bucket.append({"name": nm, "text": ""})

        features_by_level = [
            {"level": lvl, "features": grouped[lvl]}
            for lvl in sorted(grouped.keys())
        ]

        # Spell list section and raw table md
        spell_lists_map, raw_tables_md = _parse_spell_list(block)
        spell_lists_by_level: List[Dict[str, object]] = []
        for key, rows in spell_lists_map.items():
            # Detect spell level from key
            if key.lower() == "cantrips":
                lvl = 0
            else:
                m = re.search(r"(\d+)$", key)
                if not m:
                    continue
                lvl = int(m.group(1))
            # pick first column that looks like a spell name
            for r in rows:
                pass
            if rows:
                headers = list(rows[0].keys())
                spell_col = next((h for h in headers if h.lower().startswith("spell")), headers[0])
                names = [r.get(spell_col, "").strip() for r in rows if r.get(spell_col, "").strip()]
                spell_lists_by_level.append({"spell_level": lvl, "spells": names})

        # Subclasses
        subclasses_docs: List[Dict[str, object]] = []
        for sub in _parse_subclasses(block):
            # group subclass features by level
            gsub: Dict[int, List[Dict[str, str]]] = {}
            for f in sub.features:
                gsub.setdefault(f.level, []).append({"name": f.name, "text": f.text_md})
            subclasses_docs.append({
                "name": sub.name,
                "features_by_level": [
                    {"level": lvl, "features": gsub[lvl]} for lvl in sorted(gsub.keys())
                ],
            })

        # Build final doc
        doc: Dict[str, object] = {
            "name": class_name.strip(),
            "features_by_level": features_by_level,
            "source": _source_label(),
        }
        if core_traits:
            doc["core_traits"] = core_traits
        if core_md:
            doc["core_traits_md"] = core_md
        if ft_tables:
            doc["features_table_md"] = ft_tables[0]["md"]
        if spell_prog:
            doc["spellcasting_progression"] = {"by_level": spell_prog}
        if spell_lists_by_level:
            # Ensure natural order by level
            doc["spell_lists_by_level"] = sorted(spell_lists_by_level, key=lambda x: x["spell_level"])  # type: ignore
        if raw_tables_md:
            doc["spell_list_tables_md"] = "\n\n".join(raw_tables_md)
        if subclasses_docs:
            doc["subclasses"] = subclasses_docs

        docs.append(doc)
    return docs

 
