# /// script
# requires-python = ">=3.12"
# dependencies = ["pytest"]
# ///
"""Tests for spell metadata field extraction."""

from __future__ import annotations

from ..classify import SpanRole
from ..merge import ClassifiedSpan, Paragraph
from ..parsers.spells import _extract_field


def _para(text: str, role: SpanRole = SpanRole.BODY_ITALIC) -> Paragraph:
    """Create a single-span paragraph."""
    return Paragraph(
        spans=[ClassifiedSpan(text=text, role=role)],
        role=role,
        page_num=1,
    )


class TestExtractFieldSeparateParagraphs:
    """5.2.1 style: each field is its own TABLE_HEADER_SMALL paragraph."""

    def test_casting_time(self):
        paras = [
            _para("Tempo di lancio: 1 azione", SpanRole.TABLE_HEADER_SMALL),
            _para("Gittata: 9 metri", SpanRole.TABLE_HEADER_SMALL),
        ]
        assert _extract_field(paras, "Tempo di lancio") == "1 azione"

    def test_duration(self):
        paras = [
            _para("Durata: Concentrazione, fino a 1 minuto", SpanRole.TABLE_HEADER_SMALL),
        ]
        assert _extract_field(paras, "Durata") == "Concentrazione, fino a 1 minuto"


class TestExtractFieldMergedParagraph:
    """5.1 style: all metadata merged into one BODY_ITALIC paragraph."""

    def test_casting_time_from_merged(self):
        paras = [
            _para(
                "Abiurazione di 2° livello "
                "Tempo di lancio: 1 azione "
                "Gittata: 9 metri "
                "Componenti: V, S, M (tessuto) "
                "Durata: 8 ore "
                "Questo incantesimo rafforza il vigore."
            ),
        ]
        assert _extract_field(paras, "Tempo di lancio") == "1 azione"

    def test_gittata_from_merged(self):
        paras = [
            _para(
                "Abiurazione di 2° livello "
                "Tempo di lancio: 1 azione "
                "Gittata: 9 metri "
                "Componenti: V, S "
                "Durata: 8 ore "
                "Testo descrizione."
            ),
        ]
        assert _extract_field(paras, "Gittata") == "9 metri"

    def test_components_from_merged(self):
        paras = [
            _para(
                "Ammaliamento di 1° livello "
                "Tempo di lancio: 1 azione "
                "Gittata: 18 metri "
                "Componenti: V "
                "Durata: Istantanea "
                "L'incantatore insulta la creatura."
            ),
        ]
        assert _extract_field(paras, "Componenti") == "V"

    def test_duration_no_bleed(self):
        """Duration must NOT bleed into the description body."""
        paras = [
            _para(
                "Abiurazione di 2° livello "
                "Tempo di lancio: 1 azione "
                "Gittata: 9 metri "
                "Componenti: V, S, M (una minuscola striscia di tessuto bianco) "
                "Durata: 8 ore "
                "Questo incantesimo rafforza il vigore e la determinazione degli alleati."
            ),
        ]
        result = _extract_field(paras, "Durata")
        assert result == "8 ore"

    def test_duration_concentration(self):
        paras = [
            _para(
                "Illusione di 4° livello "
                "Tempo di lancio: 1 azione "
                "Gittata: 36 metri "
                "Componenti: V, S "
                "Durata: Concentrazione, fino a 1 minuto "
                "L'incantatore attinge agli incubi di una creatura."
            ),
        ]
        result = _extract_field(paras, "Durata")
        assert result == "Concentrazione, fino a 1 minuto"

    def test_duration_instantaneous(self):
        paras = [
            _para(
                "Invocazione di 3° livello "
                "Tempo di lancio: 1 azione "
                "Gittata: 45 metri "
                "Componenti: V, S, M (guano di pipistrello e zolfo) "
                "Durata: Istantanea "
                "Una scia di luce brillante parte dall'indice."
            ),
        ]
        result = _extract_field(paras, "Durata")
        assert result == "Istantanea"

    def test_duration_with_double_space(self):
        """5.1 PDF has double spaces between duration value and description."""
        paras = [
            _para(
                "Abiurazione di 2° livello "
                "Tempo di lancio: 1 azione "
                "Gittata: 9 metri "
                "Componenti: V, S "
                "Durata: 8 ore  Questo incantesimo rafforza."
            ),
        ]
        result = _extract_field(paras, "Durata")
        assert result == "8 ore"

    def test_duration_period_before_description(self):
        """Duration ends with period, then description starts with L'."""
        paras = [
            _para(
                "Trasmutazione di 2° livello "
                "Tempo di lancio: 1 azione "
                "Gittata: Contatto "
                "Componenti: V, S, M "
                "Durata: Concentrazione, fino a 1 ora. "
                "L'incantatore tocca una creatura e le concede un potenziamento."
            ),
        ]
        result = _extract_field(paras, "Durata")
        assert result == "Concentrazione, fino a 1 ora"
