"""
Domain Entities for D&D 5e Classes
Following DDD principles with rich domain models
"""
from __future__ import annotations

from dataclasses import dataclass, field
from typing import Dict, List, Optional

from .value_objects import (
    Ability, ClassSlug, EquipmentOption, HitDie, Level, 
    ClassFeature, LevelProgression, MagicProgression,
    MulticlassRequirement, ResourceProgression, SkillChoice
)


@dataclass
class Subclass:
    """D&D 5e Subclass aggregate"""
    slug: ClassSlug
    name: str
    description: Optional[str] = None
    features: List[ClassFeature] = field(default_factory=list)
    bonus_spells: Dict[int, List[str]] = field(default_factory=dict)
    
    def __post_init__(self):
        if not self.name:
            raise ValueError("Subclass must have a name")
    
    def add_feature(self, feature: ClassFeature) -> None:
        """Add a feature to this subclass"""
        if feature not in self.features:
            self.features.append(feature)
            self.features.sort(key=lambda f: f.level.value)
    
    def get_features_at_level(self, level: Level) -> List[ClassFeature]:
        """Get all features gained at a specific level"""
        return [f for f in self.features if f.level == level]
    
    def add_bonus_spells(self, level: int, spells: List[str]) -> None:
        """Add bonus spells for a level"""
        if 1 <= level <= 20:
            self.bonus_spells[level] = spells


@dataclass
class ClassProgressions:
    """All progression systems for a class"""
    weapon_mastery: Dict[int, int] = field(default_factory=dict)
    fighting_styles: Dict[int, int] = field(default_factory=dict) 
    extra_attacks: Dict[int, int] = field(default_factory=dict)
    resources: List[ResourceProgression] = field(default_factory=list)
    ability_score_improvements: List[int] = field(default_factory=lambda: [4, 8, 12, 16])
    epic_boon_level: int = 19
    
    def add_resource_progression(self, resource: ResourceProgression) -> None:
        """Add a resource progression"""
        # Remove existing progression for same resource
        self.resources = [r for r in self.resources if r.resource_key != resource.resource_key]
        self.resources.append(resource)
    
    def get_resource_at_level(self, resource_key: str, level: Level) -> Optional[int]:
        """Get resource count at specific level"""
        for resource in self.resources:
            if resource.resource_key == resource_key:
                return resource.progression.get(level.value)
        return None


@dataclass
class ClassRules:
    """Class-specific rules and limitations"""
    durations: Dict[str, str] = field(default_factory=dict)
    limitations: Dict[str, bool] = field(default_factory=dict)
    formulas: Dict[str, str] = field(default_factory=dict)
    
    def add_duration_rule(self, feature: str, duration: str) -> None:
        """Add a duration rule for a feature"""
        self.durations[feature] = duration
    
    def add_limitation(self, limitation: str, active: bool = True) -> None:
        """Add a class limitation"""
        self.limitations[limitation] = active
    
    def add_formula(self, name: str, formula: str) -> None:
        """Add a calculation formula"""
        self.formulas[name] = formula


