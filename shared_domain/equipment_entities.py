"""
Equipment entities for D&D 5e SRD following ADR data model
Includes weapons, armor, tools, magic items, and general equipment
"""
from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, List, Optional, Any, Union
import re


class WeaponCategory(Enum):
    """Weapon categories"""
    SIMPLE_MELEE = "Semplice da Mischia"
    SIMPLE_RANGED = "Semplice a Distanza"
    MARTIAL_MELEE = "Da Guerra da Mischia"
    MARTIAL_RANGED = "Da Guerra a Distanza"


class WeaponProperty(Enum):
    """Weapon properties"""
    ACCURATE = "Accurata"
    AMMUNITION = "Munizioni"
    FINESSE = "Elegante"
    HEAVY = "Pesante"
    LIGHT = "Leggera"
    LOADING = "Ricarica"
    RANGE = "Gittata"
    REACH = "Portata"
    SPECIAL = "Speciale"
    THROWN = "Da Lancio"
    TWO_HANDED = "A Due Mani"
    VERSATILE = "Versatile"


class ArmorCategory(Enum):
    """Armor categories"""
    LIGHT = "Leggera"
    MEDIUM = "Media"
    HEAVY = "Pesante"
    SHIELD = "Scudo"


class MagicItemRarity(Enum):
    """Magic item rarity"""
    COMMON = "Comune"
    UNCOMMON = "Non Comune"
    RARE = "Raro"
    VERY_RARE = "Molto Raro"
    LEGENDARY = "Leggendario"
    ARTIFACT = "Artefatto"


@dataclass(frozen=True)
class EquipmentId:
    """Equipment entity identifier"""
    value: str
    
    def __post_init__(self):
        if not self.value or not self.value.strip():
            raise ValueError("EquipmentId cannot be empty")
        if not re.match(r'^[a-z][a-z0-9-]*$', self.value):
            raise ValueError(f"Invalid equipment ID format: {self.value}")


@dataclass(frozen=True)
class Currency:
    """Currency amount in gold pieces"""
    gold_pieces: Union[int, float, str]
    
    def __post_init__(self):
        if isinstance(self.gold_pieces, str):
            # Handle cases like "2,000 mo" or "1/2 mo"
            if not re.match(r'^[\d,./\s]+\s*(mo|ma|me|mr|mp)?$', self.gold_pieces.lower()):
                raise ValueError(f"Invalid currency format: {self.gold_pieces}")
    
    def to_gold(self) -> float:
        """Convert to gold pieces as float"""
        if isinstance(self.gold_pieces, (int, float)):
            return float(self.gold_pieces)
        
        # Parse string format
        text = self.gold_pieces.lower().strip()
        # Extract numeric part
        numeric_part = re.sub(r'[^\d,./]', '', text)
        
        if '/' in numeric_part:
            parts = numeric_part.split('/')
            return float(parts[0]) / float(parts[1])
        
        return float(numeric_part.replace(',', ''))
    
    def get_text(self) -> str:
        """Get display text"""
        if isinstance(self.gold_pieces, str):
            return self.gold_pieces
        return f"{self.gold_pieces} mo"


@dataclass(frozen=True)
class Weight:
    """Item weight in kilograms"""
    kilograms: Union[float, str]
    
    def __post_init__(self):
        if isinstance(self.kilograms, str):
            if not re.match(r'^[\d,./\s]+\s*kg?$', self.kilograms.lower()):
                raise ValueError(f"Invalid weight format: {self.kilograms}")
    
    def to_kg(self) -> float:
        """Convert to kilograms as float"""
        if isinstance(self.kilograms, (int, float)):
            return float(self.kilograms)
        
        # Parse string format
        numeric_part = re.sub(r'[^\d,./]', '', self.kilograms)
        if '/' in numeric_part:
            parts = numeric_part.split('/')
            return float(parts[0]) / float(parts[1])
        
        return float(numeric_part.replace(',', ''))
    
    def get_text(self) -> str:
        """Get display text"""
        if isinstance(self.kilograms, str):
            return self.kilograms
        return f"{self.kilograms} kg"


