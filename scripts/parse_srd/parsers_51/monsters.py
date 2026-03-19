"""Monsters parser for SRD 5.1.

Structure: H2("Mostri (A)") > H5(monster) or H2 > H3(group) > H5(monster)
Stat blocks use BODY_BOLD labels + SIDEBAR values (no Optima/stat-red).
"""

from __future__ import annotations

import re
import warnings

from ..classify import SpanRole
from ..heading_tree import HeadingNode
from ..merge import Paragraph
from ..schemas import AbilityMods, AbilityScores, Feature, Monster
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

_ABILITY_NAMES = {
    "for": "strength", "des": "dexterity", "cos": "constitution",
    "int": "intelligence", "sag": "wisdom", "car": "charisma",
}


def _extract_cr_number(gs_text: str) -> str:
    if not gs_text:
        return ""
    match = re.match(r"^(\d+(?:/\d+)?)", gs_text.strip())
    return match.group(1) if match else gs_text


def _extract_stat_field(content: list[Paragraph], label: str) -> str:
    """Extract stat field from BODY_BOLD label + SIDEBAR value spans."""
    label_lower = label.lower()
    for para in content:
        for i, span in enumerate(para.spans):
            if span.role != SpanRole.BODY_BOLD:
                continue
            if not span.text.strip().lower().startswith(label_lower):
                continue
            value_parts: list[str] = []
            remaining = span.text.strip()[len(label):].strip().lstrip(":")
            if remaining:
                value_parts.append(remaining)
            for j in range(i + 1, len(para.spans)):
                next_span = para.spans[j]
                if next_span.role == SpanRole.BODY_BOLD:
                    break
                value_parts.append(next_span.text.strip())
            return " ".join(p for p in value_parts if p).strip()
    return ""


def _parse_ability_scores(content: list[Paragraph]) -> tuple[AbilityScores, AbilityMods]:
    scores = AbilityScores(
        strength=10, dexterity=10, constitution=10,
        intelligence=10, wisdom=10, charisma=10,
    )
    mods = AbilityMods(
        strength=0, dexterity=0, constitution=0,
        intelligence=0, wisdom=0, charisma=0,
    )

    # Strategy 1: look for individual BODY_BOLD paragraphs (rare — usually merged)
    # Strategy 2: find ability labels in spans across all paragraphs
    ability_order: list[str] = []
    score_values: list[str] = []

    # Scan all spans for ability abbreviations
    for para in content:
        for span in para.spans:
            text = span.text.strip()
            if span.role == SpanRole.BODY_BOLD and text.upper() in ("FOR", "DES", "COS", "INT", "SAG", "CAR"):
                ability = _ABILITY_NAMES.get(text.lower())
                if ability and ability not in ability_order:
                    ability_order.append(ability)

    if len(ability_order) < 6:
        # Fallback: search in merged paragraph text
        ability_order = []
        full_text = " ".join(p.text for p in content)
        if "FOR" in full_text and "CAR" in full_text:
            ability_order = ["strength", "dexterity", "constitution",
                             "intelligence", "wisdom", "charisma"]

    if not ability_order:
        return scores, mods

    # Find score values: "21 (+5)" patterns in SIDEBAR paragraphs or merged text
    for para in content:
        if para.role == SpanRole.SIDEBAR or para.role == SpanRole.BODY_ITALIC:
            for m in re.finditer(r"(\d+)\s*\(([\+\-\−]?\d+)\)", para.text):
                score_values.append(m.group(0))
                if len(score_values) >= 6:
                    break
        if len(score_values) >= 6:
            break

    for i, ability in enumerate(ability_order[:6]):
        if i < len(score_values):
            m = re.match(r"(\d+)\s*\(([\+\-\−]?\d+)\)", score_values[i])
            if m:
                scores[ability] = int(m.group(1))  # type: ignore[literal-required]
                mod_str = m.group(2).replace("−", "-").replace("\u2212", "-")
                mods[ability] = int(mod_str)  # type: ignore[literal-required]

    return scores, mods


def _is_feature_name_span(span) -> bool:
    """Check if a span starts a feature name.

    Feature names are BODY_BOLD_ITALIC (traits/actions) or BODY_BOLD ending
    with a period (legendary actions use Calibri-Bold instead of BoldItalic).
    """
    text = span.text.strip()
    if not text:
        return False
    if span.role == SpanRole.BODY_BOLD_ITALIC:
        return True
    if span.role == SpanRole.BODY_BOLD and text.endswith("."):
        return True
    return False


def _parse_features(content: list[Paragraph]) -> list[Feature]:
    """Parse features from BODY_BOLD_ITALIC or BODY_BOLD name + description pattern."""
    features: list[Feature] = []
    current_name = ""
    current_parts: list[str] = []

    for para in content:
        found_feature = False
        for i, span in enumerate(para.spans):
            if _is_feature_name_span(span):
                if current_name:
                    features.append(Feature(
                        name=current_name,
                        description="\n\n".join(current_parts),
                    ))
                name = span.text.strip().rstrip(".")
                desc_parts = [para.spans[j].text for j in range(i + 1, len(para.spans))]
                current_name = name
                current_parts = [" ".join(desc_parts).strip()] if desc_parts else []
                found_feature = True
                break

        if not found_feature and current_name:
            md = para.text.strip()
            if md:
                current_parts.append(md)

    if current_name:
        features.append(Feature(
            name=current_name,
            description="\n\n".join(current_parts),
        ))

    return features


def _is_monster_node(node: HeadingNode) -> bool:
    """Check if a node looks like a monster (H5 with BODY_ITALIC subtitle containing comma)."""
    if node.level != 5:
        return False
    for para in node.content:
        if para.role == SpanRole.BODY_ITALIC and "," in para.text:
            return True
    return False