@dataclass 
class DndClass:
    """D&D 5e Character Class aggregate root"""
    slug: ClassSlug
    name: str
    hit_die: HitDie
    primary_ability: Ability
    saving_throw_proficiencies: List[Ability]
    
    # Optional basic traits
    subtitle: Optional[str] = None
    skill_proficiency_options: Optional[SkillChoice] = None
    weapon_proficiencies: List[str] = field(default_factory=list)
    armor_proficiencies: List[str] = field(default_factory=list) 
    tool_proficiencies: List[str] = field(default_factory=list)
    equipment_options: List[EquipmentOption] = field(default_factory=list)
    
    # Multiclassing
    multiclass_prerequisites: List[MulticlassRequirement] = field(default_factory=list)
    multiclass_traits: List[str] = field(default_factory=list)
    multiclass_notes: Optional[str] = None
    
    # Progression systems
    progressions: ClassProgressions = field(default_factory=ClassProgressions)
    magic: Optional[MagicProgression] = None
    level_table: List[LevelProgression] = field(default_factory=list)
    
    # Features and subclasses
    class_features: List[ClassFeature] = field(default_factory=list)
    subclasses: List[Subclass] = field(default_factory=list)
    
    # Rules and spell lists
    rules: ClassRules = field(default_factory=ClassRules)
    spell_lists: Dict[int, List[str]] = field(default_factory=dict)  # level -> spells
    
    # Recommendations
    recommended_cantrips: List[str] = field(default_factory=list)
    recommended_initial_spells: List[str] = field(default_factory=list)
    recommended_equipment: Optional[str] = None
    recommended_feats: List[str] = field(default_factory=list)
    recommended_epic_boon: Optional[str] = None
    
    # Raw content for display
    markdown_content: Optional[str] = None
    
    def __post_init__(self):
        if not self.name:
            raise ValueError("Class must have a name")
        if not self.saving_throw_proficiencies or len(self.saving_throw_proficiencies) != 2:
            raise ValueError("Class must have exactly 2 saving throw proficiencies")
    
    def add_feature(self, feature: ClassFeature) -> None:
        """Add a class feature"""
        if feature not in self.class_features:
            self.class_features.append(feature)
            self.class_features.sort(key=lambda f: f.level.value)
    
    def add_subclass(self, subclass: Subclass) -> None:
        """Add a subclass"""
        if subclass not in self.subclasses:
            self.subclasses.append(subclass)
    
    def add_level_progression(self, progression: LevelProgression) -> None:
        """Add level progression data"""
        # Remove existing progression for same level
        self.level_table = [lp for lp in self.level_table if lp.level != progression.level]
        self.level_table.append(progression)
        self.level_table.sort(key=lambda lp: lp.level.value)
    
    def get_features_at_level(self, level: Level) -> List[ClassFeature]:
        """Get all class features gained at a specific level"""
        return [f for f in self.class_features if f.level == level]
    
    def get_progression_at_level(self, level: Level) -> Optional[LevelProgression]:
        """Get level progression data for a specific level"""
        for progression in self.level_table:
            if progression.level == level:
                return progression
        return None
    
    def is_spellcaster(self) -> bool:
        """Check if this class can cast spells"""
        return self.magic is not None and self.magic.has_spells
    
    def add_spell_list(self, level: int, spells: List[str]) -> None:
        """Add spell list for a level (0=cantrips, 1-9=spell levels)"""
        if 0 <= level <= 9:
            self.spell_lists[level] = spells
    
    def validate_level_table(self) -> bool:
        """Validate that level table is complete (1-20)"""
        levels = {lp.level.value for lp in self.level_table}
        return levels == set(range(1, 21))
    
    def to_dict(self) -> dict:
        """Convert to dictionary matching ADR format"""
        result = {
            "slug": self.slug.value,
            "nome": self.name,
            "dado_vita": self.hit_die.value,
            "caratteristica_primaria": self.primary_ability.value,
            "salvezze_competenze": [ability.value for ability in self.saving_throw_proficiencies],
            "armi_competenze": self.weapon_proficiencies,
            "armature_competenze": self.armor_proficiencies,
            "strumenti_competenze": self.tool_proficiencies,
        }
        
        if self.subtitle:
            result["sottotitolo"] = self.subtitle
            
        if self.skill_proficiency_options:
            result["abilità_competenze_opzioni"] = {
                "scegli": self.skill_proficiency_options.choose,
                "opzioni": self.skill_proficiency_options.options
            }
            
        if self.equipment_options:
            result["equipaggiamento_iniziale_opzioni"] = [
                {"etichetta": opt.label, "oggetti": opt.items} 
                for opt in self.equipment_options
            ]
            
        if self.multiclass_prerequisites or self.multiclass_traits:
            multiclass = {}
            if self.multiclass_prerequisites:
                multiclass["prerequisiti"] = [
                    f"{req.ability.value} {req.minimum_score}" 
                    for req in self.multiclass_prerequisites
                ]
            if self.multiclass_traits:
                multiclass["tratti_acquisiti"] = self.multiclass_traits
            if self.multiclass_notes:
                multiclass["note"] = self.multiclass_notes
            result["multiclasse"] = multiclass
            
        # Add progressions
        if any([self.progressions.weapon_mastery, self.progressions.fighting_styles, 
                self.progressions.extra_attacks, self.progressions.resources]):
            progressions = {}
            if self.progressions.weapon_mastery:
                progressions["maestria_armi"] = {"livelli": self.progressions.weapon_mastery}
            if self.progressions.fighting_styles:
                progressions["stili_combattimento"] = {"livelli": self.progressions.fighting_styles}
            if self.progressions.extra_attacks:
                progressions["attacchi_extra"] = {"livelli": self.progressions.extra_attacks}
            if self.progressions.resources:
                progressions["risorse"] = [
                    {"chiave": r.resource_key, "livelli": r.progression}
                    for r in self.progressions.resources
                ]
            progressions["aumenti_caratteristica"] = self.progressions.ability_score_improvements
            progressions["dono_epico"] = self.progressions.epic_boon_level
            result["progressioni"] = progressions
            
        # Add magic progression
        if self.magic:
            magic_dict = {
                "ha_incantesimi": self.magic.has_spells,
            }
            if self.magic.spell_list_reference:
                magic_dict["lista_riferimento"] = self.magic.spell_list_reference
            if self.magic.spellcasting_ability:
                magic_dict["caratteristica_incantatore"] = self.magic.spellcasting_ability.value
            magic_dict["preparazione"] = self.magic.preparation_method.value
            if self.magic.focus:
                magic_dict["focus"] = self.magic.focus
            magic_dict["rituali"] = self.magic.ritual_casting.value
            if self.magic.cantrip_progression:
                magic_dict["trucchetti"] = self.magic.cantrip_progression
            if self.magic.spells_known_progression:
                magic_dict["incantesimi_preparati_o_noti"] = self.magic.spells_known_progression
            if self.magic.spell_slot_progression:
                magic_dict["slot"] = {
                    str(level): slots.slots 
                    for level, slots in self.magic.spell_slot_progression.items()
                }
            result["magia"] = magic_dict
            
        # Add level table
        if self.level_table:
            result["tabella_livelli"] = [
                {
                    "livello": lp.level.value,
                    "bonus_competenza": lp.proficiency_bonus.value,
                    "privilegi_di_classe": lp.class_features,
                    "trucchetti_conosciuti": lp.cantrips_known,
                    "incantesimi_preparati": lp.spells_prepared,
                    "slot_incantesimo": {str(i+1): lp.spell_slots.slots[i] for i in range(9)} if lp.spell_slots else None,
                    "risorse": lp.resources
                }
                for lp in self.level_table
            ]
            
        # Add features
        if self.class_features:
            result["privilegi_di_classe"] = [
                {"nome": f.name, "livello": f.level.value, "descrizione": f.description}
                for f in self.class_features
            ]
            
        # Add subclasses  
        if self.subclasses:
            result["sottoclassi"] = [
                {
                    "slug": sc.slug.value,
                    "nome": sc.name,
                    "descrizione": sc.description,
                    "privilegi_sottoclasse": [
                        {"nome": f.name, "livello": f.level.value, "descrizione": f.description}
                        for f in sc.features
                    ],
                    "incantesimi_sempre_preparati": {
                        str(level): spells for level, spells in sc.bonus_spells.items()
                    }
                }
                for sc in self.subclasses
            ]
            
        # Add spell lists
        if self.spell_lists:
            result["liste_incantesimi"] = {
                str(level): spells for level, spells in self.spell_lists.items()
            }
            
        # Add rules
        if any([self.rules.durations, self.rules.limitations, self.rules.formulas]):
            result["regole_classe"] = {
                "durate": self.rules.durations,
                "limitazioni": self.rules.limitations, 
                "formule": self.rules.formulas
            }
            
        # Add recommendations
        if any([self.recommended_cantrips, self.recommended_initial_spells, 
                self.recommended_equipment, self.recommended_feats, self.recommended_epic_boon]):
            recommendations = {}
            if self.recommended_cantrips:
                recommendations["trucchetti_cons"] = self.recommended_cantrips
            if self.recommended_initial_spells:
                recommendations["incantesimi_iniziali_cons"] = self.recommended_initial_spells
            if self.recommended_equipment:
                recommendations["equip_iniziale_cons"] = self.recommended_equipment
            if self.recommended_feats:
                recommendations["talenti_cons"] = self.recommended_feats
            if self.recommended_epic_boon:
                recommendations["dono_epico_cons"] = self.recommended_epic_boon
            result["raccomandazioni"] = recommendations
            
        if self.markdown_content:
            result["markdown"] = self.markdown_content
            
        return result