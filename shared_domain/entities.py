"""
Shared Domain Entities for D&D 5e SRD
Used by both Editor (read-side) and Parser (write-side)
"""
from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, List, Optional, Protocol, Any
import re


class Ability(Enum):
    """Six core D&D abilities"""
    FORZA = "Forza"
    DESTREZZA = "Destrezza" 
    COSTITUZIONE = "Costituzione"
    INTELLIGENZA = "Intelligenza"
    SAGGEZZA = "Saggezza"
    CARISMA = "Carisma"


@dataclass(frozen=True)
class EntityId:
    """Base entity identifier"""
    value: str
    
    def __post_init__(self):
        if not self.value or not self.value.strip():
            raise ValueError("EntityId cannot be empty")


@dataclass(frozen=True)
class ClassId(EntityId):
    """Class entity identifier"""
    
    def __post_init__(self):
        super().__post_init__()
        if not re.match(r'^[a-z][a-z0-9-]*$', self.value):
            raise ValueError(f"Invalid class ID format: {self.value}")


@dataclass(frozen=True)
class Level:
    """Character level (1-20)"""
    value: int
    
    def __post_init__(self):
        if not 1 <= self.value <= 20:
            raise ValueError(f"Level must be 1-20, got {self.value}")


@dataclass(frozen=True)
class HitDie:
    """Hit die value (d6, d8, d10, d12)"""
    value: int
    
    def __post_init__(self):
        valid_dice = {6, 8, 10, 12}
        if self.value not in valid_dice:
            raise ValueError(f"Hit die must be one of {valid_dice}, got d{self.value}")
    
    def __str__(self) -> str:
        return f"d{self.value}"


# Repository Interfaces (Ports)
class ClassRepository(ABC):
    """Port for class data persistence"""
    
    @abstractmethod
    async def find_by_id(self, class_id: ClassId) -> Optional['DndClass']:
        pass
    
    @abstractmethod
    async def find_all(self) -> List['DndClass']:
        pass
    
    @abstractmethod
    async def save(self, dnd_class: 'DndClass') -> None:
        pass
    
    @abstractmethod
    async def search(self, query: str, filters: Dict[str, Any] = None) -> List['DndClass']:
        pass


class SpellRepository(ABC):
    """Port for spell data persistence"""
    
    @abstractmethod
    async def find_by_name(self, name: str) -> Optional['Spell']:
        pass
    
    @abstractmethod
    async def find_by_level(self, level: int) -> List['Spell']:
        pass


# Domain Events
@dataclass(frozen=True)
class DomainEvent:
    """Base domain event"""
    event_id: str
    timestamp: str
    aggregate_id: str


@dataclass(frozen=True) 
class ClassParsed(DomainEvent):
    """Event fired when a class is successfully parsed"""
    class_name: str
    version: str


@dataclass(frozen=True)
class ClassViewed(DomainEvent):
    """Event fired when a class is viewed"""
    class_name: str
    user_agent: Optional[str] = None


# Core Domain Entities
@dataclass
class ClassFeature:
    """Individual class feature with business logic"""
    name: str
    level: Level
    description: str
    
    def __post_init__(self):
        if not self.name or not self.name.strip():
            raise ValueError("Feature must have a name")
        if not self.description or not self.description.strip():
            raise ValueError("Feature must have a description")
    
    def is_core_feature(self) -> bool:
        """Check if this is a core class feature (levels 1-3)"""
        return self.level.value <= 3
    
    def get_summary(self, max_chars: int = 100) -> str:
        """Get feature summary for display"""
        if len(self.description) <= max_chars:
            return self.description
        return self.description[:max_chars-3] + "..."


