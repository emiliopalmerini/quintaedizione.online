from __future__ import annotations

import re
from typing import Dict, List, Optional, Tuple

from ..utils import (
    SECTION_H2_RE,
    SECTION_H3_RE,
    SECTION_H4_RE,
    clean_value,
    norm_key,
    split_sections,
)

# Regex utils specific to class parsing (Italian headings)
TAB_LABEL_RE = re.compile(r"^Tabella:\s*(?P<label>.+?)\s*$", re.IGNORECASE)
LEVEL_FEATURE_H_RE = re.compile(
    r"^(?P<level>\d{1,2})\s*°\s*livello:\s*(?P<name>.+?)\s*$",
    re.IGNORECASE,
)


def _slugify(s: str) -> str:
    x = s.strip().lower()
    # Replace spaces and apostrophes; keep simple ascii-only without transliteration
    x = x.replace(" ", "-").replace("'", "").replace("’", "").replace("/", "-")
    return x


def _parse_markdown_table(
    block: List[str], start_idx: int
) -> Tuple[List[str], List[List[str]], int]:
    """
    Parse a GitHub-style markdown table starting at start_idx (header row).
    Returns (headers, rows, next_index_after_table).
    """
    n = len(block)
    i = start_idx
    if i >= n or "|" not in block[i]:
        raise ValueError("Expected table header row starting with '|'")
    header = [c.strip() for c in block[i].strip().strip("|").split("|")]
    i += 1
    # skip delimiter row
    if i < n and "|" in block[i]:
        i += 1
    rows: List[List[str]] = []
    while i < n:
        line = block[i]
        if not line.strip().startswith("|"):
            break
        cols = [c.strip() for c in line.strip().strip("|").split("|")]
        rows.append(cols)
        i += 1
    return header, rows, i


def _find_next_table(
    block: List[str], label_contains: str
) -> Tuple[List[str], List[List[str]]]:
    label_lc = label_contains.lower()
    i = 0
    n = len(block)
    while i < n:
        m = TAB_LABEL_RE.match(block[i].strip())
        if m and label_lc in m.group("label").lower():
            # Find first '|' line after this
            j = i + 1
            while j < n and not block[j].strip().startswith("|"):
                j += 1
            headers, rows, _ = _parse_markdown_table(block, j)
            return headers, rows
        i += 1
    return [], []


def _iter_tables(block: List[str]) -> List[Tuple[str, List[str], List[List[str]]]]:
    """Return all markdown tables in a block as (label, headers, rows)."""
    out: List[Tuple[str, List[str], List[List[str]]]] = []
    n = len(block)
    i = 0
    while i < n:
        line = block[i].strip()
        m = TAB_LABEL_RE.match(line)
        if m:
            label = m.group("label").strip()
            # find header row starting with '|'
            j = i + 1
            while j < n and not block[j].strip().startswith("|"):
                j += 1
            if j < n and block[j].strip().startswith("|"):
                try:
                    headers, rows, k = _parse_markdown_table(block, j)
                    out.append((label, headers, rows))
                    i = k
                    continue
                except Exception:
                    pass
        i += 1
    return out


