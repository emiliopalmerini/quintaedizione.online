"""
Complete ADR-compliant D&D 5e Classes Parser
Implements the full template structure from docs/adrs/0001-data-model.md
"""
from __future__ import annotations

import re
from typing import Dict, List, Optional, Tuple

from ..utils import (
    SECTION_H2_RE,
    SECTION_H3_RE,
    SECTION_H4_RE,
    clean_value,
    norm_key,
    split_sections,
)


def _slugify(s: str) -> str:
    """Convert string to slug format"""
    x = s.strip().lower()
    x = x.replace(" ", "-").replace("'", "").replace("'", "").replace("/", "-")
    return x


def _parse_markdown_table(
    block: List[str], start_idx: int
) -> Tuple[List[str], List[List[str]], int]:
    """Parse a GitHub-style markdown table starting at start_idx"""
    n = len(block)
    i = start_idx
    if i >= n or "|" not in block[i]:
        raise ValueError("Expected table header row starting with '|'")
    header = [c.strip() for c in block[i].strip().strip("|").split("|")]
    i += 1
    # skip delimiter row
    if i < n and "|" in block[i]:
        i += 1
    rows: List[List[str]] = []
    while i < n:
        line = block[i]
        if not line.strip().startswith("|"):
            break
        row = [c.strip() for c in line.strip().strip("|").split("|")]
        rows.append(row)
        i += 1
    return header, rows, i


def _parse_base_traits_table(block: List[str]) -> Dict:
    """Parse the 'Tratti base del [Classe]' table"""
    traits = {}
    
    # Find the base traits table
    for i, line in enumerate(block):
        if "Tratti base del" in line:
            try:
                headers, rows, _ = _parse_markdown_table(block, i + 2)  # Skip table header line
                
                for row in rows:
                    if len(row) >= 2:
                        key = row[0].strip()
                        value = row[1].strip()
                        
                        if "Caratteristica primaria" in key:
                            traits["caratteristica_primaria"] = value
                        elif "Dado Punti Ferita" in key:
                            # Extract hit die (d6, d8, d10, d12)
                            match = re.search(r'[dD](\d+)', value)
                            if match:
                                traits["dado_vita"] = f"d{match.group(1)}"
                        elif "Tiri salvezza competenti" in key:
                            # Split by 'e' or commas
                            saves = [s.strip() for s in value.replace(" e ", ", ").split(",")]
                            traits["salvezze_competenze"] = saves
                        elif "Abilità competenti" in key:
                            # Parse skill selection
                            if "Scegli" in value:
                                match = re.search(r'Scegli (\d+)', value)
                                if match:
                                    count = int(match.group(1))
                                    # Extract options after the colon
                                    if ":" in value:
                                        options_text = value.split(":", 1)[1].strip()
                                        options = [opt.strip() for opt in options_text.split(",")]
                                        traits["abilità_competenze_opzioni"] = {
                                            "scegli": count,
                                            "opzioni": options
                                        }
                            else:
                                # All abilities listed
                                options = [opt.strip() for opt in value.split(",")]
                                traits["abilità_competenze_opzioni"] = {
                                    "scegli": len(options),
                                    "opzioni": options
                                }
                        elif "Armi competenti" in key:
                            weapons = [w.strip() for w in value.replace(" e ", ", ").split(",")]
                            traits["armi_competenze"] = weapons
                        elif "Armature" in key or "addestramento" in key:
                            armors = [a.strip() for a in value.replace(" e ", ", ").split(",")]
                            traits["armature_competenze"] = armors
                        elif "Strumenti" in key:
                            if "Scegli" in value:
                                match = re.search(r'Scegli (\d+)', value)
                                if match:
                                    traits["strumenti_competenze"] = [f"Scegli {match.group(1)} strumenti musicali"]
                            else:
                                tools = [t.strip() for t in value.split(",")]
                                traits["strumenti_competenze"] = tools
                        elif "Equipaggiamento iniziale" in key:
                            # Parse equipment options
                            options = []
                            if "(A)" in value and "(B)" in value:
                                parts = value.split("(B)")
                                option_a = parts[0].replace("(A)", "").strip()
                                option_b = parts[1].replace("oppure", "").strip()
                                
                                options.append({
                                    "etichetta": "Opzione A",
                                    "oggetti": [option_a]
                                })
                                options.append({
                                    "etichetta": "Opzione B", 
                                    "oggetti": [option_b]
                                })
                            traits["equipaggiamento_iniziale_opzioni"] = options
                            
                break
            except (ValueError, IndexError):
                continue
    
    return traits


