# /// script
# requires-python = ">=3.12"
# dependencies = ["pytest"]
# ///
"""Tests for 5.1 subclass detection inside 'Privilegi di classe'."""

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


class TestInlineSubclass51:
    """5.1 subclasses nested inside 'Privilegi di classe' as H3."""

    def test_barbaro_berserker(self):
        """Cammino del Berserker is a subclass, not a feature."""
        tree = [
            HeadingNode(
                title="Barbaro", level=1, content=[], children=[
                    HeadingNode(
                        title="Privilegi di classe", level=2, content=[], children=[
                            HeadingNode(
                                title="Ira", level=3,
                                content=[_para("Furia")], children=[],
                            ),
                            HeadingNode(
                                title="Cammino Primordiale", level=3,
                                content=[_para("A livello 3, scegli un cammino.")], children=[],
                            ),
                            HeadingNode(
                                title="Cammino del Berserker", level=3,
                                content=[_para("Il berserker entra in frenesia.")],
                                children=[
                                    HeadingNode(
                                        title="Frenesia", level=5,
                                        content=[_para("Quando entri in ira...")], children=[],
                                    ),
                                    HeadingNode(
                                        title="Ira Incontenibile", level=5,
                                        content=[_para("A partire dal 6° livello...")], children=[],
                                    ),
                                ],
                            ),
                        ],
                    ),
                ],
            ),
        ]
        result = parse_classes(_section(), [], tree)
        barbaro = result[0]
        assert barbaro["name"] == "Barbaro"
        # Berserker should be a subclass, NOT a feature
        feature_names = [f["name"] for f in barbaro["features"]]
        assert "Cammino del Berserker" not in feature_names
        assert "Ira" in feature_names
        # Should have 1 subclass
        assert len(barbaro["subclasses"]) == 1
        assert barbaro["subclasses"][0]["name"] == "Cammino del Berserker"
        # Subclass should have sub-features
        sc_features = barbaro["subclasses"][0]["features"]
        sc_feat_names = [f["name"] for f in sc_features]
        assert "Frenesia" in sc_feat_names

    def test_chierico_dominio_vita(self):
        """Dominio della Vita is a subclass."""
        tree = [
            HeadingNode(
                title="Chierico", level=1, content=[], children=[
                    HeadingNode(
                        title="Privilegi di classe", level=2, content=[], children=[
                            HeadingNode(
                                title="Incantesimi", level=3,
                                content=[_para("Il chierico è un incantatore.")],
                                children=[
                                    HeadingNode(title="Trucchetti", level=5,
                                                content=[_para("Conosci 3 trucchetti.")],
                                                children=[]),
                                ],
                            ),
                            HeadingNode(
                                title="Dominio della Vita", level=3,
                                content=[_para("Il dominio della Vita si concentra.")],
                                children=[
                                    HeadingNode(title="Competenza Bonus", level=5,
                                                content=[_para("Competenza armature pesanti.")],
                                                children=[]),
                                    HeadingNode(title="Discepolo della Vita", level=5,
                                                content=[_para("Le cure sono potenziate.")],
                                                children=[]),
                                ],
                            ),
                        ],
                    ),
                ],
            ),
        ]
        result = parse_classes(_section(), [], tree)
        chierico = result[0]
        feature_names = [f["name"] for f in chierico["features"]]
        assert "Dominio della Vita" not in feature_names
        assert "Incantesimi" in feature_names
        assert len(chierico["subclasses"]) == 1
        assert chierico["subclasses"][0]["name"] == "Dominio della Vita"

    def test_incantesimi_not_a_subclass(self):
        """'Incantesimi' has H5 children but is NOT a subclass."""
        tree = [
            HeadingNode(
                title="Bardo", level=1, content=[], children=[
                    HeadingNode(
                        title="Privilegi di classe", level=2, content=[], children=[
                            HeadingNode(
                                title="Incantesimi", level=3,
                                content=[_para("Il bardo è un incantatore.")],
                                children=[
                                    HeadingNode(title="Trucchetti", level=5,
                                                content=[_para("Conosci 2 trucchetti.")],
                                                children=[]),
                                ],
                            ),
                            HeadingNode(
                                title="Collegio della Sapienza", level=3,
                                content=[_para("Il collegio della sapienza raccoglie...")],
                                children=[
                                    HeadingNode(title="Competenze bonus", level=5,
                                                content=[_para("Tre competenze a scelta.")],
                                                children=[]),
                                ],
                            ),
                        ],
                    ),
                ],
            ),
        ]
        result = parse_classes(_section(), [], tree)
        bardo = result[0]
        feature_names = [f["name"] for f in bardo["features"]]
        assert "Incantesimi" in feature_names
        assert len(bardo["subclasses"]) == 1
        assert bardo["subclasses"][0]["name"] == "Collegio della Sapienza"
