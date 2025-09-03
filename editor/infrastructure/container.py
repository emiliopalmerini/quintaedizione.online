"""
Dependency injection container for Editor service
Configures all dependencies following hexagonal architecture principles
"""
import os
from typing import Optional
import logging

from shared_domain.entities import ClassQueryRepository
from editor.adapters.persistence.mongodb_class_query_repository import MongoDBClassQueryRepository
from editor.application.query_handlers import (
    SearchClassesHandler,
    GetClassDetailHandler,
    GetClassesByAbilityHandler,
    GetSpellcastingClassesHandler,
    GetClassFeaturesHandler
)

logger = logging.getLogger(__name__)


class EditorContainer:
    """Dependency injection container for Editor service"""
    
    def __init__(self, config: Optional[dict] = None):
        self.config = config or self._load_config()
        self._class_query_repository: Optional[ClassQueryRepository] = None
    
    def _load_config(self) -> dict:
        """Load configuration from environment variables"""
        return {
            "mongo_uri": os.getenv("MONGO_URI", "mongodb://admin:password@localhost:27017/?authSource=admin"),
            "database_name": os.getenv("DB_NAME", "dnd"),
            "enable_query_caching": os.getenv("ENABLE_QUERY_CACHING", "false").lower() == "true",
            "cache_ttl_seconds": int(os.getenv("CACHE_TTL_SECONDS", "300"))
        }
    
    def get_class_query_repository(self) -> ClassQueryRepository:
        """Get configured class query repository"""
        if self._class_query_repository is None:
            logger.info("Initializing MongoDB class query repository")
            self._class_query_repository = MongoDBClassQueryRepository(
                connection_string=self.config["mongo_uri"],
                database_name=self.config["database_name"]
            )
        return self._class_query_repository
    
    def get_search_classes_handler(self) -> SearchClassesHandler:
        """Get configured handler for searching classes"""
        return SearchClassesHandler(
            class_repository=self.get_class_query_repository()
        )
    
    def get_class_detail_handler(self) -> GetClassDetailHandler:
        """Get configured handler for class details"""
        return GetClassDetailHandler(
            class_repository=self.get_class_query_repository()
        )
    
    def get_classes_by_ability_handler(self) -> GetClassesByAbilityHandler:
        """Get configured handler for classes by ability"""
        return GetClassesByAbilityHandler(
            class_repository=self.get_class_query_repository()
        )
    
    def get_spellcasting_classes_handler(self) -> GetSpellcastingClassesHandler:
        """Get configured handler for spellcasting classes"""
        return GetSpellcastingClassesHandler(
            class_repository=self.get_class_query_repository()
        )
    
    def get_class_features_handler(self) -> GetClassFeaturesHandler:
        """Get configured handler for class features"""
        return GetClassFeaturesHandler(
            class_repository=self.get_class_query_repository()
        )
    
    async def close(self) -> None:
        """Clean up resources"""
        if self._class_query_repository and hasattr(self._class_query_repository, 'close'):
            logger.info("Closing class query repository connection")
            await self._class_query_repository.close()


# Global container instance
_container: Optional[EditorContainer] = None


def get_container() -> EditorContainer:
    """Get global container instance"""
    global _container
    if _container is None:
        _container = EditorContainer()
    return _container


async def reset_container() -> None:
    """Reset global container (useful for testing)"""
    global _container
    if _container:
        await _container.close()
    _container = None


def configure_container(config: dict) -> EditorContainer:
    """Configure container with specific config"""
    global _container
    if _container:
        # Note: In a real app, you'd await this
        pass
    _container = EditorContainer(config)
    return _container