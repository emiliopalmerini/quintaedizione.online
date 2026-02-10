"""Monsters (Mostri) and Animals (Animali) parser.

Structure in PDF:
  H2: Group heading (e.g., "Banditi", "Draghi bianchi")
    H3: Individual monster name (e.g., "Bandito", "Drago bianco adulto")
      STAT_SUBTITLE: "{Tipo} {Taglia}, {allineamento}"
      STAT_LABEL/STAT_VALUE: CA, Iniziativa, PF, Velocità
      STAT_SCORE_HEADER: MOD SALV
      STAT_SCORE_LABEL/VALUE: ability scores grid
      STAT_LABEL/VALUE: Abilità, Sensi, Lingue, GS, etc.
      H6: "Tratti" — monster traits
      H6: "Azioni" — monster actions
      H6: "Azioni bonus"
      H6: "Reazioni"
      H6: "Azioni leggendarie"

When H2 == H3 (same name), there is no group — it's a standalone monster.
Pages 289-384: Mostri, Pages 385-405: Animali.
"""

from __future__ import annotations

import re
import warnings

from ..classify import SpanRole
from ..heading_tree import HeadingNode
from ..markdown_gen import paragraph_to_markdown
from ..merge import Paragraph
from ..schemas import AbilityMods, AbilityScores, Feature, Monster
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

# Stat label names in the stat block header area
_STAT_LABELS = {"ca", "iniziativa", "pf", "velocità"}

# Section labels that appear as H6 headings
_SECTION_LABELS = {
    "tratti", "azioni", "azioni bonus", "reazioni", "azioni leggendarie",
}

# Known D&D 5e damage types (Italian)
_DAMAGE_TYPES = {
    "acido", "contundente", "freddo", "fulmine", "fuoco", "forza",
    "necrotico", "perforante", "psichico", "radioso", "tagliente",
    "tuono", "veleno",
}


# Known D&D 5e conditions (Italian)
_CONDITION_TYPES = {
    "accecato", "affascinato", "assordato", "avvelenato", "esausto",
    "incapacitato", "indebolimento", "invisibile", "paralizzato",
    "pietrificato", "privo di sensi", "prono", "spaventato", "stordito",
    "trattenuto", "afferrato",
}


def _extract_cr_number(gs_text: str) -> str:
    """Extract just the CR number from a GS field like '13 (PE 10.000; BC +5)'."""
    if not gs_text:
        return ""
    match = re.match(r"^(\d+(?:/\d+)?)", gs_text.strip())
    return match.group(1) if match else gs_text


def _classify_immunity_tokens(text: str) -> tuple[str, str]:
    """Classify comma-separated immunity tokens into damage and condition groups.

    Returns (damage_immunities, condition_immunities).
    """
    tokens = [t.strip() for t in text.split(",")]
    damage_parts: list[str] = []
    condition_parts: list[str] = []

    for token in tokens:
        if not token:
            continue
        # Strip parenthetical details for matching (e.g. "avvelenato (solo veleno)")
        clean = re.sub(r"\s*\(.*?\)", "", token).strip().lower()
        if clean in _DAMAGE_TYPES:
            damage_parts.append(token)
        elif clean in _CONDITION_TYPES:
            condition_parts.append(token)
        else:
            # Unknown token — warn and guess based on context
            warnings.warn(f"Monster immunity: unrecognized token '{token}'")
            # Default to condition if it doesn't look like a damage type
            condition_parts.append(token)

    return ", ".join(damage_parts), ", ".join(condition_parts)


# Ability score abbreviation map (Italian)
_ABILITY_NAMES = {
    "for": "strength",
    "des": "dexterity",
    "cos": "constitution",
    "int": "intelligence",
    "sag": "wisdom",
    "car": "charisma",
}


def _extract_stat_field(content: list[Paragraph], label: str) -> str:
    """Extract a stat block field value from STAT_LABEL/STAT_VALUE paragraphs."""
    for para in content:
        if para.role not in (SpanRole.STAT_LABEL, SpanRole.STAT_VALUE):
            continue
        text = para.text.strip()
        # Labels appear as: "CA 15" or "PF 52 (8d8 + 16)"
        # They may be combined: "CA 15	Iniziativa +1 (11)PF 11 (2d8 + 2)Velocità 9 m"
        # Split by known labels
        upper_label = label.upper() if label in ("ca", "pf", "gs") else label.capitalize()
        pattern = re.compile(rf"{re.escape(upper_label)}\s+(.+?)(?=\s*(?:CA|Iniziativa|PF|Velocità|Abilità|Attrezzatura|Sensi|Lingue|GS|Resistenz|Immunità|Vulnerabilit)\s|$)", re.IGNORECASE)
        m = pattern.search(text)
        if m:
            return m.group(1).strip()
    return ""


