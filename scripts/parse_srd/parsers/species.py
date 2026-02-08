"""Species (Specie) parser.

Structure in PDF (pages 93-97, shared with backgrounds):
  H2: Specie dei personaggi
  H4: Descrizioni delle specie
    H5: Species name (e.g., "Dragonide")
      TABLE_HEADER: "Tipo di creatura: umanoide"
      TABLE_HEADER: "Taglia: ..."
      TABLE_HEADER: "Velocità: ..."
      BODY paragraphs: trait descriptions with _**TraitName.**_ pattern
"""

from __future__ import annotations

import re

from ..classify import SpanRole
from ..heading_tree import HeadingNode
from ..markdown_gen import paragraphs_to_markdown
from ..merge import Paragraph
from ..schemas import Species, SpeciesTrait
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

_KNOWN_SPECIES = {
    "dragonide", "elfo", "gnomo", "goliath", "halfling",
    "nano", "orco", "tiefling", "umano",
}

# Matches trait names in bold-italic pattern: _**TraitName.**_
_TRAIT_NAME_RE = re.compile(r"_\*\*(.+?)\.\*\*_")


def _extract_field(content: list[Paragraph], label: str) -> str:
    """Extract a metadata field from TABLE_HEADER paragraphs."""
    for para in content:
        text = para.text.strip()
        if text.lower().startswith(label.lower()):
            rest = text[len(label):].strip()
            if rest.startswith(":"):
                rest = rest[1:].strip()
            return rest
    return ""


def _extract_species(name: str, content: list[Paragraph]) -> Species:
    """Extract a Species from its content paragraphs."""
    creature_type = _extract_field(content, "Tipo di creatura")
    size = _extract_field(content, "Taglia")
    speed = _extract_field(content, "Velocità")

    # Body text is everything that's not a metadata field
    body: list[Paragraph] = []
    meta_labels = {"tipo di creatura", "taglia", "velocità"}
    for para in content:
        text = para.text.strip().lower()
        is_meta = any(text.startswith(label) for label in meta_labels)
        if not is_meta and para.role not in (SpanRole.TABLE_HEADER, SpanRole.TABLE_HEADER_SMALL):
            body.append(para)

    description = paragraphs_to_markdown(body)

    # Extract traits from description using bold-italic pattern
    traits: list[SpeciesTrait] = []
    # Split on _**TraitName.**_ markers
    md = paragraphs_to_markdown(body)
    parts = _TRAIT_NAME_RE.split(md)
    if len(parts) > 1:
        # parts[0] = intro text before first trait
        # parts[1], parts[2] = trait name, trait text
        # parts[3], parts[4] = next trait name, text, etc.
        for i in range(1, len(parts), 2):
            trait_name = parts[i].strip()
            trait_text = parts[i + 1].strip() if i + 1 < len(parts) else ""
            traits.append(SpeciesTrait(name=trait_name, description=trait_text))

    return Species(
        id=slugify(name),
        name=name,
        creature_type=creature_type,
        size=size,
        speed=speed,
        traits=traits,
        description=description,
    )


def _collect_species(nodes: list[HeadingNode]) -> list[Species]:
    """Recursively find species entries in the heading tree."""
    results: list[Species] = []

    for node in nodes:
        # Descend through container nodes
        if node.title.lower().strip() in (
            "specie dei personaggi", "descrizioni delle specie",
        ):
            results.extend(_collect_species(node.children))
            continue

        # Skip backgrounds sections
        if node.title.lower().strip() in (
            "background dei personaggi", "descrizioni dei background",
            "elementi di un background",
        ):
            continue

        # H5 = individual species
        if node.level == 5 and node.title.lower().strip() in _KNOWN_SPECIES:
            results.append(_extract_species(node.title, node.content))
            continue

        # Descend other container nodes (H1, H4)
        results.extend(_collect_species(node.children))

    return results


@register("species")
def parse_species(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[Species]:
    """Parse playable species from pages 93-97."""
    return _collect_species(tree)
