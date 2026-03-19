"""Font → semantic role classification.

Maps font properties (name, size, color) to semantic roles used throughout
the parsing pipeline. Parameterized by FontProfile to support multiple PDFs.
"""

from __future__ import annotations

from enum import Enum, auto

from .extract import RawSpan
from .profiles import FontProfile, PROFILE_521


class SpanRole(Enum):
    # Headings
    H1 = auto()           # 26pt+ — top-level section titles
    H2 = auto()           # 18pt — group headings
    H3 = auto()           # 14-14.8pt — entity names
    H4 = auto()           # 14pt — sub-section headings
    H5 = auto()           # 12pt — spell names, sub-sub-sections
    H6 = auto()           # action/reaction section labels

    # Body text
    BODY = auto()
    BODY_BOLD = auto()
    BODY_ITALIC = auto()
    BODY_BOLD_ITALIC = auto()

    # Sidebar text
    SIDEBAR = auto()
    SIDEBAR_BOLD = auto()
    SIDEBAR_ITALIC = auto()
    SIDEBAR_BOLD_ITALIC = auto()

    # Stat block header area (5.2.1: Optima, #540000; 5.1: N/A)
    STAT_LABEL = auto()
    STAT_VALUE = auto()
    STAT_SUBTITLE = auto()

    # Stat block body (5.2.1: Optima, #231f20; 5.1: N/A)
    STAT_BODY = auto()
    STAT_ITALIC = auto()
    STAT_BOLD_ITALIC = auto()

    # Ability score grid (5.2.1 only)
    STAT_SCORE_LABEL = auto()
    STAT_SCORE_VALUE = auto()
    STAT_SCORE_HEADER = auto()

    # Table headers
    TABLE_HEADER = auto()
    TABLE_HEADER_SMALL = auto()

    # Table body
    TABLE_BODY = auto()

    # Links
    LINK = auto()

    # Footer
    FOOTER = auto()

    # Drop cap
    DROP_CAP = auto()

    # Unknown
    UNKNOWN = auto()

    # TOC
    TOC = auto()


# 5.2.1-specific colors not in the profile (secondary palette)
_STAT_GRAY = 0x636466
_SCORE_HEADER_GRAY = 0x8E9093


def _font_base(font_name: str) -> str:
    """Strip subset prefix (e.g., 'EZNTKP+GillSans-SemiBold' → 'GillSans-SemiBold')."""
    if "+" in font_name:
        return font_name.split("+", 1)[1]
    return font_name


def _size_match(actual: float, target: float, tol: float = 1.5) -> bool:
    return abs(actual - target) <= tol


def _classify_521(font: str, size: float, color: int, profile: FontProfile) -> SpanRole:
    """Classification rules for SRD 5.2.1 (GillSans/Optima/Cambria)."""

    # Footer — gray color
    if color == profile.footer_color:
        if size <= 6.5:
            return SpanRole.STAT_SCORE_HEADER
        return SpanRole.FOOTER

    # Score header (MOD SALV) — secondary gray
    if color == _SCORE_HEADER_GRAY:
        return SpanRole.STAT_SCORE_HEADER

    # Link
    if color == profile.link_color:
        return SpanRole.LINK

    # Stat block subtitle — Optima-Italic, #636466
    if color == _STAT_GRAY and profile.stat_font and profile.stat_font in font:
        return SpanRole.STAT_SUBTITLE

    # Stat block header labels/values — stat_color
    if profile.stat_color and color == profile.stat_color:
        if "SC700" in font:
            return SpanRole.STAT_SCORE_LABEL
        if profile.stat_font and profile.stat_font in font:
            if "Bold" in font:
                return SpanRole.STAT_LABEL
            return SpanRole.STAT_VALUE
        # GillSans score values
        return SpanRole.STAT_SCORE_VALUE

    # Red headings
    if color == profile.heading_color:
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
        # GillSans (not SemiBold) at 12pt — section action labels
        if profile.heading_font in font and _size_match(size, 12, 1.0):
            return SpanRole.H6
        return SpanRole.UNKNOWN

    # Dark body color
    if color == profile.body_color:
        # Drop cap
        if "SC700" in font and "Bold" in font and size >= 10:
            return SpanRole.DROP_CAP

        # SC700 small caps in sidebar/stat
        if "SC700" in font:
            if "SemiBold" in font:
                return SpanRole.STAT_SCORE_LABEL
            return SpanRole.SIDEBAR

        # Optima family — stat block body text
        if profile.stat_font and profile.stat_font in font:
            if "BoldItalic" in font:
                return SpanRole.STAT_BOLD_ITALIC
            if "Bold" in font:
                return SpanRole.STAT_BODY
            if "Italic" in font:
                return SpanRole.STAT_ITALIC
            return SpanRole.STAT_BODY

        # Sidebar/table font family
        if profile.sidebar_font in font:
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

        # Body font family
        if profile.body_font in font:
            if "BoldItalic" in font:
                return SpanRole.BODY_BOLD_ITALIC
            if "Bold" in font:
                return SpanRole.BODY_BOLD
            if "Italic" in font:
                return SpanRole.BODY_ITALIC
            return SpanRole.BODY

    # TOC entries (Cambria 9pt)
    if profile.body_font in font and _size_match(size, 9, 0.5):
        return SpanRole.TOC

    return SpanRole.UNKNOWN


