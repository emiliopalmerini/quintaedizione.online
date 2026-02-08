"""Backgrounds parser.

Structure in PDF (pages 93-97):
  H2: Background dei personaggi
    H4: Descrizioni dei background
      H5: Background name (e.g., "Accolito")
        TABLE_HEADER: "Punteggi di caratteristica: ..."
        TABLE_HEADER: "Talento: ..."
        TABLE_HEADER: "Competenze nelle abilità: ..."
        TABLE_HEADER: "Competenza negli strumenti: ..."
        TABLE_HEADER: "Equipaggiamento: ..."
        BODY paragraphs: flavor text

The parser skips species content (under "Specie dei personaggi" H2).
"""

from __future__ import annotations

from ..classify import SpanRole
from ..heading_tree import HeadingNode
from ..markdown_gen import paragraphs_to_markdown
from ..merge import Paragraph
from ..schemas import Background
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

_KNOWN_BACKGROUNDS = {
    "accolito", "criminale", "sapiente", "soldato",
}


def _is_background_node(node: HeadingNode) -> bool:
    """Check if a heading node is a background entry."""
    return node.title.lower().strip() in _KNOWN_BACKGROUNDS


def _extract_field(content: list[Paragraph], label: str) -> str:
    """Extract a metadata field from TABLE_HEADER paragraphs."""
    for para in content:
        text = para.text.strip()
        if text.lower().startswith(label.lower()):
            # Extract value after the label (with or without colon)
            rest = text[len(label):].strip()
            if rest.startswith(":"):
                rest = rest[1:].strip()
            return rest
    return ""


def _extract_background(name: str, content: list[Paragraph]) -> Background:
    """Extract a Background from its content paragraphs."""
    ability_scores = _extract_field(content, "Punteggi di caratteristica")
    feat = _extract_field(content, "Talento")
    skills = _extract_field(content, "Competenze nelle abilità")
    tool = _extract_field(content, "Competenza negli strumenti")
    equipment = _extract_field(content, "Equipaggiamento")

    # Body text is everything that's not a metadata field
    body: list[Paragraph] = []
    meta_labels = {
        "punteggi di caratteristica", "talento", "competenze nelle abilità",
        "competenza negli strumenti", "equipaggiamento",
    }
    for para in content:
        text = para.text.strip().lower()
        is_meta = any(text.startswith(label) for label in meta_labels)
        if not is_meta and para.role not in (SpanRole.TABLE_HEADER, SpanRole.TABLE_HEADER_SMALL):
            body.append(para)

    return Background(
        id=slugify(name),
        name=name,
        ability_scores=ability_scores,
        feat=feat,
        skill_proficiencies=skills,
        tool_proficiency=tool,
        equipment=equipment,
        description=paragraphs_to_markdown(body),
    )


def _collect_backgrounds(nodes: list[HeadingNode]) -> list[Background]:
    """Recursively find background entries in the heading tree."""
    results: list[Background] = []

    for node in nodes:
        # Only process under "Background dei personaggi" or "Descrizioni dei background"
        if node.title.lower().strip() in ("background dei personaggi", "descrizioni dei background"):
            results.extend(_collect_backgrounds(node.children))
            continue

        # Skip species sections entirely
        if node.title.lower().strip() in (
            "specie dei personaggi", "descrizioni delle specie",
            "elementi di una specie",
        ):
            continue

        # Skip intro/meta sections
        if node.title.lower().strip() in ("elementi di un background",):
            continue

        # H5 under "Descrizioni dei background" = individual background
        if node.level == 5 and _is_background_node(node):
            results.append(_extract_background(node.title, node.content))
            continue

        # Descend other nodes
        results.extend(_collect_backgrounds(node.children))

    return results


@register("backgrounds")
def parse_backgrounds(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[Background]:
    """Parse backgrounds from pages 93-97."""
    return _collect_backgrounds(tree)