def _extract_stat_fields(content: list[Paragraph]) -> dict[str, str]:
    """Extract all stat block fields from content paragraphs."""
    # Combine all STAT_LABEL/STAT_VALUE paragraph texts
    stat_text = ""
    for para in content:
        if para.role in (SpanRole.STAT_LABEL, SpanRole.STAT_VALUE):
            stat_text += para.text + " "

    # Parse individual fields using regex
    fields: dict[str, str] = {}
    labels = [
        "CA", "Iniziativa", "PF", "Velocità", "Abilità",
        "Attrezzatura", "Sensi", "Lingue", "GS",
        "Resistenze", "Vulnerabilità", "Immunità",
    ]
    for i, label in enumerate(labels):
        # Find position of this label
        pos = stat_text.find(label)
        if pos < 0:
            continue
        start = pos + len(label)
        # Skip whitespace/colon
        while start < len(stat_text) and stat_text[start] in " \t:":
            start += 1
        # Find position of next label
        end = len(stat_text)
        for next_label in labels:
            if next_label == label:
                continue
            next_pos = stat_text.find(next_label, start)
            if 0 < next_pos < end:
                end = next_pos
        value = stat_text[start:end].strip()
        fields[label.lower()] = value

    return fields


def _parse_ability_scores(content: list[Paragraph]) -> tuple[AbilityScores, AbilityMods, dict[str, str]]:
    """Parse the ability score grid from STAT_SCORE_LABEL/VALUE paragraphs."""
    scores = AbilityScores(
        strength=10, dexterity=10, constitution=10,
        intelligence=10, wisdom=10, charisma=10,
    )
    mods = AbilityMods(
        strength=0, dexterity=0, constitution=0,
        intelligence=0, wisdom=0, charisma=0,
    )
    saves: dict[str, str] = {}

    # Collect all score-related text
    score_text = ""
    for para in content:
        if para.role in (SpanRole.STAT_SCORE_LABEL, SpanRole.STAT_SCORE_VALUE):
            score_text += para.text + " "

    if not score_text:
        return scores, mods, saves

    # Parse: "For 11 +0 +0 Des 12 +1 +1 Cos 12 +1 +1 Int 10 +0 +0 Sag 10 +0 +0 Car 10 +0 +0"
    # The pattern is: AbbrScore Mod Save (or just Score Mod if no proficiency)
    pattern = re.compile(
        r"([A-Z][a-z]{2})\s*(\d+)\s*([\+\−\-]?\d+)\s*([\+\−\-]?\d+)?",
    )
    parsed_count = 0
    for m in pattern.finditer(score_text):
        abbr = m.group(1).lower()
        score_val = int(m.group(2))
        mod_val = int(m.group(3).replace("−", "-").replace("\u2212", "-"))
        save_val = m.group(4)

        ability = _ABILITY_NAMES.get(abbr)
        if ability:
            scores[ability] = score_val  # type: ignore[literal-required]
            mods[ability] = mod_val  # type: ignore[literal-required]
            parsed_count += 1
            if save_val:
                save_int = save_val.replace("−", "-").replace("\u2212", "-")
                saves[ability] = save_int

    if parsed_count < 6:
        warnings.warn(f"Monster: only {parsed_count}/6 ability scores parsed")

    return scores, mods, saves


