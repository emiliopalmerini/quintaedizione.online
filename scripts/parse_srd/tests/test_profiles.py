# /// script
# requires-python = ">=3.12"
# dependencies = ["pytest"]
# ///
"""Tests for FontProfile dataclass and profile definitions."""

from __future__ import annotations

import pytest

from ..profiles import FontProfile, PROFILE_521, PROFILE_51


class TestFontProfile:
    """FontProfile is a frozen dataclass with all required fields."""

    def test_profile_is_frozen(self):
        with pytest.raises(AttributeError):
            PROFILE_521.name = "oops"

    def test_profile_521_has_stat_font(self):
        """5.2.1 uses distinct Optima font for stat blocks."""
        assert PROFILE_521.stat_font == "Optima"
        assert PROFILE_521.stat_color == 0x540000

    def test_profile_51_has_no_stat_font(self):
        """5.1 uses body fonts for stat blocks — no distinct font/color."""
        assert PROFILE_51.stat_font is None
        assert PROFILE_51.stat_color is None

    def test_profile_521_colors(self):
        assert PROFILE_521.heading_color == 0x8C2220
        assert PROFILE_521.body_color == 0x231F20
        assert PROFILE_521.link_color == 0x1E5E9E

    def test_profile_51_colors(self):
        assert PROFILE_51.heading_color == 0x943634
        assert PROFILE_51.body_color == 0x000000
        assert PROFILE_51.link_color == 0x0000FF

    def test_profile_521_fonts(self):
        assert PROFILE_521.heading_font == "GillSans"
        assert PROFILE_521.body_font == "Cambria"
        assert PROFILE_521.sidebar_font == "GillSans"

    def test_profile_51_fonts(self):
        assert PROFILE_51.heading_font == "Calibri"
        assert PROFILE_51.body_font == "Cambria"
        assert PROFILE_51.sidebar_font == "Calibri"

    def test_profile_521_footer(self):
        assert PROFILE_521.footer_color == 0x808285

    def test_profile_51_footer(self):
        """5.1 footer uses body color — footer_color is None."""
        assert PROFILE_51.footer_color is None
