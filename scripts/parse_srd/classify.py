"""Font → semantic role classification.

Maps font properties (name, size, color) to semantic roles used throughout
the parsing pipeline. Font values calibrated against IT_SRD_CC_v5.2.1.pdf.
"""

from __future__ import annotations

from enum import Enum, auto

from .extract import RawSpan


class SpanRole(Enum):
    # Headings — GillSans-SemiBold, color #8c2220
    H1 = auto()           # 26pt — top-level section titles
    H2 = auto()           # 18pt — group headings (dragon families, NPC groups)
    H3 = auto()           # 14.8pt — entity names (monster, spell entry)
    H4 = auto()           # 14pt — sub-section headings
    H5 = auto()           # 12pt — GillSans-SemiBold, #8c2220 (spell names, sub-sub-sections)
    H6 = auto()           # 12pt — GillSans (not SemiBold), #8c2220 (Azioni, Reazioni labels)

    # Body text — Cambria family, color #231f20
    BODY = auto()         # 10pt Cambria
    BODY_BOLD = auto()    # 10pt Cambria-Bold
    BODY_ITALIC = auto()  # 10pt Cambria-Italic
    BODY_BOLD_ITALIC = auto()  # 10pt Cambria-BoldItalic

    # Sidebar text — GillSans family, color #231f20, 9-9.5pt
    SIDEBAR = auto()      # GillSans
    SIDEBAR_BOLD = auto()  # GillSans-Bold
    SIDEBAR_ITALIC = auto()  # GillSans-Italic
    SIDEBAR_BOLD_ITALIC = auto()  # GillSans-BoldItalic

    # Stat block header area — Optima family, color #540000
    STAT_LABEL = auto()   # 9pt Optima-Bold #540000 (CA, PF, Velocità)
    STAT_VALUE = auto()   # 9pt Optima-Regular #540000

    # Stat block subtitle — Optima-Italic, color #636466
    STAT_SUBTITLE = auto()

    # Stat block body — Optima family, color #231f20, 9.5pt
    STAT_BODY = auto()          # Optima-Regular
    STAT_ITALIC = auto()        # Optima-Italic
    STAT_BOLD_ITALIC = auto()   # Optima-BoldItalic

    # Ability score grid — GillSans family, color #540000
    STAT_SCORE_LABEL = auto()   # SC700 variants
    STAT_SCORE_VALUE = auto()   # GillSans 9.8pt #540000
    STAT_SCORE_HEADER = auto()  # GillSans 6pt #8e9093 ("MOD SALV")

    # Table headers — GillSans-SemiBold, color #231f20
    TABLE_HEADER = auto()       # 10.5pt
    TABLE_HEADER_SMALL = auto()  # 9.2pt

    # Table body — GillSans, color #231f20, 9.5pt
    TABLE_BODY = auto()

    # Links
    LINK = auto()         # Cambria #1e5e9e

    # Footer
    FOOTER = auto()       # 11pt GillSans/GillSans-SemiBold #808285

    # Drop cap — GillSans-Bold-SC700
    DROP_CAP = auto()

    # Unknown
    UNKNOWN = auto()

    # TOC
    TOC = auto()


# Colors
_RED = 0x8c2220
_DARK = 0x231f20
_STAT_RED = 0x540000
_STAT_GRAY = 0x636466
_FOOTER_GRAY = 0x808285
_SCORE_HEADER_GRAY = 0x8e9093
_LINK_BLUE = 0x1e5e9e


def _font_base(font_name: str) -> str:
    """Strip subset prefix (e.g., 'EZNTKP+GillSans-SemiBold' → 'GillSans-SemiBold')."""
    if "+" in font_name:
        return font_name.split("+", 1)[1]
    return font_name


def _size_match(actual: float, target: float, tol: float = 1.5) -> bool:
    return abs(actual - target) <= tol


def classify_span(span: RawSpan) -> SpanRole:
    """Classify a raw span into a semantic role based on font metadata."""
    font = _font_base(span.font_name)
    size = span.font_size
    color = span.color

    # Footer — GillSans variants at 11pt with gray color
    if color == _FOOTER_GRAY:
        if size <= 6.5:
            return SpanRole.STAT_SCORE_HEADER
        return SpanRole.FOOTER

    # Score header (MOD SALV) — 6pt GillSans #8e9093
    if color == _SCORE_HEADER_GRAY:
        return SpanRole.STAT_SCORE_HEADER

    # Link — Cambria at 10pt with blue color
    if color == _LINK_BLUE:
        return SpanRole.LINK

    # Stat block subtitle — Optima-Italic, #636466
    if color == _STAT_GRAY and "Optima" in font:
        return SpanRole.STAT_SUBTITLE

    # Stat block header labels/values — color #540000
    if color == _STAT_RED:
        if "SC700" in font:
            return SpanRole.STAT_SCORE_LABEL
        if "Optima" in font:
            if "Bold" in font:
                return SpanRole.STAT_LABEL
            return SpanRole.STAT_VALUE
        # GillSans score values
        return SpanRole.STAT_SCORE_VALUE

    # Red headings — GillSans-SemiBold, #8c2220
    if color == _RED:
        if "SemiBold" in font or "Semibold" in font:
            if size >= 23:
                return SpanRole.H1
            if size >= 16:
                return SpanRole.H2
            if _size_match(size, 14.8, 0.5):
                return SpanRole.H3
            if _size_match(size, 14, 1.0):
                return SpanRole.H4
            if _size_match(size, 12, 1.0):
                return SpanRole.H5
            return SpanRole.H5
        # GillSans (not SemiBold) at 12pt #8c2220 — section action labels
        if "GillSans" in font and _size_match(size, 12, 1.0):
            return SpanRole.H6
        return SpanRole.UNKNOWN

    # Dark text (#231f20)
    if color == _DARK:
        # Drop cap
        if "SC700" in font and "Bold" in font and size >= 10:
            return SpanRole.DROP_CAP

        # SC700 small caps in sidebar/stat
        if "SC700" in font:
            if "SemiBold" in font:
                return SpanRole.STAT_SCORE_LABEL
            return SpanRole.SIDEBAR

        # Optima family — stat block body text
        if "Optima" in font:
            if "BoldItalic" in font:
                return SpanRole.STAT_BOLD_ITALIC
            if "Bold" in font:
                return SpanRole.STAT_BODY  # rare in body, treat as body
            if "Italic" in font:
                return SpanRole.STAT_ITALIC
            return SpanRole.STAT_BODY

        # GillSans family — sidebar/table text
        if "GillSans" in font:
            if "SemiBold" in font:
                if _size_match(size, 10.5, 0.5):
                    return SpanRole.TABLE_HEADER
                if _size_match(size, 9.2, 0.5):
                    return SpanRole.TABLE_HEADER_SMALL
                return SpanRole.SIDEBAR_BOLD
            if "BoldItalic" in font:
                return SpanRole.SIDEBAR_BOLD_ITALIC
            if "Bold" in font:
                return SpanRole.SIDEBAR_BOLD
            if "Italic" in font:
                return SpanRole.SIDEBAR_ITALIC
            return SpanRole.SIDEBAR

        # Cambria family — main body text
        if "Cambria" in font:
            if "BoldItalic" in font:
                return SpanRole.BODY_BOLD_ITALIC
            if "Bold" in font:
                return SpanRole.BODY_BOLD
            if "Italic" in font:
                return SpanRole.BODY_ITALIC
            return SpanRole.BODY

    # TOC entries (Cambria 9pt)
    if "Cambria" in font and _size_match(size, 9, 0.5):
        return SpanRole.TOC

    return SpanRole.UNKNOWN
