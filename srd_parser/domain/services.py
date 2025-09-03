"""
Domain Services for D&D 5e Class parsing
Business logic that doesn't belong to a specific entity
"""
from typing import Dict, List, Optional, Tuple
import re

from .entities import DndClass, Subclass, ClassProgressions
from .value_objects import (
    Ability, ClassSlug, HitDie, Level, ClassFeature, LevelProgression,
    MagicProgression, MulticlassRequirement, SkillChoice, EquipmentOption,
    SpellSlots, ProficiencyBonus, ResourceProgression, SpellPreparation, Rituals
)


class ClassParsingService:
    """Domain service for parsing class data from markdown"""
    
    @staticmethod
    def parse_primary_ability(text: str) -> Ability:
        """Parse primary ability from Italian text"""
        ability_map = {
            "forza": Ability.FORZA,
            "destrezza": Ability.DESTREZZA,
            "costituzione": Ability.COSTITUZIONE,
            "intelligenza": Ability.INTELLIGENZA,
            "saggezza": Ability.SAGGEZZA,
            "carisma": Ability.CARISMA
        }
        return ability_map.get(text.lower(), Ability.FORZA)
    
    @staticmethod
    def parse_hit_die(text: str) -> HitDie:
        """Parse hit die from text like 'D12 per livello da Barbaro'"""
        match = re.search(r'd(\d+)', text.lower())
        if not match:
            return HitDie.D8  # default
            
        die_size = match.group(1)
        die_map = {
            "6": HitDie.D6,
            "8": HitDie.D8, 
            "10": HitDie.D10,
            "12": HitDie.D12
        }
        return die_map.get(die_size, HitDie.D8)
    
    @staticmethod
    def parse_saving_throws(text: str) -> List[Ability]:
        """Parse saving throw proficiencies"""
        abilities = []
        for ability_name in text.split(" e "):
            ability_name = ability_name.strip()
            try:
                ability = ClassParsingService.parse_primary_ability(ability_name)
                abilities.append(ability)
            except:
                continue
        return abilities
    
    @staticmethod
    def parse_skill_choice(text: str) -> Optional[SkillChoice]:
        """Parse skill proficiency options like 'Scegli 2: Addestrare Animali, Atletica, ...'"""
        match = re.search(r'Scegli\s+(\d+):\s*(.+)', text, re.IGNORECASE)
        if not match:
            return None
            
        choose_count = int(match.group(1))
        options_text = match.group(2)
        options = [opt.strip() for opt in re.split(r',\s*', options_text) if opt.strip()]
        
        return SkillChoice(choose=choose_count, options=options)
    
    @staticmethod
    def parse_equipment_options(text: str) -> List[EquipmentOption]:
        """Parse equipment options from text"""
        options = []
        
        # Look for "Scegli A o B:" pattern
        match = re.search(
            r'Scegli\s*A\s*o\s*B:\s*\(A\)\s*(.+?);\s*oppure\s*\(B\)\s*(.+?)$',
            text,
            re.IGNORECASE | re.DOTALL
        )
        if match:
            option_a_text = match.group(1).strip()
            option_b_text = match.group(2).strip()
            
            option_a_items = [item.strip() for item in option_a_text.split(',') if item.strip()]
            option_b_items = [item.strip() for item in option_b_text.split(',') if item.strip()]
            
            options.append(EquipmentOption(label="Opzione A", items=option_a_items))
            options.append(EquipmentOption(label="Opzione B", items=option_b_items))
        else:
            # Simple comma-separated list
            items = [item.strip() for item in text.split(',') if item.strip()]
            if items:
                options.append(EquipmentOption(label="Default", items=items))
                
        return options
    
    @staticmethod
    def parse_multiclass_requirements(text: str) -> List[MulticlassRequirement]:
        """Parse multiclass prerequisites"""
        requirements = []
        # This would need implementation based on actual text patterns
        return requirements
    
    @staticmethod
    def extract_class_progressions(level_table: List[Dict]) -> ClassProgressions:
        """Extract progression systems from level table"""
        progressions = ClassProgressions()
        
        # Extract weapon mastery progression
        weapon_mastery = {}
        for row in level_table:
            if "maestria_nelle_armi" in row or "Maestria nelle armi" in row.get("note", ""):
                mastery_count = row.get("maestria_nelle_armi")
                if mastery_count:
                    weapon_mastery[row["livello"]] = mastery_count
        progressions.weapon_mastery = weapon_mastery
        
        # Extract extra attacks
        extra_attacks = {}
        for row in level_table:
            features = row.get("privilegi_di_classe", [])
            for feature in features:
                if "attacco extra" in feature.lower():
                    # Determine number of extra attacks based on level
                    level = row["livello"]
                    if level == 5:
                        extra_attacks[level] = 1
                    elif level == 11:
                        extra_attacks[level] = 2  
                    elif level == 20:
                        extra_attacks[level] = 3
        progressions.extra_attacks = extra_attacks
        
        return progressions
    
    @staticmethod
    def create_spell_slots(slot_data) -> Optional[SpellSlots]:
        """Create SpellSlots from various slot data formats"""
        if not slot_data:
            return None
            
        if isinstance(slot_data, dict):
            # Convert dict format to list
            slots = [0] * 9
            for level_str, count in slot_data.items():
                try:
                    level = int(level_str)
                    if 1 <= level <= 9:
                        slots[level - 1] = int(count) if count is not None else 0
                except (ValueError, TypeError):
                    continue
            return SpellSlots(slots)
        elif isinstance(slot_data, list) and len(slot_data) == 9:
            return SpellSlots(slot_data)
        
        return None
    
    @staticmethod
    def determine_spellcasting_method(class_features: List[str]) -> SpellPreparation:
        """Determine spell preparation method from class features"""
        features_text = " ".join(class_features).lower()
        
        if "preparare" in features_text or "prepari" in features_text:
            return SpellPreparation.PREPARED
        elif "conosci" in features_text or "impari" in features_text:
            return SpellPreparation.KNOWN
        
        return SpellPreparation.NONE
    
    @staticmethod
    def extract_spellcasting_focus(text: str) -> Optional[str]:
        """Extract spellcasting focus from text"""
        focus_patterns = [
            r"focus:\s*([^;]+)",
            r"componente\s+materiale:\s*([^;]+)",
            r"strumento\s+musicale",
            r"simbolo\s+sacro",
            r"focus\s+druidico",
            r"focus\s+arcano"
        ]
        
        for pattern in focus_patterns:
            match = re.search(pattern, text, re.IGNORECASE)
            if match:
                return match.group(1).strip() if match.groups() else match.group(0)
        
        return None
    
    @staticmethod
    def detect_ritual_casting(features_text: str) -> Rituals:
        """Detect ritual casting capability from features"""
        text_lower = features_text.lower()
        
        if "libro degli incantesimi" in text_lower:
            return Rituals.DA_LIBRO
        elif "rituali" in text_lower:
            return Rituals.SOLO_LISTA
        
        return Rituals.NESSUNO