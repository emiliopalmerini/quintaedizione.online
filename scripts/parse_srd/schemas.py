"""Output TypedDicts for all JSON files."""

from __future__ import annotations

from typing import TypedDict


# ── Spells ──────────────────────────────────────────────────────────────────


class Spell(TypedDict):
    id: str
    name: str
    level: int  # 0 for cantrips
    school: str
    classes: list[str]
    casting_time: str
    range: str
    components: str
    duration: str
    description: str  # markdown
    at_higher_levels: str  # markdown, empty if N/A
    ritual: bool
    source: str


# ── Monsters ────────────────────────────────────────────────────────────────


class AbilityScores(TypedDict):
    strength: int
    dexterity: int
    constitution: int
    intelligence: int
    wisdom: int
    charisma: int


class AbilityMods(TypedDict):
    strength: int
    dexterity: int
    constitution: int
    intelligence: int
    wisdom: int
    charisma: int


class Feature(TypedDict):
    name: str
    description: str  # markdown


class Monster(TypedDict):
    id: str
    name: str
    group: str  # parent group heading if any (e.g., "Banditi")
    type: str
    subtype: str
    size: str
    alignment: str
    ac: str
    initiative: str
    hp: str
    speed: str
    ability_scores: AbilityScores
    ability_mods: AbilityMods
    saving_throws: dict[str, str]
    skills: str
    resistances: str
    damage_immunities: str
    condition_immunities: str
    senses: str
    languages: str
    cr: str
    cr_detail: str
    equipment: str
    traits: list[Feature]
    actions: list[Feature]
    bonus_actions: list[Feature]
    reactions: list[Feature]
    legendary_actions: list[Feature]
    source: str  # "monsters" or "animals"


# ── Classes ─────────────────────────────────────────────────────────────────


class ClassFeature(TypedDict):
    name: str
    level: int
    description: str  # markdown


class Subclass(TypedDict):
    name: str
    description: str  # markdown
    features: list[ClassFeature]


class PlayerClass(TypedDict):
    id: str
    name: str
    hit_die: str
    proficiencies: str  # markdown
    level_table: list[dict[str, str]]  # list of rows: {"level": "1", "feature": "...", ...}
    features: list[ClassFeature]
    subclasses: list[Subclass]
    spell_list: list[str]  # spell names for casters
    description: str  # markdown


# ── Equipment ───────────────────────────────────────────────────────────────


class EquipmentItem(TypedDict):
    id: str
    name: str
    category: str  # "weapons", "armor", "tools", "gear", "mounts", "services"
    subcategory: str
    properties: dict[str, str]
    description: str  # markdown


# ── Magic Items ─────────────────────────────────────────────────────────────


class MagicItem(TypedDict):
    id: str
    name: str
    type: str
    rarity: str
    attunement: bool
    attunement_details: str  # e.g., "da un chierico"
    description: str  # markdown


# ── Feats ───────────────────────────────────────────────────────────────────


class Feat(TypedDict):
    id: str
    name: str
    category: str  # Origini, Generale, Combattimento, Epico
    prerequisite: str
    repeatable: bool
    benefit: str  # markdown


# ── Backgrounds ─────────────────────────────────────────────────────────────


class Background(TypedDict):
    id: str
    name: str
    ability_scores: str
    feat: str
    skill_proficiencies: str
    tool_proficiency: str
    equipment: str
    description: str  # markdown


# ── Species ─────────────────────────────────────────────────────────────────


class SpeciesTrait(TypedDict):
    name: str
    description: str  # markdown


class Species(TypedDict):
    id: str
    name: str
    creature_type: str
    size: str
    speed: str
    traits: list[SpeciesTrait]
    description: str  # markdown


# ── Rules ───────────────────────────────────────────────────────────────────


class RuleEntry(TypedDict):
    id: str
    title: str
    content: str  # markdown
    children: list[RuleEntry]


# ── Glossary ────────────────────────────────────────────────────────────────


class GlossaryEntry(TypedDict):
    id: str
    term: str
    category: str  # optional descriptor in brackets
    definition: str  # markdown
    see_also: list[str]


# ── Attribution ─────────────────────────────────────────────────────────────


class Attribution(TypedDict):
    title: str
    license: str
    text: str  # full legal notice markdown
