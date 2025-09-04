"""
MongoDB implementation for background and feat queries optimized for read operations (CQRS Query side)
Handles character backgrounds and feats
"""
from typing import Dict, List, Optional, Any
import logging
from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorCollection

from shared_domain.background_entities import BackgroundQueryRepository, FeatQueryRepository, BackgroundId, FeatId
from shared_domain.query_models import (
    BackgroundSearchQuery, FeatSearchQuery,
    BackgroundSummary, FeatSummary, BackgroundDetail, FeatDetail
)

logger = logging.getLogger(__name__)


class MongoDBBackgroundQueryRepository(BackgroundQueryRepository):
    """MongoDB implementation optimized for background read operations"""
    
    def __init__(self, connection_string: str, database_name: str):
        self.client = AsyncIOMotorClient(connection_string)
        self.db = self.client[database_name]
        self.collection: AsyncIOMotorCollection = self.db.backgrounds
        
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
            await self.collection.create_index("competenze_abilita")
            await self.collection.create_index("competenze_strumenti")
            await self.collection.create_index("linguaggi")
            
            logger.info("Background read-optimized indexes ensured")
            
        except Exception as e:
            logger.warning(f"Could not create background indexes: {e}")
    
    async def search_backgrounds(self, query: BackgroundSearchQuery) -> List[BackgroundSummary]:
        """Search backgrounds with filtering and return summaries"""
        try:
            mongo_query = self._build_search_query(query)
            
            # Use projection for performance - only summary fields
            projection = {
                "_id": 1,
                "name": 1,
                "nome": 1,
                "competenze_abilita": 1,
                "competenze_strumenti": 1,
                "linguaggi": 1,
                "equipaggiamento": 1,
                "description": 1,
                "descrizione": 1
            }
            
            # Apply sorting and limits
            cursor = self.collection.find(mongo_query, projection)
            
            if query.sort_by == "name":
                cursor = cursor.sort([("nome", 1), ("name", 1)])
            elif query.sort_by == "skills":
                cursor = cursor.sort("competenze_abilita", 1)
            
            if query.limit:
                cursor = cursor.limit(query.limit)
            if query.offset:
                cursor = cursor.skip(query.offset)
            
            docs = await cursor.to_list(length=None)
            return [self._document_to_summary(doc) for doc in docs]
            
        except Exception as e:
            logger.error(f"Error in background search: {e}")
            return []
    
    async def get_backgrounds_by_skill(self, skill: str) -> List[BackgroundSummary]:
        """Get all backgrounds that provide a specific skill"""
        try:
            docs = await self.collection.find(
                {"competenze_abilita": skill},
                {
                    "_id": 1, "name": 1, "nome": 1, "competenze_abilita": 1,
                    "competenze_strumenti": 1, "linguaggi": 1, "equipaggiamento": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting backgrounds by skill {skill}: {e}")
            return []
    
    async def get_backgrounds_by_tool_proficiency(self, tool: str) -> List[BackgroundSummary]:
        """Get all backgrounds that provide specific tool proficiency"""
        try:
            docs = await self.collection.find(
                {"competenze_strumenti": tool},
                {
                    "_id": 1, "name": 1, "nome": 1, "competenze_abilita": 1,
                    "competenze_strumenti": 1, "linguaggi": 1, "equipaggiamento": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting backgrounds by tool {tool}: {e}")
            return []
    
    async def get_background_detail(self, background_id: BackgroundId) -> Optional[BackgroundDetail]:
        """Get detailed background information by ID"""
        try:
            doc = await self.collection.find_one({"_id": background_id.value})
            if not doc:
                return None
            return self._document_to_detail(doc)
        except Exception as e:
            logger.error(f"Error getting background detail for {background_id.value}: {e}")
            return None
    
    async def get_backgrounds_by_feat(self, feat_name: str) -> List[BackgroundSummary]:
        """Get backgrounds that grant specific feat"""
        try:
            docs = await self.collection.find(
                {"talento": feat_name},
                {
                    "_id": 1, "name": 1, "nome": 1, "competenze_abilita": 1,
                    "competenze_strumenti": 1, "linguaggi": 1, "equipaggiamento": 1,
                    "talento": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting backgrounds by feat {feat_name}: {e}")
            return []
    
    def _build_search_query(self, query: BackgroundSearchQuery) -> Dict[str, Any]:
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
        if query.skill_proficiencies:
            mongo_query["competenze_abilita"] = {"$in": query.skill_proficiencies}
        
        if query.tool_proficiencies:
            mongo_query["competenze_strumenti"] = {"$in": query.tool_proficiencies}
        
        if query.languages:
            mongo_query["linguaggi"] = {"$in": query.languages}
        
        return mongo_query
    
    def _document_to_summary(self, doc: Dict[str, Any]) -> BackgroundSummary:
        """Convert MongoDB document to BackgroundSummary"""
        name = doc.get("name", doc.get("nome", ""))
        description = (doc.get("description") or doc.get("descrizione") or "")
        description_preview = description[:100] + "..." if len(description) > 100 else description
        
        return BackgroundSummary(
            id=str(doc["_id"]),
            name=name,
            italian_name=doc.get("nome", name),
            skill_proficiencies=doc.get("competenze_abilita", []),
            tool_proficiencies=doc.get("competenze_strumenti", []),
            languages=doc.get("linguaggi", []),
            equipment=doc.get("equipaggiamento", []),
            description_preview=description_preview
        )
    
    def _document_to_detail(self, doc: Dict[str, Any]) -> BackgroundDetail:
        """Convert MongoDB document to BackgroundDetail"""
        return BackgroundDetail(
            id=str(doc["_id"]),
            nome=doc.get("nome", doc.get("name", "")),
            descrizione=doc.get("descrizione", doc.get("description", "")),
            skill_proficiencies=doc.get("competenze_abilita", []),
            tool_proficiencies=doc.get("competenze_strumenti", []),
            languages=doc.get("linguaggi", []),
            equipment_options=doc.get("equipaggiamento", []),
            feat=doc.get("talento", ""),
            special_features=doc.get("caratteristiche_speciali", []),
            suggested_characteristics=doc.get("caratteristiche_suggerite", [])
        )
    
    async def close(self) -> None:
        """Close database connection"""
        self.client.close()


class MongoDBFeatQueryRepository(FeatQueryRepository):
    """MongoDB implementation optimized for feat read operations"""
    
    def __init__(self, connection_string: str, database_name: str):
        self.client = AsyncIOMotorClient(connection_string)
        self.db = self.client[database_name]
        self.collection: AsyncIOMotorCollection = self.db.talenti
        
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
            await self.collection.create_index("aumento_caratteristica")
            await self.collection.create_index("prerequisiti")
            
            logger.info("Feat read-optimized indexes ensured")
            
        except Exception as e:
            logger.warning(f"Could not create feat indexes: {e}")
    
    async def search_feats(self, query: FeatSearchQuery) -> List[FeatSummary]:
        """Search feats with filtering and return summaries"""
        try:
            mongo_query = self._build_search_query(query)
            
            # Use projection for performance - only summary fields
            projection = {
                "_id": 1,
                "name": 1,
                "nome": 1,
                "aumento_caratteristica": 1,
                "prerequisiti": 1,
                "description": 1,
                "descrizione": 1
            }
            
            # Apply sorting and limits
            cursor = self.collection.find(mongo_query, projection)
            
            if query.sort_by == "name":
                cursor = cursor.sort([("nome", 1), ("name", 1)])
            elif query.sort_by == "ability_score":
                cursor = cursor.sort("aumento_caratteristica", 1)
            
            if query.limit:
                cursor = cursor.limit(query.limit)
            if query.offset:
                cursor = cursor.skip(query.offset)
            
            docs = await cursor.to_list(length=None)
            return [self._document_to_summary(doc) for doc in docs]
            
        except Exception as e:
            logger.error(f"Error in feat search: {e}")
            return []
    
    async def get_feats_by_ability_score_increase(self, ability: str) -> List[FeatSummary]:
        """Get all feats that provide specific ability score increase"""
        try:
            docs = await self.collection.find(
                {"aumento_caratteristica": ability},
                {
                    "_id": 1, "name": 1, "nome": 1, "aumento_caratteristica": 1,
                    "prerequisiti": 1, "description": 1, "descrizione": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting feats by ability {ability}: {e}")
            return []
    
    async def get_feats_without_prerequisites(self) -> List[FeatSummary]:
        """Get all feats that don't have prerequisites"""
        try:
            docs = await self.collection.find(
                {"$or": [{"prerequisiti": {"$exists": False}}, {"prerequisiti": []}]},
                {
                    "_id": 1, "name": 1, "nome": 1, "aumento_caratteristica": 1,
                    "prerequisiti": 1, "description": 1, "descrizione": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting feats without prerequisites: {e}")
            return []
    
    # Implement missing abstract methods
    async def get_feat_detail(self, feat_id: FeatId) -> Optional[FeatDetail]:
        """Get detailed feat information by ID"""
        try:
            doc = await self.collection.find_one({"_id": feat_id.value})
            if not doc:
                return None
            return self._document_to_feat_detail(doc)
        except Exception as e:
            logger.error(f"Error getting feat detail for {feat_id.value}: {e}")
            return None
    
    async def get_feats_by_category(self, category) -> List[FeatSummary]:
        """Get feats by category"""
        try:
            # Convert enum to string if needed
            category_str = category.value if hasattr(category, 'value') else str(category)
            docs = await self.collection.find(
                {"categoria": category_str},
                {
                    "_id": 1, "name": 1, "nome": 1, "aumento_caratteristica": 1,
                    "prerequisiti": 1, "description": 1, "descrizione": 1, "categoria": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting feats by category {category}: {e}")
            return []
    
    async def get_origin_feats(self) -> List[FeatSummary]:
        """Get origin feats"""
        try:
            docs = await self.collection.find(
                {"categoria": "origin"},
                {
                    "_id": 1, "name": 1, "nome": 1, "aumento_caratteristica": 1,
                    "prerequisiti": 1, "description": 1, "descrizione": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting origin feats: {e}")
            return []
    
    async def get_epic_boons(self) -> List[FeatSummary]:
        """Get epic boons"""
        try:
            docs = await self.collection.find(
                {"categoria": "epic_boon"},
                {
                    "_id": 1, "name": 1, "nome": 1, "aumento_caratteristica": 1,
                    "prerequisiti": 1, "description": 1, "descrizione": 1
                }
            ).sort([("nome", 1), ("name", 1)]).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting epic boons: {e}")
            return []
    
    def _build_search_query(self, query: FeatSearchQuery) -> Dict[str, Any]:
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
        if query.ability_score_increases:
            mongo_query["aumento_caratteristica"] = {"$in": query.ability_score_increases}
        
        if query.has_prerequisites is not None:
            if query.has_prerequisites:
                mongo_query["prerequisiti"] = {"$exists": True, "$ne": []}
            else:
                mongo_query["$or"] = [
                    {"prerequisiti": {"$exists": False}}, 
                    {"prerequisiti": []}
                ]
        
        return mongo_query
    
    def _document_to_summary(self, doc: Dict[str, Any]) -> FeatSummary:
        """Convert MongoDB document to FeatSummary"""
        name = doc.get("name", doc.get("nome", ""))
        description = (doc.get("description") or doc.get("descrizione") or "")
        description_preview = description[:100] + "..." if len(description) > 100 else description
        
        return FeatSummary(
            id=str(doc["_id"]),
            name=name,
            italian_name=doc.get("nome", name),
            ability_score_increases=doc.get("aumento_caratteristica", []),
            prerequisites=doc.get("prerequisiti", []),
            description_preview=description_preview
        )
    
    def _document_to_feat_detail(self, doc: Dict[str, Any]) -> FeatDetail:
        """Convert MongoDB document to FeatDetail"""
        return FeatDetail(
            id=str(doc["_id"]),
            nome=doc.get("nome", doc.get("name", "")),
            descrizione=doc.get("descrizione", doc.get("description", "")),
            categoria=doc.get("categoria", ""),
            prerequisites=doc.get("prerequisiti", []),
            ability_score_increases=doc.get("aumento_caratteristica", []),
            benefits=doc.get("benefici", []),
            source=doc.get("fonte", "SRD")
        )
    
    async def close(self) -> None:
        """Close database connection"""
        self.client.close()