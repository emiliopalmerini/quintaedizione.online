"""Classes parser for SRD 5.1.

Structure: H1(ClassName) > H2("Privilegi di classe") > H3(features) + H5(proficiencies)
Subclasses: either separate H2 siblings or inline H3 from _INLINE_SUBCLASS_NAMES.
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

_INLINE_SUBCLASS_NAMES = {
    "cammino del berserker", "collegio della sapienza", "dominio della vita",
    "circolo della terra", "campione", "furfante", "scuola di invocazione",
    "via della mano aperta", "giuramento di devozione", "cacciatore",
    "discendenza draconica", "l'immondo",
}

_LEVEL_FEATURE_RE = re.compile(r"Livello\s+(\d+):\s*(.+)", re.IGNORECASE)

_PROFICIENCY_TITLES = {"punti ferita", "competenze", "equipaggiamento"}


def _find_hit_die(*texts: str) -> str:
    for text in texts:
        clean = text.replace("**", "").replace("*", "")
        m = re.search(r"Dado Vita\s*\|\s*[Dd](\d+)", clean)
        if m:
            return f"d{m.group(1)}"
        m = re.search(r"[Dd]adi [Vv]ita:?\s*\d*[Dd](\d+)", clean)
        if m:
            return f"d{m.group(1)}"
    return ""


def _parse_feature(node: HeadingNode) -> ClassFeature | None:
    title = node.title.strip()
    m = _LEVEL_FEATURE_RE.match(title)
    if m:
        level = int(m.group(1))
        name = m.group(2).strip()
    else:
        level = 0
        name = title

    description = paragraphs_to_markdown(node.content)
    for child in node.children:
        child_md = paragraphs_to_markdown(child.content)
        if child_md:
            description += f"\n\n#### {child.title}\n\n{child_md}"

    return ClassFeature(name=name, level=level, description=description)


def _parse_subclass(name: str, nodes: list[HeadingNode]) -> Subclass:
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
    name = class_node.title.strip()

    features: list[ClassFeature] = []
    subclasses: list[Subclass] = []
    proficiencies = ""
    description = ""
    spell_list: list[str] = []

    for h2 in class_node.children:
        title_lower = h2.title.lower().strip()

        if "privilegi di classe" in title_lower:
            desc_md = paragraphs_to_markdown(h2.content)
            if desc_md:
                description = desc_md

            for child in h2.children:
                child_title_lower = child.title.lower().strip()

                if child.level == 5 and child_title_lower in _PROFICIENCY_TITLES:
                    prof_md = paragraphs_to_markdown(child.content)
                    if prof_md:
                        proficiencies += f"**{child.title.strip()}**\n\n{prof_md}\n\n"
                    continue

                if child_title_lower in _INLINE_SUBCLASS_NAMES:
                    sc_nodes = list(child.children)
                    if child.content:
                        sc_nodes = [child] + list(child.children)
                    subclasses.append(_parse_subclass(child.title.strip(), sc_nodes))
                    continue

                feat = _parse_feature(child)
                if feat:
                    features.append(feat)
        else:
            for child in h2.children:
                if child.level == 3:
                    sc_nodes = list(child.children)
                    if child.content:
                        sc_nodes = [child] + list(child.children)
                    subclasses.append(_parse_subclass(child.title.strip(), sc_nodes))

    hit_die = _find_hit_die(description, proficiencies)

    return PlayerClass(
        id=slugify(name),
        name=name,
        hit_die=hit_die,
        proficiencies=proficiencies.strip(),
        level_table=[],
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
    """Parse classes from 5.1 SRD. Class names are at H1 level."""
    classes: list[PlayerClass] = []

    for node in tree:
        if node.level == 1 and node.title.strip() in _CLASS_NAMES:
            classes.append(_extract_class(node))

    return classes
