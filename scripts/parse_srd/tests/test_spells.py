# /// script
# requires-python = ">=3.12"
# dependencies = ["pytest"]
# ///
"""Tests for spell subtitle parsing across SRD versions."""

from __future__ import annotations

from ..parsers.spells import _parse_subtitle


class TestParseSubtitle521:
    """SRD 5.2.1 subtitle format: includes class list in parentheses."""

    def test_cantrip(self):
        level, school, classes = _parse_subtitle(
            "Trucchetto di Ammaliamento (Bardo, Stregone, Warlock)"
        )
        assert level == 0
        assert school == "Ammaliamento"
        assert classes == ["Bardo", "Stregone", "Warlock"]

    def test_leveled_with_ordinal_indicator(self):
        """5.2.1 uses º (U+00BA, masculine ordinal indicator)."""
        level, school, classes = _parse_subtitle(
            "Invocazione di 3º livello (Mago, Stregone)"
        )
        assert level == 3
        assert school == "Invocazione"
        assert classes == ["Mago", "Stregone"]


class TestParseSubtitle51:
    """SRD 5.1 subtitle format: no class list, uses ° (degree sign)."""

    def test_cantrip_no_classes(self):
        level, school, classes = _parse_subtitle("Trucchetto di Trasmutazione")
        assert level == 0
        assert school == "Trasmutazione"
        assert classes == []

    def test_leveled_degree_sign(self):
        """5.1 uses ° (U+00B0, degree sign) instead of º."""
        level, school, classes = _parse_subtitle("Illusione di 4° livello")
        assert level == 4
        assert school == "Illusione"
        assert classes == []

    def test_leveled_1st(self):
        level, school, classes = _parse_subtitle("Ammaliamento di 1° livello")
        assert level == 1
        assert school == "Ammaliamento"
        assert classes == []

    def test_leveled_with_ritual(self):
        level, school, classes = _parse_subtitle(
            "Abiurazione di 1° livello (rituale)"
        )
        assert level == 1
        assert school == "Abiurazione"
        assert classes == []

    def test_leveled_9th(self):
        level, school, classes = _parse_subtitle("Evocazione di 9° livello")
        assert level == 9
        assert school == "Evocazione"
        assert classes == []

    def test_subtitle_with_trailing_metadata(self):
        """In 5.1, subtitle + metadata can merge into one paragraph."""
        level, school, classes = _parse_subtitle(
            "Ammaliamento di 1° livello Tempo di lancio: 1 azione Gittata: 9 metri"
        )
        assert level == 1
        assert school == "Ammaliamento"

    def test_cantrip_with_trailing_metadata(self):
        level, school, classes = _parse_subtitle(
            "Trucchetto di Invocazione Tempo di lancio: 1 azione"
        )
        assert level == 0
        assert school == "Invocazione"

    def test_unknown_returns_negative(self):
        level, school, classes = _parse_subtitle("Not a spell subtitle")
        assert level == -1
        assert school == ""
        assert classes == []


class TestSpellOverrides:
    """Manual overrides for column-break artifacts."""

    def test_creazione_duration(self):
        from ..parsers.spells import _SPELL_OVERRIDES
        assert _SPELL_OVERRIDES["creazione"]["duration"] == "Speciale"

    def test_dominare_persone_school(self):
        from ..parsers.spells import _SPELL_OVERRIDES
        assert _SPELL_OVERRIDES["dominare persone"]["school"] == "Ammaliamento"
