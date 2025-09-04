"""
Dependency injection container for Editor service
Configures all dependencies following hexagonal architecture principles
"""
import os
from typing import Optional
import logging

# Repository interfaces from shared domain
from shared_domain.entities import ClassQueryRepository
from shared_domain.spell_entities import SpellQueryRepository
from shared_domain.monster_entities import MonsterQueryRepository
from shared_domain.document_entities import DocumentQueryRepository
from shared_domain.equipment_entities import EquipmentQueryRepository
from shared_domain.background_entities import BackgroundQueryRepository, FeatQueryRepository

# Repository implementations
from adapters.persistence.mongodb_class_query_repository import MongoDBClassQueryRepository
from adapters.persistence.mongodb_spell_query_repository import MongoDBSpellQueryRepository
from adapters.persistence.mongodb_monster_query_repository import MongoDBMonsterQueryRepository
from adapters.persistence.mongodb_document_query_repository import MongoDBDocumentQueryRepository
from adapters.persistence.mongodb_equipment_query_repository import MongoDBEquipmentQueryRepository
from adapters.persistence.mongodb_background_query_repository import (
    MongoDBBackgroundQueryRepository, MongoDBFeatQueryRepository
)

# Use cases from shared domain
from shared_domain.use_cases import (
    CompleteUseCaseFactory, SearchClassesUseCase, SearchSpellsUseCase,
    SearchMonstersUseCase, SearchDocumentsUseCase, GetDocumentUseCase,
    GetFilterOptionsUseCase, SearchEquipmentUseCase, SearchBackgroundsUseCase,
    SearchFeatsUseCase, UseCaseResult
)

