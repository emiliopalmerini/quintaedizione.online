"""Simplified query builder for filtering and searching."""

from typing import Dict, Any, List, Optional, Mapping
import re


def build_text_search(query: str, fields: List[str]) -> Dict[str, Any]:
    """Build simple text search across specified fields."""
    if not query or not fields:
        return {}
    
    # Simple regex search (case-insensitive)
    regex_pattern = {"$regex": re.escape(query), "$options": "i"}
    
    return {
        "$or": [
            {field: regex_pattern} for field in fields
        ]
    }


def build_collection_filters(collection: str, params: Mapping[str, str]) -> Dict[str, Any]:
    """Build MongoDB filter based on collection type and parameters."""
    filters = {}
    
    if collection == "incantesimi":
        # Spell filters
        if level := params.get("level"):
            try:
                level_val = int(level)
                filters["livello"] = level_val
            except ValueError:
                pass
                
        if school := params.get("school"):
            school_regex = {"$regex": re.escape(school), "$options": "i"}
            filters.setdefault("$and", []).append({"scuola": school_regex})
            
        if ritual := params.get("ritual"):
            ritual_bool = ritual.lower() in ("true", "1", "yes", "si")
            filters.setdefault("$and", []).append({"rituale": ritual_bool})
            
        if classes := params.get("classes"):
            classes_regex = {"$regex": re.escape(classes), "$options": "i"}
            filters.setdefault("$and", []).append({"classi": classes_regex})
    
    elif collection == "oggetti_magici":
        # Magic item filters  
        if rarity := params.get("rarity"):
            rarity_regex = {"$regex": re.escape(rarity), "$options": "i"}
            filters.setdefault("$and", []).append({"rarita": rarity_regex})
            
        if item_type := params.get("type"):
            type_regex = {"$regex": re.escape(item_type), "$options": "i"}
            filters.setdefault("$and", []).append({"tipo": type_regex})
            
        if attunement := params.get("attunement"):
            attunement_bool = attunement.lower() in ("true", "1", "yes", "si")
            filters.setdefault("$and", []).append({"sintonizzazione": attunement_bool})
    
    elif collection == "armature":
        # Armor filters
        if category := params.get("category"):
            cat_regex = {"$regex": re.escape(category), "$options": "i"}
            filters.setdefault("$and", []).append({"categoria": cat_regex})
            
        if stealth := params.get("stealth"):
            stealth_bool = stealth.lower() in ("true", "1", "yes", "si")
            filters.setdefault("$and", []).append({"svantaggio_furtivita": stealth_bool})
            
        if ac_base := params.get("ac_base"):
            try:
                ac_val = int(ac_base)
                filters.setdefault("$and", []).append({"ca_base": ac_val})
            except ValueError:
                pass
                
        if strength := params.get("strength"):
            if strength == "null":
                filters.setdefault("$and", []).append({
                    "$or": [
                        {"forza_richiesta": {"$exists": False}},
                        {"forza_richiesta": None},
                        {"forza_richiesta": 0}
                    ]
                })
            else:
                try:
                    str_val = int(strength)
                    filters.setdefault("$and", []).append({"forza_richiesta": str_val})
                except ValueError:
                    pass
                    
        if cost_range := params.get("cost_range"):
            cost_filter = _build_range_filter(cost_range, ["costo"])
            if cost_filter:
                filters.setdefault("$and", []).append(cost_filter)
                
        if weight_range := params.get("weight_range"):
            weight_filter = _build_range_filter(weight_range, ["peso"])
            if weight_filter:
                filters.setdefault("$and", []).append(weight_filter)
    
    elif collection == "armi":
        # Weapon filters
        if category := params.get("category"):
            cat_regex = {"$regex": re.escape(category), "$options": "i"}
            filters.setdefault("$and", []).append({"categoria": cat_regex})
            
        if mastery := params.get("mastery"):
            mastery_regex = {"$regex": re.escape(mastery), "$options": "i"}
            filters.setdefault("$and", []).append({"maestria": mastery_regex})
            
        if weapon_property := params.get("property"):
            prop_regex = {"$regex": re.escape(weapon_property), "$options": "i"}
            filters.setdefault("$and", []).append({"proprieta": prop_regex})
    
    elif collection == "mostri":
        # Monster filters
        if size := params.get("size"):
            size_regex = {"$regex": re.escape(size), "$options": "i"}
            filters.setdefault("$and", []).append({"taglia": size_regex})
            
        if monster_type := params.get("type"):
            type_regex = {"$regex": re.escape(monster_type), "$options": "i"}
            filters.setdefault("$and", []).append({"tipo": type_regex})
            
        if alignment := params.get("alignment"):
            align_regex = {"$regex": re.escape(alignment), "$options": "i"}
            filters.setdefault("$and", []).append({"allineamento": align_regex})
            
        if cr := params.get("cr"):
            # Try numeric first, then text
            try:
                cr_val = float(cr)
                filters.setdefault("$and", []).append({"gs": cr_val})
            except ValueError:
                cr_regex = {"$regex": re.escape(cr), "$options": "i"}
                filters.setdefault("$and", []).append({"gs": cr_regex})
    
    elif collection == "strumenti":
        # Tool filters
        if ability := params.get("ability"):
            ability_regex = {"$regex": re.escape(ability), "$options": "i"}
            filters.setdefault("$and", []).append({"caratteristica": ability_regex})
            
        if category := params.get("category"):
            cat_regex = {"$regex": re.escape(category), "$options": "i"}
            filters.setdefault("$and", []).append({"categoria": cat_regex})
            
        if craft := params.get("craft"):
            craft_regex = {"$regex": re.escape(craft), "$options": "i"}
            filters.setdefault("$and", []).append({"puo_creare": craft_regex})
    
    elif collection == "servizi":
        # Service filters
        if category := params.get("category"):
            cat_regex = {"$regex": re.escape(category), "$options": "i"}
            filters.setdefault("$and", []).append({"categoria": cat_regex})
            
        if availability := params.get("availability"):
            avail_regex = {"$regex": re.escape(availability), "$options": "i"}
            filters.setdefault("$and", []).append({"disponibilita": avail_regex})
    
    elif collection == "equipaggiamento":
        # Equipment filters
        if weight := params.get("weight"):
            weight_regex = {"$regex": re.escape(weight), "$options": "i"}
            filters.setdefault("$and", []).append({"peso": weight_regex})
    
    elif collection == "backgrounds":
        # Background filters
        if skill_prof := params.get("skill_proficiencies"):
            skill_regex = {"$regex": re.escape(skill_prof), "$options": "i"}
            filters.setdefault("$and", []).append({"competenze_abilita": skill_regex})
            
        if tool_prof := params.get("tool_proficiencies"):
            tool_regex = {"$regex": re.escape(tool_prof), "$options": "i"}
            filters.setdefault("$and", []).append({"competenze_strumenti": tool_regex})
            
        if languages := params.get("languages"):
            lang_regex = {"$regex": re.escape(languages), "$options": "i"}
            filters.setdefault("$and", []).append({"linguaggi": lang_regex})
    
    elif collection == "specie":
        # Species filters
        if size := params.get("size"):
            size_regex = {"$regex": re.escape(size), "$options": "i"}
            filters.setdefault("$and", []).append({"taglia": size_regex})
            
        if creature_type := params.get("creature_type"):
            type_regex = {"$regex": re.escape(creature_type), "$options": "i"}
            filters.setdefault("$and", []).append({"tipo_creatura": type_regex})
            
        if movement_speed := params.get("movement_speed"):
            speed_regex = {"$regex": re.escape(movement_speed), "$options": "i"}
            filters.setdefault("$and", []).append({"velocita_movimento": speed_regex})
            
        if ability_increase := params.get("ability_score_increase"):
            ability_regex = {"$regex": re.escape(ability_increase), "$options": "i"}
            filters.setdefault("$and", []).append({"aumento_caratteristica": ability_regex})
    
    elif collection == "talenti":
        # Feat filters
        if prereq_type := params.get("prerequisite_type"):
            if prereq_type == "none":
                filters.setdefault("$and", []).append({
                    "$or": [
                        {"prerequisiti": {"$exists": False}},
                        {"prerequisiti": None},
                        {"prerequisiti": ""}
                    ]
                })
            else:
                prereq_regex = {"$regex": re.escape(prereq_type), "$options": "i"}
                filters.setdefault("$and", []).append({"prerequisiti": prereq_regex})
                
        if ability_increase := params.get("ability_increase"):
            ability_regex = {"$regex": re.escape(ability_increase), "$options": "i"}
            filters.setdefault("$and", []).append({"aumento_caratteristica": ability_regex})
            
        if category := params.get("category"):
            cat_regex = {"$regex": re.escape(category), "$options": "i"}
            filters.setdefault("$and", []).append({"categoria": cat_regex})
    
    elif collection == "classi":
        # Class filters
        if primary_ability := params.get("caratteristica_primaria"):
            ability_regex = {"$regex": re.escape(primary_ability), "$options": "i"}
            filters.setdefault("$and", []).append({"caratteristica_primaria": ability_regex})
            
        if hit_die := params.get("dado_vita"):
            die_regex = {"$regex": re.escape(hit_die), "$options": "i"}
            filters.setdefault("$and", []).append({"dado_vita": die_regex})
            
        if has_spells := params.get("ha_incantesimi"):
            spells_bool = has_spells.lower() in ("true", "1", "yes", "si")
            filters.setdefault("$and", []).append({"ha_incantesimi": spells_bool})
            
        if spellcasting_ability := params.get("caratteristica_incantatore"):
            spell_ability_regex = {"$regex": re.escape(spellcasting_ability), "$options": "i"}
            filters.setdefault("$and", []).append({"caratteristica_incantatore": spell_ability_regex})
            
        if preparation := params.get("preparazione"):
            prep_regex = {"$regex": re.escape(preparation), "$options": "i"}
            filters.setdefault("$and", []).append({"preparazione_incantesimi": prep_regex})
            
        if weapon_prof := params.get("armi_competenze"):
            weapon_regex = {"$regex": re.escape(weapon_prof), "$options": "i"}
            filters.setdefault("$and", []).append({"competenze_armi": weapon_regex})
    
    elif collection == "documenti":
        # Document filters
        if page_num := params.get("numero_di_pagina"):
            page_regex = {"$regex": re.escape(page_num), "$options": "i"}
            filters.setdefault("$and", []).append({"numero_di_pagina": page_regex})
            
        if category := params.get("category"):
            cat_regex = {"$regex": re.escape(category), "$options": "i"}
            filters.setdefault("$and", []).append({"categoria": cat_regex})
            
        if content := params.get("content"):
            content_regex = {"$regex": re.escape(content), "$options": "i"}
            filters.setdefault("$and", []).append({"contenuto": content_regex})
    
    return filters


def _build_range_filter(range_param: str, field_names: List[str]) -> Optional[Dict[str, Any]]:
    """Helper to build range filters for cost and weight."""
    try:
        if range_param.endswith("+"):
            min_val = int(range_param[:-1])
            return {"$or": [{field: {"$gte": min_val}} for field in field_names]}
        elif "-" in range_param and not range_param.startswith("-"):
            parts = range_param.split("-")
            if len(parts) == 2:
                min_val = int(parts[0])
                max_val = int(parts[1])
                return {"$or": [{field: {"$gte": min_val, "$lte": max_val}} for field in field_names]}
    except (ValueError, IndexError):
        pass
    return None


def build_sort_criteria(sort_type: str = "alpha") -> List[tuple]:
    """Build MongoDB sort criteria."""
    if sort_type == "alpha":
        return [
            ("_sortkey_alpha", 1),
            ("slug", 1), 
            ("name", 1),
            ("nome", 1),
            ("title", 1),
            ("titolo", 1)
        ]
    elif sort_type == "level":
        return [("level", 1), ("livello", 1), ("name", 1)]
    else:
        return [("_id", 1)]  # Default sort