def _find_feature_boundaries(spans: list) -> list[int]:
    """Find span indices where a new feature name starts.

    A feature boundary is a STAT_BOLD_ITALIC span that begins a name
    (consecutive STAT_BOLD_ITALIC spans ending with a period form the name).
    We detect boundaries by finding STAT_BOLD_ITALIC spans that follow
    a non-STAT_BOLD_ITALIC span (or are the first span).
    """
    boundaries: list[int] = []
    for i, span in enumerate(spans):
        if span.role != SpanRole.STAT_BOLD_ITALIC or not span.text.strip():
            continue
        # This is a bold-italic span. It starts a new feature if:
        # - it's the first span, or
        # - the previous non-empty span is not STAT_BOLD_ITALIC
        is_boundary = True
        for j in range(i - 1, -1, -1):
            if not spans[j].text.strip():
                continue
            if spans[j].role == SpanRole.STAT_BOLD_ITALIC:
                is_boundary = False
            break
        if is_boundary:
            boundaries.append(i)
    return boundaries


def _extract_feature_at(spans: list, start: int, end: int, page_num: int) -> tuple[str, str]:
    """Extract feature name and description markdown from a span slice.

    Args:
        spans: Full span list of the paragraph.
        start: Index of the first STAT_BOLD_ITALIC span of the feature name.
        end: Index one past the last span belonging to this feature.
        page_num: Page number for constructing temporary Paragraphs.

    Returns:
        (name, description_markdown) tuple.
    """
    name_parts: list[str] = []
    desc_start = start
    for i in range(start, end):
        span = spans[i]
        if span.role == SpanRole.STAT_BOLD_ITALIC and span.text.strip():
            name_parts.append(span.text.strip().rstrip("."))
            desc_start = i + 1
        elif not span.text.strip() and not name_parts:
            desc_start = i + 1
            continue
        else:
            desc_start = i
            break

    name = " ".join(name_parts).strip()
    remaining = spans[desc_start:end]
    desc_md = ""
    if remaining:
        desc_para = Paragraph(spans=remaining, role=SpanRole.STAT_BODY, page_num=page_num)
        desc_md = paragraph_to_markdown(desc_para).strip()
    return name, desc_md


def _parse_features(content: list[Paragraph]) -> list[Feature]:
    """Parse features from action section paragraphs.

    Features follow the pattern: _**Feature Name.**_ Description text
    Detected via STAT_BOLD_ITALIC spans. A single paragraph may contain
    multiple features when the merger groups consecutive stat lines together.
    """
    features: list[Feature] = []
    current_name = ""
    current_parts: list[str] = []

    for para in content:
        boundaries = _find_feature_boundaries(para.spans)

        if not boundaries:
            # No feature start — continuation of current feature
            if current_name:
                md = paragraph_to_markdown(para)
                if md.strip():
                    current_parts.append(md)
            continue

        # Process each feature found in this paragraph
        for idx, boundary in enumerate(boundaries):
            # Determine the end of this feature's spans
            if idx + 1 < len(boundaries):
                span_end = boundaries[idx + 1]
            else:
                span_end = len(para.spans)

            # Save previous feature before starting a new one
            if current_name:
                features.append(Feature(
                    name=current_name,
                    description="\n\n".join(current_parts),
                ))

            name, desc_md = _extract_feature_at(
                para.spans, boundary, span_end, para.page_num,
            )
            current_name = name
            current_parts = [desc_md] if desc_md else []

    # Don't forget last feature
    if current_name:
        features.append(Feature(
            name=current_name,
            description="\n\n".join(current_parts),
        ))

    return features


