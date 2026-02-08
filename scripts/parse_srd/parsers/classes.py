"""Classes (Classi) parser.

Structure in PDF:
  H1: Classi
    H2: {ClassName} (e.g., "Barbaro")
      H4: "Diventare un {classe}..." — prerequisites
      H4: "Privilegi di classe del {classe}" — class features
        H5: "Livello N: Feature Name" — individual features
      H4: "Sottoclasse del {classe}:" — subclass header
      H4: "{SubclassName}" — subclass details
        H5: "Livello N: Feature Name"

Content under H2 includes the level table as TABLE_HEADER/SIDEBAR text.
Hit die is in TABLE_HEADER_SMALL "Dado Vita" paragraph, value in next SIDEBAR.

Subclass header patterns:
  A) "Sottoclasse del X:" (0 children) → next H4 = subclass name
  B) "Sottoclasse del X: Name" (has children) → name after colon, children are features
  C) "Sottoclasse del X: Partial" (0 children) → next H4 has rest of name
"""

from __future__ import annotations

import re

from ..heading_tree import HeadingNode
from ..markdown_gen import paragraphs_to_markdown
from ..merge import Paragraph
from ..schemas import ClassFeature, PlayerClass, Subclass
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

_CLASS_NAMES = [
    "Barbaro", "Bardo", "Chierico", "Druido", "Guerriero", "Ladro",
    "Mago", "Monaco", "Paladino", "Ranger", "Stregone", "Warlock",
]

# Match "Livello N: Feature Name"
_LEVEL_FEATURE_RE = re.compile(r"Livello\s+(\d+):\s*(.+)", re.IGNORECASE)


def _parse_feature(node: HeadingNode) -> ClassFeature | None:
    """Parse a class feature from an H5 node."""
    title = node.title.strip()
    m = _LEVEL_FEATURE_RE.match(title)
    if m:
        level = int(m.group(1))
        name = m.group(2).strip()
    else:
        level = 0
        name = title

    description = paragraphs_to_markdown(node.content)
    # Include children content
    for child in node.children:
        child_md = paragraphs_to_markdown(child.content)
        if child_md:
            description += f"\n\n#### {child.title}\n\n{child_md}"

    return ClassFeature(name=name, level=level, description=description)


def _parse_subclass(name: str, nodes: list[HeadingNode]) -> Subclass:
    """Parse a subclass from its H4 node and H5 feature children."""
    features: list[ClassFeature] = []
    desc_parts: list[str] = []

    for node in nodes:
        if node.level == 5:
            feat = _parse_feature(node)
            if feat:
                features.append(feat)
        elif node.content:
            desc_parts.append(paragraphs_to_markdown(node.content))

    return Subclass(
        name=name,
        description="\n\n".join(desc_parts),
        features=features,
    )


def _extract_class(class_node: HeadingNode) -> PlayerClass:
    """Extract a PlayerClass from an H2 heading node."""
    name = class_node.title.strip()

    # Content under H2 = intro text + level table + trait summary
    description = paragraphs_to_markdown(class_node.content)

    # Extract hit die from the markdown table in description
    # Table contains "Dado Vita | D12 per ogni livello da barbaro" etc.
    hit_die = ""
    m = re.search(r"Dado Vita\s*\|\s*[Dd](\d+)", description)
    if m:
        hit_die = f"d{m.group(1)}"

    # Process H4 children
    features: list[ClassFeature] = []
    subclasses: list[Subclass] = []
    proficiencies = ""
    spell_list: list[str] = []

    # State: None = normal, "" = Pattern A (next H4 = full name),
    # "Partial" = Pattern C (next H4 completes the name)
    pending_sc: str | None = None

    for h4 in class_node.children:
        title_lower = h4.title.lower().strip()

        # If expecting subclass name continuation (Pattern A or C)
        if pending_sc is not None:
            if pending_sc:
                sc_name = f"{pending_sc} {h4.title.strip()}"
            else:
                sc_name = h4.title.strip()
            pending_sc = None
            sc_nodes = list(h4.children)
            if h4.content:
                sc_nodes = [h4] + list(h4.children)
            subclasses.append(_parse_subclass(sc_name, sc_nodes))
            continue

        # Prerequisites section
        if "diventare" in title_lower:
            prof_md = paragraphs_to_markdown(h4.content)
            for child in h4.children:
                prof_md += "\n\n" + paragraphs_to_markdown(child.content)
            proficiencies = prof_md
            continue

        # Class features section
        if "privilegi di classe" in title_lower:
            for child in h4.children:
                feat = _parse_feature(child)
                if feat:
                    features.append(feat)
            continue

        # Spell list section
        if "lista degli incantesimi" in title_lower:
            continue

        # Options sections (metamagic, invocations) — include as features
        if "opzioni di" in title_lower:
            for child in h4.children:
                feat = _parse_feature(child)
                if feat:
                    features.append(feat)
            continue

        # Subclass header ("Sottoclasse del {classe}:...")
        if "sottoclasse" in title_lower:
            colon_idx = h4.title.find(":")
            after_colon = h4.title[colon_idx + 1:].strip() if colon_idx >= 0 else ""

            if after_colon and h4.children:
                # Pattern B: full name + features on same H4
                sc_nodes = list(h4.children)
                if h4.content:
                    sc_nodes = [h4] + list(h4.children)
                subclasses.append(_parse_subclass(after_colon, sc_nodes))
            elif after_colon:
                # Pattern C: partial name, next H4 has rest
                pending_sc = after_colon
            else:
                # Pattern A: next H4 is the full subclass name
                pending_sc = ""
            continue

    return PlayerClass(
        id=slugify(name),
        name=name,
        hit_die=hit_die,
        proficiencies=proficiencies,
        level_table=[],  # Level table extraction is complex; stored in description
        features=features,
        subclasses=subclasses,
        spell_list=spell_list,
        description=description,
    )


@register("classes")
def parse_classes(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[PlayerClass]:
    """Parse player classes from the Classi section (pages 32-92)."""
    classes: list[PlayerClass] = []

    def _collect(nodes: list[HeadingNode]) -> None:
        for node in nodes:
            if node.level == 1:
                _collect(node.children)
                continue
            if node.level == 2 and node.title.strip() in _CLASS_NAMES:
                classes.append(_extract_class(node))

    _collect(tree)
    return classes
