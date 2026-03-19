"""Races (Razze) parser for SRD 5.1.

Structure in PDF (pages 2-7):
  H1: Razze
    H3: Tratti razziali — generic intro (skip)
    H2: Race name (e.g., "Elfo")
      H3: "Tratti degli elfi" — traits
        BODY paragraphs with _**TraitName.**_ pattern
        H5: Subrace name (e.g., "Elfo alto")
          BODY paragraphs with subrace traits

Outputs Species schema to species.json.
"""

from __future__ import annotations

import re

from ..heading_tree import HeadingNode
from ..markdown_gen import paragraphs_to_markdown
from ..merge import Paragraph
from ..schemas import Species, SpeciesTrait
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

# Matches trait names in bold-italic: _**TraitName.**_
_TRAIT_NAME_RE = re.compile(r"_\*\*(.+?)\.\*\*_")

_KNOWN_RACES = {
    "elfo", "halfling", "nano", "umano", "dragonide",
    "gnomo", "mezzelfo", "mezzorco", "tiefling",
}


def _extract_traits_from_markdown(md: str) -> list[SpeciesTrait]:
    """Extract traits from markdown using _**TraitName.**_ pattern."""
    traits: list[SpeciesTrait] = []
    parts = _TRAIT_NAME_RE.split(md)
    if len(parts) > 1:
        for i in range(1, len(parts), 2):
            trait_name = parts[i].strip()
            trait_text = parts[i + 1].strip() if i + 1 < len(parts) else ""
            traits.append(SpeciesTrait(name=trait_name, description=trait_text))
    return traits


def _extract_race(name: str, node: HeadingNode) -> Species:
    """Extract a Species from a race H2 node."""
    # Find the "Tratti dei/degli..." H3 child
    traits_node: HeadingNode | None = None
    for child in node.children:
        if child.level == 3 and child.title.lower().startswith("tratti"):
            traits_node = child
            break

    if not traits_node:
        return Species(
            id=slugify(name),
            name=name,
            creature_type="umanoide",
            size="",
            speed="",
            traits=[],
            description="",
        )

    # Build description from traits content
    md = paragraphs_to_markdown(traits_node.content)
    traits = _extract_traits_from_markdown(md)

    # Extract metadata from traits
    speed = ""
    size = ""
    for trait in traits:
        if "velocità" in trait["name"].lower():
            speed = trait["description"]
        elif "taglia" in trait["name"].lower():
            size = trait["description"]

    # Add subrace content
    subrace_parts: list[str] = []
    for child in traits_node.children:
        if child.level == 5:
            child_md = paragraphs_to_markdown(child.content)
            subrace_parts.append(f"##### {child.title.strip()}\n\n{child_md}")

    full_description = md
    if subrace_parts:
        full_description += "\n\n" + "\n\n".join(subrace_parts)

    return Species(
        id=slugify(name),
        name=name,
        creature_type="umanoide",
        size=size,
        speed=speed,
        traits=traits,
        description=full_description,
    )


@register("races")
def parse_races(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[Species]:
    """Parse races from 5.1 SRD into Species schema."""
    results: list[Species] = []

    for node in tree:
        # H1("Razze") — descend into children
        if node.level == 1:
            for child in node.children:
                if child.level == 2 and child.title.lower().strip() in _KNOWN_RACES:
                    results.append(_extract_race(child.title.strip(), child))
            continue

        # Direct H2 race (if no H1 wrapper)
        if node.level == 2 and node.title.lower().strip() in _KNOWN_RACES:
            results.append(_extract_race(node.title.strip(), node))

    return results
