"""
MongoDB implementation optimized for read operations (CQRS Query side)
Editor service focuses on fast, complex queries with projections
"""
from typing import Dict, List, Optional, Any
import logging
from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorCollection

from shared_domain.entities import DndClass, ClassId, ClassQueryRepository
from shared_domain.query_models import ClassSearchQuery, ClassSummary, ClassDetail

logger = logging.getLogger(__name__)


class MongoDBClassQueryRepository(ClassQueryRepository):
    """MongoDB implementation optimized for read operations and complex queries"""
    
    def __init__(self, connection_string: str, database_name: str):
        self.client = AsyncIOMotorClient(connection_string)
        self.db = self.client[database_name]
        self.collection: AsyncIOMotorCollection = self.db.classi
        
        # Ensure read-optimized indexes (async initialization will be handled elsewhere)
        # self._ensure_read_indexes()
    
    async def _ensure_read_indexes(self) -> None:
        """Create indexes optimized for read operations"""
        try:
            # Text search index for name and description
            await self.collection.create_index([
                ("name", "text"),
                ("features.description", "text")
            ])
            
            # Compound index for filtering
            await self.collection.create_index([
                ("primary_ability", 1),
                ("hit_die", 1)
            ])
            
            # Index for spell progression queries
            await self.collection.create_index("spell_progression.cantrips_by_level.1")
            
            # Index for level-based feature queries  
            await self.collection.create_index("features.level")
            
            logger.info("Read-optimized indexes ensured")
            
        except Exception as e:
            logger.warning(f"Could not create indexes: {e}")
    
    async def find_by_id(self, class_id: ClassId) -> Optional[DndClass]:
        """Find class by ID with full details"""
        try:
            doc = await self.collection.find_one({"slug": class_id.value})
            if not doc:
                return None
            return self._document_to_entity(doc)
        except Exception as e:
            logger.error(f"Error finding class by ID {class_id.value}: {e}")
            return None
    
    async def search_classes(self, query: ClassSearchQuery) -> List[ClassSummary]:
        """Search classes with filtering and return summaries"""
        try:
            mongo_query = self._build_search_query(query)
            logger.info(f"Class search query: {mongo_query}")
            
            # Use projection for performance - only summary fields
            projection = {
                "slug": 1,
                "nome": 1,
                "caratteristica_primaria": 1,
                "dado_vita": 1,
                "source": 1,
                "sottoclassi.nome": 1,  # Only subclass names for summary
                "magia": 1  # Magic/spellcasting info
            }
            
            # Apply sorting and limits
            cursor = self.collection.find(mongo_query, projection)
            
            if query.sort_by == "name" or query.sort_by == "alpha":
                cursor = cursor.sort("nome", 1)
            elif query.sort_by == "hit_die":
                cursor = cursor.sort("dado_vita", -1)
            elif query.sort_by == "primary_ability":
                cursor = cursor.sort("caratteristica_primaria", 1)
            
            if query.limit:
                cursor = cursor.limit(query.limit)
            if query.offset:
                cursor = cursor.skip(query.offset)
            
            docs = await cursor.to_list(length=None)
            logger.info(f"Found {len(docs)} classes")
            if docs:
                logger.info(f"First class doc: {docs[0]}")
            summaries = [self._document_to_summary(doc) for doc in docs]
            if summaries:
                logger.info(f"First summary: {summaries[0].__dict__}")
            return summaries
            
        except Exception as e:
            logger.error(f"Error in class search: {e}")
            return []
    
    async def get_class_detail(self, class_id: ClassId) -> Optional[ClassDetail]:
        """Get detailed class information for viewing"""
        try:
            doc = await self.collection.find_one({"id": class_id.value})
            if not doc:
                return None
            return self._document_to_detail(doc)
        except Exception as e:
            logger.error(f"Error getting class detail for {class_id.value}: {e}")
            return None
    
    async def get_classes_by_ability(self, primary_ability: str) -> List[ClassSummary]:
        """Get all classes with specific primary ability"""
        try:
            docs = await self.collection.find(
                {"primary_ability": primary_ability},
                {"id": 1, "name": 1, "primary_ability": 1, "hit_die": 1, "source": 1}
            ).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting classes by ability {primary_ability}: {e}")
            return []
    
    async def get_spellcasting_classes(self) -> List[ClassSummary]:
        """Get all classes with spellcasting progression"""
        try:
            docs = await self.collection.find(
                {"spell_progression": {"$exists": True}},
                {"id": 1, "name": 1, "primary_ability": 1, "hit_die": 1, "source": 1,
                 "spell_progression.cantrips_by_level": 1}
            ).to_list(length=None)
            
            return [self._document_to_summary(doc) for doc in docs]
        except Exception as e:
            logger.error(f"Error getting spellcasting classes: {e}")
            return []
    
    async def get_class_features_by_level(self, class_id: ClassId, level: int) -> List[Dict[str, Any]]:
        """Get class features available at specific level"""
        try:
            doc = await self.collection.find_one(
                {"id": class_id.value},
                {"features": 1, "subclasses.features": 1}
            )
            
            if not doc:
                return []
            
            features = []
            
            # Main class features
            for feature in doc.get("features", []):
                if feature.get("level", 99) <= level:
                    features.append({
                        "name": feature["name"],
                        "level": feature["level"],
                        "description": feature["description"],
                        "source": "class"
                    })
            
            return features
            
        except Exception as e:
            logger.error(f"Error getting features for level {level}: {e}")
            return []
    
    def _build_search_query(self, query: ClassSearchQuery) -> Dict[str, Any]:
        """Build MongoDB query from search parameters"""
        mongo_query = {}
        
        # Text search - use regex for Italian documents
        if query.text_query:
            mongo_query["$or"] = [
                {"nome": {"$regex": query.text_query, "$options": "i"}},
                {"slug": {"$regex": query.text_query, "$options": "i"}}
            ]
        
        # Filters
        if query.primary_ability:
            mongo_query["caratteristica_primaria"] = query.primary_ability
        
        if query.min_hit_die:
            mongo_query["dado_vita"] = {"$gte": f"d{query.min_hit_die}"}
        if query.max_hit_die:
            if "dado_vita" in mongo_query:
                mongo_query["dado_vita"]["$lte"] = f"d{query.max_hit_die}"
            else:
                mongo_query["dado_vita"] = {"$lte": f"d{query.max_hit_die}"}
        
        if query.is_spellcaster is not None:
            if query.is_spellcaster:
                mongo_query["magia.ha_incantesimi"] = True
            else:
                mongo_query["magia.ha_incantesimi"] = {"$ne": True}
        
        if query.source:
            mongo_query["source"] = query.source
        
        return mongo_query
    
    def _document_to_summary(self, doc: Dict[str, Any]) -> ClassSummary:
        """Convert MongoDB document to ClassSummary"""
        is_spellcaster = bool(doc.get("magia", {}).get("ha_incantesimi", False))
        
        subclass_names = [sc["nome"] for sc in doc.get("sottoclassi", [])]
        
        return ClassSummary(
            id=doc["slug"],
            name=doc["nome"],
            primary_ability=doc["caratteristica_primaria"],
            hit_die=doc["dado_vita"],
            source=doc.get("source", "SRD"),
            is_spellcaster=is_spellcaster,
            subclass_count=len(subclass_names),
            subclass_names=subclass_names[:3]  # Show first 3 subclasses
        )
    
    def _document_to_detail(self, doc: Dict[str, Any]) -> ClassDetail:
        """Convert MongoDB document to ClassDetail"""
        # Group features by level
        features_by_level = {}
        for feature in doc.get("features", []):
            level = feature.get("level", 1)
            if level not in features_by_level:
                features_by_level[level] = []
            features_by_level[level].append({
                "name": feature["name"],
                "description": feature["description"]
            })
        
        # Parse spell progression
        spell_slots = {}
        if "spell_progression" in doc:
            spell_slots = doc["spell_progression"].get("spell_slots_by_level", {})
        
        return ClassDetail(
            id=doc["id"],
            name=doc["name"],
            primary_ability=doc["primary_ability"],
            hit_die=doc["hit_die"],
            source=doc.get("source", "SRD"),
            saving_throw_proficiencies=doc.get("saving_throw_proficiencies", []),
            armor_proficiencies=doc.get("armor_proficiencies", []),
            weapon_proficiencies=doc.get("weapon_proficiencies", []),
            skill_options=doc.get("skill_options"),
            features_by_level=features_by_level,
            spell_slots_by_level=spell_slots,
            subclasses=[
                {
                    "id": sc["id"],
                    "name": sc["name"],
                    "description": sc.get("description", ""),
                    "feature_count": len(sc.get("features", []))
                }
                for sc in doc.get("subclasses", [])
            ]
        )
    
    def _document_to_entity(self, doc: Dict[str, Any]) -> DndClass:
        """Convert MongoDB document to full domain entity (reuse parser logic)"""
        from srd_parser.adapters.persistence.mongodb_class_repository import MongoDBClassRepository
        temp_repo = MongoDBClassRepository("", "")
        return temp_repo._document_to_entity(doc)
    
    async def close(self) -> None:
        """Close database connection"""
        self.client.close()