@dataclass
class SpellProgression:
    """Spellcasting progression for a class"""
    cantrips_by_level: Dict[int, int]
    spells_by_level: Dict[int, int] 
    spell_slots_by_level: Dict[int, List[int]]
    
    def get_cantrips_at_level(self, level: Level) -> int:
        """Get number of cantrips known at level"""
        return self.cantrips_by_level.get(level.value, 0)
    
    def get_spell_slots_at_level(self, character_level: Level, spell_level: int) -> int:
        """Get spell slots for specific spell level at character level"""
        slots = self.spell_slots_by_level.get(character_level.value, [])
        if spell_level < 1 or spell_level > len(slots):
            return 0
        return slots[spell_level - 1]
    
    def is_full_caster(self) -> bool:
        """Check if this represents a full spellcaster progression"""
        # Full casters get 9th level slots at level 17
        slots_at_17 = self.spell_slots_by_level.get(17, [])
        return len(slots_at_17) >= 9 and slots_at_17[8] > 0


@dataclass
class DndClass:
    """Core D&D Class aggregate root"""
    id: ClassId
    name: str
    primary_ability: Ability
    hit_die: str  # e.g., "d12"
    
    # Collections
    features: List[ClassFeature] = field(default_factory=list)
    subclasses: List['Subclass'] = field(default_factory=list)
    
    # Optional data
    spell_progression: Optional[SpellProgression] = None
    saving_throw_proficiencies: List[Ability] = field(default_factory=list)
    armor_proficiencies: List[str] = field(default_factory=list)
    weapon_proficiencies: List[str] = field(default_factory=list)
    skill_options: Optional[Dict[str, Any]] = None
    
    # Metadata
    version: str = "1.0"
    source: str = "SRD"
    
    def __post_init__(self):
        if not self.name or not self.name.strip():
            raise ValueError("Class must have a name")
    
    def add_feature(self, feature: ClassFeature) -> None:
        """Add a class feature with domain logic"""
        # Business rule: No duplicate features at same level
        existing = [f for f in self.features if f.name == feature.name and f.level == feature.level]
        if existing:
            raise ValueError(f"Feature '{feature.name}' already exists at level {feature.level.value}")
        
        self.features.append(feature)
        self._sort_features()
    
    def get_features_at_level(self, level: Level) -> List[ClassFeature]:
        """Get all features available at a specific level"""
        return [f for f in self.features if f.level.value <= level.value]
    
    def get_core_features(self) -> List[ClassFeature]:
        """Get defining features of the class (levels 1-3)"""
        return [f for f in self.features if f.is_core_feature()]
    
    def is_spellcaster(self) -> bool:
        """Check if class can cast spells"""
        return self.spell_progression is not None
    
    def is_full_caster(self) -> bool:
        """Check if class is a full spellcaster"""
        if not self.spell_progression:
            return False
        return self.spell_progression.is_full_caster()
    
    def get_max_spell_level_at(self, character_level: Level) -> int:
        """Get highest spell level castable at character level"""
        if not self.spell_progression:
            return 0
        
        slots = self.spell_progression.spell_slots_by_level.get(character_level.value, [])
        for spell_level in range(9, 0, -1):
            if spell_level <= len(slots) and slots[spell_level - 1] > 0:
                return spell_level
        return 0
    
    def _sort_features(self) -> None:
        """Keep features sorted by level then name"""
        self.features.sort(key=lambda f: (f.level.value, f.name))
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "id": self.id.value,
            "name": self.name,
            "primary_ability": self.primary_ability.value,
            "hit_die": self.hit_die,
            "is_spellcaster": self.is_spellcaster(),
            "is_full_caster": self.is_full_caster(),
            "feature_count": len(self.features),
            "subclass_count": len(self.subclasses),
            "version": self.version,
            "source": self.source
        }


