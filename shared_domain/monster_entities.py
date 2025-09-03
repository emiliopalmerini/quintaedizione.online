"""
Monster entities for D&D 5e SRD following ADR data model
"""
from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, List, Optional, Any, Union
import re


class MonsterSize(Enum):
    """Creature sizes"""
    TINY = "Minuscola"
    SMALL = "Piccola"
    MEDIUM = "Media"
    LARGE = "Grande"
    HUGE = "Enorme"
    GARGANTUAN = "Mastodontica"


class MonsterType(Enum):
    """Creature types"""
    ABERRATION = "Aberrazione"
    BEAST = "Bestia"
    CELESTIAL = "Celestiale"
    CONSTRUCT = "Costrutto"
    DRAGON = "Drago"
    ELEMENTAL = "Elementale"
    FEY = "Fata"
    FIEND = "Immondo"
    GIANT = "Gigante"
    HUMANOID = "Umanoide"
    MONSTROSITY = "Mostruosità"
    OOZE = "Melma"
    PLANT = "Vegetale"
    UNDEAD = "Non Morto"


class Alignment(Enum):
    """Creature alignments"""
    LAWFUL_GOOD = "Legale Buono"
    NEUTRAL_GOOD = "Neutrale Buono"
    CHAOTIC_GOOD = "Caotico Buono"
    LAWFUL_NEUTRAL = "Legale Neutrale"
    TRUE_NEUTRAL = "Neutrale Puro"
    CHAOTIC_NEUTRAL = "Caotico Neutrale"
    LAWFUL_EVIL = "Legale Malvagio"
    NEUTRAL_EVIL = "Neutrale Malvagio"
    CHAOTIC_EVIL = "Caotico Malvagio"
    UNALIGNED = "Senza Allineamento"
    ANY_ALIGNMENT = "Qualsiasi Allineamento"


class DamageType(Enum):
    """Damage types"""
    ACID = "Acido"
    BLUDGEONING = "Contundente"
    COLD = "Freddo"
    FIRE = "Fuoco"
    FORCE = "Forza"
    LIGHTNING = "Fulmine"
    NECROTIC = "Necrotico"
    PIERCING = "Perforante"
    POISON = "Veleno"
    PSYCHIC = "Psichico"
    RADIANT = "Radiante"
    SLASHING = "Tagliente"
    THUNDER = "Tuono"


class ConditionType(Enum):
    """Condition types"""
    BLINDED = "Accecato"
    CHARMED = "Affascinato"
    DEAFENED = "Assordato"
    EXHAUSTION = "Sfinimento"
    FRIGHTENED = "Spaventato"
    GRAPPLED = "Afferrato"
    INCAPACITATED = "Incapacitato"
    INVISIBLE = "Invisibile"
    PARALYZED = "Paralizzato"
    PETRIFIED = "Pietrificato"
    POISONED = "Avvelenato"
    PRONE = "Prono"
    RESTRAINED = "Trattenuto"
    STUNNED = "Stordito"
    UNCONSCIOUS = "Privo di Sensi"


@dataclass(frozen=True)
class MonsterId:
    """Monster entity identifier"""
    value: str
    
    def __post_init__(self):
        if not self.value or not self.value.strip():
            raise ValueError("MonsterId cannot be empty")
        if not re.match(r'^[a-z][a-z0-9-]*$', self.value):
            raise ValueError(f"Invalid monster ID format: {self.value}")


@dataclass(frozen=True)
class ChallengeRating:
    """Challenge Rating with XP value"""
    cr: Union[float, str]  # Can be fraction like "1/4" or number
    xp: int
    
    def __post_init__(self):
        if isinstance(self.cr, str):
            # Validate fraction format
            if not re.match(r'^\d+/\d+$', self.cr):
                raise ValueError(f"Invalid CR fraction format: {self.cr}")
        elif isinstance(self.cr, (int, float)):
            if self.cr < 0:
                raise ValueError(f"CR cannot be negative: {self.cr}")
        
        if self.xp < 0:
            raise ValueError(f"XP cannot be negative: {self.xp}")
    
    def get_cr_text(self) -> str:
        return str(self.cr)
    
    def get_numeric_cr(self) -> float:
        if isinstance(self.cr, str):
            parts = self.cr.split('/')
            return float(parts[0]) / float(parts[1])
        return float(self.cr)


