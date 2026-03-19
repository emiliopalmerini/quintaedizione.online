# /// script
# requires-python = ">=3.12"
# dependencies = ["pytest"]
# ///
"""Tests for classify_span with profile-based classification.

Tests cover both SRD 5.2.1 and 5.1 font patterns, ensuring the same
semantic roles are produced from different font families/colors.
"""

from __future__ import annotations

import pytest

from ..classify import SpanRole, classify_span
from ..extract import RawSpan
from ..profiles import PROFILE_521, PROFILE_51


def _span(
    text: str = "test",
    font_name: str = "Cambria",
    font_size: float = 10.0,
    color: int = 0x231F20,
) -> RawSpan:
    """Create a RawSpan with sensible defaults."""
    return RawSpan(
        text=text,
        font_name=font_name,
        font_size=font_size,
        color=color,
        bbox=(0, 0, 100, 10),
        page_num=1,
    )


# ── SRD 5.2.1 (existing behavior, now profile-parameterized) ──


class TestClassify521Headings:
    """Heading classification for SRD 5.2.1 (GillSans-SemiBold, #8c2220)."""

    def test_h1_26pt(self):
        span = _span(font_name="GillSans-SemiBold", font_size=26.0, color=0x8C2220)
        assert classify_span(span, PROFILE_521) == SpanRole.H1

    def test_h2_18pt(self):
        span = _span(font_name="GillSans-SemiBold", font_size=18.0, color=0x8C2220)
        assert classify_span(span, PROFILE_521) == SpanRole.H2

    def test_h3_14_8pt(self):
        span = _span(font_name="GillSans-SemiBold", font_size=14.8, color=0x8C2220)
        assert classify_span(span, PROFILE_521) == SpanRole.H3

    def test_h4_14pt(self):
        span = _span(font_name="GillSans-SemiBold", font_size=14.0, color=0x8C2220)
        assert classify_span(span, PROFILE_521) == SpanRole.H4

    def test_h5_12pt(self):
        span = _span(font_name="GillSans-SemiBold", font_size=12.0, color=0x8C2220)
        assert classify_span(span, PROFILE_521) == SpanRole.H5

    def test_h6_gillsans_not_semibold_12pt(self):
        span = _span(font_name="GillSans", font_size=12.0, color=0x8C2220)
        assert classify_span(span, PROFILE_521) == SpanRole.H6

    def test_subset_prefix_stripped(self):
        span = _span(font_name="EZNTKP+GillSans-SemiBold", font_size=18.0, color=0x8C2220)
        assert classify_span(span, PROFILE_521) == SpanRole.H2


class TestClassify521Body:
    """Body text classification for SRD 5.2.1 (Cambria, #231f20)."""

    def test_body(self):
        span = _span(font_name="Cambria", font_size=10.0, color=0x231F20)
        assert classify_span(span, PROFILE_521) == SpanRole.BODY

    def test_body_bold(self):
        span = _span(font_name="Cambria-Bold", font_size=10.0, color=0x231F20)
        assert classify_span(span, PROFILE_521) == SpanRole.BODY_BOLD

    def test_body_italic(self):
        span = _span(font_name="Cambria-Italic", font_size=10.0, color=0x231F20)
        assert classify_span(span, PROFILE_521) == SpanRole.BODY_ITALIC

    def test_body_bold_italic(self):
        span = _span(font_name="Cambria-BoldItalic", font_size=10.0, color=0x231F20)
        assert classify_span(span, PROFILE_521) == SpanRole.BODY_BOLD_ITALIC