# Legacy handlers (to be deprecated)
from application.query_handlers import (
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
        
        # Repository instances (lazy initialization)
        self._class_query_repository: Optional[ClassQueryRepository] = None
        self._spell_query_repository: Optional[SpellQueryRepository] = None
        self._monster_query_repository: Optional[MonsterQueryRepository] = None
        self._document_query_repository: Optional[DocumentQueryRepository] = None
        self._equipment_query_repository: Optional[EquipmentQueryRepository] = None
        self._background_query_repository: Optional[BackgroundQueryRepository] = None
        self._feat_query_repository: Optional[FeatQueryRepository] = None
        
        # Use case factory
        self._use_case_factory: Optional[CompleteUseCaseFactory] = None
    
    def _load_config(self) -> dict:
        """Load configuration from environment variables"""
        return {
            "mongo_uri": os.getenv("MONGO_URI", "mongodb://admin:password@localhost:27017/?authSource=admin"),
            "database_name": os.getenv("DB_NAME", "dnd"),
            "enable_query_caching": os.getenv("ENABLE_QUERY_CACHING", "false").lower() == "true",
            "cache_ttl_seconds": int(os.getenv("CACHE_TTL_SECONDS", "300"))
        }
    
    # Repository getters
    def get_class_query_repository(self) -> ClassQueryRepository:
        """Get configured class query repository"""
        if self._class_query_repository is None:
            logger.info("Initializing MongoDB class query repository")
            self._class_query_repository = MongoDBClassQueryRepository(
                connection_string=self.config["mongo_uri"],
                database_name=self.config["database_name"]
            )
        return self._class_query_repository
    
    def get_spell_query_repository(self) -> SpellQueryRepository:
        """Get configured spell query repository"""
        if self._spell_query_repository is None:
            logger.info("Initializing MongoDB spell query repository")
            self._spell_query_repository = MongoDBSpellQueryRepository(
                connection_string=self.config["mongo_uri"],
                database_name=self.config["database_name"]
            )
        return self._spell_query_repository
    
    def get_monster_query_repository(self) -> MonsterQueryRepository:
        """Get configured monster query repository"""
        if self._monster_query_repository is None:
            logger.info("Initializing MongoDB monster query repository")
            self._monster_query_repository = MongoDBMonsterQueryRepository(
                connection_string=self.config["mongo_uri"],
                database_name=self.config["database_name"]
            )
        return self._monster_query_repository
    
    def get_document_query_repository(self) -> DocumentQueryRepository:
        """Get configured document query repository"""
        if self._document_query_repository is None:
            logger.info("Initializing MongoDB document query repository")
            self._document_query_repository = MongoDBDocumentQueryRepository(
                connection_string=self.config["mongo_uri"],
                database_name=self.config["database_name"]
            )
        return self._document_query_repository
    
    def get_equipment_query_repository(self) -> EquipmentQueryRepository:
        """Get configured equipment query repository"""
        if self._equipment_query_repository is None:
            logger.info("Initializing MongoDB equipment query repository")
            self._equipment_query_repository = MongoDBEquipmentQueryRepository(
                connection_string=self.config["mongo_uri"],
                database_name=self.config["database_name"]
            )
        return self._equipment_query_repository
    
    def get_background_query_repository(self) -> BackgroundQueryRepository:
        """Get configured background query repository"""
        if self._background_query_repository is None:
            logger.info("Initializing MongoDB background query repository")
            self._background_query_repository = MongoDBBackgroundQueryRepository(
                connection_string=self.config["mongo_uri"],
                database_name=self.config["database_name"]
            )
        return self._background_query_repository
    
    def get_feat_query_repository(self) -> FeatQueryRepository:
        """Get configured feat query repository"""
        if self._feat_query_repository is None:
            logger.info("Initializing MongoDB feat query repository")
            self._feat_query_repository = MongoDBFeatQueryRepository(
                connection_string=self.config["mongo_uri"],
                database_name=self.config["database_name"]
            )
        return self._feat_query_repository
    
    def get_use_case_factory(self) -> CompleteUseCaseFactory:
        """Get configured use case factory"""
        if self._use_case_factory is None:
            logger.info("Initializing complete use case factory")
            # For now, we create minimal dependencies for the factory
            # In a full implementation, we'd have proper event publisher and repositories
            from shared_domain.use_cases import EventPublisher
            
            class NoOpEventPublisher(EventPublisher):
                async def publish(self, event) -> None:
                    pass  # No-op for now
            
            # Create dummy repositories for the base factory
            # These won't be used since we override the methods we need
            class DummyClassRepo:
                pass
            
            class DummySpellRepo:
                pass
            
            self._use_case_factory = CompleteUseCaseFactory(
                class_repository=DummyClassRepo(),
                spell_repository=DummySpellRepo(),
                event_publisher=NoOpEventPublisher(),
                spell_query_repository=self.get_spell_query_repository(),
                monster_query_repository=self.get_monster_query_repository(),
                document_query_repository=self.get_document_query_repository(),
                equipment_query_repository=self.get_equipment_query_repository(),
                background_query_repository=self.get_background_query_repository(),
                feat_query_repository=self.get_feat_query_repository()
            )
        return self._use_case_factory
    
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
    
    # New Use Case Methods
    def get_search_spells_use_case(self) -> SearchSpellsUseCase:
        """Get search spells use case"""
        return self.get_use_case_factory().create_search_spells_use_case()
    
    def get_search_monsters_use_case(self) -> SearchMonstersUseCase:
        """Get search monsters use case"""
        return self.get_use_case_factory().create_search_monsters_use_case()
    
    def get_search_documents_use_case(self) -> SearchDocumentsUseCase:
        """Get search documents use case"""
        return self.get_use_case_factory().create_search_documents_use_case()
    
    def get_get_document_use_case(self) -> GetDocumentUseCase:
        """Get single document use case"""
        return self.get_use_case_factory().create_get_document_use_case()
    
    def get_filter_options_use_case(self) -> GetFilterOptionsUseCase:
        """Get filter options use case"""
        return self.get_use_case_factory().create_get_filter_options_use_case()
    
    # Equipment use cases
    def get_search_equipment_use_case(self) -> SearchEquipmentUseCase:
        """Get search equipment use case"""
        return self.get_use_case_factory().create_search_equipment_use_case()
    
    # Background and feat use cases
    def get_search_backgrounds_use_case(self) -> SearchBackgroundsUseCase:
        """Get search backgrounds use case"""
        return self.get_use_case_factory().create_search_backgrounds_use_case()
    
    def get_search_feats_use_case(self) -> SearchFeatsUseCase:
        """Get search feats use case"""
        return self.get_use_case_factory().create_search_feats_use_case()
    
    # Advanced Use Cases (if available)
    def get_advanced_navigation_use_case(self):
        """Get advanced navigation use case"""
        try:
            return self.get_use_case_factory().create_advanced_navigation_use_case()
        except ValueError as e:
            logger.warning(f"Advanced navigation use case not available: {e}")
            return None
    
    def get_content_discovery_use_case(self):
        """Get content discovery use case"""
        try:
            return self.get_use_case_factory().create_content_discovery_use_case()
        except ValueError as e:
            logger.warning(f"Content discovery use case not available: {e}")
            return None
    
    def get_search_suggestion_use_case(self):
        """Get search suggestion use case"""
        try:
            return self.get_use_case_factory().create_search_suggestion_use_case()
        except ValueError as e:
            logger.warning(f"Search suggestion use case not available: {e}")
            return None
    
    async def close(self) -> None:
        """Clean up resources"""
        repositories = [
            ("class_query_repository", self._class_query_repository),
            ("spell_query_repository", self._spell_query_repository),
            ("monster_query_repository", self._monster_query_repository),
            ("document_query_repository", self._document_query_repository),
            ("equipment_query_repository", self._equipment_query_repository),
            ("background_query_repository", self._background_query_repository),
            ("feat_query_repository", self._feat_query_repository)
        ]
        
        for name, repo in repositories:
            if repo and hasattr(repo, 'close'):
                logger.info(f"Closing {name} connection")
                await repo.close()


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