@dataclass(frozen=True)
class AbilityScores:
    """Six ability scores"""
    forza: int
    destrezza: int
    costituzione: int
    intelligenza: int
    saggezza: int
    carisma: int
    
    def __post_init__(self):
        for ability, score in [
            ("Forza", self.forza), ("Destrezza", self.destrezza),
            ("Costituzione", self.costituzione), ("Intelligenza", self.intelligenza),
            ("Saggezza", self.saggezza), ("Carisma", self.carisma)
        ]:
            if not 1 <= score <= 30:
                raise ValueError(f"{ability} score must be 1-30, got {score}")
    
    def get_modifier(self, ability: str) -> int:
        """Get ability modifier"""
        score = getattr(self, ability.lower())
        return (score - 10) // 2
    
    def to_dict(self) -> Dict[str, int]:
        return {
            "str": self.forza,
            "dex": self.destrezza,
            "con": self.costituzione,
            "int": self.intelligenza,
            "wis": self.saggezza,
            "cha": self.carisma
        }


@dataclass
class DamageResistance:
    """Damage resistance/immunity/vulnerability"""
    damage_types: List[DamageType]
    conditions: Optional[str] = None  # e.g., "from nonmagical attacks"
    
    def get_text(self) -> str:
        types_text = ", ".join(dt.value for dt in self.damage_types)
        if self.conditions:
            return f"{types_text} {self.conditions}"
        return types_text


@dataclass
class Speed:
    """Creature speed"""
    walking: int = 9  # Default walking speed in meters
    flying: Optional[int] = None
    swimming: Optional[int] = None
    climbing: Optional[int] = None
    burrowing: Optional[int] = None
    hovering: bool = False
    
    def __post_init__(self):
        if self.walking < 0:
            raise ValueError("Walking speed cannot be negative")
        
        speeds = [self.flying, self.swimming, self.climbing, self.burrowing]
        for speed in speeds:
            if speed is not None and speed < 0:
                raise ValueError("Speed values cannot be negative")
    
    def get_text(self) -> str:
        """Get speed text representation"""
        parts = [f"{self.walking} m"]
        
        if self.flying:
            fly_text = f"volo {self.flying} m"
            if self.hovering:
                fly_text += " (librarsi)"
            parts.append(fly_text)
        
        if self.swimming:
            parts.append(f"nuoto {self.swimming} m")
        
        if self.climbing:
            parts.append(f"scalare {self.climbing} m")
        
        if self.burrowing:
            parts.append(f"scavare {self.burrowing} m")
        
        return ", ".join(parts)


@dataclass
class MonsterAttack:
    """Monster attack"""
    nome: str
    tipo: str  # "Mischia", "Distanza", "Incantesimo"
    bonus_attacco: int
    gittata: str
    bersagli: int = 1
    danni: str = ""  # e.g., "1d8 + 3 perforante"
    danni_aggiuntivi: Optional[str] = None
    effetti_speciali: Optional[str] = None
    
    def __post_init__(self):
        if not self.nome.strip():
            raise ValueError("Attack name cannot be empty")


@dataclass
class MonsterAction:
    """Monster action (Attack, Legendary Action, etc.)"""
    nome: str
    descrizione: str
    tipo: str = "Azione"  # "Azione", "Azione Leggendaria", "Azione del Covo", etc.
    ricarica: Optional[str] = None  # e.g., "Ricarica 5-6"
    limitazioni: Optional[str] = None  # e.g., "1/Giorno"
    
    def __post_init__(self):
        if not self.nome.strip():
            raise ValueError("Action name cannot be empty")
        if not self.descrizione.strip():
            raise ValueError("Action description cannot be empty")