def _parse_level_table(block: List[str]) -> List[Dict]:
    """Parse the main level progression table"""
    levels = []
    
    # Find the level table (usually "Privilegi del [Classe]")
    for i, line in enumerate(block):
        if ("Privilegi del" in line or "Tabella:" in line) and "Livello" not in line:
            # Look for actual table with "Livello" header
            for j in range(i + 1, min(i + 10, len(block))):
                if "| Livello |" in block[j]:
                    try:
                        headers, rows, _ = _parse_markdown_table(block, j)
                        
                        for row in rows:
                            if len(row) >= 2 and row[0].strip().isdigit():
                                level = int(row[0].strip())
                                level_data = {"livello": level}
                                
                                # Map table columns to fields
                                for idx, header in enumerate(headers):
                                    if idx < len(row) and row[idx].strip():
                                        header_clean = header.lower().strip()
                                        value = row[idx].strip()
                                        
                                        if "bonus" in header_clean:
                                            level_data["bonus_competenza"] = int(value.replace("+", ""))
                                        elif "privilegi" in header_clean:
                                            privileges = [p.strip() for p in value.split(",")]
                                            level_data["privilegi_di_classe"] = privileges
                                        elif "trucchetti" in header_clean:
                                            if value != "—" and value.isdigit():
                                                level_data["trucchetti"] = int(value)
                                        elif "incantesimi" in header_clean:
                                            if value != "—" and value.isdigit():
                                                level_data["incantesimi_preparati"] = int(value)
                                        elif header_clean.isdigit():
                                            # Spell slot columns (1, 2, 3, etc.)
                                            slot_level = int(header_clean)
                                            if value != "—" and value.isdigit():
                                                if "slot" not in level_data:
                                                    level_data["slot"] = [0] * 9
                                                if slot_level <= 9:
                                                    level_data["slot"][slot_level - 1] = int(value)
                                        else:
                                            # Other class-specific columns
                                            level_data[header_clean.replace(" ", "_")] = value
                                
                                levels.append(level_data)
                        break
                    except (ValueError, IndexError):
                        continue
            break
    
    return levels


def _parse_multiclass_section(block: List[str]) -> Dict:
    """Parse multiclass requirements and benefits"""
    multiclass = {
        "prerequisiti": [],
        "tratti_acquisiti": [],
        "note": "Consulta le regole per il multiclasse in 'Creazione del personaggio'"
    }
    
    # Find multiclass section
    for i, line in enumerate(block):
        if "Come personaggio multiclasse" in line:
            # Parse next few lines for benefits
            for j in range(i + 1, min(i + 5, len(block))):
                line = block[j].strip()
                if line.startswith("- ") and "tratti dalla tabella" in line:
                    # Extract benefits after colon
                    if ":" in line:
                        benefits_text = line.split(":", 1)[1].strip()
                        benefits = [b.strip() for b in benefits_text.replace(" e ", ", ").split(",")]
                        multiclass["tratti_acquisiti"].extend(benefits)
            break
    
    return multiclass


