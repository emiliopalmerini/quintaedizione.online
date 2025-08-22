import os

MONGO_URI = os.environ.get("MONGO_URI", "mongodb://localhost:27017")
DB_NAME = os.environ.get("DB_NAME", "dnd")

COLLECTIONS = [
    "spells","magic_items","armor","weapons","tools",
    "mounts_vehicles","services","rules_glossary","monsters","animals","classes",
]

# Italian labels for collections
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
