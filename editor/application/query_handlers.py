"""
Query handlers for Editor service following CQRS pattern
Handles all read operations with optimized queries and caching
"""
from typing import List, Optional, Dict, Any
import logging
from dataclasses import dataclass

from shared_domain.entities import ClassId, ClassQueryRepository
from shared_domain.query_models import (
    ClassSearchQuery, ClassSummary, ClassDetail, QueryResult
)

logger = logging.getLogger(__name__)


@dataclass
class SearchClassesQuery:
    """Query to search classes with filtering"""
    text_query: Optional[str] = None
    primary_ability: Optional[str] = None
    min_hit_die: Optional[int] = None
    max_hit_die: Optional[int] = None
    is_spellcaster: Optional[bool] = None
    source: Optional[str] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


@dataclass
class GetClassDetailQuery:
    """Query to get detailed class information"""
    class_id: str


@dataclass
class GetClassesByAbilityQuery:
    """Query to get classes filtered by primary ability"""
    primary_ability: str


@dataclass
class GetSpellcastingClassesQuery:
    """Query to get all spellcasting classes"""
    pass


class SearchClassesHandler:
    """Handler for searching classes"""
    
    def __init__(self, class_repository: ClassQueryRepository):
        self.class_repository = class_repository
    
    async def handle(self, query: SearchClassesQuery) -> QueryResult[List[ClassSummary]]:
        """Execute class search with filtering"""
        try:
            logger.info(f"Searching classes with query: {query.text_query}")
            
            search_query = ClassSearchQuery(
                text_query=query.text_query,
                primary_ability=query.primary_ability,
                min_hit_die=query.min_hit_die,
                max_hit_die=query.max_hit_die,
                is_spellcaster=query.is_spellcaster,
                source=query.source,
                sort_by=query.sort_by,
                limit=query.limit,
                offset=query.offset
            )
            
            classes = await self.class_repository.search_classes(search_query)
            
            logger.info(f"Found {len(classes)} classes")
            
            return QueryResult(
                success=True,
                data=classes,
                metadata={
                    "total_results": len(classes),
                    "query_parameters": {
                        "text": query.text_query,
                        "filters_applied": bool(
                            query.primary_ability or query.min_hit_die or 
                            query.max_hit_die or query.is_spellcaster is not None
                        )
                    }
                }
            )
            
        except Exception as e:
            logger.error(f"Error searching classes: {e}", exc_info=True)
            return QueryResult(
                success=False,
                error=f"Search failed: {str(e)}"
            )


class GetClassDetailHandler:
    """Handler for getting detailed class information"""
    
    def __init__(self, class_repository: ClassQueryRepository):
        self.class_repository = class_repository
    
    async def handle(self, query: GetClassDetailQuery) -> QueryResult[Optional[ClassDetail]]:
        """Get detailed class information"""
        try:
            logger.info(f"Getting class detail for ID: {query.class_id}")
            
            class_id = ClassId(query.class_id)
            class_detail = await self.class_repository.get_class_detail(class_id)
            
            if class_detail:
                logger.info(f"Retrieved class detail: {class_detail.name}")
                return QueryResult(
                    success=True,
                    data=class_detail,
                    metadata={
                        "features_count": sum(len(features) for features in class_detail.features_by_level.values()),
                        "subclasses_count": len(class_detail.subclasses),
                        "is_spellcaster": bool(class_detail.spell_slots_by_level)
                    }
                )
            else:
                logger.warning(f"Class not found: {query.class_id}")
                return QueryResult(
                    success=True,
                    data=None,
                    error="Class not found"
                )
                
        except Exception as e:
            logger.error(f"Error getting class detail: {e}", exc_info=True)
            return QueryResult(
                success=False,
                error=f"Failed to get class detail: {str(e)}"
            )


class GetClassesByAbilityHandler:
    """Handler for getting classes by primary ability"""
    
    def __init__(self, class_repository: ClassQueryRepository):
        self.class_repository = class_repository
    
    async def handle(self, query: GetClassesByAbilityQuery) -> QueryResult[List[ClassSummary]]:
        """Get all classes with specific primary ability"""
        try:
            logger.info(f"Getting classes by ability: {query.primary_ability}")
            
            classes = await self.class_repository.get_classes_by_ability(query.primary_ability)
            
            return QueryResult(
                success=True,
                data=classes,
                metadata={
                    "primary_ability": query.primary_ability,
                    "count": len(classes)
                }
            )
            
        except Exception as e:
            logger.error(f"Error getting classes by ability: {e}", exc_info=True)
            return QueryResult(
                success=False,
                error=f"Failed to get classes by ability: {str(e)}"
            )


class GetSpellcastingClassesHandler:
    """Handler for getting all spellcasting classes"""
    
    def __init__(self, class_repository: ClassQueryRepository):
        self.class_repository = class_repository
    
    async def handle(self, query: GetSpellcastingClassesQuery) -> QueryResult[List[ClassSummary]]:
        """Get all classes with spellcasting progression"""
        try:
            logger.info("Getting spellcasting classes")
            
            classes = await self.class_repository.get_spellcasting_classes()
            
            return QueryResult(
                success=True,
                data=classes,
                metadata={
                    "spellcaster_count": len(classes),
                    "abilities_represented": list(set(c.primary_ability for c in classes))
                }
            )
            
        except Exception as e:
            logger.error(f"Error getting spellcasting classes: {e}", exc_info=True)
            return QueryResult(
                success=False,
                error=f"Failed to get spellcasting classes: {str(e)}"
            )


@dataclass 
class GetClassFeaturesQuery:
    """Query to get class features at specific level"""
    class_id: str
    level: int


class GetClassFeaturesHandler:
    """Handler for getting class features by level"""
    
    def __init__(self, class_repository: ClassQueryRepository):
        self.class_repository = class_repository
    
    async def handle(self, query: GetClassFeaturesQuery) -> QueryResult[List[Dict[str, Any]]]:
        """Get class features available at specific level"""
        try:
            logger.info(f"Getting features for class {query.class_id} at level {query.level}")
            
            class_id = ClassId(query.class_id)
            features = await self.class_repository.get_class_features_by_level(class_id, query.level)
            
            return QueryResult(
                success=True,
                data=features,
                metadata={
                    "class_id": query.class_id,
                    "level": query.level,
                    "features_count": len(features)
                }
            )
            
        except Exception as e:
            logger.error(f"Error getting class features: {e}", exc_info=True)
            return QueryResult(
                success=False,
                error=f"Failed to get class features: {str(e)}"
            )