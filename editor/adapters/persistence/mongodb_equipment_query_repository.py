"""
MongoDB implementation for equipment queries optimized for read operations (CQRS Query side)
Handles weapons, armor, magic items, and general equipment
"""
from typing import Dict, List, Optional, Any
import logging
from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorCollection

from shared_domain.equipment_entities import EquipmentQueryRepository
from shared_domain.query_models import (
    WeaponSearchQuery, ArmorSearchQuery, MagicItemSearchQuery,
    WeaponSummary, ArmorSummary, MagicItemSummary
)

logger = logging.getLogger(__name__)


class MongoDBEquipmentQueryRepository(EquipmentQueryRepository):
    """MongoDB implementation optimized for equipment read operations"""
    
    def __init__(self, connection_string: str, database_name: str):
        self.client = AsyncIOMotorClient(connection_string)
        self.db = self.client[database_name]
        
        # Collection mapping for different equipment types
        self.collection_mapping = {
            "weapons": "armi",
            "armor": "armature", 
            "magic_items": "oggetti_magici",
            "equipment": "equipaggiamento",
            "tools": "strumenti"
        }
        
        # Ensure read-optimized indexes
        self._ensure_read_indexes()
    
    async def _ensure_read_indexes(self) -> None:
        """Create indexes optimized for read operations on all equipment collections"""
        try:
            for collection_name in self.collection_mapping.values():
                collection = self.db[collection_name]
                
                # Text search index for common fields
                await collection.create_index([
                    ("name", "text"),
                    ("nome", "text"),
                    ("description", "text"),
                    ("descrizione", "text")
                ])
                
                # Common filter indexes
                await collection.create_index("categoria")
                await collection.create_index("tipo")
                await collection.create_index("rarita")
                await collection.create_index("costo")
                
            logger.info("Equipment read-optimized indexes ensured")
            
        except Exception as e:
            logger.warning(f"Could not create equipment indexes: {e}")
    
    def _get_collection(self, equipment_type: str) -> AsyncIOMotorCollection:
        """Get collection for equipment type"""
        collection_name = self.collection_mapping.get(equipment_type, equipment_type)
        return self.db[collection_name]
    
    async def search_weapons(self, query: WeaponSearchQuery) -> List[WeaponSummary]:
        """Search weapons with filtering and return summaries"""
        try:
            collection = self._get_collection("weapons")
            mongo_query = self._build_weapon_search_query(query)
            
            # Use projection for performance - only summary fields
            projection = {
                "_id": 1,
                "name": 1,
                "nome": 1,
                "categoria": 1,
                "tipo": 1,
                "danni": 1,
                "proprieta": 1,
                "costo": 1,
                "peso": 1,
                "maestria": 1
            }
            
            # Apply sorting and limits
            cursor = collection.find(mongo_query, projection)
            
            if query.sort_by == "name":
                cursor = cursor.sort([("nome", 1), ("name", 1)])
            elif query.sort_by == "damage":
                cursor = cursor.sort("danni", -1)
            elif query.sort_by == "cost":
                cursor = cursor.sort("costo", 1)
            elif query.sort_by == "category":
                cursor = cursor.sort("categoria", 1)
            
            if query.limit:
                cursor = cursor.limit(query.limit)
            if query.offset:
                cursor = cursor.skip(query.offset)
            
            docs = await cursor.to_list(length=None)
            return [self._document_to_weapon_summary(doc) for doc in docs]
            
        except Exception as e:
            logger.error(f"Error in weapon search: {e}")
            return []
    
    async def search_armor(self, query: ArmorSearchQuery) -> List[ArmorSummary]:
        """Search armor with filtering and return summaries"""
        try:
            collection = self._get_collection("armor")
            mongo_query = self._build_armor_search_query(query)
            
            # Use projection for performance - only summary fields
            projection = {
                "_id": 1,
                "name": 1,
                "nome": 1,
                "categoria": 1,
                "tipo": 1,
                "ca": 1,
                "max_des_mod": 1,
                "forza_min": 1,
                "stealth_disv": 1,
                "costo": 1,
                "peso": 1
            }
            
            # Apply sorting and limits
            cursor = collection.find(mongo_query, projection)
            
            if query.sort_by == "name":
                cursor = cursor.sort([("nome", 1), ("name", 1)])
            elif query.sort_by == "armor_class":
                cursor = cursor.sort("ca", -1)
            elif query.sort_by == "cost":
                cursor = cursor.sort("costo", 1)
            elif query.sort_by == "category":
                cursor = cursor.sort("categoria", 1)
            
            if query.limit:
                cursor = cursor.limit(query.limit)
            if query.offset:
                cursor = cursor.skip(query.offset)
            
            docs = await cursor.to_list(length=None)
            return [self._document_to_armor_summary(doc) for doc in docs]
            
        except Exception as e:
            logger.error(f"Error in armor search: {e}")
            return []
    
    async def search_magic_items(self, query: MagicItemSearchQuery) -> List[MagicItemSummary]:
        """Search magic items with filtering and return summaries"""
        try:
            collection = self._get_collection("magic_items")
            mongo_query = self._build_magic_item_search_query(query)
            
            # Use projection for performance - only summary fields
            projection = {
                "_id": 1,
                "name": 1,
                "nome": 1,
                "tipo": 1,
                "rarita": 1,
                "attunement": 1,
                "richiede_sintonia": 1,
                "description": 1,
                "descrizione": 1
            }
            
            # Apply sorting and limits
            cursor = collection.find(mongo_query, projection)
            
            if query.sort_by == "name":
                cursor = cursor.sort([("nome", 1), ("name", 1)])
            elif query.sort_by == "rarity":
                cursor = cursor.sort("rarita", 1)
            elif query.sort_by == "type":
                cursor = cursor.sort("tipo", 1)
            
            if query.limit:
                cursor = cursor.limit(query.limit)
            if query.offset:
                cursor = cursor.skip(query.offset)
            
            docs = await cursor.to_list(length=None)
            return [self._document_to_magic_item_summary(doc) for doc in docs]
            
        except Exception as e:
            logger.error(f"Error in magic item search: {e}")
            return []
    
    async def get_weapons_by_category(self, category: str) -> List[WeaponSummary]:
        """Get all weapons of specific category"""
        try:
            collection = self._get_collection("weapons")
            docs = await collection.find(
                {"categoria": category},
                {
                    "_id": 1, "name": 1, "nome": 1, "categoria": 1, 
                    "tipo": 1, "danni": 1, "proprieta": 1,
                    "costo": 1, "peso": 1, "maestria": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_weapon_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting weapons by category {category}: {e}")
            return []
    
    async def get_armor_by_category(self, category: str) -> List[ArmorSummary]:
        """Get all armor of specific category"""
        try:
            collection = self._get_collection("armor")
            docs = await collection.find(
                {"categoria": category},
                {
                    "_id": 1, "name": 1, "nome": 1, "categoria": 1,
                    "tipo": 1, "ca": 1, "max_des_mod": 1,
                    "forza_min": 1, "stealth_disv": 1, "costo": 1, "peso": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_armor_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting armor by category {category}: {e}")
            return []
    
    async def get_magic_items_by_rarity(self, rarity: str) -> List[MagicItemSummary]:
        """Get all magic items of specific rarity"""
        try:
            collection = self._get_collection("magic_items")
            docs = await collection.find(
                {"rarita": rarity},
                {
                    "_id": 1, "name": 1, "nome": 1, "tipo": 1,
                    "rarita": 1, "attunement": 1, "richiede_sintonia": 1,
                    "description": 1, "descrizione": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_magic_item_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting magic items by rarity {rarity}: {e}")
            return []
    
    def _build_weapon_search_query(self, query: WeaponSearchQuery) -> Dict[str, Any]:
        """Build MongoDB query from weapon search parameters"""
        mongo_query = {}
        
        # Text search
        if query.text_query:
            mongo_query["$or"] = [
                {"name": {"$regex": query.text_query, "$options": "i"}},
                {"nome": {"$regex": query.text_query, "$options": "i"}},
                {"description": {"$regex": query.text_query, "$options": "i"}},
                {"descrizione": {"$regex": query.text_query, "$options": "i"}}
            ]
        
        # Filters
        if query.category:
            mongo_query["categoria"] = query.category
        
        if query.weapon_type:
            mongo_query["tipo"] = query.weapon_type
        
        if query.properties:
            mongo_query["proprieta"] = {"$in": query.properties}
        
        if query.proficiency:
            mongo_query["maestria"] = query.proficiency
        
        return mongo_query
    
    def _build_armor_search_query(self, query: ArmorSearchQuery) -> Dict[str, Any]:
        """Build MongoDB query from armor search parameters"""
        mongo_query = {}
        
        # Text search
        if query.text_query:
            mongo_query["$or"] = [
                {"name": {"$regex": query.text_query, "$options": "i"}},
                {"nome": {"$regex": query.text_query, "$options": "i"}},
                {"description": {"$regex": query.text_query, "$options": "i"}},
                {"descrizione": {"$regex": query.text_query, "$options": "i"}}
            ]
        
        # Filters
        if query.category:
            mongo_query["categoria"] = query.category
        
        if query.armor_type:
            mongo_query["tipo"] = query.armor_type
        
        if query.min_armor_class is not None:
            mongo_query["ca"] = {"$gte": query.min_armor_class}
        if query.max_armor_class is not None:
            if "ca" in mongo_query:
                mongo_query["ca"]["$lte"] = query.max_armor_class
            else:
                mongo_query["ca"] = {"$lte": query.max_armor_class}
        
        if query.stealth_disadvantage is not None:
            mongo_query["stealth_disv"] = query.stealth_disadvantage
        
        return mongo_query
    
    def _build_magic_item_search_query(self, query: MagicItemSearchQuery) -> Dict[str, Any]:
        """Build MongoDB query from magic item search parameters"""
        mongo_query = {}
        
        # Text search
        if query.text_query:
            mongo_query["$or"] = [
                {"name": {"$regex": query.text_query, "$options": "i"}},
                {"nome": {"$regex": query.text_query, "$options": "i"}},
                {"description": {"$regex": query.text_query, "$options": "i"}},
                {"descrizione": {"$regex": query.text_query, "$options": "i"}}
            ]
        
        # Filters
        if query.item_type:
            mongo_query["tipo"] = query.item_type
        
        if query.rarity:
            mongo_query["rarita"] = query.rarity
        
        if query.requires_attunement is not None:
            mongo_query["$or"] = [
                {"attunement": query.requires_attunement},
                {"richiede_sintonia": query.requires_attunement}
            ]
        
        return mongo_query
    
    def _document_to_weapon_summary(self, doc: Dict[str, Any]) -> WeaponSummary:
        """Convert MongoDB document to WeaponSummary"""
        name = doc.get("name", doc.get("nome", ""))
        
        return WeaponSummary(
            id=str(doc["_id"]),
            name=name,
            italian_name=doc.get("nome", name),
            category=doc.get("categoria", ""),
            weapon_type=doc.get("tipo", ""),
            damage=doc.get("danni", ""),
            properties=doc.get("proprieta", []),
            cost=doc.get("costo", ""),
            weight=doc.get("peso", ""),
            proficiency=doc.get("maestria", "")
        )
    
    def _document_to_armor_summary(self, doc: Dict[str, Any]) -> ArmorSummary:
        """Convert MongoDB document to ArmorSummary"""
        name = doc.get("name", doc.get("nome", ""))
        
        return ArmorSummary(
            id=str(doc["_id"]),
            name=name,
            italian_name=doc.get("nome", name),
            category=doc.get("categoria", ""),
            armor_type=doc.get("tipo", ""),
            armor_class=doc.get("ca", 0),
            max_dex_modifier=doc.get("max_des_mod"),
            min_strength=doc.get("forza_min"),
            stealth_disadvantage=doc.get("stealth_disv", False),
            cost=doc.get("costo", ""),
            weight=doc.get("peso", "")
        )
    
    def _document_to_magic_item_summary(self, doc: Dict[str, Any]) -> MagicItemSummary:
        """Convert MongoDB document to MagicItemSummary"""
        name = doc.get("name", doc.get("nome", ""))
        description = (doc.get("description") or doc.get("descrizione") or "")
        description_preview = description[:100] + "..." if len(description) > 100 else description
        
        return MagicItemSummary(
            id=str(doc["_id"]),
            name=name,
            italian_name=doc.get("nome", name),
            item_type=doc.get("tipo", ""),
            rarity=doc.get("rarita", ""),
            requires_attunement=doc.get("attunement", doc.get("richiede_sintonia", False)),
            description_preview=description_preview
        )
    
    async def close(self) -> None:
        """Close database connection"""
        self.client.close()