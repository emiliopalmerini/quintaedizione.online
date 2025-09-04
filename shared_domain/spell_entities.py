"""
Spell entities for D&D 5e SRD following ADR data model
"""
from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, List, Optional, Any
import re


class SpellSchool(Enum):
    """Schools of magic"""
    ABJURATION = "Abiurazione"
    CONJURATION = "Evocazione"
    DIVINATION = "Divinazione"
    ENCHANTMENT = "Ammaliamento"
    EVOCATION = "Invocazione"
    ILLUSION = "Illusione"
    NECROMANCY = "Necromanzia"
    TRANSMUTATION = "Trasmutazione"


class CastingTime(Enum):
    """Spell casting times"""
    ACTION = "1 azione"
    BONUS_ACTION = "1 azione bonus"
    REACTION = "1 reazione"
    RITUAL = "10 minuti (rituale)"
    MINUTE_1 = "1 minuto"
    MINUTE_10 = "10 minuti"
    HOUR_1 = "1 ora"
    HOUR_8 = "8 ore"
    HOUR_24 = "24 ore"


class SpellRange(Enum):
    """Spell ranges"""
    SELF = "Personale"
    TOUCH = "Contatto"
    FEET_30 = "9 metri"
    FEET_60 = "18 metri"
    FEET_90 = "27 metri"
    FEET_120 = "36 metri"
    FEET_150 = "45 metri"
    FEET_300 = "90 metri"
    FEET_500 = "150 metri"
    FEET_1000 = "300 metri"
    MILE_1 = "1,5 chilometri"
    UNLIMITED = "Illimitata"
    SIGHT = "A vista"


class SpellDuration(Enum):
    """Spell durations"""
    INSTANTANEOUS = "Istantanea"
    ROUND_1 = "1 round"
    MINUTE_1 = "1 minuto"
    MINUTE_10 = "10 minuti"
    HOUR_1 = "1 ora"
    HOUR_8 = "8 ore"
    HOUR_24 = "24 ore"
    DAY_7 = "7 giorni"
    DAY_30 = "30 giorni"
    PERMANENT = "Permanente"
    CONCENTRATION = "Concentrazione"


@dataclass(frozen=True)
class SpellId:
    """Spell entity identifier"""
    value: str
    
    def __post_init__(self):
        if not self.value or not self.value.strip():
            raise ValueError("SpellId cannot be empty")
        if not re.match(r'^[a-z][a-z0-9-]*$', self.value):
            raise ValueError(f"Invalid spell ID format: {self.value}")


@dataclass(frozen=True)
class SpellLevel:
    """Spell level (0-9)"""
    value: int
    
    def __post_init__(self):
        if not 0 <= self.value <= 9:
            raise ValueError(f"Spell level must be 0-9, got {self.value}")
    
    def is_cantrip(self) -> bool:
        return self.value == 0


@dataclass(frozen=True)
class SpellComponent:
    """Spell casting component"""
    type: str  # "V", "S", "M"
    description: Optional[str] = None
    cost_gp: Optional[int] = None
    consumed: bool = False
    
    def __post_init__(self):
        valid_types = {"V", "S", "M"}
        if self.type not in valid_types:
            raise ValueError(f"Invalid component type: {self.type}")


@dataclass
class SpellCasting:
    """Spell casting information"""
    tempo: CastingTime
    gittata: SpellRange
    durata: SpellDuration
    gittata_custom: Optional[str] = None
    componenti: List[SpellComponent] = field(default_factory=list)
    durata_custom: Optional[str] = None
    concentrazione: bool = False
    rituale: bool = False
    
    def get_range_text(self) -> str:
        return self.gittata_custom or self.gittata.value
    
    def get_duration_text(self) -> str:
        if self.concentrazione:
            base_duration = self.durata_custom or self.durata.value
            return f"Concentrazione, fino a {base_duration.lower()}"
        return self.durata_custom or self.durata.value
    
    def get_components_text(self) -> str:
        components = []
        for comp in self.componenti:
            if comp.description:
                components.append(f"{comp.type} ({comp.description})")
            else:
                components.append(comp.type)
        return ", ".join(components)


