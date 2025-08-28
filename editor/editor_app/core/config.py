import os

MONGO_URI = os.getenv("MONGO_URI", "mongodb://localhost:27017")
DB_NAME = os.getenv("DB_NAME", "dnd")

IT_COLLECTIONS = [
    # Collezioni italiane principali (documenti resta fuori dal menu; è in homepage)
    "classi",
    "backgrounds",
]

EN_COLLECTIONS = [
    # Collezioni inglesi
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

COLLECTIONS = IT_COLLECTIONS + EN_COLLECTIONS

COLLECTION_LABELS = {
    "documenti": "Documenti",
    # ITA
    "classi": "Classi",
    "backgrounds": "Background",
    # ENG (usa direttamente il nome inglese; rimosso suffisso "(EN)")
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
