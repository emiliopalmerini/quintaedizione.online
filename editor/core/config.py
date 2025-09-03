"""Simplified configuration for D&D 5e SRD Editor - Italian only."""

import os
from typing import Dict, List

# Database configuration
MONGO_URI = os.getenv("MONGO_URI", "mongodb://localhost:27017")
DB_NAME = os.getenv("DB_NAME", "dnd")

# Italian-only collections (simplified)
COLLECTIONS: List[str] = [
    "documenti",
    "classi",
    "backgrounds", 
    "incantesimi",
    "oggetti_magici",
    "armature",
    "armi",
    "strumenti",
    "equipaggiamento",
    "servizi",
    "mostri",
]

# Collection display labels (Italian only)
COLLECTION_LABELS: Dict[str, str] = {
    "documenti": "Documenti SRD",
    "classi": "Classi",
    "backgrounds": "Background", 
    "incantesimi": "Incantesimi",
    "oggetti_magici": "Oggetti Magici",
    "armature": "Armature",
    "armi": "Armi",
    "strumenti": "Strumenti", 
    "equipaggiamento": "Equipaggiamento",
    "servizi": "Servizi",
    "mostri": "Mostri",
}

# Database collection mapping (Italian collections only)
DB_COLLECTIONS: Dict[str, str] = {
    "documenti": "documenti",
    "classi": "classi",
    "backgrounds": "backgrounds",
    "incantesimi": "incantesimi", 
    "oggetti_magici": "oggetti_magici",
    "armature": "armature",
    "armi": "armi",
    "strumenti": "strumenti",
    "equipaggiamento": "equipaggiamento", 
    "servizi": "servizi",
    "mostri": "mostri",
}


def get_collection_label(collection: str) -> str:
    """Get display label for collection."""
    return COLLECTION_LABELS.get(collection, collection.title())


def get_db_collection(collection: str) -> str:
    """Get database collection name for logical collection."""
    return DB_COLLECTIONS.get(collection, collection)


def is_valid_collection(collection: str) -> bool:
    """Check if collection name is valid."""
    return collection in COLLECTIONS


# Backwards compatibility exports
LOGICAL_COLLECTIONS = COLLECTIONS
LABELS_IT = COLLECTION_LABELS  
DB_COLLECTIONS_IT = DB_COLLECTIONS
label_for = lambda col, lang=None: get_collection_label(col)
db_collection_for = lambda col, lang=None: get_db_collection(col)