@dataclass
class Spell:
    """D&D 5e Spell entity"""
    id: SpellId
    nome: str
    livello: SpellLevel
    scuola: SpellSchool
    classi: List[str]
    lancio: SpellCasting
    descrizione: str
    contenuto_markdown: str
    
    # Optional fields
    sottoscuole: List[str] = field(default_factory=list)
    livelli_superiori: Optional[str] = None
    fonte: str = "SRD"
    versione: str = "1.0"
    
    def __post_init__(self):
        if not self.nome.strip():
            raise ValueError("Spell name cannot be empty")
        if not self.classi:
            raise ValueError("Spell must have at least one class")
        if not self.descrizione.strip():
            raise ValueError("Spell description cannot be empty")
    
    def is_cantrip(self) -> bool:
        return self.livello.is_cantrip()
    
    def is_ritual(self) -> bool:
        return self.lancio.rituale
    
    def requires_concentration(self) -> bool:
        return self.lancio.concentrazione
    
    def has_material_component(self) -> bool:
        return any(comp.type == "M" for comp in self.lancio.componenti)
    
    def get_expensive_components(self) -> List[SpellComponent]:
        """Get components with gold cost"""
        return [comp for comp in self.lancio.componenti 
                if comp.type == "M" and comp.cost_gp]
    
    def is_available_to_class(self, class_name: str) -> bool:
        """Check if spell is available to specific class"""
        return class_name in self.classi
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "id": self.id.value,
            "nome": self.nome,
            "livello": self.livello.value,
            "scuola": self.scuola.value,
            "is_cantrip": self.is_cantrip(),
            "is_ritual": self.is_ritual(),
            "requires_concentration": self.requires_concentration(),
            "classi": self.classi,
            "fonte": self.fonte,
            "versione": self.versione
        }


# Repository interface for spells
class SpellRepository(ABC):
    """Repository interface for spell write operations"""
    
    @abstractmethod
    async def find_by_id(self, spell_id: SpellId) -> Optional[Spell]:
        pass
    
    @abstractmethod
    async def find_by_name(self, name: str) -> Optional[Spell]:
        pass
    
    @abstractmethod
    async def find_by_class(self, class_name: str) -> List[Spell]:
        pass
    
    @abstractmethod
    async def find_by_level(self, level: SpellLevel) -> List[Spell]:
        pass
    
    @abstractmethod
    async def find_by_school(self, school: SpellSchool) -> List[Spell]:
        pass
    
    @abstractmethod
    async def save(self, spell: Spell) -> None:
        pass
    
    @abstractmethod
    async def find_all(self) -> List[Spell]:
        pass


class SpellQueryRepository(ABC):
    """Repository interface for spell read operations (CQRS Query)"""
    
    @abstractmethod
    async def search_spells(self, query: 'SpellSearchQuery') -> List['SpellSummary']:
        pass
    
    @abstractmethod
    async def get_spell_detail(self, spell_id: SpellId) -> Optional['SpellDetail']:
        pass
    
    @abstractmethod
    async def get_spells_by_class_and_level(self, class_name: str, level: SpellLevel) -> List['SpellSummary']:
        pass
    
    @abstractmethod
    async def get_cantrips_by_class(self, class_name: str) -> List['SpellSummary']:
        pass
    
    @abstractmethod
    async def get_ritual_spells(self) -> List['SpellSummary']:
        pass


@dataclass
class SpellValidationService:
    """Domain service for spell validation"""
    
    @staticmethod
    def validate_spell(spell: Spell) -> List[str]:
        """Validate spell business rules"""
        errors = []
        
        # Cantrips shouldn't have higher level effects
        if spell.is_cantrip() and spell.livelli_superiori:
            errors.append("Cantrips cannot have higher level effects")
        
        # Concentration spells should have reasonable durations
        if spell.requires_concentration():
            instant_duration = spell.lancio.durata == SpellDuration.INSTANTANEOUS
            if instant_duration:
                errors.append("Concentration spells cannot have instantaneous duration")
        
        # Ritual spells should be reasonable levels
        if spell.is_ritual() and spell.livello.value > 6:
            errors.append("High-level ritual spells are unusual")
        
        # Material components with cost should be described
        expensive_components = spell.get_expensive_components()
        for comp in expensive_components:
            if not comp.description:
                errors.append(f"Expensive material component lacks description")
        
        return errors
    
    @staticmethod
    def suggest_missing_data(spell: Spell) -> List[str]:
        """Suggest potentially missing data"""
        suggestions = []
        
        if not spell.sottoscuole and spell.scuola in [SpellSchool.ENCHANTMENT, SpellSchool.ILLUSION]:
            suggestions.append("Consider adding subschool information")
        
        if not spell.livelli_superiori and spell.livello.value > 0 and spell.livello.value < 9:
            suggestions.append("Consider adding higher level effects")
        
        if spell.is_ritual() and not any(comp.type == "M" for comp in spell.lancio.componenti):
            suggestions.append("Most ritual spells have material components")
        
        return suggestions