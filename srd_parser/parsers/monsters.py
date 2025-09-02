from __future__ import annotations

import re
from typing import Dict, List

from .items_common import collect_labeled_fields, slugify, split_items


TAG_RE = re.compile(r"^(?P<size>[^,]+)\s+(?P<type>[^,]+)(?:,\s*(?P<alignment>.+))?$")

# Additional regex patterns for monster-specific fields like GS (without colon)
GS_BOLD_RE = re.compile(r"^\*\*(?P<field>[^:*]+?)\*\*\s+(?P<value>.+?)\s*$")
GS_DASH_RE = re.compile(r"^-+\s*\*\*(?P<field>[^:*]+?)\*\*\s+(?P<value>.+?)\s*$")


def collect_monster_fields(block: List[str]) -> Dict[str, str]:
    """Enhanced field collection for monsters that handles GS field without colon"""
    fields: Dict[str, str] = {}
    
    # Import the standard regex patterns
    from ..utils import BOLD_FIELD_RE, DASH_FIELD_RE
    
    for ln in block:
        ln_s = ln.strip()
        
        # Try standard patterns first (with colon)
        m = BOLD_FIELD_RE.match(ln_s) or DASH_FIELD_RE.match(ln_s)
        
        # If no match, try GS-style patterns (without colon)  
        if not m:
            m = GS_BOLD_RE.match(ln_s) or GS_DASH_RE.match(ln_s)
        
        if not m:
            continue
            
        key = m.group("field").strip()
        val = m.group("value").strip()
        if key and val:
            fields[key] = val
    return fields


def _first_meta(block: List[str]) -> str:
    for ln in block:
        s = ln.strip()
        if s.startswith("*") and s.endswith("*"):
            return s.strip("*").strip()
    return ""


def _parse_meta(s: str) -> Dict:
    out: Dict = {}
    m = TAG_RE.match(s)
    if not m:
        return out
    out["tag"] = {
        "taglia": m.group("size").strip(),
        "tipo": m.group("type").strip(),
        "allineamento": (m.group("alignment") or "").strip(),
    }
    return out


def _map_fields(fields: Dict[str, str]) -> Dict:
    # Dataset uses English keys even in IT monsters file for core stats
    key_map = {
        "Armor Class": "ac",
        "Hit Points": "hp",
        "Speed": "velocita",
        "Initiative": "iniziativa",
    }
    out: Dict = {}
    for k, v in fields.items():
        if k in key_map:
            out[key_map[k]] = v
    # Challenge Rating (CR) or GS (IT)
    cr_val = fields.get("CR") or fields.get("Gs") or fields.get("GS") or fields.get("Grado di Sfida")
    if cr_val:
        raw = str(cr_val).strip()
        
        # Parse GS and PE from format like "14 (PE 11,500, o 13,000 nella tana)"
        gs_pe_match = re.search(r"(\d+(?:\.\d+)?|(\d+/\d+))\s*\(PE\s*([\d,\.]+)(?:,\s*o\s*([\d,\.]+)\s*nella\s*tana)?\)", raw)
        if gs_pe_match:
            # Extract GS
            gs_part = gs_pe_match.group(1)
            if "/" in gs_part:
                a, b = gs_part.split("/")
                gs_val = float(int(a) / int(b))
            else:
                gs_val = float(gs_part)
            out["cr"] = gs_val
            
            # Extract PE values (normalize numbers: remove dots/commas used as thousand separators)
            pe_base = gs_pe_match.group(3).replace(",", "").replace(".", "")
            out["xp"] = {"base": int(pe_base)}
            
            # Extract lair PE if present
            if gs_pe_match.group(4):
                pe_lair = gs_pe_match.group(4).replace(",", "").replace(".", "")
                out["xp"]["lair"] = int(pe_lair)
                
            out["cr_raw"] = raw
        else:
            # Fallback: Extract first numeric or fraction like 1/2, 3, 4.5
            m = re.search(r"(\d+/(\d+))|(\d+(?:\.\d+)?)", raw)
            if m:
                frac = m.group(1)
                num = m.group(3)
                try:
                    if frac:
                        a, b = frac.split("/")
                        val = float(int(a) / int(b))
                    else:
                        val = float(num)
                    out["cr"] = val
                    out["cr_raw"] = raw
                except Exception:
                    out["cr_raw"] = raw
    return out


def parse_monsters(md_lines: List[str], *, namespace: str = "monster") -> List[Dict]:
    items = split_items(md_lines, level="h2")
    docs: List[Dict] = []
    for idx, (title, block) in enumerate(items, start=1):
        if not block:
            continue
        fields = collect_monster_fields(block)  # Use enhanced field collection
        base = _map_fields(fields)
        meta = _parse_meta(_first_meta(block))
        content = (f"## {title}\n" + "\n".join(block)).strip() + "\n"
        doc: Dict = {
            "slug": slugify(title),
            "nome": title.strip(),
            **meta,
            **base,
            "content": content,
        }
        docs.append(doc)
    return docs
