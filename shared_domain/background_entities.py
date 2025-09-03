"""
Background and Feat entities for D&D 5e SRD following ADR data model
"""
from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, List, Optional, Any
import re


class AbilityName(Enum):
    """Ability score names"""
    FORZA = "Forza"
    DESTREZZA = "Destrezza"
    COSTITUZIONE = "Costituzione"
    INTELLIGENZA = "Intelligenza"
    SAGGEZZA = "Saggezza"
    CARISMA = "Carisma"


class FeatCategory(Enum):
    """Feat categories"""
    ORIGIN = "Talento di Origine"
    GENERAL = "Talento Generale"
    FIGHTING_STYLE = "Stile di Combattimento"
    EPIC_BOON = "Dono Epico"


@dataclass(frozen=True)
class BackgroundId:
    """Background entity identifier"""
    value: str
    
    def __post_init__(self):
        if not self.value or not self.value.strip():
            raise ValueError("BackgroundId cannot be empty")
        if not re.match(r'^[a-z][a-z0-9-]*$', self.value):
            raise ValueError(f"Invalid background ID format: {self.value}")


@dataclass(frozen=True)
class FeatId:
    """Feat entity identifier"""
    value: str
    
    def __post_init__(self):
        if not self.value or not self.value.strip():
            raise ValueError("FeatId cannot be empty")
        if not re.match(r'^[a-z][a-z0-9-]*$', self.value):
            raise ValueError(f"Invalid feat ID format: {self.value}")


@dataclass
class EquipmentOption:
    """Equipment choice option"""
    etichetta: str
    oggetti: List[str]
    
    def __post_init__(self):
        if not self.etichetta.strip():
            raise ValueError("Equipment option label cannot be empty")
        if not self.oggetti:
            raise ValueError("Equipment option must have items")


@dataclass
class Background:
    """D&D 5e Background entity"""
    id: BackgroundId
    nome: str
    punteggi_caratteristica: List[AbilityName]
    talento: str
    abilita_competenze: List[str]
    strumenti_competenze: List[str]
    equipaggiamento_iniziale_opzioni: List[EquipmentOption]
    contenuto_markdown: str
    
    # Optional fields
    descrizione: Optional[str] = None
    linguaggi: List[str] = field(default_factory=list)
    caratteristiche_speciali: List[str] = field(default_factory=list)
    fonte: str = "SRD"
    versione: str = "1.0"
    
    def __post_init__(self):
        if not self.nome.strip():
            raise ValueError("Background name cannot be empty")
        if len(self.punteggi_caratteristica) < 2 or len(self.punteggi_caratteristica) > 3:
            raise ValueError("Background must have 2-3 ability score options")
        if not self.talento.strip():
            raise ValueError("Background must have a feat")
        if len(self.abilita_competenze) < 1:
            raise ValueError("Background must grant at least one skill proficiency")
        if not self.equipaggiamento_iniziale_opzioni:
            raise ValueError("Background must have equipment options")
    
    def get_ability_names(self) -> List[str]:
        """Get ability score names as strings"""
        return [ability.value for ability in self.punteggi_caratteristica]
    
    def grants_language_choice(self) -> bool:
        """Check if background grants language choice"""
        return bool(self.linguaggi)
    
    def has_special_features(self) -> bool:
        """Check if background has special features"""
        return bool(self.caratteristiche_speciali)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "id": self.id.value,
            "nome": self.nome,
            "punteggi_caratteristica": self.get_ability_names(),
            "talento": self.talento,
            "abilita_competenze_count": len(self.abilita_competenze),
            "strumenti_competenze_count": len(self.strumenti_competenze),
            "grants_language_choice": self.grants_language_choice(),
            "has_special_features": self.has_special_features(),
            "fonte": self.fonte,
            "versione": self.versione
        }


@dataclass
class FeatBenefit:
    """Individual benefit granted by a feat"""
    nome: str
    descrizione: str
    tipo: str = "Beneficio"  # "Beneficio", "Aumento Caratteristica", etc.
    
    def __post_init__(self):
        if not self.nome.strip():
            raise ValueError("Benefit name cannot be empty")
        if not self.descrizione.strip():
            raise ValueError("Benefit description cannot be empty")


