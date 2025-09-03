"""
MongoDB implementation of ClassRepository for Parser service (Write-optimized)
"""
from typing import Dict, List, Optional, Any
import pymongo
from pymongo import MongoClient
from pymongo.collection import Collection

from shared_domain.entities import DndClass, ClassId, ClassRepository, Spell


class MongoDBClassRepository(ClassRepository):
    """MongoDB implementation optimized for write operations"""
    
    def __init__(self, connection_string: str, database_name: str):
        self.client = MongoClient(connection_string)
        self.db = self.client[database_name]
        self.collection: Collection = self.db.classes
        
        # Create indexes for write performance
        self._ensure_indexes()
    
    def _ensure_indexes(self) -> None:
        """Create indexes optimized for write operations"""
        # Unique index on class ID
        self.collection.create_index("id", unique=True)
        # Compound index for updates
        self.collection.create_index([("id", 1), ("version", 1)])
        # Index for name searches during parsing
        self.collection.create_index("name")
    
    async def find_by_id(self, class_id: ClassId) -> Optional[DndClass]:
        """Find class by ID"""
        doc = self.collection.find_one({"id": class_id.value})
        if not doc:
            return None
        return self._document_to_entity(doc)
    
    async def find_all(self) -> List[DndClass]:
        """Find all classes (used for batch operations in parser)"""
        docs = list(self.collection.find())
        return [self._document_to_entity(doc) for doc in docs]
    
    async def save(self, dnd_class: DndClass) -> None:
        """Save class with upsert (parser typically overwrites)"""
        doc = self._entity_to_document(dnd_class)
        
        # Upsert based on ID
        self.collection.replace_one(
            {"id": dnd_class.id.value},
            doc,
            upsert=True
        )
    
    async def search(self, query: str, filters: Dict[str, Any] = None) -> List[DndClass]:
        """Search classes (basic implementation for parser needs)"""
        mongo_query = {}
        
        if query:
            mongo_query["name"] = {"$regex": query, "$options": "i"}
        
        if filters:
            if "is_spellcaster" in filters:
                mongo_query["spell_progression"] = {"$exists": filters["is_spellcaster"]}
            if "primary_ability" in filters:
                mongo_query["primary_ability"] = filters["primary_ability"]
        
        docs = list(self.collection.find(mongo_query))
        return [self._document_to_entity(doc) for doc in docs]
    
    def _entity_to_document(self, dnd_class: DndClass) -> Dict[str, Any]:
        """Convert domain entity to MongoDB document"""
        from shared_domain.entities import ClassFeature, SpellProgression, Subclass
        
        doc = {
            "id": dnd_class.id.value,
            "name": dnd_class.name,
            "primary_ability": dnd_class.primary_ability.value,
            "hit_die": dnd_class.hit_die,
            "version": dnd_class.version,
            "source": dnd_class.source,
            
            # Collections
            "features": [
                {
                    "name": f.name,
                    "level": f.level.value,
                    "description": f.description
                }
                for f in dnd_class.features
            ],
            
            "subclasses": [
                {
                    "id": sc.id.value,
                    "name": sc.name,
                    "parent_class_id": sc.parent_class_id.value,
                    "description": sc.description,
                    "features": [
                        {
                            "name": f.name,
                            "level": f.level.value,
                            "description": f.description
                        }
                        for f in sc.features
                    ]
                }
                for sc in dnd_class.subclasses
            ],
            
            # Optional data
            "saving_throw_proficiencies": [ability.value for ability in dnd_class.saving_throw_proficiencies],
            "armor_proficiencies": dnd_class.armor_proficiencies,
            "weapon_proficiencies": dnd_class.weapon_proficiencies,
            "skill_options": dnd_class.skill_options,
        }
        
        # Spell progression if present
        if dnd_class.spell_progression:
            doc["spell_progression"] = {
                "cantrips_by_level": dnd_class.spell_progression.cantrips_by_level,
                "spells_by_level": dnd_class.spell_progression.spells_by_level,
                "spell_slots_by_level": dnd_class.spell_progression.spell_slots_by_level
            }
        
        return doc
    
    def _document_to_entity(self, doc: Dict[str, Any]) -> DndClass:
        """Convert MongoDB document to domain entity"""
        from shared_domain.entities import (
            ClassFeature, SpellProgression, Subclass, Level, 
            Ability, EntityId
        )
        
        # Parse primary ability
        primary_ability = Ability(doc["primary_ability"])
        
        # Create class entity
        dnd_class = DndClass(
            id=ClassId(doc["id"]),
            name=doc["name"],
            primary_ability=primary_ability,
            hit_die=doc["hit_die"],
            version=doc.get("version", "1.0"),
            source=doc.get("source", "SRD")
        )
        
        # Parse features
        for feature_doc in doc.get("features", []):
            feature = ClassFeature(
                name=feature_doc["name"],
                level=Level(feature_doc["level"]),
                description=feature_doc["description"]
            )
            dnd_class.features.append(feature)
        
        # Parse subclasses
        for subclass_doc in doc.get("subclasses", []):
            subclass = Subclass(
                id=EntityId(subclass_doc["id"]),
                name=subclass_doc["name"],
                parent_class_id=ClassId(subclass_doc["parent_class_id"]),
                description=subclass_doc.get("description")
            )
            
            # Parse subclass features
            for feature_doc in subclass_doc.get("features", []):
                feature = ClassFeature(
                    name=feature_doc["name"],
                    level=Level(feature_doc["level"]),
                    description=feature_doc["description"]
                )
                subclass.features.append(feature)
            
            dnd_class.subclasses.append(subclass)
        
        # Parse optional data
        if "saving_throw_proficiencies" in doc:
            dnd_class.saving_throw_proficiencies = [
                Ability(ability) for ability in doc["saving_throw_proficiencies"]
            ]
        
        dnd_class.armor_proficiencies = doc.get("armor_proficiencies", [])
        dnd_class.weapon_proficiencies = doc.get("weapon_proficiencies", [])
        dnd_class.skill_options = doc.get("skill_options")
        
        # Parse spell progression
        if "spell_progression" in doc:
            spell_prog_doc = doc["spell_progression"]
            dnd_class.spell_progression = SpellProgression(
                cantrips_by_level=spell_prog_doc["cantrips_by_level"],
                spells_by_level=spell_prog_doc["spells_by_level"],
                spell_slots_by_level=spell_prog_doc["spell_slots_by_level"]
            )
        
        return dnd_class
    
    def close(self) -> None:
        """Close database connection"""
        self.client.close()