class TestClassify521StatBlock:
    """Stat block classification for SRD 5.2.1 (Optima, #540000)."""

    def test_stat_label(self):
        span = _span(font_name="Optima-Bold", font_size=9.0, color=0x540000)
        assert classify_span(span, PROFILE_521) == SpanRole.STAT_LABEL

    def test_stat_value(self):
        span = _span(font_name="Optima-Regular", font_size=9.0, color=0x540000)
        assert classify_span(span, PROFILE_521) == SpanRole.STAT_VALUE

    def test_stat_subtitle(self):
        span = _span(font_name="Optima-Italic", font_size=10.0, color=0x636466)
        assert classify_span(span, PROFILE_521) == SpanRole.STAT_SUBTITLE

    def test_stat_body(self):
        span = _span(font_name="Optima-Regular", font_size=9.5, color=0x231F20)
        assert classify_span(span, PROFILE_521) == SpanRole.STAT_BODY

    def test_stat_italic(self):
        span = _span(font_name="Optima-Italic", font_size=9.5, color=0x231F20)
        assert classify_span(span, PROFILE_521) == SpanRole.STAT_ITALIC


class TestClassify521Other:
    """Links, footer, table headers for SRD 5.2.1."""

    def test_link(self):
        span = _span(font_name="Cambria", font_size=10.0, color=0x1E5E9E)
        assert classify_span(span, PROFILE_521) == SpanRole.LINK

    def test_footer(self):
        span = _span(font_name="GillSans-SemiBold", font_size=11.0, color=0x808285)
        assert classify_span(span, PROFILE_521) == SpanRole.FOOTER

    def test_table_header(self):
        span = _span(font_name="GillSans-SemiBold", font_size=10.5, color=0x231F20)
        assert classify_span(span, PROFILE_521) == SpanRole.TABLE_HEADER


# ── SRD 5.1 (new, Calibri-based) ──


class TestClassify51Headings:
    """Heading classification for SRD 5.1 (Calibri, #943634)."""

    def test_h1_26pt(self):
        span = _span(font_name="Calibri", font_size=26.0, color=0x943634)
        assert classify_span(span, PROFILE_51) == SpanRole.H1

    def test_h1_30pt(self):
        """The title page uses 30pt Calibri."""
        span = _span(font_name="Calibri", font_size=30.0, color=0x943634)
        assert classify_span(span, PROFILE_51) == SpanRole.H1

    def test_h2_18pt(self):
        span = _span(font_name="Calibri", font_size=18.0, color=0x943634)
        assert classify_span(span, PROFILE_51) == SpanRole.H2

    def test_h3_14pt(self):
        """5.1 uses 14pt for entity names (vs 14.8pt in 5.2.1)."""
        span = _span(font_name="Calibri", font_size=14.0, color=0x943634)
        assert classify_span(span, PROFILE_51) == SpanRole.H3

    def test_h5_12pt(self):
        span = _span(font_name="Calibri", font_size=12.0, color=0x943634)
        assert classify_span(span, PROFILE_51) == SpanRole.H5

    def test_h3_bold_monster_name(self):
        """Monster names in 5.1 are Calibri-Bold 12pt #943634."""
        span = _span(font_name="Calibri-Bold", font_size=12.0, color=0x943634)
        assert classify_span(span, PROFILE_51) == SpanRole.H5


class TestClassify51Body:
    """Body text classification for SRD 5.1 (Cambria, #000000)."""

    def test_body(self):
        span = _span(font_name="Cambria", font_size=10.0, color=0x000000)
        assert classify_span(span, PROFILE_51) == SpanRole.BODY

    def test_body_bold(self):
        span = _span(font_name="Cambria-Bold", font_size=10.0, color=0x000000)
        assert classify_span(span, PROFILE_51) == SpanRole.BODY_BOLD

    def test_body_italic(self):
        span = _span(font_name="Cambria-Italic", font_size=10.0, color=0x000000)
        assert classify_span(span, PROFILE_51) == SpanRole.BODY_ITALIC

    def test_body_bold_italic(self):
        span = _span(font_name="Cambria-BoldItalic", font_size=10.0, color=0x000000)
        assert classify_span(span, PROFILE_51) == SpanRole.BODY_BOLD_ITALIC