@dataclass
class AbilityScoreIncrease:
    """Ability score increase from feat"""
    abilita: List[AbilityName]
    aumento: int = 1
    scelta_multipla: bool = False  # Can choose different abilities
    
    def __post_init__(self):
        if not self.abilita:
            raise ValueError("Must specify at least one ability")
        if not 1 <= self.aumento <= 2:
            raise ValueError("Ability increase must be 1 or 2")
    
    def get_ability_names(self) -> List[str]:
        """Get ability names as strings"""
        return [ability.value for ability in self.abilita]
    
    def get_text(self) -> str:
        """Get increase description text"""
        if len(self.abilita) == 1:
            return f"+{self.aumento} {self.abilita[0].value}"
        elif self.scelta_multipla:
            abilities_text = " o ".join(self.get_ability_names())
            return f"+{self.aumento} {abilities_text} (a scelta)"
        else:
            abilities_text = " e ".join(self.get_ability_names())
            return f"+{self.aumento} {abilities_text}"


@dataclass
class Feat:
    """D&D 5e Feat entity"""
    id: FeatId
    nome: str
    categoria: FeatCategory
    prerequisiti: str
    benefici: List[FeatBenefit]
    contenuto_markdown: str
    
    # Optional fields
    aumenti_caratteristica: List[AbilityScoreIncrease] = field(default_factory=list)
    descrizione: Optional[str] = None
    livello_minimo: Optional[int] = None
    restrizioni_classe: List[str] = field(default_factory=list)
    restrizioni_razza: List[str] = field(default_factory=list)
    fonte: str = "SRD"
    versione: str = "1.0"
    
    def __post_init__(self):
        if not self.nome.strip():
            raise ValueError("Feat name cannot be empty")
        if not self.benefici:
            raise ValueError("Feat must have at least one benefit")
    
    def has_prerequisites(self) -> bool:
        """Check if feat has prerequisites"""
        return bool(self.prerequisiti and self.prerequisiti.strip())
    
    def grants_ability_increase(self) -> bool:
        """Check if feat grants ability score increases"""
        return bool(self.aumenti_caratteristica)
    
    def is_origin_feat(self) -> bool:
        """Check if this is an origin feat"""
        return self.categoria == FeatCategory.ORIGIN
    
    def is_epic_boon(self) -> bool:
        """Check if this is an epic boon"""
        return self.categoria == FeatCategory.EPIC_BOON
    
    def is_fighting_style(self) -> bool:
        """Check if this is a fighting style"""
        return self.categoria == FeatCategory.FIGHTING_STYLE
    
    def has_class_restrictions(self) -> bool:
        """Check if feat is restricted to certain classes"""
        return bool(self.restrizioni_classe)
    
    def has_race_restrictions(self) -> bool:
        """Check if feat is restricted to certain races"""
        return bool(self.restrizioni_razza)
    
    def is_available_to_class(self, class_name: str) -> bool:
        """Check if feat is available to specific class"""
        if not self.has_class_restrictions():
            return True
        return class_name in self.restrizioni_classe
    
    def get_total_ability_increases(self) -> int:
        """Get total ability score increases granted"""
        return sum(inc.aumento for inc in self.aumenti_caratteristica)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "id": self.id.value,
            "nome": self.nome,
            "categoria": self.categoria.value,
            "has_prerequisites": self.has_prerequisites(),
            "grants_ability_increase": self.grants_ability_increase(),
            "total_ability_increases": self.get_total_ability_increases(),
            "is_origin_feat": self.is_origin_feat(),
            "is_epic_boon": self.is_epic_boon(),
            "has_class_restrictions": self.has_class_restrictions(),
            "has_race_restrictions": self.has_race_restrictions(),
            "benefici_count": len(self.benefici),
            "fonte": self.fonte,
            "versione": self.versione
        }


@dataclass
class Service:
    """Service entity (from services collection)"""
    id: EquipmentId  # Reuse equipment ID format
    nome: str
    costo: str  # Services have varied cost formats
    categoria: str
    contenuto_markdown: str
    
    # Optional fields
    descrizione: Optional[str] = None
    durata: Optional[str] = None
    disponibilita: Optional[str] = None
    fonte: str = "SRD"
    versione: str = "1.0"
    
    def __post_init__(self):
        if not self.nome.strip():
            raise ValueError("Service name cannot be empty")
        if not self.costo.strip():
            raise ValueError("Service cost cannot be empty")
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "id": self.id.value,
            "nome": self.nome,
            "costo": self.costo,
            "categoria": self.categoria,
            "fonte": self.fonte,
            "versione": self.versione
        }


