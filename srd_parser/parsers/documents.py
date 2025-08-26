from __future__ import annotations

import os
import re
from typing import Dict, List, Optional

H1_RE = re.compile(r"^#\s+(?P<title>.+?)\s*$")


def _slug_from_filename(filename: str) -> str:
    base = os.path.basename(filename)
    # Drop extension
    if base.lower().endswith(".md"):
        base = base[:-3]
    # Strip leading page number and underscore, e.g., 01_foo_bar -> foo_bar
    m = re.match(r"^(\d+)_+(.*)$", base)
    name = m.group(2) if m else base
    # Slugify: underscores to hyphens, spaces to hyphens, lowercase
    slug = name.strip().lower().replace(" ", "-").replace("_", "-")
    # Collapse multiple hyphens
    slug = re.sub(r"-+", "-", slug)
    return slug


def _page_from_filename(filename: str) -> Optional[int]:
    base = os.path.basename(filename)
    m = re.match(r"^(\d+)_", base)
    if m:
        try:
            return int(m.group(1))
        except Exception:
            return None
    return None


def parse_document(md_lines: List[str], filename: Optional[str] = None) -> List[Dict]:
    """
    Parse a full markdown page into a single 'documento' entry, with
    - slug: derived from filename
    - titolo: first H1 line
    - content: full markdown (including H1)
    - numero_di_pagina: leading number from filename, if any
    """
    content = "\n".join(md_lines)
    title = None
    for ln in md_lines:
        m = H1_RE.match(ln.rstrip("\n"))
        if m:
            title = m.group("title").strip()
            break
    slug = _slug_from_filename(filename or "") if filename else None
    numero = _page_from_filename(filename or "") if filename else None
    doc: Dict = {
        "slug": slug or "",
        "titolo": title or "",
        "content": content,
    }
    if numero is not None:
        doc["numero_di_pagina"] = numero
    return [doc]