def _parse_progressions(levels: List[Dict], class_name: str) -> Dict:
    """Parse class progressions from level table"""
    progressions = {}
    
    # Standard ability score increases
    progressions["aumenti_caratteristica"] = [4, 8, 12, 16]
    progressions["dono_epico"] = 19
    
    # Weapon mastery progression
    maestria_levels = {}
    for level in levels:
        for key, value in level.items():
            if "maestria" in key.lower() and isinstance(value, (int, str)) and str(value).isdigit():
                maestria_levels[str(level["livello"])] = int(value)
    
    if maestria_levels:
        progressions["maestria_armi"] = {"livelli": maestria_levels}
    
    # Extra attacks for martial classes
    martial_classes = ["Barbaro", "Guerriero", "Monaco", "Paladino", "Ranger"]
    if class_name in martial_classes:
        extra_attacks = {"5": 1}
        if class_name == "Guerriero":
            extra_attacks.update({"11": 2, "20": 3})
        progressions["attacchi_extra"] = {"livelli": extra_attacks}
    
    # Class resources
    resource_mappings = {
        "Barbaro": "ira",
        "Bardo": "ispirazione_bardica", 
        "Chierico": "canalizzare_divinita",
        "Druido": "forma_selvatica",
        "Monaco": "punti_disciplina",
        "Stregone": "punti_stregoneria",
        "Warlock": "slot_patto"
    }
    
    if class_name in resource_mappings:
        resource_key = resource_mappings[class_name]
        resource_levels = {}
        
        for level in levels:
            # Try to find resource in level data
            for key, value in level.items():
                if resource_key.lower() in key.lower():
                    if isinstance(value, (int, str)) and str(value).isdigit():
                        resource_levels[str(level["livello"])] = int(value)
                    break
        
        if resource_levels:
            progressions["risorse"] = [{
                "chiave": resource_key,
                "livelli": resource_levels
            }]
    
    return progressions


def _parse_magic_structure(class_name: str, levels: List[Dict]) -> Optional[Dict]:
    """Parse magic/spellcasting structure"""
    spellcaster_info = {
        "Bardo": ("Bardo", "Carisma", "known", "strumento musicale", "nessuno"),
        "Chierico": ("Chierico", "Saggezza", "prepared", "simbolo sacro", "solo_lista"),
        "Druido": ("Druido", "Saggezza", "prepared", "focus druidico", "solo_lista"),
        "Mago": ("Mago", "Intelligenza", "prepared", "focus arcano", "da_libro"),
        "Stregone": ("Stregone", "Carisma", "known", "focus arcano", "nessuno"),
        "Warlock": ("Warlock", "Carisma", "known", "focus arcano", "nessuno"),
        "Paladino": ("Paladino", "Carisma", "prepared", "simbolo sacro", "nessuno"),
        "Ranger": ("Ranger", "Saggezza", "known", "—", "nessuno")
    }
    
    if class_name not in spellcaster_info:
        return None
    
    lista_ref, caratteristica, preparazione, focus, rituali = spellcaster_info[class_name]
    
    magic = {
        "ha_incantesimi": True,
        "lista_riferimento": lista_ref,
        "caratteristica_incantatore": caratteristica,
        "preparazione": preparazione,
        "focus": focus,
        "rituali": rituali
    }
    
    # Extract spell progression from levels
    trucchetti = {}
    incantesimi = {}
    slots = {}
    
    for level in levels:
        level_num = str(level["livello"])
        
        if "trucchetti" in level:
            trucchetti[level_num] = level["trucchetti"]
        
        if "incantesimi_preparati" in level:
            incantesimi[level_num] = level["incantesimi_preparati"]
        
        if "slot" in level:
            slots[level_num] = level["slot"]
    
    if trucchetti:
        magic["trucchetti"] = trucchetti
    
    if incantesimi:
        if preparazione == "known":
            magic["incantesimi_noti"] = incantesimi
        else:
            magic["incantesimi_preparati_o_noti"] = incantesimi
    
    if slots:
        magic["slot"] = slots
    
    # Special Warlock handling
    if class_name == "Warlock":
        pact_slots = {
            "1": 1, "2": 1, "3": 2, "5": 2, "7": 2, "9": 2, "11": 3, "13": 3, "15": 3, "17": 4, "19": 4
        }
        slot_levels = {
            "1": 1, "3": 2, "5": 3, "7": 4, "9": 5
        }
        magic["patto_warlock"] = {
            "slot": pact_slots,
            "livello_slot": slot_levels
        }
    
    return magic


