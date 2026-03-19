# /// script
# requires-python = ">=3.12"
# dependencies = ["pytest"]
# ///
"""Tests for magic item subtitle parsing."""

from __future__ import annotations

from ..parsers.magic_items import _parse_subtitle


class TestParseSubtitle:
    """Magic item subtitle parsing."""

    def test_standard(self):
        t, r, att, det = _parse_subtitle("Oggetto meraviglioso, raro")
        assert t == "Oggetto meraviglioso"
        assert r == "raro"
        assert att is False

    def test_with_attunement(self):
        t, r, att, det = _parse_subtitle(
            "Oggetto meraviglioso, non comune (richiede sintonia)"
        )
        assert t == "Oggetto meraviglioso"
        assert r == "non comune"
        assert att is True
        assert det == ""

    def test_with_attunement_details(self):
        t, r, att, det = _parse_subtitle(
            "Arma (qualsiasi spada), rara (richiede sintonia da un paladino)"
        )
        assert t == "Arma (qualsiasi spada)"
        assert r == "rara"
        assert att is True
        assert det == "da un paladino"

    def test_variable_rarity(self):
        """5.1 uses 'rarità variabile' for multi-variant items."""
        t, r, att, det = _parse_subtitle("Pergamena, rarità variabile")
        assert t == "Pergamena"
        assert r == "rarità variabile"

    def test_rarity_varia(self):
        """Some items use 'varia' as rarity."""
        t, r, att, det = _parse_subtitle("Pozione, varia")
        assert t == "Pozione"
        assert r == "varia"

    def test_empty_on_no_match(self):
        t, r, att, det = _parse_subtitle("Not a subtitle")
        assert t == ""
        assert r == ""


class TestOverrides:
    """Manual overrides for items where subtitle parsing fails."""

    def test_avatar_della_morte(self):
        """Avatar della morte is a stat block — override provides type/rarity."""
        from ..parsers.magic_items import _OVERRIDES
        override = _OVERRIDES.get("avatar della morte")
        assert override is not None
        item_type, rarity, attunement, _ = override
        assert item_type == "Creatura"
        assert rarity == "leggendario"
        assert attunement is False