@dataclass
class DamageInfo:
    """Weapon damage information"""
    dice: str  # e.g., "1d8"
    damage_type: str  # e.g., "Perforante"
    versatile_dice: Optional[str] = None  # e.g., "1d10" for versatile
    
    def __post_init__(self):
        if not self.dice.strip():
            raise ValueError("Damage dice cannot be empty")
        if not self.damage_type.strip():
            raise ValueError("Damage type cannot be empty")
    
    def get_damage_text(self) -> str:
        """Get damage display text"""
        base_damage = f"{self.dice} {self.damage_type}"
        if self.versatile_dice:
            return f"{base_damage} ({self.versatile_dice} versatile)"
        return base_damage


@dataclass
class WeaponRange:
    """Weapon range information"""
    normale: Optional[str] = None  # e.g., "6 m"
    lunga: Optional[str] = None   # e.g., "18 m"
    
    def get_range_text(self) -> str:
        """Get range display text"""
        if self.normale and self.lunga:
            return f"{self.normale}/{self.lunga}"
        elif self.normale:
            return self.normale
        return "—"


@dataclass
class Weapon:
    """D&D 5e Weapon entity"""
    id: EquipmentId
    nome: str
    costo: Currency
    peso: Weight
    danno: DamageInfo
    categoria: WeaponCategory
    proprieta: List[WeaponProperty]
    maestria: str
    contenuto_markdown: str
    
    # Optional fields
    gittata: Optional[WeaponRange] = None
    descrizione: Optional[str] = None
    fonte: str = "SRD"
    versione: str = "1.0"
    
    def __post_init__(self):
        if not self.nome.strip():
            raise ValueError("Weapon name cannot be empty")
        if not self.maestria.strip():
            raise ValueError("Weapon mastery cannot be empty")
    
    def is_melee(self) -> bool:
        """Check if weapon is melee"""
        return self.categoria in [WeaponCategory.SIMPLE_MELEE, WeaponCategory.MARTIAL_MELEE]
    
    def is_ranged(self) -> bool:
        """Check if weapon is ranged"""
        return self.categoria in [WeaponCategory.SIMPLE_RANGED, WeaponCategory.MARTIAL_RANGED]
    
    def is_martial(self) -> bool:
        """Check if weapon is martial"""
        return self.categoria in [WeaponCategory.MARTIAL_MELEE, WeaponCategory.MARTIAL_RANGED]
    
    def has_property(self, prop: WeaponProperty) -> bool:
        """Check if weapon has specific property"""
        return prop in self.proprieta
    
    def is_versatile(self) -> bool:
        """Check if weapon is versatile"""
        return self.has_property(WeaponProperty.VERSATILE)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "id": self.id.value,
            "nome": self.nome,
            "categoria": self.categoria.value,
            "is_melee": self.is_melee(),
            "is_ranged": self.is_ranged(),
            "is_martial": self.is_martial(),
            "is_versatile": self.is_versatile(),
            "proprieta": [prop.value for prop in self.proprieta],
            "fonte": self.fonte,
            "versione": self.versione
        }


