"""
Query models for CQRS read-side operations
Used by Editor service for optimized read queries
"""
from dataclasses import dataclass
from typing import Dict, List, Optional, Any, Generic, TypeVar

T = TypeVar('T')


@dataclass
class ClassSearchQuery:
    """Query parameters for searching classes"""
    text_query: Optional[str] = None
    primary_ability: Optional[str] = None
    min_hit_die: Optional[int] = None
    max_hit_die: Optional[int] = None
    is_spellcaster: Optional[bool] = None
    source: Optional[str] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


@dataclass
class ClassSummary:
    """Lightweight class summary for listings"""
    id: str
    name: str
    primary_ability: str
    hit_die: int
    source: str
    is_spellcaster: bool = False
    subclass_count: int = 0
    subclass_names: List[str] = None
    
    def __post_init__(self):
        if self.subclass_names is None:
            self.subclass_names = []


@dataclass
class ClassDetail:
    """Detailed class information for viewing"""
    id: str
    name: str
    primary_ability: str
    hit_die: int
    source: str
    saving_throw_proficiencies: List[str]
    armor_proficiencies: List[str]
    weapon_proficiencies: List[str]
    skill_options: Optional[str]
    features_by_level: Dict[int, List[Dict[str, str]]]
    spell_slots_by_level: Dict[str, Dict[str, int]]
    subclasses: List[Dict[str, Any]]
    
    def __post_init__(self):
        if self.features_by_level is None:
            self.features_by_level = {}
        if self.spell_slots_by_level is None:
            self.spell_slots_by_level = {}
        if self.subclasses is None:
            self.subclasses = []


@dataclass
class QueryResult(Generic[T]):
    """Generic query result wrapper"""
    success: bool
    data: Optional[T] = None
    error: Optional[str] = None
    metadata: Optional[Dict[str, Any]] = None
    
    def __post_init__(self):
        if self.metadata is None:
            self.metadata = {}


@dataclass
class ClassFeatureDetail:
    """Detailed information about a class feature"""
    name: str
    level: int
    description: str
    source: str  # "class" or "subclass"
    subclass_name: Optional[str] = None


@dataclass
class SpellSlotProgression:
    """Spell slot progression for a spellcasting class"""
    level: int
    spell_slots: Dict[int, int]  # spell_level -> slots
    cantrips_known: int = 0
    spells_known: int = 0


@dataclass
class SubclassDetail:
    """Detailed subclass information"""
    id: str
    name: str
    parent_class_id: str
    description: str
    features: List[ClassFeatureDetail]
    
    def __post_init__(self):
        if self.features is None:
            self.features = []


# ===== SPELL QUERY MODELS =====

@dataclass
class SpellSearchQuery:
    """Query parameters for searching spells"""
    text_query: Optional[str] = None
    class_name: Optional[str] = None
    level: Optional[int] = None
    school: Optional[str] = None
    ritual_only: Optional[bool] = None
    concentration_only: Optional[bool] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


@dataclass
class SpellSummary:
    """Lightweight spell summary for listings"""
    id: str
    nome: str
    livello: int
    scuola: str
    classi: List[str]
    is_ritual: bool = False
    requires_concentration: bool = False
    casting_time: str = ""
    
    def __post_init__(self):
        if self.classi is None:
            self.classi = []


@dataclass
class SpellDetail:
    """Detailed spell information for viewing"""
    id: str
    nome: str
    livello: int
    scuola: str
    classi: List[str]
    casting_time: str
    range: str
    duration: str
    components: str
    description: str
    higher_levels: Optional[str]
    is_ritual: bool
    requires_concentration: bool
    
    def __post_init__(self):
        if self.classi is None:
            self.classi = []


# ===== MONSTER QUERY MODELS =====

@dataclass
class MonsterSearchQuery:
    """Query parameters for searching monsters"""
    text_query: Optional[str] = None
    monster_type: Optional[str] = None
    size: Optional[str] = None
    min_cr: Optional[float] = None
    max_cr: Optional[float] = None
    alignment: Optional[str] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


@dataclass
class MonsterSummary:
    """Lightweight monster summary for listings"""
    id: str
    nome: str
    size: str
    monster_type: str
    alignment: str
    challenge_rating: str
    armor_class: int
    hit_points: str
    is_spellcaster: bool = False
    has_legendary_actions: bool = False
    
    def __post_init__(self):
        pass


@dataclass
class MonsterDetail:
    """Detailed monster information for viewing"""
    id: str
    nome: str
    size: str
    monster_type: str
    alignment: str
    armor_class: int
    hit_points: str
    speed: str
    challenge_rating: str
    xp_value: int
    abilities: Dict[str, int]
    saving_throws: Dict[str, int]
    skills: Dict[str, int]
    damage_resistances: List[str]
    damage_immunities: List[str]
    condition_immunities: List[str]
    senses: List[str]
    languages: List[str]
    traits: List[Dict[str, str]]
    actions: List[Dict[str, str]]
    legendary_actions: List[Dict[str, str]]
    is_spellcaster: bool
    spellcasting_info: Optional[Dict[str, Any]] = None
    
    def __post_init__(self):
        for attr in ['abilities', 'saving_throws', 'skills']:
            if getattr(self, attr) is None:
                setattr(self, attr, {})
        
        for attr in ['damage_resistances', 'damage_immunities', 'condition_immunities', 
                    'senses', 'languages', 'traits', 'actions', 'legendary_actions']:
            if getattr(self, attr) is None:
                setattr(self, attr, [])