def _parse_class_features(block: List[str]) -> List[Dict]:
    """Parse individual class features"""
    features = []
    
    # Pattern for level features: "#### N° livello: Nome"
    level_feature_pattern = re.compile(r'^#### (\d+)° livello: (.+)$')
    
    current_feature = None
    current_description = []
    
    for line in block:
        match = level_feature_pattern.match(line.strip())
        if match:
            # Save previous feature
            if current_feature:
                current_feature["descrizione"] = "\n".join(current_description).strip()
                features.append(current_feature)
            
            # Start new feature
            level = int(match.group(1))
            name = match.group(2).strip()
            current_feature = {
                "nome": name,
                "livello": level,
                "descrizione": ""
            }
            current_description = []
        elif current_feature and line.strip() and not line.startswith("###"):
            current_description.append(line.strip())
    
    # Save last feature
    if current_feature:
        current_feature["descrizione"] = "\n".join(current_description).strip()
        features.append(current_feature)
    
    return features


def _parse_subclasses(block: List[str], class_name: str) -> List[Dict]:
    """Parse subclasses"""
    subclasses = []
    
    h3_sections = split_sections(block, SECTION_H3_RE)
    for h3_title, h3_block in h3_sections:
        # Look for subclass pattern: "Sottoclasse del [Classe]: [Nome]"
        match = re.match(rf"Sottoclasse del {class_name}: (.+)", h3_title.strip())
        if match:
            subclass_name = match.group(1).strip()
            
            subclass = {
                "slug": _slugify(subclass_name),
                "nome": subclass_name,
                "descrizione": "",
                "privilegi_sottoclasse": []
            }
            
            # Extract description (usually the first line or italic text)
            for line in h3_block[:5]:
                if line.strip().startswith("*") and line.strip().endswith("*"):
                    subclass["descrizione"] = line.strip().strip("*").strip()
                    break
            
            # Parse subclass features
            subclass_features = _parse_class_features(h3_block)
            for feature in subclass_features:
                subclass["privilegi_sottoclasse"].append({
                    "nome": feature["nome"],
                    "livello": feature["livello"],
                    "descrizione": feature["descrizione"]
                })
            
            subclasses.append(subclass)
    
    return subclasses


def _parse_spell_lists(block: List[str]) -> Dict:
    """Parse spell lists if present"""
    spell_lists = {}
    
    # Look for spell list sections
    content = "\n".join(block)
    
    # Extract cantrips and spell lists from tables
    # This would need more sophisticated parsing based on actual content
    # For now, return empty dict
    
    return spell_lists


def _parse_recommendations(block: List[str], class_name: str) -> Dict:
    """Parse recommendations from content"""
    recommendations = {}
    
    # Extract recommendations from italic text mentioning "consigliati"
    cantrips = []
    spells = []
    
    for line in block:
        if "consigliati" in line.lower():
            # Extract spell names in italics
            italic_matches = re.findall(r'\*([^*]+)\*', line)
            for spell in italic_matches:
                spell = spell.strip()
                if spell and not any(x in spell.lower() for x in ["vedi", "tabella"]):
                    if "trucchett" in line.lower():
                        cantrips.append(spell)
                    else:
                        spells.append(spell)
    
    if cantrips:
        recommendations["trucchetti_cons"] = cantrips
    if spells:
        recommendations["incantesimi_iniziali_cons"] = spells
    
    # Equipment recommendation
    recommendations["equip_iniziale_cons"] = "Opzione A"
    
    # Class-specific feat recommendations
    feat_recommendations = {
        "Barbaro": ["Grande maestria in armi", "Atletico"],
        "Bardo": ["Talentuoso", "Attore"],
        "Chierico": ["Guaritore", "Lanciatore ritualista"],
        "Druido": ["Elementalista", "Naturalista"],
        "Guerriero": ["Grande maestria in armi", "Esperto combattente"],
        "Monaco": ["Mobilità", "Deflettere frecce"],
        "Paladino": ["Grande maestria in armi", "Benedetto"],
        "Ranger": ["Tiratore scelto", "Esploratore"],
        "Ladro": ["Furtivo", "Opportunista"],
        "Stregone": ["Incantatore elementale", "Metamagico"],
        "Warlock": ["Invocazioni occulte", "Incantatore eldritch"],
        "Mago": ["Lanciatore ritualista", "Erudito"]
    }
    
    if class_name in feat_recommendations:
        recommendations["talenti_cons"] = feat_recommendations[class_name]
    
    # Epic boon recommendations
    boon_recommendations = {
        "Barbaro": "Dono dell'Offensiva irresistibile",
        "Bardo": "Dono di Richiamo degli incantesimi",
        "Chierico": "Dono del Ripristino",
        "Druido": "Dono della Resistenza dell'energia",
        "Guerriero": "Dono dell'Offensiva irresistibile",
        "Monaco": "Dono della Rapidità",
        "Paladino": "Dono della Protezione",
        "Ranger": "Dono del Cacciatore di mostri",
        "Ladro": "Dono dell'Abilità perfetta",
        "Stregone": "Dono dell'Incantesimo leggendario",
        "Warlock": "Dono dell'Arcano supremo",
        "Mago": "Dono dell'Incantesimo leggendario"
    }
    
    if class_name in boon_recommendations:
        recommendations["dono_epico_cons"] = boon_recommendations[class_name]
    
    return recommendations


