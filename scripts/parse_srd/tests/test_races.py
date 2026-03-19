# /// script
# requires-python = ">=3.12"
# dependencies = ["pytest"]
# ///
"""Tests for races parser (5.1 → species.json)."""

from __future__ import annotations

from ..classify import SpanRole
from ..heading_tree import HeadingNode as _RealHeadingNode
from ..merge import ClassifiedSpan, Paragraph
from ..parsers.races import parse_races
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
    return SectionDef("Razze", (2, 7), "species.json", "races")


class TestParseRaces:
    """5.1: H1(Razze) > H2(RaceName) > H3(Tratti) with traits as bold-italic."""

    def test_finds_race(self):
        tree = [
            HeadingNode(title="Razze", level=1, content=[], children=[
                HeadingNode(title="Elfo", level=2, content=[], children=[
                    HeadingNode(
                        title="Tratti degli elfi", level=3,
                        content=[
                            _para("_**Incremento dei punteggi di caratteristica.**_ "
                                  "Il punteggio di Destrezza aumenta di 2."),
                            _para("_**Velocità.**_ La velocità base è di 9 metri."),
                        ],
                        children=[],
                    ),
                ]),
            ]),
        ]
        result = parse_races(_section(), [], tree)
        assert len(result) == 1
        assert result[0]["name"] == "Elfo"
        assert len(result[0]["traits"]) >= 2

    def test_extracts_subrace(self):
        """Subraces are H5 children of the traits H3."""
        tree = [
            HeadingNode(title="Razze", level=1, content=[], children=[
                HeadingNode(title="Elfo", level=2, content=[], children=[
                    HeadingNode(
                        title="Tratti degli elfi", level=3,
                        content=[
                            _para("_**Velocità.**_ 9 metri."),
                        ],
                        children=[
                            HeadingNode(
                                title="Elfo alto", level=5,
                                content=[
                                    _para("_**Incremento dei punteggi.**_ Int +1."),
                                    _para("_**Trucchetto.**_ Conosci un trucchetto."),
                                ],
                                children=[],
                            ),
                        ],
                    ),
                ]),
            ]),
        ]
        result = parse_races(_section(), [], tree)
        assert len(result) == 1
        # Subrace traits should be included in description
        desc = result[0]["description"]
        assert "Elfo alto" in desc

    def test_multiple_races(self):
        tree = [
            HeadingNode(title="Razze", level=1, content=[], children=[
                HeadingNode(title="Elfo", level=2, content=[], children=[
                    HeadingNode(title="Tratti degli elfi", level=3,
                                content=[_para("Tratti.")], children=[]),
                ]),
                HeadingNode(title="Nano", level=2, content=[], children=[
                    HeadingNode(title="Tratti dei nani", level=3,
                                content=[_para("Tratti.")], children=[]),
                ]),
                HeadingNode(title="Umano", level=2, content=[], children=[
                    HeadingNode(title="Tratti degli umani", level=3,
                                content=[_para("Tratti.")], children=[]),
                ]),
            ]),
        ]
        result = parse_races(_section(), [], tree)
        assert len(result) == 3
        names = [r["name"] for r in result]
        assert "Elfo" in names
        assert "Nano" in names
        assert "Umano" in names
