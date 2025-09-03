"""
Dependency injection container for Parser service
Configures all dependencies following hexagonal architecture principles
"""
import os
from typing import Optional
import logging

from shared_domain.entities import ClassRepository, EventPublisher
from srd_parser.adapters.persistence.mongodb_class_repository import MongoDBClassRepository
from srd_parser.adapters.events.in_memory_event_publisher import InMemoryEventPublisher, LoggingEventPublisher, CompositeEventPublisher
from srd_parser.domain.services import ClassParsingService
from srd_parser.application.command_handlers import ParseMultipleClassesHandler, ValidateClassDataHandler

logger = logging.getLogger(__name__)


class ParserContainer:
    """Dependency injection container for Parser service"""
    
    def __init__(self, config: Optional[dict] = None):
        self.config = config or self._load_config()
        self._class_repository: Optional[ClassRepository] = None
        self._event_publisher: Optional[EventPublisher] = None
        self._parsing_service: Optional[ClassParsingService] = None
    
    def _load_config(self) -> dict:
        """Load configuration from environment variables"""
        return {
            "mongo_uri": os.getenv("MONGO_URI", "mongodb://admin:password@localhost:27017/?authSource=admin"),
            "database_name": os.getenv("DB_NAME", "dnd"),
            "enable_event_logging": os.getenv("ENABLE_EVENT_LOGGING", "true").lower() == "true",
            "enable_event_store": os.getenv("ENABLE_EVENT_STORE", "true").lower() == "true"
        }
    
    def get_class_repository(self) -> ClassRepository:
        """Get configured class repository"""
        if self._class_repository is None:
            logger.info("Initializing MongoDB class repository")
            self._class_repository = MongoDBClassRepository(
                connection_string=self.config["mongo_uri"],
                database_name=self.config["database_name"]
            )
        return self._class_repository
    
    def get_event_publisher(self) -> EventPublisher:
        """Get configured event publisher"""
        if self._event_publisher is None:
            logger.info("Initializing event publisher")
            
            publishers = []
            
            # Add event store publisher if enabled
            if self.config["enable_event_store"]:
                publishers.append(InMemoryEventPublisher())
            
            # Add logging publisher if enabled
            if self.config["enable_event_logging"]:
                publishers.append(LoggingEventPublisher())
            
            # Use composite publisher if multiple publishers, otherwise single publisher
            if len(publishers) > 1:
                self._event_publisher = CompositeEventPublisher(publishers)
            elif len(publishers) == 1:
                self._event_publisher = publishers[0]
            else:
                # Fallback to logging publisher
                self._event_publisher = LoggingEventPublisher()
                
        return self._event_publisher
    
    def get_parsing_service(self) -> ClassParsingService:
        """Get configured parsing service"""
        if self._parsing_service is None:
            logger.info("Initializing class parsing service")
            self._parsing_service = ClassParsingService()
        return self._parsing_service
    
    def get_parse_multiple_classes_handler(self) -> ParseMultipleClassesHandler:
        """Get configured handler for parsing multiple classes"""
        return ParseMultipleClassesHandler(
            class_repository=self.get_class_repository(),
            event_publisher=self.get_event_publisher(),
            parsing_service=self.get_parsing_service()
        )
    
    def get_validate_class_data_handler(self) -> ValidateClassDataHandler:
        """Get configured handler for validating class data"""
        return ValidateClassDataHandler(
            parsing_service=self.get_parsing_service()
        )
    
    def close(self) -> None:
        """Clean up resources"""
        if self._class_repository and hasattr(self._class_repository, 'close'):
            logger.info("Closing class repository connection")
            self._class_repository.close()


# Global container instance
_container: Optional[ParserContainer] = None


def get_container() -> ParserContainer:
    """Get global container instance"""
    global _container
    if _container is None:
        _container = ParserContainer()
    return _container


def reset_container() -> None:
    """Reset global container (useful for testing)"""
    global _container
    if _container:
        _container.close()
    _container = None


def configure_container(config: dict) -> ParserContainer:
    """Configure container with specific config"""
    global _container
    if _container:
        _container.close()
    _container = ParserContainer(config)
    return _container