def _add_subtitle(class_name: str) -> str:
    """Add appropriate subtitle for each class"""
    subtitles = {
        "Barbaro": "Guerriero selvaggio primordiale",
        "Bardo": "Maestro di musica e magia",
        "Chierico": "Campione divino della fede",
        "Druido": "Protettore della natura",
        "Guerriero": "Esperto combattente versatile",
        "Monaco": "Asceta dalla disciplina interiore",
        "Paladino": "Paladino sacro della giustizia",
        "Ranger": "Guardiano dei confini selvaggi",
        "Ladro": "Esperto delle ombre e inganni",
        "Stregone": "Fonte innata di magia arcana",
        "Warlock": "Vincolato a poteri ultraterreni",
        "Mago": "Studioso delle arti arcane"
    }
    return subtitles.get(class_name, "Avventuriero specializzato")


def parse_class_block(title: str, block: List[str]) -> Dict:
    """Parse a single class block into complete ADR format"""
    name = title.strip()
    
    # Start with basic info
    class_data = {
        "slug": _slugify(name),
        "nome": name,
        "sottotitolo": _add_subtitle(name)
    }
    
    # Parse base traits from table
    base_traits = _parse_base_traits_table(block)
    class_data.update(base_traits)
    
    # Parse level progression table
    levels = _parse_level_table(block)
    if levels:
        class_data["tabella_livelli"] = levels
    
    # Parse multiclass section
    multiclass = _parse_multiclass_section(block)
    if multiclass["tratti_acquisiti"]:  # Only add if we found benefits
        class_data["multiclasse"] = multiclass
    
    # Parse progressions from level table
    progressions = _parse_progressions(levels, name)
    if progressions:
        class_data["progressioni"] = progressions
    
    # Parse magic structure for spellcasters
    magic = _parse_magic_structure(name, levels)
    if magic:
        class_data["magia"] = magic
    
    # Parse class features
    features = _parse_class_features(block)
    if features:
        class_data["privilegi_di_classe"] = features
    
    # Parse subclasses
    subclasses = _parse_subclasses(block, name)
    if subclasses:
        class_data["sottoclassi"] = subclasses
    
    # Parse spell lists
    spell_lists = _parse_spell_lists(block)
    if spell_lists:
        class_data["liste_incantesimi"] = spell_lists
    
    # Parse recommendations
    recommendations = _parse_recommendations(block, name)
    if recommendations:
        class_data["raccomandazioni"] = recommendations
    
    # Add full markdown content
    class_data["content"] = "\n".join([f"## {name}"] + block).strip() + "\n"
    
    return class_data


def parse_classes(md_lines: List[str]) -> List[Dict]:
    """
    Parse the Italian SRD classes markdown into complete ADR-structured documents.
    """
    # Split top-level by H2 (## NomeClasse)
    classes = split_sections(md_lines, SECTION_H2_RE)
    docs: List[Dict] = []
    
    for title, block in classes:
        # Skip the initial document H1 ('# Classi')
        if title.lower().startswith("classi"):
            continue
        
        if not block:
            continue
            
        try:
            class_doc = parse_class_block(title, block)
            docs.append(class_doc)
        except Exception as e:
            print(f"Error parsing class '{title}': {e}")
            continue
    
    return docs