@dataclass
class Subclass:
    """Class specialization"""
    id: EntityId
    name: str
    parent_class_id: ClassId
    features: List[ClassFeature] = field(default_factory=list)
    description: Optional[str] = None
    
    def __post_init__(self):
        if not self.name or not self.name.strip():
            raise ValueError("Subclass must have a name")
    
    def add_feature(self, feature: ClassFeature) -> None:
        """Add subclass feature with validation"""
        # Business rule: Subclass features typically start at level 3
        if feature.level.value < 3:
            raise ValueError("Subclass features typically start at level 3 or higher")
        
        self.features.append(feature)
        self.features.sort(key=lambda f: f.level.value)


@dataclass
class Spell:
    """Spell entity for shared use"""
    id: EntityId
    name: str
    level: int  # 0-9 (0 = cantrip)
    school: str
    classes: List[str]  # Classes that can learn this spell
    
    def __post_init__(self):
        if not 0 <= self.level <= 9:
            raise ValueError("Spell level must be 0-9")
        if not self.classes:
            raise ValueError("Spell must be available to at least one class")
    
    def is_cantrip(self) -> bool:
        """Check if spell is a cantrip"""
        return self.level == 0
    
    def can_be_cast_by(self, class_name: str) -> bool:
        """Check if spell can be cast by given class"""
        return class_name.lower() in [c.lower() for c in self.classes]


# Domain Services
class ClassValidationService:
    """Domain service for class validation business logic"""
    
    @staticmethod
    def validate_class_consistency(dnd_class: DndClass) -> List[str]:
        """Validate class internal consistency"""
        errors = []
        
        # Spellcasters must have spell-related features
        if dnd_class.is_spellcaster():
            spell_features = [f for f in dnd_class.features if "incantesimi" in f.name.lower()]
            if not spell_features:
                errors.append("Spellcasting class missing spell-related features")
        
        # Full casters should have primary mental ability
        if dnd_class.is_full_caster():
            mental_abilities = [Ability.INTELLIGENZA, Ability.SAGGEZZA, Ability.CARISMA]
            if dnd_class.primary_ability not in mental_abilities:
                errors.append("Full caster should have mental primary ability")
        
        # Classes should have features at levels 1, 2, 3
        core_levels = {f.level.value for f in dnd_class.get_core_features()}
        missing_levels = {1, 2, 3} - core_levels
        if missing_levels:
            errors.append(f"Missing core features at levels: {sorted(missing_levels)}")
        
        return errors
    
    @staticmethod
    def suggest_missing_data(dnd_class: DndClass) -> List[str]:
        """Suggest potentially missing data"""
        suggestions = []
        
        if not dnd_class.saving_throw_proficiencies:
            suggestions.append("Consider adding saving throw proficiencies")
        
        if not dnd_class.armor_proficiencies:
            suggestions.append("Consider adding armor proficiencies")
        
        if not dnd_class.weapon_proficiencies:
            suggestions.append("Consider adding weapon proficiencies")
        
        return suggestions


# CQRS Query Side Repository
class ClassQueryRepository(ABC):
    """Repository interface for read-side operations (CQRS Query)"""
    
    @abstractmethod
    async def search_classes(self, query: 'ClassSearchQuery') -> List['ClassSummary']:
        """Search classes with filtering"""
        pass
    
    @abstractmethod
    async def get_class_detail(self, class_id: ClassId) -> Optional['ClassDetail']:
        """Get detailed class information for viewing"""
        pass
    
    @abstractmethod
    async def get_classes_by_ability(self, primary_ability: str) -> List['ClassSummary']:
        """Get all classes with specific primary ability"""
        pass
    
    @abstractmethod
    async def get_spellcasting_classes(self) -> List['ClassSummary']:
        """Get all classes with spellcasting progression"""
        pass
    
    @abstractmethod  
    async def get_class_features_by_level(self, class_id: ClassId, level: int) -> List[Dict[str, Any]]:
        """Get class features available at specific level"""
        pass


# Event Publishing Interface
class EventPublisher(ABC):
    """Interface for publishing domain events"""
    
    @abstractmethod
    async def publish(self, event: DomainEvent) -> None:
        """Publish a domain event"""
        pass