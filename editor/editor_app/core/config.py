import os

MONGO_URI = os.getenv("MONGO_URI", "mongodb://localhost:27017")
DB_NAME = os.getenv("DB_NAME", "dnd")

COLLECTIONS = [
    # Collezioni italiane principali (documenti resta fuori dal menu; è in homepage)
    "classi",
    # Collezioni legacy/extra (visibili se presenti)
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

COLLECTION_LABELS = {
    "documenti": "Documenti",
    "classi": "Classi",
    "spells": "Incantesimi (EN)",
    "magic_items": "Oggetti magici (EN)",
    "armor": "Armature (EN)",
    "weapons": "Armi (EN)",
    "tools": "Strumenti (EN)",
    "mounts_vehicles": "Cavalcature e veicoli (EN)",
    "services": "Servizi (EN)",
    "rules_glossary": "Glossario regole (EN)",
    "monsters": "Mostri (EN)",
    "animals": "Animali (EN)",
}