class TestClassify51StatBlock:
    """Stat blocks in 5.1 use Calibri at body color — no distinct stat roles.

    Stat labels (Classe Armatura, PF, etc.) appear as Calibri-Bold 10pt #000000.
    These classify as BODY_BOLD, not STAT_LABEL — the monster parser handles
    the distinction by text pattern matching.
    """

    def test_stat_label_is_body_bold(self):
        """CA, PF, Velocità labels are Calibri-Bold 10pt — classified as BODY_BOLD."""
        span = _span(font_name="Calibri-Bold", font_size=10.0, color=0x000000)
        assert classify_span(span, PROFILE_51) == SpanRole.BODY_BOLD

    def test_stat_value_is_sidebar(self):
        """Stat values are Calibri 10pt — classified as SIDEBAR (same as page headers)."""
        span = _span(font_name="Calibri", font_size=10.0, color=0x000000)
        assert classify_span(span, PROFILE_51) == SpanRole.SIDEBAR

    def test_ability_score_labels_are_body_bold(self):
        """FOR, DES, COS etc. are Calibri-Bold 10pt — BODY_BOLD."""
        span = _span(text="FOR", font_name="Calibri-Bold", font_size=10.0, color=0x000000)
        assert classify_span(span, PROFILE_51) == SpanRole.BODY_BOLD


class TestClassify51ActionHeaders:
    """Action section headers in 5.1 monster stat blocks."""

    def test_azioni_11pt_bold(self):
        """'Azioni' is Calibri-Bold 11pt #000000 — classified as H6."""
        span = _span(text="Azioni", font_name="Calibri-Bold", font_size=11.0, color=0x000000)
        assert classify_span(span, PROFILE_51) == SpanRole.H6

    def test_reazioni_11pt_bold(self):
        span = _span(text="Reazioni", font_name="Calibri-Bold", font_size=11.0, color=0x000000)
        assert classify_span(span, PROFILE_51) == SpanRole.H6


class TestClassify51Other:
    """Links, footer, tables for SRD 5.1."""

    def test_link(self):
        span = _span(font_name="Calibri", font_size=9.0, color=0x0000FF)
        assert classify_span(span, PROFILE_51) == SpanRole.LINK

    def test_footer_8pt_italic(self):
        """Footer in 5.1 is Calibri-Italic 8pt #000000."""
        span = _span(font_name="Calibri-Italic", font_size=8.0, color=0x000000)
        assert classify_span(span, PROFILE_51) == SpanRole.FOOTER

    def test_table_header_bold_9pt(self):
        """Table headers in 5.1 are Calibri-Bold 9pt #000000."""
        span = _span(font_name="Calibri-Bold", font_size=9.0, color=0x000000)
        assert classify_span(span, PROFILE_51) == SpanRole.TABLE_HEADER_SMALL

    def test_table_body_9pt(self):
        """Table body in 5.1 is Calibri 9pt #000000."""
        span = _span(font_name="Calibri", font_size=9.0, color=0x000000)
        assert classify_span(span, PROFILE_51) == SpanRole.TABLE_BODY

    def test_monster_subtitle_italic(self):
        """Monster subtitle (Aberrazione Grande, legale malvagio) is Calibri-Italic 10pt."""
        span = _span(font_name="Calibri-Italic", font_size=10.0, color=0x000000)
        assert classify_span(span, PROFILE_51) == SpanRole.BODY_ITALIC


# ── Backward compatibility ──


class TestClassifyDefaultProfile:
    """classify_span without profile arg uses PROFILE_521 (backward compat)."""

    def test_default_profile_is_521(self):
        span = _span(font_name="GillSans-SemiBold", font_size=26.0, color=0x8C2220)
        assert classify_span(span) == SpanRole.H1

    def test_default_profile_body(self):
        span = _span(font_name="Cambria", font_size=10.0, color=0x231F20)
        assert classify_span(span) == SpanRole.BODY
