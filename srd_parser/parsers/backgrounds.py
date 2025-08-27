from __future__ import annotations

import re
from typing import Dict, List, Optional, Tuple

from ..utils import (
    BOLD_FIELD_RE,
    SECTION_H2_RE,
    SECTION_H3_RE,
    SECTION_H4_RE,
    clean_value,
    split_sections,
)


def _slugify(name: str) -> str:
    x = name.strip().lower()
    x = x.replace(" ", "-").replace("'", "").replace("’", "").replace("/", "-")
    x = re.sub(r"-+", "-", x)
    return x


def _parse_equipment_options(value: str) -> List[Dict]:
    """Parse equipment line like:
    "*Scegli A o B:* (A) ...; oppure (B) ..."
    Returns a list of two option dicts.
    """
    if not value:
        return []
    # Strip emphasis and helper notes
    v = value.strip()
    v = v.replace("**", "").replace("*", "")
    # Normalize punctuation
    v = v.replace("  ", " ")
    # Try to extract (A) ... ; oppure (B) ...
    m = re.search(
        r"\(A\)\s*(.+?);\s*(?:oppure|o)\s*\(B\)\s*(.+)$",
        v,
        re.IGNORECASE,
    )
    if not m:
        # Fallback: treat whole as a single option
        items = [s.strip() for s in v.split(",") if s.strip()]
        if items:
            return [{"etichetta": "Default", "oggetti": items}]
        return []
    a_items = [s.strip() for s in m.group(1).split(",") if s.strip()]
    b_items = [s.strip() for s in m.group(2).split(",") if s.strip()]
    return [
        {"etichetta": "Opzione A", "oggetti": a_items},
        {"etichetta": "Opzione B", "oggetti": b_items},
    ]


def _clean_parenthetical_refs(v: str) -> str:
    # Remove trailing references like (vedi "Talenti") or similar
    return re.sub(r"\s*\(vedi[^)]*\)\s*$", "", v, flags=re.IGNORECASE).strip()


def _parse_background_block(title: str, block: List[str]) -> Dict:
    doc: Dict = {
        "slug": _slugify(title),
        "nome": title.strip(),
    }

    # Include raw markdown content for the background section
    try:
        full_md = "\n".join([f"#### {title}"] + block).strip() + "\n"
        if full_md:
            doc["content"] = full_md
    except Exception:
        pass

    # Scan labeled lines
    fields: Dict[str, str] = {}
    for ln in block:
        m = BOLD_FIELD_RE.match(ln.strip())
        if not m:
            continue
        key = m.group("field").strip()
        val = m.group("value").strip()
        if key and val:
            fields[key] = val

    # Punteggi di Caratteristica -> list of 3
    abil = fields.get("Punteggi di Caratteristica")
    if abil:
        parts = [p.strip() for p in re.split(r",| e ", abil) if p.strip()]
        if parts:
            doc["punteggi_caratteristica"] = parts

    # Talento -> string (strip trailing refs)
    tal = fields.get("Talento")
    if tal:
        doc["talento"] = _clean_parenthetical_refs(tal)

    # Competenze in Abilità -> list
    comp_abi = fields.get("Competenze in Abilità") or fields.get("Competenza in Abilità")
    if comp_abi:
        parts = [p.strip() for p in re.split(r",| e ", comp_abi) if p.strip()]
        if parts:
            doc["abilità_competenze"] = parts

    # Competenza negli Strumenti -> list (normalize optional 'scegli ...')
    comp_str = fields.get("Competenza negli Strumenti") or fields.get("Competenze negli Strumenti")
    if comp_str:
        val = comp_str.replace("**", "").replace("*", "").strip()
        val = _clean_parenthetical_refs(val)
        # Normalize common choice phrasing
        if re.search(r"scegli\s+un\s+tipo\s+di\s+set\s+da\s+gioco", val, re.IGNORECASE):
            parts = ["Set da Gioco (scegli un tipo)"]
        else:
            parts = [p.strip() for p in re.split(r",| e ", val) if p.strip()]
        if parts:
            doc["strumenti_competenze"] = parts

    # Equipaggiamento -> options A/B
    eq = fields.get("Equipaggiamento")
    if eq:
        opts = _parse_equipment_options(eq)
        if opts:
            doc["equipaggiamento_iniziale_opzioni"] = opts

    return doc


def parse_backgrounds(md_lines: List[str]) -> List[Dict]:
    """
    Parse Italian backgrounds from '05_origini_personaggio.md'.
    Extracts the section under H3 'Descrizioni dei Background' and returns one
    document per background with fields aligned to ADR 0001.
    """
    # Narrow to the H3 section for background descriptions
    h3_sections = split_sections(md_lines, SECTION_H3_RE)
    target_block: Optional[List[str]] = None
    for h3_title, h3_block in h3_sections:
        t = h3_title.strip().lower()
        if "descrizioni" in t and "background" in t:
            target_block = h3_block
            break
    if not target_block:
        return []

    # Trim at the first H2 encountered (e.g., '## Specie del Personaggio')
    trimmed: List[str] = []
    for ln in target_block:
        if SECTION_H2_RE.match(ln):
            break
        trimmed.append(ln)

    # Background entries are H4 sections
    entries = split_sections(trimmed, SECTION_H4_RE)
    docs: List[Dict] = []
    for bg_title, bg_block in entries:
        if not bg_block:
            continue
        docs.append(_parse_background_block(bg_title, bg_block))
    return docs
