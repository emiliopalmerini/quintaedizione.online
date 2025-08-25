import os

MONGO_URI = os.getenv("MONGO_URI", "mongodb://localhost:27017")
DB_NAME = os.getenv("DB_NAME", "dnd")

COLLECTIONS = [
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
    "classes",
]

COLLECTION_LABELS = {
    "spells": "Incantesimi",
    "magic_items": "Oggetti magici",
    "armor": "Armature",
    "weapons": "Armi",
    "tools": "Strumenti",
    "mounts_vehicles": "Cavalcature e veicoli",
    "services": "Servizi",
    "rules_glossary": "Glossario regole",
    "monsters": "Mostri",
    "animals": "Animali",
    "classes": "Classi",
}
