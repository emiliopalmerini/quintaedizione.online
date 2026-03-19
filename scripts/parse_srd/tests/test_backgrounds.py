# /// script
# requires-python = ">=3.12"
# dependencies = ["pytest"]
# ///
"""Tests for backgrounds parser across SRD versions."""

from __future__ import annotations

from ..classify import SpanRole
from ..heading_tree import HeadingNode as _RealHeadingNode
from ..merge import ClassifiedSpan, Paragraph
from ..parsers.backgrounds import parse_backgrounds
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
    return SectionDef("Backgrounds", (65, 67), "backgrounds.json", "backgrounds")


class TestParseBackgrounds521:
    """5.2.1: H2 > H4("Descrizioni") > H5(BackgroundName)."""

    def test_finds_accolito(self):
        tree = [
            HeadingNode(
                title="Background dei personaggi", level=2,
                content=[], children=[
                    HeadingNode(
                        title="Descrizioni dei background", level=4,
                        content=[], children=[
                            HeadingNode(
                                title="Accolito", level=5,
                                content=[
                                    _para("Punteggi di caratteristica: +1 Sag, +1 Int, +1 Car",
                                          SpanRole.TABLE_HEADER),
                                    _para("Talento: Iniziato alla Magia (Divina)",
                                          SpanRole.TABLE_HEADER),
                                    _para("Un accolito ha dedicato la propria vita al servizio."),
                                ],
                                children=[],
                            ),
                        ],
                    ),
                ],
            ),
        ]
        result = parse_backgrounds(_section(), [], tree)
        assert len(result) == 1
        assert result[0]["name"] == "Accolito"


class TestParseBackgrounds51:
    """5.1: H2("Background") > H3(BackgroundName)."""

    def test_finds_accolito_at_h3(self):
        tree = [
            HeadingNode(
                title="Background", level=2,
                content=[_para("Ogni storia ha un inizio.")],
                children=[
                    HeadingNode(
                        title="Competenze", level=5,
                        content=[_para("Le competenze del background.")],
                        children=[],
                    ),
                    HeadingNode(
                        title="Accolito", level=3,
                        content=[
                            _para("Competenze nelle abilità: Intuizione, Religione",
                                  SpanRole.BODY_BOLD),
                            _para("Linguaggi: Due a scelta",
                                  SpanRole.BODY_BOLD),
                            _para("Equipaggiamento: Un simbolo sacro",
                                  SpanRole.BODY_BOLD),
                            _para("Un accolito ha dedicato la propria vita."),
                        ],
                        children=[
                            HeadingNode(
                                title="Privilegio: Rifugio dei Fedeli", level=5,
                                content=[_para("L'accolito può trovare rifugio.")],
                                children=[],
                            ),
                        ],
                    ),
                ],
            ),
        ]
        result = parse_backgrounds(_section(), [], tree)
        assert len(result) == 1
        assert result[0]["name"] == "Accolito"
        assert "Rifugio dei Fedeli" in result[0]["description"]

    def test_extracts_skills_from_body_bold(self):
        """5.1 uses BODY_BOLD for metadata, not TABLE_HEADER."""
        tree = [
            HeadingNode(
                title="Background", level=2,
                content=[],
                children=[
                    HeadingNode(
                        title="Accolito", level=3,
                        content=[
                            _para("Competenze nelle abilità: Intuizione, Religione",
                                  SpanRole.BODY_BOLD),
                            _para("Equipaggiamento: Un simbolo sacro",
                                  SpanRole.BODY_BOLD),
                            _para("Un accolito ha dedicato la vita."),
                        ],
                        children=[],
                    ),
                ],
            ),
        ]
        result = parse_backgrounds(_section(), [], tree)
        assert result[0]["skill_proficiencies"] == "Intuizione, Religione"
        assert result[0]["equipment"] == "Un simbolo sacro"