def _parse_base_traits_table(block: List[str]) -> Dict:
    headers, rows = _find_next_table(block, "Tratti base")
    if not headers or not rows:
        return {}
    # Base traits tables are 2 columns: label | value
    traits: Dict[str, str] = {}
    for r in rows:
        if len(r) < 2:
            continue
        key = r[0].strip().strip(":")
        val = r[1].strip()
        if key:
            traits[key] = val

    out: Dict = {}

    # Caratteristica primaria
    cp = clean_value(traits.get("Caratteristica primaria"))
    if cp:
        out["caratteristica_primaria"] = cp

    # Dado Punti Ferita -> dado_vita (normalize like d12)
    dv = traits.get("Dado Punti Ferita")
    if dv:
        m = re.search(r"d\s*\d+", dv, re.IGNORECASE)
        if m:
            out["dado_vita"] = m.group(0).lower().replace(" ", "")

    # Tiri salvezza competenti -> list
    saves = traits.get("Tiri salvezza competenti")
    if saves:
        parts = re.split(r",| e ", saves)
        out["salvezze_competenze"] = [p.strip() for p in parts if p.strip()]

    # Abilità competenti -> choose + options
    skills = traits.get("Abilità competenti")
    if skills:
        m = re.search(r"Scegli\s+(\d+):\s*(.+)", skills, re.IGNORECASE)
        if m:
            scegli = int(m.group(1))
            opts = [o.strip() for o in re.split(r",\s*", m.group(2)) if o.strip()]
            out["abilità_competenze_opzioni"] = {"scegli": scegli, "opzioni": opts}
        else:
            out["abilità_competenze_opzioni"] = {
                "scegli": 0,
                "opzioni": [skills.strip()],
            }

    # Armi competenti -> list
    weapons = traits.get("Armi competenti")
    if weapons:
        parts = [p.strip() for p in re.split(r",| e ", weapons) if p.strip()]
        out["armi_competenze"] = parts

    # Armature addestramento -> list mapped to armature_competenze
    arm = traits.get("Armature addestramento")
    if arm:
        # Split on commas and ' e '
        parts = [p.strip() for p in re.split(r",| e ", arm) if p.strip()]
        out["armature_competenze"] = parts

    # Equipaggiamento iniziale -> options A/B if present
    eq = traits.get("Equipaggiamento iniziale")
    if eq:
        m = re.search(
            r"Scegli\s*A\s*o\s*B:\s*\(A\)\s*(.+?);\s*oppure\s*\(B\)\s*(.+)$",
            eq,
            re.IGNORECASE,
        )
        if m:
            a_items = [s.strip() for s in m.group(1).split(",") if s.strip()]
            b_items = [s.strip() for s in m.group(2).split(",") if s.strip()]
            out["equipaggiamento_iniziale_opzioni"] = [
                {"etichetta": "Opzione A", "oggetti": a_items},
                {"etichetta": "Opzione B", "oggetti": b_items},
            ]
        else:
            out["equipaggiamento_iniziale_opzioni"] = [
                {
                    "etichetta": "Default",
                    "oggetti": [s.strip() for s in eq.split(",") if s.strip()],
                }
            ]

    return out


def _parse_levels_table(block: List[str]) -> List[Dict]:
    headers, rows = _find_next_table(block, "Privilegi")
    if not headers or not rows:
        return []

    # Normalize header keys while keeping original labels for numeric columns
    def map_header(h: str) -> str:
        h = h.strip()
        if h.isdigit():
            return h
        mapping = {
            "Livello": "livello",
            "Bonus competenza": "bonus_competenza",
            # Normalize privilege column(s)
            "Privilegi di classe": "privilegi_di_classe",
            "Privilegi": "privilegi_di_classe",
            "Capacità": "privilegi_di_classe",  # backward compat
            "Trucchetti": "trucchetti_conosciuti",
            "Incantesimi preparati": "incantesimi_preparati",
        }
        return mapping.get(h, norm_key(h))

    keys = [map_header(h) for h in headers]
    out: List[Dict] = []
    for r in rows:
        if not any(c.strip() for c in r):
            continue
        row = {keys[i]: r[i].strip() if i < len(r) else "" for i in range(len(keys))}
        item: Dict = {}
        # livello
        try:
            item["livello"] = int(row.get("livello", "").strip() or 0)
        except Exception:
            continue
        # bonus_competenza
        bc = row.get("bonus_competenza") or ""
        m = re.search(r"(\+|−|-)?\s*(\d+)", bc)
        if m:
            item["bonus_competenza"] = int(m.group(2))
        # privilegi di classe (list of names per level row)
        privs = clean_value(row.get("privilegi_di_classe"))
        if privs:
            privs = privs.replace("—", "").strip()
            if privs:
                item["privilegi_di_classe"] = [
                    c.strip() for c in privs.split(",") if c.strip()
                ]
        # trucchetti/incantesimi_preparati
        for k in ("trucchetti_conosciuti", "incantesimi_preparati"):
            v = clean_value(row.get(k))
            if v:
                try:
                    item[k] = int(v)
                except Exception:
                    item[k] = v
        # slot_incantesimo (numeric headers 1..9)
        slots: Dict[str, int] = {}
        for hk, hv in row.items():
            if hk.isdigit():
                v = clean_value(hv)
                if v and v != "—":
                    try:
                        slots[hk] = int(v)
                    except Exception:
                        pass
        if slots:
            item["slot_incantesimo"] = slots

        # Copy across other known useful columns if present
        for maybe in [
            "furie",
            "danni_da_furia",
            "maestria_nelle_armi",
            "canalizzare_divinità",
            "dado",
        ]:
            v = clean_value(row.get(maybe))
            if v is not None:
                # Numbers where appropriate
                if v.isdigit():
                    item[maybe] = int(v)
                else:
                    item[maybe] = v

        out.append(item)
    return out