# Repository interfaces
class BackgroundRepository(ABC):
    """Repository interface for background write operations"""
    
    @abstractmethod
    async def find_by_id(self, background_id: BackgroundId) -> Optional[Background]:
        pass
    
    @abstractmethod
    async def find_by_name(self, name: str) -> Optional[Background]:
        pass
    
    @abstractmethod
    async def save(self, background: Background) -> None:
        pass
    
    @abstractmethod
    async def find_all(self) -> List[Background]:
        pass


class FeatRepository(ABC):
    """Repository interface for feat write operations"""
    
    @abstractmethod
    async def find_by_id(self, feat_id: FeatId) -> Optional[Feat]:
        pass
    
    @abstractmethod
    async def find_by_name(self, name: str) -> Optional[Feat]:
        pass
    
    @abstractmethod
    async def find_by_category(self, category: FeatCategory) -> List[Feat]:
        pass
    
    @abstractmethod
    async def save(self, feat: Feat) -> None:
        pass
    
    @abstractmethod
    async def find_all(self) -> List[Feat]:
        pass


class BackgroundQueryRepository(ABC):
    """Repository interface for background read operations (CQRS Query)"""
    
    @abstractmethod
    async def search_backgrounds(self, query: 'BackgroundSearchQuery') -> List['BackgroundSummary']:
        pass
    
    @abstractmethod
    async def get_background_detail(self, background_id: BackgroundId) -> Optional['BackgroundDetail']:
        pass
    
    @abstractmethod
    async def get_backgrounds_by_feat(self, feat_name: str) -> List['BackgroundSummary']:
        pass
    
    @abstractmethod
    async def get_backgrounds_by_skill(self, skill_name: str) -> List['BackgroundSummary']:
        pass


class FeatQueryRepository(ABC):
    """Repository interface for feat read operations (CQRS Query)"""
    
    @abstractmethod
    async def search_feats(self, query: 'FeatSearchQuery') -> List['FeatSummary']:
        pass
    
    @abstractmethod
    async def get_feat_detail(self, feat_id: FeatId) -> Optional['FeatDetail']:
        pass
    
    @abstractmethod
    async def get_feats_by_category(self, category: FeatCategory) -> List['FeatSummary']:
        pass
    
    @abstractmethod
    async def get_origin_feats(self) -> List['FeatSummary']:
        pass
    
    @abstractmethod
    async def get_epic_boons(self) -> List['FeatSummary']:
        pass


@dataclass
class BackgroundValidationService:
    """Domain service for background validation"""
    
    @staticmethod
    def validate_background(background: Background) -> List[str]:
        """Validate background business rules"""
        errors = []
        
        # Should have reasonable number of skills
        if len(background.abilita_competenze) > 4:
            errors.append("Backgrounds rarely grant more than 4 skill proficiencies")
        
        # Equipment options should be reasonable
        if len(background.equipaggiamento_iniziale_opzioni) > 3:
            errors.append("Too many equipment options")
        
        # Should not duplicate ability scores
        abilities = background.get_ability_names()
        if len(abilities) != len(set(abilities)):
            errors.append("Background cannot have duplicate ability score options")
        
        return errors


@dataclass
class FeatValidationService:
    """Domain service for feat validation"""
    
    @staticmethod
    def validate_feat(feat: Feat) -> List[str]:
        """Validate feat business rules"""
        errors = []
        
        # Epic boons should not have prerequisites
        if feat.is_epic_boon() and feat.has_prerequisites():
            errors.append("Epic boons typically do not have prerequisites")
        
        # Origin feats should grant ability increases
        if feat.is_origin_feat() and not feat.grants_ability_increase():
            errors.append("Origin feats typically grant ability score increases")
        
        # Ability increases should be reasonable
        total_increases = feat.get_total_ability_increases()
        if total_increases > 2:
            errors.append("Feats rarely grant more than +2 total ability increases")
        
        # Fighting styles should be restricted to appropriate classes
        if feat.is_fighting_style() and not feat.has_class_restrictions():
            errors.append("Fighting styles should be restricted to martial classes")
        
        return errors