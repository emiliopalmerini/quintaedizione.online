# /// script
# requires-python = ">=3.12"
# dependencies = ["pytest"]
# ///
"""Tests for monster parser — 5.1 body-font stat blocks."""

from __future__ import annotations

from ..classify import SpanRole
from ..heading_tree import HeadingNode as _RealHeadingNode
from ..merge import ClassifiedSpan, Paragraph
from ..parsers.monsters import parse_monsters
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


def _span(text: str, role: SpanRole) -> ClassifiedSpan:
    return ClassifiedSpan(text=text, role=role)


def _para(text: str, role: SpanRole = SpanRole.BODY) -> Paragraph:
    return Paragraph(
        spans=[ClassifiedSpan(text=text, role=role)],
        role=role,
        page_num=1,
    )


def _multi_para(spans: list[ClassifiedSpan], role: SpanRole | None = None) -> Paragraph:
    """Paragraph with multiple spans; role defaults to first span's role."""
    return Paragraph(
        spans=spans,
        role=role or spans[0].role,
        page_num=1,
    )


def _section() -> SectionDef:
    return SectionDef("Mostri", (298, 410), "monsters.json", "monsters")


def _aboleth_content() -> list[Paragraph]:
    """Minimal Aboleth stat block in 5.1 format."""
    return [
        _para("Aberrazione Grande, legale malvagio", SpanRole.BODY_ITALIC),
        _multi_para([
            _span("Classe Armatura", SpanRole.BODY_BOLD),
            _span(" 17 (armatura naturale)", SpanRole.SIDEBAR),
        ], SpanRole.BODY_BOLD),
        _multi_para([
            _span("Punti Ferita", SpanRole.BODY_BOLD),
            _span(" 135 (18d10 + 36)", SpanRole.SIDEBAR),
        ], SpanRole.BODY_BOLD),
        _multi_para([
            _span("Velocità", SpanRole.BODY_BOLD),
            _span(" 3 m, nuotare 12 m", SpanRole.SIDEBAR),
        ], SpanRole.BODY_BOLD),
        # Ability scores
        _para("FOR", SpanRole.BODY_BOLD),
        _para("DES", SpanRole.BODY_BOLD),
        _para("COS", SpanRole.BODY_BOLD),
        _para("INT", SpanRole.BODY_BOLD),
        _para("SAG", SpanRole.BODY_BOLD),
        _para("CAR", SpanRole.BODY_BOLD),
        _para("21 (+5)", SpanRole.SIDEBAR),
        _para("9 (-1)", SpanRole.SIDEBAR),
        _para("15 (+2) 18 (+4) 15 (+2) 18 (+4)", SpanRole.SIDEBAR),
        # More stat fields
        _multi_para([
            _span("Tiri Salvezza", SpanRole.BODY_BOLD),
            _span(" Cos +6, Int +8, Sag +6", SpanRole.SIDEBAR),
        ], SpanRole.BODY_BOLD),
        _multi_para([
            _span("Sensi", SpanRole.BODY_BOLD),
            _span(" Percezione passiva 20, scurovisione 36 m", SpanRole.SIDEBAR),
        ], SpanRole.BODY_BOLD),
        _multi_para([
            _span("Linguaggi", SpanRole.BODY_BOLD),
            _span(" Gergo delle profondità, telepatia 36 m", SpanRole.SIDEBAR),
        ], SpanRole.BODY_BOLD),
        _multi_para([
            _span("Sfida", SpanRole.BODY_BOLD),
            _span(" 10 (5.900 PE)", SpanRole.SIDEBAR),
        ], SpanRole.BODY_BOLD),
        # Traits
        _multi_para([
            _span("Anfibio.", SpanRole.BODY_BOLD_ITALIC),
            _span(" L'aboleth può respirare in aria e in acqua.", SpanRole.SIDEBAR),
        ], SpanRole.BODY_BOLD_ITALIC),
    ]


class TestParseMonsters51:
    """5.1: H2(group) > H5(monster) with body-font stat blocks."""

    def test_finds_monster_at_h5(self):
        tree = [
            HeadingNode(
                title="Mostri (A)", level=2, content=[], children=[
                    HeadingNode(
                        title="Aboleth", level=5,
                        content=_aboleth_content(),
                        children=[
                            HeadingNode(
                                title="Azioni", level=6,
                                content=[
                                    _multi_para([
                                        _span("Multiattacco.", SpanRole.BODY_BOLD_ITALIC),
                                        _span(" L'aboleth effettua tre attacchi.", SpanRole.SIDEBAR),
                                    ], SpanRole.BODY_BOLD_ITALIC),
                                ],
                                children=[],
                            ),
                        ],
                    ),
                ],
            ),
        ]
        result = parse_monsters(_section(), [], tree)
        assert len(result) == 1
        m = result[0]
        assert m["name"] == "Aboleth"
        assert m["type"] == "Aberrazione"
        assert m["size"] == "Grande"
        assert "legale malvagio" in m["alignment"]
        assert "17" in m["ac"]
        assert "135" in m["hp"]
        assert m["cr"] == "10"
        assert len(m["actions"]) >= 1

    def test_grouped_monsters(self):
        """H3 group > H5 monsters."""
        tree = [
            HeadingNode(
                title="Mostri (A)", level=2, content=[], children=[
                    HeadingNode(
                        title="Angeli", level=3, content=[], children=[
                            HeadingNode(
                                title="Deva", level=5,
                                content=[
                                    _para("Celestiale Medio, legale buono", SpanRole.BODY_ITALIC),
                                    _multi_para([
                                        _span("Classe Armatura", SpanRole.BODY_BOLD),
                                        _span(" 17 (armatura naturale)", SpanRole.SIDEBAR),
                                    ], SpanRole.BODY_BOLD),
                                    _multi_para([
                                        _span("Punti Ferita", SpanRole.BODY_BOLD),
                                        _span(" 136 (16d8 + 64)", SpanRole.SIDEBAR),
                                    ], SpanRole.BODY_BOLD),
                                    _multi_para([
                                        _span("Velocità", SpanRole.BODY_BOLD),
                                        _span(" 9 m, volare 27 m", SpanRole.SIDEBAR),
                                    ], SpanRole.BODY_BOLD),
                                    _multi_para([
                                        _span("Sfida", SpanRole.BODY_BOLD),
                                        _span(" 10 (5.900 PE)", SpanRole.SIDEBAR),
                                    ], SpanRole.BODY_BOLD),
                                ],
                                children=[],
                            ),
                        ],
                    ),
                ],
            ),
        ]
        result = parse_monsters(_section(), [], tree)
        assert len(result) == 1
        assert result[0]["name"] == "Deva"
        assert result[0]["group"] == "Angeli"