@dataclass
class MonsterTrait:
    """Monster special trait/feature"""
    nome: str
    descrizione: str
    tipo: str = "Tratto"  # "Tratto", "Resistenza Magica", etc.
    
    def __post_init__(self):
        if not self.nome.strip():
            raise ValueError("Trait name cannot be empty")
        if not self.descrizione.strip():
            raise ValueError("Trait description cannot be empty")


@dataclass
class Monster:
    """D&D 5e Monster entity"""
    id: MonsterId
    nome: str
    tag: Dict[str, str]  # taglia, tipo, allineamento
    classe_armatura: int
    punti_ferita: str  # e.g., "150 (20d10 + 40)"
    velocita: Speed
    caratteristiche: AbilityScores
    challenge_rating: ChallengeRating
    contenuto_markdown: str
    
    # Optional attributes
    tiri_salvezza: Dict[str, int] = field(default_factory=dict)
    competenze: Dict[str, int] = field(default_factory=dict)
    resistenze_danni: List[DamageResistance] = field(default_factory=list)
    immunita_danni: List[DamageResistance] = field(default_factory=list)
    vulnerabilita_danni: List[DamageResistance] = field(default_factory=list)
    immunita_condizioni: List[ConditionType] = field(default_factory=list)
    sensi: List[str] = field(default_factory=list)
    linguaggi: List[str] = field(default_factory=list)
    
    # Actions and abilities
    tratti: List[MonsterTrait] = field(default_factory=list)
    azioni: List[MonsterAction] = field(default_factory=list)
    azioni_bonus: List[MonsterAction] = field(default_factory=list)
    reazioni: List[MonsterAction] = field(default_factory=list)
    azioni_leggendarie: List[MonsterAction] = field(default_factory=list)
    azioni_del_covo: List[MonsterAction] = field(default_factory=list)
    
    # Spellcasting
    incantesimi_innati: Dict[str, List[str]] = field(default_factory=dict)
    incantesimi_preparati: Dict[str, List[str]] = field(default_factory=dict)
    caratteristica_incantatore: Optional[str] = None
    cd_incantesimo: Optional[int] = None
    bonus_attacco_incantesimo: Optional[int] = None
    
    # Metadata
    fonte: str = "SRD"
    versione: str = "1.0"
    
    def __post_init__(self):
        if not self.nome.strip():
            raise ValueError("Monster name cannot be empty")
        if self.classe_armatura < 1:
            raise ValueError("AC must be at least 1")
        if not self.punti_ferita.strip():
            raise ValueError("HP string cannot be empty")
    
    def get_size(self) -> MonsterSize:
        """Get monster size enum"""
        size_map = {size.value: size for size in MonsterSize}
        return size_map.get(self.tag.get("taglia"), MonsterSize.MEDIUM)
    
    def get_type(self) -> MonsterType:
        """Get monster type enum"""
        type_map = {mtype.value: mtype for mtype in MonsterType}
        return type_map.get(self.tag.get("tipo"), MonsterType.HUMANOID)
    
    def get_alignment(self) -> Alignment:
        """Get monster alignment enum"""
        align_map = {align.value: align for align in Alignment}
        return align_map.get(self.tag.get("allineamento"), Alignment.UNALIGNED)
    
    def is_spellcaster(self) -> bool:
        """Check if monster can cast spells"""
        return bool(self.incantesimi_innati or self.incantesimi_preparati)
    
    def has_legendary_actions(self) -> bool:
        """Check if monster has legendary actions"""
        return bool(self.azioni_leggendarie)
    
    def has_lair_actions(self) -> bool:
        """Check if monster has lair actions"""
        return bool(self.azioni_del_covo)
    
    def get_proficiency_bonus(self) -> int:
        """Calculate proficiency bonus based on CR"""
        cr_numeric = self.challenge_rating.get_numeric_cr()
        if cr_numeric < 5:
            return 2
        elif cr_numeric < 9:
            return 3
        elif cr_numeric < 13:
            return 4
        elif cr_numeric < 17:
            return 5
        else:
            return 6
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "id": self.id.value,
            "nome": self.nome,
            "taglia": self.get_size().value,
            "tipo": self.get_type().value,
            "allineamento": self.get_alignment().value,
            "challenge_rating": self.challenge_rating.get_cr_text(),
            "xp": self.challenge_rating.xp,
            "is_spellcaster": self.is_spellcaster(),
            "has_legendary_actions": self.has_legendary_actions(),
            "has_lair_actions": self.has_lair_actions(),
            "fonte": self.fonte,
            "versione": self.versione
        }


