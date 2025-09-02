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
                filters["$or"] = [{"level": level_val}, {"livello": level_val}]
            except ValueError:
                pass
                
        if school := params.get("school"):
            school_regex = {"$regex": re.escape(school), "$options": "i"}
            filters.setdefault("$and", []).append({
                "$or": [{"school": school_regex}, {"scuola": school_regex}]
            })
            
        if ritual := params.get("ritual"):
            ritual_bool = ritual.lower() in ("true", "1", "yes", "si")
            filters.setdefault("$and", []).append({
                "$or": [{"ritual": ritual_bool}, {"rituale": ritual_bool}]
            })
    
    elif collection == "oggetti_magici":
        # Magic item filters  
        if rarity := params.get("rarity"):
            rarity_regex = {"$regex": re.escape(rarity), "$options": "i"}
            filters.setdefault("$and", []).append({
                "$or": [{"rarity": rarity_regex}, {"rarita": rarity_regex}]
            })
            
        if item_type := params.get("type"):
            type_regex = {"$regex": re.escape(item_type), "$options": "i"}
            filters.setdefault("$and", []).append({
                "$or": [{"type": type_regex}, {"tipo": type_regex}]
            })
    
    elif collection == "armature":
        # Armor filters
        if category := params.get("category"):
            cat_regex = {"$regex": re.escape(category), "$options": "i"}
            filters.setdefault("$and", []).append({
                "$or": [{"category": cat_regex}, {"categoria": cat_regex}]
            })
            
        if stealth := params.get("stealth"):
            stealth_bool = stealth.lower() in ("true", "1", "yes", "si")
            filters.setdefault("$and", []).append({
                "$or": [{"stealth_disadvantage": stealth_bool}, {"svantaggio_furtivita": stealth_bool}]
            })
            
        if ac_base := params.get("ac_base"):
            try:
                ac_val = int(ac_base)
                filters.setdefault("$and", []).append({
                    "$or": [{"ac_base": ac_val}, {"ca_base": ac_val}]
                })
            except ValueError:
                pass
    
    elif collection == "armi":
        # Weapon filters
        if category := params.get("category"):
            cat_regex = {"$regex": re.escape(category), "$options": "i"}
            filters.setdefault("$and", []).append({
                "$or": [{"category": cat_regex}, {"categoria": cat_regex}]
            })
    
    elif collection == "mostri":
        # Monster filters
        if size := params.get("size"):
            size_regex = {"$regex": re.escape(size), "$options": "i"}
            filters.setdefault("$and", []).append({
                "$or": [{"size": size_regex}, {"taglia": size_regex}]
            })
            
        if cr := params.get("cr"):
            # Try numeric first, then text
            try:
                cr_val = float(cr)
                filters.setdefault("$and", []).append({
                    "$or": [{"challenge_rating": cr_val}, {"cr": cr_val}]
                })
            except ValueError:
                cr_regex = {"$regex": re.escape(cr), "$options": "i"}
                filters.setdefault("$and", []).append({
                    "$or": [{"challenge_rating": cr_regex}, {"cr": cr_regex}]
                })
    
    return filters


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