def _parse_features(block: List[str]) -> List[Dict]:
    # Find all h4 sections that look like "X° livello: Nome"
    sections = split_sections(block, SECTION_H4_RE)
    feats: List[Dict] = []
    for title, content in sections:
        m = LEVEL_FEATURE_H_RE.match(title)
        if not m:
            continue
        level = int(m.group("level"))
        name = m.group("name").strip()
        # Join consecutive lines until next feature header; content is already scoped
        text = "\n".join([ln for ln in content]).strip()
        feats.append({"nome": name, "livello": level, "descrizione": text})
    return feats


def _parse_class_block(title: str, block: List[str]) -> Dict:
    name = title.strip()
    doc: Dict = {
        "slug": _slugify(name),
        "nome": name,
    }

    # Include full markdown content of the class section (from H2 to before next H2)
    try:
        full_md = "\n".join([f"## {name}"] + block).strip() + "\n"
        if full_md:
            doc["content"] = full_md
    except Exception:
        pass

    # Base traits (Tratti base del <Classe>)
    core = _parse_base_traits_table(block)
    doc.update(core)

    # Levels table (Privilegi del <Classe>)
    levels = _parse_levels_table(block)
    if levels:
        doc["tabella_livelli"] = levels

    # Features by level (top-level, exclude subclass sections)
    # Collect lines until first H3 that is a subclass heading
    pre_sub_lines: List[str] = []
    subclass_h3_idx: Optional[int] = None
    for idx, ln in enumerate(block):
        m3 = SECTION_H3_RE.match(ln)
        if m3 and m3.group("title").strip().lower().startswith("sottoclasse del"):
            subclass_h3_idx = idx
            break
        pre_sub_lines.append(ln)
    feats = _parse_features(pre_sub_lines)
    if feats:
        doc["privilegi_di_classe"] = feats

    # Class spell lists under H3 "Lista incantesimi del(lo) <Classe>"
    lanciare = _parse_class_spell_lists(block)
    if lanciare:
        doc["lanciare_incantesimi"] = lanciare

    # Subclasses under H3 "Sottoclasse del <Classe>: <Nome>"
    subclasses: List[Dict] = []
    h3_sections = split_sections(block, SECTION_H3_RE)
    for h3_title, h3_block in h3_sections:
        t = h3_title.strip()
        m = re.match(r"^Sottoclasse\s+del\s+[^:]+:\s*(?P<name>.+)$", t, re.IGNORECASE)
        if not m:
            continue
        sub_name = m.group("name").strip()
        sub_feats = _parse_features(h3_block)
        inc_aggiuntivi = _parse_additional_spells(h3_block)
        sub_doc = {
            "slug": _slugify(sub_name),
            "nome": sub_name,
        }
        if sub_feats:
            sub_doc["privilegi_sottoclasse"] = sub_feats
        if inc_aggiuntivi:
            sub_doc["incantesimi_aggiuntivi"] = inc_aggiuntivi
        subclasses.append(sub_doc)
    if subclasses:
        doc["sottoclassi"] = subclasses

    return doc


