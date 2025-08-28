import os
from typing import Dict

MONGO_URI = os.getenv("MONGO_URI", "mongodb://localhost:27017")
DB_NAME = os.getenv("DB_NAME", "dnd")

# Logical collections (shared across languages)
LOGICAL_COLLECTIONS = [
    "classes",
    "backgrounds",
    "spells",
    "magic_items",
    "armor",
    "weapons",
    "tools",
    "mounts_vehicles",
    "services",
    "rules_glossary",
    "monsters",
    "animals",
]

# Labels per language
LABELS_IT: Dict[str, str] = {
    "documenti": "Documenti",
    "classes": "Classi",
    "backgrounds": "Background",
    "spells": "Incantesimi",
    "magic_items": "Oggetti Magici",
    "armor": "Armature",
    "weapons": "Armi",
    "tools": "Strumenti",
    "mounts_vehicles": "Cavalcature e Veicoli",
    "services": "Servizi",
    "rules_glossary": "Glossario Regole",
    "monsters": "Mostri",
    "animals": "Animali",
}

LABELS_EN: Dict[str, str] = {
    "documenti": "Documents",
    "classes": "Classes",
    "backgrounds": "Backgrounds",
    "spells": "Spells",
    "magic_items": "Magic Items",
    "armor": "Armor",
    "weapons": "Weapons",
    "tools": "Tools",
    "mounts_vehicles": "Mounts & Vehicles",
    "services": "Services",
    "rules_glossary": "Rules Glossary",
    "monsters": "Monsters",
    "animals": "Animals",
}

def label_for(collection: str, lang: str | None) -> str:
    l = (lang or "it").lower()
    table = LABELS_EN if l.startswith("en") else LABELS_IT
    return table.get(collection, collection)

# Database collection mapping per language.
# For IT we prepare target names (e.g., *_it) for future ingestion.
DB_COLLECTIONS_IT: Dict[str, str] = {
    "classes": "classi",
    "backgrounds": "backgrounds",  # currently same name
    "spells": "spells_it",
    "magic_items": "magic_items_it",
    "armor": "armor_it",
    "weapons": "weapons_it",
    "tools": "tools_it",
    "mounts_vehicles": "mounts_vehicles_it",
    "services": "services_it",
    "rules_glossary": "rules_glossary_it",
    "monsters": "monsters_it",
    "animals": "animals_it",
}

DB_COLLECTIONS_EN: Dict[str, str] = {
    "classes": "classes",
    "backgrounds": "backgrounds_en",
    "spells": "spells",
    "magic_items": "magic_items",
    "armor": "armor",
    "weapons": "weapons",
    "tools": "tools",
    "mounts_vehicles": "mounts_vehicles",
    "services": "services",
    "rules_glossary": "rules_glossary",
    "monsters": "monsters",
    "animals": "animals",
}

def db_collection_for(collection: str, lang: str | None) -> str:
    l = (lang or "it").lower()
    table = DB_COLLECTIONS_EN if l.startswith("en") else DB_COLLECTIONS_IT
    return table.get(collection, collection)

# Back-compat exports
COLLECTIONS = LOGICAL_COLLECTIONS
COLLECTION_LABELS = {c: LABELS_IT.get(c, c) for c in LOGICAL_COLLECTIONS}
