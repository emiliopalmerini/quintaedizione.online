from __future__ import annotations

import re
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional

__all__ = ["QFilterOptions", "q_filter"]


@dataclass(frozen=True)
class QFilterOptions:
    """Opzioni per costruire il filtro testuale."""

    fields: List[str] = field(
        default_factory=lambda: [
            "name",
            "term",
            "description",
            "description_md",
            "title",
        ]
    )
    use_text: bool = False  # usa {$text: {$search: q}} se c'è un indice testo
    min_len: int = 2  # lunghezza minima della query dopo strip
    case_insensitive: bool = True  # aggiunge $options: "i" al regex
    raw_regex: bool = False  # non eseguire re.escape(q)
    whole_words: bool = False  # racchiude con \b...\b
    prefix: bool = False  # match da inizio stringa ^q


def _regex_payload(q: str, opt: QFilterOptions) -> Dict[str, Any]:
    pat = q if opt.raw_regex else re.escape(q)
    if opt.whole_words:
        pat = rf"\b{pat}\b"
    if opt.prefix:
        pat = rf"^{pat}"
    payload: Dict[str, Any] = {"$regex": pat}
    if opt.case_insensitive:
        payload["$options"] = "i"
    return payload


def q_filter(
    q: str,
    *,
    options: Optional[QFilterOptions] = None,
    extra: Optional[Dict[str, Any]] = None,
) -> Dict[str, Any]:
    """
    Costruisce un filtro MongoDB per ricerca testuale.
    - Se q è vuota o corta: ritorna `extra` o {}.
    - Se use_text=True: usa {$text: {$search: q}} (richiede indice di testo).
    - Altrimenti: usa OR di regex sui campi.
    - Se `extra` è passato: combina con AND.

    Esempio:
        q_filter("dragon")
        q_filter("dra", options=QFilterOptions(prefix=True))
        q_filter("mago", options=QFilterOptions(use_text=True), extra={"type": "spell"})
    """
    opt = options or QFilterOptions()
    qn = (q or "").strip()
    if len(qn) < opt.min_len:
        return extra or {}

    if opt.use_text:
        base: Dict[str, Any] = {"$text": {"$search": qn}}
    else:
        rx = _regex_payload(qn, opt)
        base = {"$or": [{f: rx} for f in opt.fields]}

    if extra:
        return {"$and": [base, extra]}
    return base