# ===== EQUIPMENT QUERY MODELS =====

@dataclass
class WeaponSearchQuery:
    """Query parameters for searching weapons"""
    text_query: Optional[str] = None
    category: Optional[str] = None
    damage_type: Optional[str] = None
    properties: List[str] = None
    is_martial: Optional[bool] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None
    
    def __post_init__(self):
        if self.properties is None:
            self.properties = []


@dataclass
class ArmorSearchQuery:
    """Query parameters for searching armor"""
    text_query: Optional[str] = None
    category: Optional[str] = None
    stealth_disadvantage: Optional[bool] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


@dataclass
class MagicItemSearchQuery:
    """Query parameters for searching magic items"""
    text_query: Optional[str] = None
    item_type: Optional[str] = None
    rarity: Optional[str] = None
    requires_attunement: Optional[bool] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


@dataclass
class WeaponSummary:
    """Lightweight weapon summary for listings"""
    id: str
    nome: str
    categoria: str
    damage: str
    properties: List[str]
    cost: str
    weight: str
    
    def __post_init__(self):
        if self.properties is None:
            self.properties = []


@dataclass
class ArmorSummary:
    """Lightweight armor summary for listings"""
    id: str
    nome: str
    categoria: str
    armor_class: str
    cost: str
    weight: str
    stealth_disadvantage: bool = False


@dataclass
class MagicItemSummary:
    """Lightweight magic item summary for listings"""
    id: str
    nome: str
    item_type: str
    rarity: str
    requires_attunement: bool = False
    estimated_cost: Optional[str] = None


# ===== BACKGROUND & FEAT QUERY MODELS =====

@dataclass
class BackgroundSearchQuery:
    """Query parameters for searching backgrounds"""
    text_query: Optional[str] = None
    ability_score: Optional[str] = None
    skill: Optional[str] = None
    feat: Optional[str] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


@dataclass
class FeatSearchQuery:
    """Query parameters for searching feats"""
    text_query: Optional[str] = None
    category: Optional[str] = None
    has_prerequisites: Optional[bool] = None
    grants_ability_increase: Optional[bool] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


@dataclass
class BackgroundSummary:
    """Lightweight background summary for listings"""
    id: str
    nome: str
    ability_scores: List[str]
    feat: str
    skills: List[str]
    tools: List[str]
    
    def __post_init__(self):
        for attr in ['ability_scores', 'skills', 'tools']:
            if getattr(self, attr) is None:
                setattr(self, attr, [])


@dataclass
class BackgroundDetail:
    """Detailed background information for viewing"""
    id: str
    nome: str
    ability_scores: List[str]
    feat: str
    skills: List[str]
    tools: List[str]
    languages: List[str]
    equipment_options: List[Dict[str, Any]]
    special_features: List[str]
    
    def __post_init__(self):
        for attr in ['ability_scores', 'skills', 'tools', 'languages', 'equipment_options', 'special_features']:
            if getattr(self, attr) is None:
                setattr(self, attr, [])


@dataclass
class FeatSummary:
    """Lightweight feat summary for listings"""
    id: str
    nome: str
    categoria: str
    prerequisites: str
    benefits_count: int
    grants_ability_increase: bool = False
    total_ability_increases: int = 0


@dataclass
class FeatDetail:
    """Detailed feat information for viewing"""
    id: str
    nome: str
    categoria: str
    prerequisites: str
    benefits: List[Dict[str, str]]
    ability_increases: List[Dict[str, Any]]
    class_restrictions: List[str]
    race_restrictions: List[str]
    minimum_level: Optional[int]
    
    def __post_init__(self):
        for attr in ['benefits', 'ability_increases', 'class_restrictions', 'race_restrictions']:
            if getattr(self, attr) is None:
                setattr(self, attr, [])


# ===== DOCUMENT QUERY MODELS =====

@dataclass
class DocumentSearchQuery:
    """Query parameters for searching documents"""
    text_query: Optional[str] = None
    category: Optional[str] = None
    has_page_number: Optional[bool] = None
    keywords: List[str] = None
    sort_by: str = "title"
    limit: Optional[int] = None
    offset: Optional[int] = None
    
    def __post_init__(self):
        if self.keywords is None:
            self.keywords = []


@dataclass
class DocumentSummary:
    """Lightweight document summary for listings"""
    id: str
    titolo: str
    categoria: Optional[str]
    pagina: Optional[int]
    paragraph_count: int
    content_length: int
    keywords: List[str]
    
    def __post_init__(self):
        if self.keywords is None:
            self.keywords = []


@dataclass
class DocumentDetail:
    """Detailed document information for viewing"""
    id: str
    titolo: str
    categoria: Optional[str]
    pagina: Optional[int]
    sommario: Optional[str]
    paragraph_count: int
    subparagraph_count: int
    total_content_length: int
    keywords: List[str]
    table_of_contents: List[Dict[str, Any]]
    
    def __post_init__(self):
        for attr in ['keywords', 'table_of_contents']:
            if getattr(self, attr) is None:
                setattr(self, attr, [])


@dataclass
class DocumentContentMatch:
    """Search result for document content"""
    document_id: str
    document_title: str
    paragraph_title: str
    subparagraph_title: Optional[str]
    content_excerpt: str
    match_score: float = 1.0