@dataclass
class Armor:
    """D&D 5e Armor entity"""
    id: EquipmentId
    nome: str
    costo: Currency
    peso: Weight
    classe_armatura: str  # e.g., "11 + mod Des"
    categoria: ArmorCategory
    contenuto_markdown: str
    
    # Optional fields
    forza_richiesta: Optional[int] = None
    svantaggio_furtivita: bool = False
    descrizione: Optional[str] = None
    fonte: str = "SRD"
    versione: str = "1.0"
    
    def __post_init__(self):
        if not self.nome.strip():
            raise ValueError("Armor name cannot be empty")
        if not self.classe_armatura.strip():
            raise ValueError("AC cannot be empty")
    
    def is_shield(self) -> bool:
        """Check if this is a shield"""
        return self.categoria == ArmorCategory.SHIELD
    
    def requires_strength(self) -> bool:
        """Check if armor has strength requirement"""
        return self.forza_richiesta is not None
    
    def imposes_stealth_disadvantage(self) -> bool:
        """Check if armor imposes stealth disadvantage"""
        return self.svantaggio_furtivita
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "id": self.id.value,
            "nome": self.nome,
            "categoria": self.categoria.value,
            "is_shield": self.is_shield(),
            "requires_strength": self.requires_strength(),
            "imposes_stealth_disadvantage": self.imposes_stealth_disadvantage(),
            "fonte": self.fonte,
            "versione": self.versione
        }


@dataclass
class Tool:
    """D&D 5e Tool entity"""
    id: EquipmentId
    nome: str
    costo: Currency
    peso: Weight
    categoria: str  # e.g., "Strumenti da Artigiano", "Kit"
    contenuto_markdown: str
    
    # Optional fields
    descrizione: Optional[str] = None
    abilita_associate: List[str] = field(default_factory=list)
    fonte: str = "SRD"
    versione: str = "1.0"
    
    def __post_init__(self):
        if not self.nome.strip():
            raise ValueError("Tool name cannot be empty")
        if not self.categoria.strip():
            raise ValueError("Tool category cannot be empty")
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "id": self.id.value,
            "nome": self.nome,
            "categoria": self.categoria,
            "abilita_associate": self.abilita_associate,
            "fonte": self.fonte,
            "versione": self.versione
        }


@dataclass
class AdventuringGear:
    """D&D 5e General equipment entity"""
    id: EquipmentId
    nome: str
    costo: Currency
    peso: Weight
    categoria: str  # e.g., "Equipaggiamento di Avventura"
    contenuto_markdown: str
    
    # Optional fields
    descrizione: Optional[str] = None
    confezione: Optional[str] = None  # e.g., "50 pezzi"
    fonte: str = "SRD"
    versione: str = "1.0"
    
    def __post_init__(self):
        if not self.nome.strip():
            raise ValueError("Equipment name cannot be empty")
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "id": self.id.value,
            "nome": self.nome,
            "categoria": self.categoria,
            "fonte": self.fonte,
            "versione": self.versione
        }


@dataclass
class MagicItem:
    """D&D 5e Magic Item entity"""
    id: EquipmentId
    nome: str
    tipo: str  # e.g., "Armor (Any Medium or Heavy, Except Hide Armor)"
    rarita: MagicItemRarity
    sintonizzazione: bool
    contenuto_markdown: str
    
    # Optional fields
    descrizione: Optional[str] = None
    costo_stimato: Optional[Currency] = None
    slot_equipaggiamento: Optional[str] = None  # e.g., "Torso", "Mani"
    scuola_magica: Optional[str] = None
    fonte: str = "SRD"
    versione: str = "1.0"
    
    def __post_init__(self):
        if not self.nome.strip():
            raise ValueError("Magic item name cannot be empty")
        if not self.tipo.strip():
            raise ValueError("Magic item type cannot be empty")
    
    def requires_attunement(self) -> bool:
        """Check if item requires attunement"""
        return self.sintonizzazione
    
    def is_consumable(self) -> bool:
        """Check if item is consumable (rough heuristic)"""
        consumable_types = ["potion", "scroll", "pozione", "pergamena"]
        return any(ctype in self.tipo.lower() for ctype in consumable_types)
    
    def get_rarity_tier(self) -> int:
        """Get rarity as numeric tier (1-6)"""
        rarity_tiers = {
            MagicItemRarity.COMMON: 1,
            MagicItemRarity.UNCOMMON: 2,
            MagicItemRarity.RARE: 3,
            MagicItemRarity.VERY_RARE: 4,
            MagicItemRarity.LEGENDARY: 5,
            MagicItemRarity.ARTIFACT: 6
        }
        return rarity_tiers.get(self.rarita, 1)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "id": self.id.value,
            "nome": self.nome,
            "tipo": self.tipo,
            "rarita": self.rarita.value,
            "rarity_tier": self.get_rarity_tier(),
            "requires_attunement": self.requires_attunement(),
            "is_consumable": self.is_consumable(),
            "fonte": self.fonte,
            "versione": self.versione
        }


