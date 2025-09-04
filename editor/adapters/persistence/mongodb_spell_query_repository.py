"""
MongoDB implementation for spell queries optimized for read operations (CQRS Query side)
Editor service focuses on fast, complex queries with projections
"""
from typing import Dict, List, Optional, Any
import logging
from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorCollection

from shared_domain.spell_entities import Spell, SpellId, SpellQueryRepository
from shared_domain.query_models import SpellSearchQuery, SpellSummary

logger = logging.getLogger(__name__)


class MongoDBSpellQueryRepository(SpellQueryRepository):
    """MongoDB implementation optimized for spell read operations"""
    
    def __init__(self, connection_string: str, database_name: str):
        self.client = AsyncIOMotorClient(connection_string)
        self.db = self.client[database_name]
        self.collection: AsyncIOMotorCollection = self.db.incantesimi
        
        # Ensure read-optimized indexes
        self._ensure_read_indexes()
    
    async def _ensure_read_indexes(self) -> None:
        """Create indexes optimized for read operations"""
        try:
            # Text search index for name and description
            await self.collection.create_index([
                ("name", "text"),
                ("nome", "text"),
                ("description", "text"),
                ("descrizione", "text")
            ])
            
            # Compound indexes for filtering
            await self.collection.create_index([
                ("scuola", 1),
                ("livello", 1),
                ("tempo_lancio", 1)
            ])
            
            # Classes filter index
            await self.collection.create_index("classi")
            
            logger.info("Spell read-optimized indexes ensured")
            
        except Exception as e:
            logger.warning(f"Could not create spell indexes: {e}")
    
    async def find_by_id(self, spell_id: SpellId) -> Optional[Spell]:
        """Find spell by ID with full details"""
        try:
            doc = await self.collection.find_one({"_id": spell_id.value})
            if not doc:
                return None
            return self._document_to_entity(doc)
        except Exception as e:
            logger.error(f"Error finding spell by ID {spell_id.value}: {e}")
            return None
    
    async def search_spells(self, query: SpellSearchQuery) -> List[SpellSummary]:
        """Search spells with filtering and return summaries"""
        try:
            mongo_query = self._build_search_query(query)
            
            # Use projection for performance - only summary fields
            projection = {
                "_id": 1,
                "name": 1,
                "nome": 1,
                "livello": 1,
                "scuola": 1,
                "tempo_lancio": 1,
                "gittata": 1,
                "durata": 1,
                "classi": 1,
                "rituale": 1,
                "concentrazione": 1
            }
            
            # Apply sorting and limits
            cursor = self.collection.find(mongo_query, projection)
            
            if query.sort_by == "name":
                cursor = cursor.sort([("nome", 1), ("name", 1)])
            elif query.sort_by == "level":
                cursor = cursor.sort("livello", 1)
            elif query.sort_by == "school":
                cursor = cursor.sort("scuola", 1)
            
            if query.limit:
                cursor = cursor.limit(query.limit)
            if query.offset:
                cursor = cursor.skip(query.offset)
            
            docs = await cursor.to_list(length=None)
            return [self._document_to_summary(doc) for doc in docs]
            
        except Exception as e:
            logger.error(f"Error in spell search: {e}")
            return []
    
    async def get_spells_by_class(self, character_class: str) -> List[SpellSummary]:
        """Get all spells available to a specific class"""
        try:
            docs = await self.collection.find(
                {"classi": character_class},
                {
                    "_id": 1, "name": 1, "nome": 1, "livello": 1, 
                    "scuola": 1, "tempo_lancio": 1, "gittata": 1, 
                    "durata": 1, "rituale": 1, "concentrazione": 1
                }
            ).sort("livello", 1).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting spells for class {character_class}: {e}")
            return []
    
    async def get_spells_by_level(self, level: int) -> List[SpellSummary]:
        """Get all spells of specific level"""
        try:
            docs = await self.collection.find(
                {"livello": level},
                {
                    "_id": 1, "name": 1, "nome": 1, "livello": 1,
                    "scuola": 1, "tempo_lancio": 1, "gittata": 1,
                    "durata": 1, "classi": 1, "rituale": 1, "concentrazione": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting spells for level {level}: {e}")
            return []
    
    async def get_ritual_spells(self) -> List[SpellSummary]:
        """Get all ritual spells"""
        try:
            docs = await self.collection.find(
                {"rituale": True},
                {
                    "_id": 1, "name": 1, "nome": 1, "livello": 1,
                    "scuola": 1, "tempo_lancio": 1, "gittata": 1,
                    "durata": 1, "classi": 1, "rituale": 1, "concentrazione": 1
                }
            ).sort("livello", 1).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting ritual spells: {e}")
            return []
    
    # Abstract methods implementation
    async def get_cantrips_by_class(self, character_class: str) -> List[SpellSummary]:
        """Get cantrips (level 0 spells) for a specific class"""
        try:
            docs = await self.collection.find(
                {"classi": character_class, "livello": 0},
                {
                    "_id": 1, "name": 1, "nome": 1, "livello": 1,
                    "scuola": 1, "tempo_lancio": 1, "gittata": 1,
                    "durata": 1, "classi": 1, "rituale": 1, "concentrazione": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting cantrips for class {character_class}: {e}")
            return []
    
    async def get_spell_detail(self, spell_id: SpellId) -> Optional[Spell]:
        """Get detailed spell information by ID"""
        try:
            # Use find_by_id if available, otherwise search by name
            return await self.find_by_id(spell_id)
        except Exception as e:
            logger.error(f"Error getting spell detail for {spell_id}: {e}")
            return None
    
    async def get_spells_by_class_and_level(self, character_class: str, level: int) -> List[SpellSummary]:
        """Get spells for specific class and level"""
        try:
            docs = await self.collection.find(
                {"classi": character_class, "livello": level},
                {
                    "_id": 1, "name": 1, "nome": 1, "livello": 1,
                    "scuola": 1, "tempo_lancio": 1, "gittata": 1,
                    "durata": 1, "classi": 1, "rituale": 1, "concentrazione": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting spells for class {character_class} level {level}: {e}")
            return []
    
    def _build_search_query(self, query: SpellSearchQuery) -> Dict[str, Any]:
        """Build MongoDB query from search parameters"""
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
        if query.level is not None:
            mongo_query["livello"] = query.level
        
        
        if query.school:
            mongo_query["scuola"] = query.school
        
        if query.class_name:
            mongo_query["classi"] = query.class_name
        
        if query.ritual_only is not None:
            mongo_query["rituale"] = query.ritual_only
        
        if query.concentration_only is not None:
            mongo_query["concentrazione"] = query.concentration_only
        
        return mongo_query
    
    def _document_to_summary(self, doc: Dict[str, Any]) -> SpellSummary:
        """Convert MongoDB document to SpellSummary"""
        return SpellSummary(
            id=str(doc["_id"]),
            name=doc.get("name", doc.get("nome", "")),
            italian_name=doc.get("nome", doc.get("name", "")),
            level=doc.get("livello", 0),
            school=doc.get("scuola", ""),
            casting_time=doc.get("tempo_lancio", ""),
            range_distance=doc.get("gittata", ""),
            duration=doc.get("durata", ""),
            classes=doc.get("classi", []),
            is_ritual=doc.get("rituale", False),
            requires_concentration=doc.get("concentrazione", False)
        )
    
    def _document_to_entity(self, doc: Dict[str, Any]) -> Spell:
        """Convert MongoDB document to full domain entity"""
        # For now, use a simplified conversion
        # In production, this should match the parser's entity conversion
        from shared_domain.spell_entities import SpellLevel, SpellSchool, CastingTime, SpellRange, SpellDuration
        
        return Spell(
            id=SpellId(str(doc["_id"])),
            name=doc.get("name", doc.get("nome", "")),
            italian_name=doc.get("nome", doc.get("name", "")),
            level=SpellLevel(doc.get("livello", 0)),
            school=SpellSchool(doc.get("scuola", "")),
            casting_time=CastingTime(doc.get("tempo_lancio", "")),
            range=SpellRange(doc.get("gittata", "")),
            duration=SpellDuration(doc.get("durata", "")),
            description=doc.get("description", doc.get("descrizione", "")),
            classes=doc.get("classi", []),
            is_ritual=doc.get("rituale", False),
            requires_concentration=doc.get("concentrazione", False),
            components=doc.get("componenti", {}),
            source="SRD"
        )
    
    async def close(self) -> None:
        """Close database connection"""
        self.client.close()