# Repository interfaces for monsters
class MonsterRepository(ABC):
    """Repository interface for monster write operations"""
    
    @abstractmethod
    async def find_by_id(self, monster_id: MonsterId) -> Optional[Monster]:
        pass
    
    @abstractmethod
    async def find_by_name(self, name: str) -> Optional[Monster]:
        pass
    
    @abstractmethod
    async def find_by_type(self, monster_type: MonsterType) -> List[Monster]:
        pass
    
    @abstractmethod
    async def find_by_cr_range(self, min_cr: float, max_cr: float) -> List[Monster]:
        pass
    
    @abstractmethod
    async def save(self, monster: Monster) -> None:
        pass
    
    @abstractmethod
    async def find_all(self) -> List[Monster]:
        pass


class MonsterQueryRepository(ABC):
    """Repository interface for monster read operations (CQRS Query)"""
    
    @abstractmethod
    async def search_monsters(self, query: 'MonsterSearchQuery') -> List['MonsterSummary']:
        pass
    
    @abstractmethod
    async def get_monster_detail(self, monster_id: MonsterId) -> Optional['MonsterDetail']:
        pass
    
    @abstractmethod
    async def get_monsters_by_cr(self, challenge_rating: str) -> List['MonsterSummary']:
        pass
    
    @abstractmethod
    async def get_spellcasting_monsters(self) -> List['MonsterSummary']:
        pass
    
    @abstractmethod
    async def get_legendary_monsters(self) -> List['MonsterSummary']:
        pass


@dataclass
class MonsterValidationService:
    """Domain service for monster validation"""
    
    @staticmethod
    def validate_monster(monster: Monster) -> List[str]:
        """Validate monster business rules"""
        errors = []
        
        # CR and XP should match
        expected_xp = MonsterValidationService._get_expected_xp(monster.challenge_rating.get_numeric_cr())
        if expected_xp and monster.challenge_rating.xp != expected_xp:
            errors.append(f"XP value {monster.challenge_rating.xp} doesn't match CR {monster.challenge_rating.cr}")
        
        # Spellcasters should have spell stats
        if monster.is_spellcaster():
            if not monster.caratteristica_incantatore:
                errors.append("Spellcasting monster missing spellcasting ability")
            if not monster.cd_incantesimo:
                errors.append("Spellcasting monster missing spell save DC")
        
        # Legendary creatures should have reasonable CR
        if monster.has_legendary_actions():
            if monster.challenge_rating.get_numeric_cr() < 1:
                errors.append("Legendary monsters should have CR 1 or higher")
        
        # Size and HP should be reasonable
        size = monster.get_size()
        if size == MonsterSize.TINY and "200" in monster.punti_ferita:
            errors.append("Tiny creatures shouldn't have very high HP")
        
        return errors
    
    @staticmethod
    def _get_expected_xp(cr: float) -> Optional[int]:
        """Get expected XP for CR"""
        xp_table = {
            0: 10, 0.125: 25, 0.25: 50, 0.5: 100,
            1: 200, 2: 450, 3: 700, 4: 1100, 5: 1800,
            6: 2300, 7: 2900, 8: 3900, 9: 5000, 10: 5900,
            11: 7200, 12: 8400, 13: 10000, 14: 11500, 15: 13000,
            16: 15000, 17: 18000, 18: 20000, 19: 22000, 20: 25000,
            21: 33000, 22: 41000, 23: 50000, 24: 62000, 25: 75000,
            26: 90000, 27: 105000, 28: 120000, 29: 135000, 30: 155000
        }
        return xp_table.get(cr)