def _extract_monster(
    name: str, group: str, node: HeadingNode, source: str,
) -> Monster | None:
    content = node.content

    # Find subtitle — truncate at first stat label
    subtitle = ""
    for para in content:
        if para.role == SpanRole.BODY_ITALIC:
            full_text = para.text.strip()
            for marker in ("Classe Armatura", "Punti Ferita", "CA "):
                idx = full_text.find(marker)
                if idx > 0:
                    full_text = full_text[:idx].strip()
                    break
            subtitle = full_text
            break

    if not subtitle:
        return None

    # Parse subtitle: "Aberrazione Grande, legale malvagio"
    creature_type = ""
    subtype = ""
    size = ""
    alignment = ""
    if "," in subtitle:
        type_size, alignment = subtitle.rsplit(",", 1)
        alignment = alignment.strip()
        subtype_match = re.search(r"\(([^)]+)\)", type_size)
        subtype = subtype_match.group(1) if subtype_match else ""
        type_size_clean = re.sub(r"\s*\(.*?\)", "", type_size).strip()
        parts = type_size_clean.split()
        if len(parts) >= 2:
            size_words = {"minuscola", "piccola", "media", "medio", "grande",
                          "enorme", "mastodontica", "piccolo", "minuscolo"}
            size_start = len(parts)
            for i in range(len(parts) - 1, -1, -1):
                if parts[i].lower().rstrip(",") in size_words or parts[i].lower() == "o":
                    size_start = i
                else:
                    break
            if size_start < len(parts):
                creature_type = " ".join(parts[:size_start])
                size = " ".join(parts[size_start:])
            else:
                creature_type = parts[0]
                size = " ".join(parts[1:])

    ac = _extract_stat_field(content, "Classe Armatura")
    hp = _extract_stat_field(content, "Punti Ferita")
    speed = _extract_stat_field(content, "Velocità")
    saves_text = _extract_stat_field(content, "Tiri Salvezza")
    skills = _extract_stat_field(content, "Abilità")
    senses = _extract_stat_field(content, "Sensi")
    languages = _extract_stat_field(content, "Linguaggi")
    cr_text = _extract_stat_field(content, "Sfida")
    resistances = _extract_stat_field(content, "Resistenze")
    immunities = _extract_stat_field(content, "Immunità")

    # Parse saving throws: "Cos +6, Int +8, Sag +6" → {"constitution": "+6", ...}
    saving_throws: dict[str, str] = {}
    if saves_text:
        for part in saves_text.split(","):
            part = part.strip()
            tokens = part.split()
            if len(tokens) >= 2:
                abbr = tokens[0].lower().rstrip(".")
                ability = _ABILITY_NAMES.get(abbr, "")
                if ability:
                    saving_throws[ability] = tokens[1]

    ability_scores, ability_mods = _parse_ability_scores(content)

    traits: list[Feature] = []
    actions: list[Feature] = []
    bonus_actions: list[Feature] = []
    reactions: list[Feature] = []
    legendary_actions: list[Feature] = []

    for child in node.children:
        section_name = child.title.lower().strip()
        features = _parse_features(child.content)
        if section_name == "tratti":
            traits = features
        elif section_name == "azioni":
            actions = features
        elif section_name == "azioni bonus":
            bonus_actions = features
        elif section_name == "reazioni":
            reactions = features
        elif section_name == "azioni leggendarie":
            legendary_actions = features

    if not traits:
        trait_content: list[Paragraph] = []
        past_stats = False
        for para in content:
            if para.role == SpanRole.BODY_BOLD_ITALIC:
                past_stats = True
                trait_content.append(para)
            elif past_stats and para.role in (SpanRole.SIDEBAR, SpanRole.BODY):
                trait_content.append(para)
        if trait_content:
            traits = _parse_features(trait_content)

    damage_immunities = ""
    condition_immunities = ""

    return Monster(
        id=slugify(name),
        name=name,
        group=group if group != name else "",
        type=creature_type,
        subtype=subtype,
        size=size,
        alignment=alignment,
        ac=ac,
        initiative="",
        hp=hp,
        speed=speed,
        ability_scores=ability_scores,
        ability_mods=ability_mods,
        saving_throws=saving_throws,
        skills=skills,
        resistances=resistances,
        damage_immunities=damage_immunities,
        condition_immunities=condition_immunities,
        senses=senses,
        languages=languages,
        cr=_extract_cr_number(cr_text),
        cr_detail=cr_text,
        equipment="",
        traits=traits,
        actions=actions,
        bonus_actions=bonus_actions,
        reactions=reactions,
        legendary_actions=legendary_actions,
        source=source,
    )


@register("monsters")
def parse_monsters(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[Monster]:
    """Parse monsters from 5.1 SRD."""
    source = "monsters"
    monsters: list[Monster] = []

    def _collect(nodes: list[HeadingNode], group: str = "") -> None:
        for node in nodes:
            if node.level == 1:
                _collect(node.children, group)
                continue

            if node.level == 2:
                group_name = node.title.strip()
                _collect(node.children, group_name)
                continue

            if node.level == 3:
                has_h5_monsters = any(_is_monster_node(c) for c in node.children)
                if has_h5_monsters:
                    _collect(node.children, node.title.strip())
                continue

            if node.level == 5 and _is_monster_node(node):
                monster = _extract_monster(
                    node.title.strip(), group, node, source,
                )
                if monster:
                    monsters.append(monster)
                continue

            _collect(node.children, group)

    _collect(tree)
    return monsters
