# /// script
# requires-python = ">=3.12"
# dependencies = ["pytest"]
# ///
"""Tests for classes parser across SRD versions."""

from __future__ import annotations

from ..classify import SpanRole
from ..heading_tree import HeadingNode as _RealHeadingNode
from ..merge import ClassifiedSpan, Paragraph
from ..parsers.classes import parse_classes
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
    return SectionDef("Classi", (8, 59), "classes.json", "classes")


class TestParseClasses521:
    """5.2.1: H1("Classi") > H2(ClassName) > H4(features/subclasses)."""

    def test_finds_barbaro(self):
        tree = [
            HeadingNode(
                title="Classi", level=1,
                content=[], children=[
                    HeadingNode(
                        title="Barbaro", level=2,
                        content=[_para("Dado Vita | D12 per ogni livello",
                                       SpanRole.SIDEBAR)],
                        children=[
                            HeadingNode(
                                title="Privilegi di classe del Barbaro", level=4,
                                content=[], children=[
                                    HeadingNode(
                                        title="Livello 1: Ira", level=5,
                                        content=[_para("Il barbaro entra in uno stato di furia.")],
                                        children=[],
                                    ),
                                ],
                            ),
                        ],
                    ),
                ],
            ),
        ]
        result = parse_classes(_section(), [], tree)
        assert len(result) == 1
        assert result[0]["name"] == "Barbaro"
        assert result[0]["hit_die"] == "d12"
        assert len(result[0]["features"]) == 1
        assert result[0]["features"][0]["name"] == "Ira"


class TestParseClasses51:
    """5.1: H1(ClassName) > H2("Privilegi di classe") > H3(features)."""

    def test_finds_barbaro_at_h1(self):
        """Class names are at H1 level in 5.1."""
        tree = [
            HeadingNode(
                title="Barbaro", level=1,
                content=[], children=[
                    HeadingNode(
                        title="Privilegi di classe", level=2,
                        content=[_para("Dado Vita | D12 per ogni livello",
                                       SpanRole.SIDEBAR)],
                        children=[
                            HeadingNode(
                                title="Punti ferita", level=5,
                                content=[_para("Dadi Vita: 1d12 per livello")],
                                children=[],
                            ),
                            HeadingNode(
                                title="Ira", level=3,
                                content=[_para("Il barbaro entra in furia.")],
                                children=[],
                            ),
                            HeadingNode(
                                title="Difesa Senza Armatura", level=3,
                                content=[_para("Finché non indossa armatura.")],
                                children=[],
                            ),
                        ],
                    ),
                ],
            ),
        ]
        result = parse_classes(_section(), [], tree)
        assert len(result) == 1
        assert result[0]["name"] == "Barbaro"
        assert result[0]["hit_die"] == "d12"

    def test_extracts_features_at_h3(self):
        """Features are H3 in 5.1, not H5."""
        tree = [
            HeadingNode(
                title="Barbaro", level=1,
                content=[], children=[
                    HeadingNode(
                        title="Privilegi di classe", level=2,
                        content=[],
                        children=[
                            HeadingNode(
                                title="Ira", level=3,
                                content=[_para("Il barbaro entra in furia.")],
                                children=[],
                            ),
                            HeadingNode(
                                title="Attacco Irruento", level=3,
                                content=[_para("L'avventatezza del barbaro.")],
                                children=[],
                            ),
                        ],
                    ),
                ],
            ),
        ]
        result = parse_classes(_section(), [], tree)
        assert len(result) == 1
        features = result[0]["features"]
        assert len(features) >= 2
        names = [f["name"] for f in features]
        assert "Ira" in names
        assert "Attacco Irruento" in names

    def test_finds_subclass_under_separate_h2(self):
        """Subclasses in 5.1 are under a separate H2 (e.g. 'Archetipi Marziali')."""
        tree = [
            HeadingNode(
                title="Guerriero", level=1,
                content=[], children=[
                    HeadingNode(
                        title="Privilegi di classe", level=2,
                        content=[],
                        children=[
                            HeadingNode(
                                title="Archetipo Marziale", level=3,
                                content=[_para("A livello 3, scegli un archetipo.")],
                                children=[],
                            ),
                        ],
                    ),
                    HeadingNode(
                        title="Archetipi Marziali", level=2,
                        content=[_para("Intro text")],
                        children=[
                            HeadingNode(
                                title="Campione", level=3,
                                content=[_para("Il campione si concentra.")],
                                children=[
                                    HeadingNode(
                                        title="Critico Migliorato", level=5,
                                        content=[_para("I tuoi attacchi criccano con 19.")],
                                        children=[],
                                    ),
                                ],
                            ),
                        ],
                    ),
                ],
            ),
        ]
        result = parse_classes(_section(), [], tree)
        assert len(result) == 1
        assert result[0]["name"] == "Guerriero"
        subclasses = result[0]["subclasses"]
        assert len(subclasses) == 1
        assert subclasses[0]["name"] == "Campione"

    def test_multiple_classes(self):
        """Multiple H1 class nodes produce multiple classes."""
        tree = [
            HeadingNode(
                title="Barbaro", level=1,
                content=[], children=[
                    HeadingNode(title="Privilegi di classe", level=2,
                                content=[], children=[]),
                ],
            ),
            HeadingNode(
                title="Bardo", level=1,
                content=[], children=[
                    HeadingNode(title="Privilegi di classe", level=2,
                                content=[], children=[]),
                ],
            ),
        ]
        result = parse_classes(_section(), [], tree)
        assert len(result) == 2
        names = [c["name"] for c in result]
        assert "Barbaro" in names
        assert "Bardo" in names
