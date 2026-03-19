# /// script
# requires-python = ">=3.12"
# dependencies = ["pytest"]
# ///
"""Tests for equipment parser — 5.1 sequential table parsing."""

from __future__ import annotations

from ..classify import SpanRole
from ..heading_tree import HeadingNode as _RealHeadingNode
from ..merge import ClassifiedSpan, Paragraph
from ..parsers.equipment import parse_equipment
from ..section_split import SectionDef

_LEVEL_TO_ROLE = {
    1: SpanRole.H1, 2: SpanRole.H2, 3: SpanRole.H3,
    4: SpanRole.H4, 5: SpanRole.H5, 6: SpanRole.H6,
}


def HeadingNode(title: str, level: int, content: list, children: list) -> _RealHeadingNode:
    return _RealHeadingNode(
        level=level, title=title, role=_LEVEL_TO_ROLE[level],
        content=content, children=children,
    )


def _para(text: str, role: SpanRole = SpanRole.BODY) -> Paragraph:
    return Paragraph(
        spans=[ClassifiedSpan(text=text, role=role)],
        role=role,
        page_num=1,
    )


def _section() -> SectionDef:
    return SectionDef("Equipaggiamento", (68, 84), "equipment.json", "equipment")


class TestParseArmorFromSequentialParagraphs:
    """5.1: armor data as sequential TABLE_HEADER_SMALL + TABLE_BODY paragraphs."""

    def test_extracts_armor_items(self):
        content = [
            _para("Descrizione armature.", SpanRole.BODY),
            # Table title
            _para("Armature", SpanRole.UNKNOWN),
            # Column headers
            _para("Armatura", SpanRole.TABLE_HEADER_SMALL),
            _para("Costo", SpanRole.TABLE_HEADER_SMALL),
            _para("Classe Armatura (CA)", SpanRole.TABLE_HEADER_SMALL),
            _para("Forza", SpanRole.TABLE_HEADER_SMALL),
            _para("Furtività", SpanRole.TABLE_HEADER_SMALL),
            _para("Peso", SpanRole.TABLE_HEADER_SMALL),
            # Subcategory
            _para("Armature leggere", SpanRole.UNKNOWN),
            # Row 1: Imbottita
            _para("Imbottita", SpanRole.TABLE_BODY),
            _para("5 mo", SpanRole.TABLE_BODY),
            _para("11 + modificatore di Des", SpanRole.TABLE_BODY),
            _para("—", SpanRole.TABLE_BODY),
            _para("Svantaggio", SpanRole.TABLE_BODY),
            _para("4 kg", SpanRole.TABLE_BODY),
            # Row 2: Cuoio
            _para("Cuoio", SpanRole.TABLE_BODY),
            _para("10 mo", SpanRole.TABLE_BODY),
            _para("11 + modificatore di Des", SpanRole.TABLE_BODY),
            _para("—", SpanRole.TABLE_BODY),
            _para("—", SpanRole.TABLE_BODY),
            _para("5 kg", SpanRole.TABLE_BODY),
        ]
        tree = [
            HeadingNode(title="Equipaggiamento", level=1, content=[], children=[
                HeadingNode(title="Armature", level=2, content=[], children=[
                    HeadingNode(title="Armature pesanti", level=3,
                                content=content, children=[]),
                ]),
            ]),
        ]
        result = parse_equipment(_section(), [], tree)
        armor = [i for i in result if i["category"] == "armor"]
        assert len(armor) >= 2
        names = [i["name"] for i in armor]
        assert "Imbottita" in names
        assert "Cuoio" in names

        imbottita = next(i for i in armor if i["name"] == "Imbottita")
        assert imbottita["subcategory"] == "Armature leggere"
        assert "11 + modificatore di Des" in imbottita["properties"].get("classe_armatura", "")


class TestParseWeaponsFromSequentialParagraphs:
    """5.1: weapon data as sequential TABLE_HEADER_SMALL + TABLE_BODY paragraphs."""

    def test_extracts_weapon_items(self):
        content = [
            # Column headers
            _para("Nome", SpanRole.TABLE_HEADER_SMALL),
            _para("Costo", SpanRole.TABLE_HEADER_SMALL),
            _para("Danni", SpanRole.TABLE_HEADER_SMALL),
            _para("Peso", SpanRole.TABLE_HEADER_SMALL),
            _para("Proprietà", SpanRole.TABLE_HEADER_SMALL),
            # Subcategory
            _para("Armi semplici da mischia", SpanRole.UNKNOWN),
            # Row 1
            _para("Bastone ferrato", SpanRole.TABLE_BODY),
            _para("2 ma", SpanRole.TABLE_BODY),
            _para("1d6 contundenti", SpanRole.TABLE_BODY),
            _para("2 kg", SpanRole.TABLE_BODY),
            _para("Versatile (1d8)", SpanRole.TABLE_BODY),
        ]
        tree = [
            HeadingNode(title="Equipaggiamento", level=1, content=[], children=[
                HeadingNode(title="Armi", level=2, content=[], children=[
                    HeadingNode(title="Competenza nelle armi", level=3,
                                content=content, children=[]),
                ]),
            ]),
        ]
        result = parse_equipment(_section(), [], tree)
        weapons = [i for i in result if i["category"] == "weapons"]
        assert len(weapons) >= 1
        assert weapons[0]["name"] == "Bastone ferrato"
        assert weapons[0]["properties"].get("danni") == "1d6 contundenti"