def _extract_monster(
    name: str,
    group: str,
    node: HeadingNode,
    source: str,
) -> Monster | None:
    """Extract a Monster from a heading node."""
    content = node.content

    # Find subtitle (STAT_SUBTITLE paragraph)
    subtitle = ""
    for para in content:
        if para.role == SpanRole.STAT_SUBTITLE:
            subtitle = para.text.strip()
            break

    if not subtitle:
        return None  # Not a real monster entry (probably a group intro)

    # Parse subtitle: "{Tipo} {Taglia}, {allineamento}"
    # e.g., "Umanoide Medio o Piccolo, neutrale"
    # e.g., "Mostruosità Media, senza allineamento"
    creature_type = ""
    subtype = ""
    size = ""
    alignment = ""
    if "," in subtitle:
        type_size, alignment = subtitle.rsplit(",", 1)
        alignment = alignment.strip()
        # Extract and strip parenthetical suffixes like "(mago)"
        subtype_match = re.search(r"\(([^)]+)\)", type_size)
        subtype = subtype_match.group(1) if subtype_match else ""
        type_size_clean = re.sub(r"\s*\(.*?\)", "", type_size).strip()
        parts = type_size_clean.split()
        if len(parts) >= 2:
            # Size indicators: Minuscola, Piccola, Media, Grande, Enorme, Mastodontica
            size_words = {"minuscola", "piccola", "media", "medio", "grande",
                          "enorme", "mastodontica", "piccolo", "minuscolo"}
            # Find where size starts (scan from end to handle multi-word types like "Non morto")
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

    # Extract stat fields
    fields = _extract_stat_fields(content)

    # Parse ability scores
    ability_scores, ability_mods, saving_throws = _parse_ability_scores(content)

    # Build sections from H6 children
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

    # Also check for traits in the main content (before any H6)
    # Some monsters have traits directly under H3 without a "Tratti" H6
    if not traits:
        # Look for STAT_BOLD_ITALIC spans in content after the stat block
        stat_block_done = False
        trait_content: list[Paragraph] = []
        for para in content:
            if para.role in (SpanRole.STAT_LABEL, SpanRole.STAT_VALUE,
                             SpanRole.STAT_SUBTITLE, SpanRole.STAT_SCORE_LABEL,
                             SpanRole.STAT_SCORE_VALUE, SpanRole.STAT_SCORE_HEADER):
                stat_block_done = True
                continue
            if stat_block_done and para.role in (SpanRole.STAT_BODY, SpanRole.STAT_BOLD_ITALIC, SpanRole.STAT_ITALIC):
                trait_content.append(para)
        if trait_content:
            traits = _parse_features(trait_content)

    # Split immunities into damage and condition immunities (separated by ";")
    raw_immunities = fields.get("immunità", "")
    damage_immunities = ""
    condition_immunities = ""
    if ";" in raw_immunities:
        parts = raw_immunities.split(";", 1)
        # Each part may still contain mixed tokens, so classify each
        dmg1, cond1 = _classify_immunity_tokens(parts[0].strip())
        dmg2, cond2 = _classify_immunity_tokens(parts[1].strip())
        damage_immunities = ", ".join(filter(None, [dmg1, dmg2]))
        condition_immunities = ", ".join(filter(None, [cond1, cond2]))
    elif raw_immunities:
        # No semicolon — classify each token individually
        damage_immunities, condition_immunities = _classify_immunity_tokens(raw_immunities)

    return Monster(
        id=slugify(name),
        name=name,
        group=group if group != name else "",
        type=creature_type,
        subtype=subtype,
        size=size,
        alignment=alignment,
        ac=fields.get("ca", ""),
        initiative=fields.get("iniziativa", ""),
        hp=fields.get("pf", ""),
        speed=fields.get("velocità", ""),
        ability_scores=ability_scores,
        ability_mods=ability_mods,
        saving_throws=saving_throws,
        skills=fields.get("abilità", ""),
        resistances=fields.get("resistenze", ""),
        damage_immunities=damage_immunities,
        condition_immunities=condition_immunities,
        senses=fields.get("sensi", ""),
        languages=fields.get("lingue", ""),
        cr=_extract_cr_number(fields.get("gs", "")),
        cr_detail=fields.get("gs", ""),
        equipment=fields.get("attrezzatura", ""),
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
    """Parse monsters from the Mostri/Animali sections."""
    source = "animals" if section.name == "Animali" else "monsters"
    monsters: list[Monster] = []

    def _collect(nodes: list[HeadingNode]) -> None:
        for node in nodes:
            # H1 — descend (e.g., "Mostri", "Mostri A-Z")
            if node.level == 1:
                _collect(node.children)
                continue

            # H2 — monster group
            if node.level == 2:
                group_name = node.title.strip()
                has_h3 = any(c.level == 3 for c in node.children)

                if has_h3:
                    for child in node.children:
                        if child.level == 3:
                            monster = _extract_monster(
                                child.title.strip(), group_name, child, source,
                            )
                            if monster:
                                monsters.append(monster)
                else:
                    # H2 with no H3 children — check if it has H4+ children
                    # or just intro text (skip)
                    _collect(node.children)
                continue

            # H3 — standalone monster
            if node.level == 3:
                monster = _extract_monster(
                    node.title.strip(), "", node, source,
                )
                if monster:
                    monsters.append(monster)
                continue

            # Other levels — descend
            _collect(node.children)

    _collect(tree)
    return monsters
