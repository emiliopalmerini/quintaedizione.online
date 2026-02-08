"""Port of internal/domain/slug.go — slugification for Italian text.

Algorithm:
  1. Lowercase + trim
  2. NFD normalize → remove combining marks → NFC
  3. Spaces → hyphens
  4. Remove non-alphanumeric (except hyphens)
  5. Collapse consecutive hyphens
  6. Trim leading/trailing hyphens
  7. Return "n-a" if empty
"""

from __future__ import annotations

import re
import unicodedata

_NON_ALNUM = re.compile(r"[^a-z0-9\-]+")
_MULTI_DASH = re.compile(r"-+")


def slugify(value: str) -> str:
    """Convert a title string to a URL-safe slug.

    Mirrors the Go implementation in internal/domain/slug.go exactly.
    """
    s = value.lower().strip()

    # NFD → strip combining marks (accents) → NFC
    s = unicodedata.normalize("NFD", s)
    s = "".join(c for c in s if unicodedata.category(c) != "Mn")
    s = unicodedata.normalize("NFC", s)

    s = s.replace(" ", "-")
    s = _NON_ALNUM.sub("", s)
    s = _MULTI_DASH.sub("-", s)
    s = s.strip("-")

    return s or "n-a"