def _parse_additional_spells(block: List[str]) -> List[Dict]:
    """Parse subclass-specific additional spells as an array of tables.
    Each entry: { "nome": <label>, "per_livello": { "3": [..], ... } }
    Detects any table within the block that has an 'Incantesimi' column.
    """
    tables = _iter_tables(block)
    out: List[Dict] = []
    for label, headers, rows in tables:
        # Identify spells column by header
        lvl_idx = None
        spells_idx = None
        for i, h in enumerate(headers):
            h_l = h.strip().lower()
            if lvl_idx is None and ("livello" in h_l):
                lvl_idx = i
            if spells_idx is None and ("incantesimi" in h_l):
                spells_idx = i
        if spells_idx is None or lvl_idx is None:
            continue
        per: Dict[str, List[str]] = {}
        for r in rows:
            if spells_idx >= len(r) or lvl_idx >= len(r):
                continue
            spells_cell = r[spells_idx].strip()
            if not spells_cell:
                continue
            m = re.search(r"(\d{1,2})", r[lvl_idx])
            if not m:
                continue
            key = m.group(1)
            spells = [s.strip() for s in spells_cell.split(",") if s.strip()]
            if spells:
                per[key] = spells
        if per:
            out.append({"nome": label, "per_livello": per})
    return out


def _parse_class_spell_lists(block: List[str]) -> Dict:
    """Parse class-level spell lists section (H3 'Lista incantesimi del...').
    Returns a dict with optional 'trucchetti' and 'lista_incantesimi'.
    """
    # Find the H3 section titled like 'Lista incantesimi del...' or 'dello'
    h3_sections = split_sections(block, SECTION_H3_RE)
    target_block: Optional[List[str]] = None
    for h3_title, h3_block in h3_sections:
        t = h3_title.strip().lower()
        if t.startswith("lista incantesimi del") or t.startswith(
            "lista incantesimi dello"
        ):
            target_block = h3_block
            break
    if not target_block:
        return {}

    trucchetti: List[str] = []
    lista: Dict[str, List[str]] = {}

    for label, headers, rows in _iter_tables(target_block):
        lab_l = label.lower()
        # Identify Incantesimo column if present
        inc_idx = None
        for i, h in enumerate(headers):
            if h.strip().lower().startswith("incantesimo"):
                inc_idx = i
                break
        if inc_idx is None:
            # Some broken tables may omit incantesimo names; skip
            continue
        names = [
            r[inc_idx].strip() for r in rows if inc_idx < len(r) and r[inc_idx].strip()
        ]
        if not names:
            continue
        if "trucchetti" in lab_l:
            trucchetti.extend(names)
            continue
        # Try to extract level from label, e.g., "di 3° livello"
        m = re.search(r"di\s+(\d{1,2})\s*°\s*livello", label, re.IGNORECASE)
        if m:
            lvl = m.group(1)
            lista[lvl] = names

    out: Dict = {}
    if trucchetti:
        out["trucchetti"] = trucchetti
    if lista:
        out["lista_incantesimi"] = lista
    return out


def parse_classes(md_lines: List[str]) -> List[Dict]:
    """
    Parse the Italian SRD classes markdown into structured documents, one per class.
    Produces fields aligned with docs/adrs/0001-data-model.md where possible.
    """
    # Split top-level by H2 (## NomeClasse)
    classes = split_sections(md_lines, SECTION_H2_RE)
    docs: List[Dict] = []
    for title, block in classes:
        # Skip the initial document H1 ('# Classi') which would parse no H2 title
        if title.lower().startswith("classi"):
            continue
        # Further trim block up to next H2 (already done) and ignore empty
        if not block:
            continue
        docs.append(_parse_class_block(title, block))
    return docs