# Repository interfaces
class EquipmentRepository(ABC):
    """Repository interface for equipment write operations"""
    
    @abstractmethod
    async def find_weapon_by_id(self, weapon_id: EquipmentId) -> Optional[Weapon]:
        pass
    
    @abstractmethod
    async def find_armor_by_id(self, armor_id: EquipmentId) -> Optional[Armor]:
        pass
    
    @abstractmethod
    async def find_magic_item_by_id(self, item_id: EquipmentId) -> Optional[MagicItem]:
        pass
    
    @abstractmethod
    async def save_weapon(self, weapon: Weapon) -> None:
        pass
    
    @abstractmethod
    async def save_armor(self, armor: Armor) -> None:
        pass
    
    @abstractmethod
    async def save_magic_item(self, item: MagicItem) -> None:
        pass


class EquipmentQueryRepository(ABC):
    """Repository interface for equipment read operations (CQRS Query)"""
    
    @abstractmethod
    async def search_weapons(self, query: 'WeaponSearchQuery') -> List['WeaponSummary']:
        pass
    
    @abstractmethod
    async def search_armor(self, query: 'ArmorSearchQuery') -> List['ArmorSummary']:
        pass
    
    @abstractmethod
    async def search_magic_items(self, query: 'MagicItemSearchQuery') -> List['MagicItemSummary']:
        pass
    
    @abstractmethod
    async def get_weapons_by_category(self, category: WeaponCategory) -> List['WeaponSummary']:
        pass
    
    @abstractmethod
    async def get_armor_by_category(self, category: ArmorCategory) -> List['ArmorSummary']:
        pass
    
    @abstractmethod
    async def get_magic_items_by_rarity(self, rarity: MagicItemRarity) -> List['MagicItemSummary']:
        pass


@dataclass
class EquipmentValidationService:
    """Domain service for equipment validation"""
    
    @staticmethod
    def validate_weapon(weapon: Weapon) -> List[str]:
        """Validate weapon business rules"""
        errors = []
        
        # Ranged weapons should have range
        if weapon.is_ranged() and not weapon.gittata:
            errors.append("Ranged weapons must have range specified")
        
        # Versatile weapons should have versatile damage
        if weapon.is_versatile() and not weapon.danno.versatile_dice:
            errors.append("Versatile weapons must specify versatile damage")
        
        # Heavy weapons shouldn't be light
        if weapon.has_property(WeaponProperty.HEAVY) and weapon.has_property(WeaponProperty.LIGHT):
            errors.append("Weapons cannot be both Heavy and Light")
        
        # Two-handed weapons shouldn't be light
        if weapon.has_property(WeaponProperty.TWO_HANDED) and weapon.has_property(WeaponProperty.LIGHT):
            errors.append("Two-handed weapons cannot be Light")
        
        return errors
    
    @staticmethod
    def validate_magic_item(item: MagicItem) -> List[str]:
        """Validate magic item business rules"""
        errors = []
        
        # Legendary items usually require attunement
        if item.rarita == MagicItemRarity.LEGENDARY and not item.requires_attunement():
            errors.append("Most legendary items require attunement")
        
        # Common items rarely require attunement
        if item.rarita == MagicItemRarity.COMMON and item.requires_attunement():
            errors.append("Common items rarely require attunement")
        
        return errors