"""
Value Objects for D&D 5e SRD Domain Model
Following DDD principles with immutable value objects
"""
from __future__ import annotations

import re
from dataclasses import dataclass
from enum import Enum
from typing import Dict, List, Optional


class Ability(Enum):
    """Six core D&D abilities"""
    FORZA = "Forza"
    DESTREZZA = "Destrezza" 
    COSTITUZIONE = "Costituzione"
    INTELLIGENZA = "Intelligenza"
    SAGGEZZA = "Saggezza"
    CARISMA = "Carisma"


class HitDie(Enum):
    """Valid hit dice for classes"""
    D6 = "d6"
    D8 = "d8" 
    D10 = "d10"
    D12 = "d12"


class SpellPreparation(Enum):
    """How class prepares spells"""
    PREPARED = "prepared"
    KNOWN = "known"
    NONE = "none"


class Rituals(Enum):
    """Ritual casting capability"""
    NESSUNO = "nessuno"
    SOLO_LISTA = "solo_lista"
    DA_LIBRO = "da_libro"


@dataclass(frozen=True)
class ClassSlug:
    """Unique identifier for a class"""
    value: str
    
    def __post_init__(self):
        if not re.match(r'^[a-z][a-z0-9-]*$', self.value):
            raise ValueError(f"Invalid class slug: {self.value}")


@dataclass(frozen=True)
class Level:
    """Character level (1-20)"""
    value: int
    
    def __post_init__(self):
        if not 1 <= self.value <= 20:
            raise ValueError(f"Level must be 1-20, got {self.value}")


@dataclass(frozen=True)
class ProficiencyBonus:
    """Proficiency bonus by level"""
    value: int
    
    def __post_init__(self):
        if not 2 <= self.value <= 6:
            raise ValueError(f"Proficiency bonus must be 2-6, got {self.value}")
    
    @classmethod
    def for_level(cls, level: Level) -> ProficiencyBonus:
        """Calculate proficiency bonus for level"""
        bonus = 2 + ((level.value - 1) // 4)
        return cls(bonus)


@dataclass(frozen=True)
class SkillChoice:
    """Skill proficiency selection"""
    choose: int
    options: List[str]
    
    def __post_init__(self):
        if self.choose < 0:
            raise ValueError("Choose count cannot be negative")
        if not self.options:
            raise ValueError("Options cannot be empty")


@dataclass(frozen=True)
class EquipmentOption:
    """Equipment choice option"""
    label: str
    items: List[str]
    
    def __post_init__(self):
        if not self.label:
            raise ValueError("Equipment option must have a label")
        if not self.items:
            raise ValueError("Equipment option must have items")


@dataclass(frozen=True)
class MulticlassRequirement:
    """Prerequisite for multiclassing"""
    ability: Ability
    minimum_score: int
    
    def __post_init__(self):
        if not 1 <= self.minimum_score <= 30:
            raise ValueError(f"Ability score must be 1-30, got {self.minimum_score}")


@dataclass(frozen=True)
class SpellSlots:
    """Spell slots by level (1st through 9th)"""
    slots: List[int]
    
    def __post_init__(self):
        if len(self.slots) != 9:
            raise ValueError("Must have exactly 9 spell slot levels")
        if any(slot < 0 for slot in self.slots):
            raise ValueError("Spell slots cannot be negative")
    
    def get_slot_count(self, spell_level: int) -> int:
        """Get spell slots for a specific spell level (1-9)"""
        if not 1 <= spell_level <= 9:
            raise ValueError(f"Spell level must be 1-9, got {spell_level}")
        return self.slots[spell_level - 1]


@dataclass(frozen=True)
class ClassFeature:
    """Individual class feature"""
    name: str
    level: Level
    description: str
    
    def __post_init__(self):
        if not self.name:
            raise ValueError("Feature must have a name")
        if not self.description:
            raise ValueError("Feature must have a description")


@dataclass(frozen=True)
class LevelProgression:
    """Complete progression data for a level"""
    level: Level
    proficiency_bonus: ProficiencyBonus
    class_features: List[str]
    cantrips_known: Optional[int] = None
    spells_prepared: Optional[int] = None 
    spell_slots: Optional[SpellSlots] = None
    resources: Optional[Dict[str, int]] = None
    
    def __post_init__(self):
        if not self.class_features:
            raise ValueError("Level progression must have at least one feature")


@dataclass(frozen=True)
class MagicProgression:
    """Spellcasting progression for class"""
    has_spells: bool
    spell_list_reference: Optional[str] = None
    spellcasting_ability: Optional[Ability] = None
    preparation_method: SpellPreparation = SpellPreparation.NONE
    focus: Optional[str] = None
    ritual_casting: Rituals = Rituals.NESSUNO
    cantrip_progression: Optional[Dict[int, int]] = None
    spells_known_progression: Optional[Dict[int, int]] = None
    spell_slot_progression: Optional[Dict[int, SpellSlots]] = None
    
    def __post_init__(self):
        if self.has_spells:
            if not self.spellcasting_ability:
                raise ValueError("Spellcasting classes must have ability")
            if self.preparation_method == SpellPreparation.NONE:
                raise ValueError("Spellcasting classes must specify preparation method")


@dataclass(frozen=True)
class ResourceProgression:
    """Class resource progression (e.g., rage uses, bardic inspiration)"""
    resource_key: str
    progression: Dict[int, int]  # level -> count
    
    def __post_init__(self):
        if not self.resource_key:
            raise ValueError("Resource must have a key")
        if not self.progression:
            raise ValueError("Resource must have progression data")
        if any(level < 1 or level > 20 for level in self.progression.keys()):
            raise ValueError("Resource progression levels must be 1-20")