def _classify_51(font: str, size: float, color: int, profile: FontProfile) -> SpanRole:
    """Classification rules for SRD 5.1 (Calibri/Cambria).

    Key differences from 5.2.1:
    - No distinct stat block fonts/colors — stat fields use body fonts
    - Headings are Calibri (not GillSans-SemiBold) at heading_color
    - Footer is Calibri-Italic 8pt at body_color (not a separate footer_color)
    - Action headers (Azioni, Reazioni) are Calibri-Bold 11pt at body_color
    """

    # Link
    if color == profile.link_color:
        return SpanRole.LINK

    # Headings — Calibri at heading_color (#943634)
    if color == profile.heading_color:
        if size >= 23:
            return SpanRole.H1
        if size >= 16:
            return SpanRole.H2
        if _size_match(size, 14, 1.0):
            return SpanRole.H3
        if _size_match(size, 12, 1.0):
            return SpanRole.H5
        return SpanRole.UNKNOWN

    # Everything else is body_color (#000000)
    if color == profile.body_color:
        # Footer: Calibri-Italic 8pt
        if "Italic" in font and profile.heading_font in font and _size_match(size, 8, 0.5):
            return SpanRole.FOOTER

        # Action headers: Calibri-Bold 11pt (Azioni, Reazioni)
        if "Bold" in font and profile.heading_font in font and _size_match(size, 11, 0.5):
            return SpanRole.H6

        # Table header: Calibri-Bold 9pt
        if "Bold" in font and profile.heading_font in font and _size_match(size, 9, 0.5):
            return SpanRole.TABLE_HEADER_SMALL

        # Table body / sidebar: Calibri 9pt (not bold, not italic)
        if (
            profile.heading_font in font
            and "Bold" not in font
            and "Italic" not in font
            and _size_match(size, 9, 0.5)
        ):
            return SpanRole.TABLE_BODY

        # Calibri 10pt body (stat labels etc.) — classified as BODY_BOLD
        if "Bold" in font and profile.heading_font in font and _size_match(size, 10, 0.5):
            if "Italic" in font:
                return SpanRole.BODY_BOLD_ITALIC
            return SpanRole.BODY_BOLD

        # Calibri-Italic 10pt (monster subtitles etc.)
        if "Italic" in font and profile.heading_font in font and _size_match(size, 10, 0.5):
            return SpanRole.BODY_ITALIC

        # Calibri 10pt plain (page numbers, SRD header line)
        if profile.heading_font in font and _size_match(size, 10, 0.5):
            return SpanRole.SIDEBAR

        # Body font (Cambria)
        if profile.body_font in font:
            if "BoldItalic" in font:
                return SpanRole.BODY_BOLD_ITALIC
            if "Bold" in font:
                return SpanRole.BODY_BOLD
            if "Italic" in font:
                return SpanRole.BODY_ITALIC
            return SpanRole.BODY

        # SymbolMT (bullet points)
        if "Symbol" in font:
            return SpanRole.BODY

    return SpanRole.UNKNOWN


def classify_span(span: RawSpan, profile: FontProfile | None = None) -> SpanRole:
    """Classify a raw span into a semantic role based on font metadata.

    Args:
        span: The raw span to classify.
        profile: Font profile for the target PDF. Defaults to PROFILE_521.
    """
    if profile is None:
        profile = PROFILE_521

    font = _font_base(span.font_name)
    size = span.font_size
    color = span.color

    if profile is PROFILE_521:
        return _classify_521(font, size, color, profile)
    else:
        return _classify_51(font, size, color, profile)
