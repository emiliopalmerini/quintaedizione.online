"""
Application Use Cases for both Editor and Parser
Implements hexagonal architecture application layer
"""
from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Dict, List, Optional, Any, Protocol
from .entities import DndClass, ClassId, ClassRepository, SpellRepository, Level, DomainEvent


# Command/Query separation
class Command(Protocol):
    """Marker for command operations (write-side)"""
    pass


class Query(Protocol):
    """Marker for query operations (read-side)"""
    pass


# Event Publisher Port
class EventPublisher(ABC):
    """Port for publishing domain events"""
    
    @abstractmethod
    async def publish(self, event: DomainEvent) -> None:
        pass


# Use Case Results
@dataclass
class UseCaseResult:
    """Base result for use case operations"""
    success: bool
    message: str
    data: Optional[Any] = None
    errors: List[str] = None


# Parser Use Cases (Write-Side)
@dataclass
class ParseClassCommand:
    """Command to parse a class from markdown"""
    class_name: str
    markdown_content: str
    source_version: str


class ParseClassUseCase:
    """Use case for parsing D&D classes from markdown"""
    
    def __init__(
        self,
        class_repository: ClassRepository,
        event_publisher: EventPublisher
    ):
        self.class_repository = class_repository
        self.event_publisher = event_publisher
    
    async def execute(self, command: ParseClassCommand) -> UseCaseResult:
        """Execute class parsing"""
        try:
            # Check if class already exists
            class_id = ClassId(self._slugify(command.class_name))
            existing = await self.class_repository.find_by_id(class_id)
            
            if existing and existing.version >= command.source_version:
                return UseCaseResult(
                    success=False,
                    message=f"Class {command.class_name} already up to date",
                    data=existing.to_dict()
                )
            
            # Parse the class (this would use the domain parser)
            parsed_class = await self._parse_class_from_markdown(
                command.class_name,
                command.markdown_content,
                command.source_version
            )
            
            # Validate business rules
            validation_errors = self._validate_class(parsed_class)
            if validation_errors:
                return UseCaseResult(
                    success=False,
                    message="Class validation failed",
                    errors=validation_errors
                )
            
            # Save to repository
            await self.class_repository.save(parsed_class)
            
            # Publish domain event
            from .entities import ClassParsed
            import uuid
            from datetime import datetime
            
            event = ClassParsed(
                event_id=str(uuid.uuid4()),
                timestamp=datetime.utcnow().isoformat(),
                aggregate_id=parsed_class.id.value,
                class_name=parsed_class.name,
                version=command.source_version
            )
            await self.event_publisher.publish(event)
            
            return UseCaseResult(
                success=True,
                message=f"Successfully parsed class {command.class_name}",
                data=parsed_class.to_dict()
            )
            
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Failed to parse class: {str(e)}",
                errors=[str(e)]
            )
    
    async def _parse_class_from_markdown(
        self, 
        name: str, 
        content: str, 
        version: str
    ) -> DndClass:
        """Parse class from markdown (would integrate with domain parser)"""
        # This would use the improved parser we created
        from srd_parser.parsers.classes_improved import ImprovedClassParser
        
        parser = ImprovedClassParser()
        # Convert to domain entity... (simplified for example)
        class_id = ClassId(self._slugify(name))
        
        # This would be much more sophisticated in reality
        from .entities import Ability
        return DndClass(
            id=class_id,
            name=name,
            primary_ability=Ability.FORZA,  # Would be parsed
            hit_die="d8",  # Would be parsed
            version=version
        )
    
    def _validate_class(self, dnd_class: DndClass) -> List[str]:
        """Validate parsed class"""
        from .entities import ClassValidationService
        return ClassValidationService.validate_class_consistency(dnd_class)
    
    def _slugify(self, text: str) -> str:
        """Convert text to slug"""
        import re
        return re.sub(r'[^a-z0-9-]', '', text.lower().replace(' ', '-'))


# Editor Use Cases (Read-Side)
@dataclass
class GetClassQuery:
    """Query to get a specific class"""
    class_id: str


@dataclass  
class SearchClassesQuery:
    """Query to search classes"""
    search_term: Optional[str] = None
    filters: Optional[Dict[str, Any]] = None
    page: int = 1
    page_size: int = 20


@dataclass
class GetClassStatsQuery:
    """Query to get class statistics"""
    class_id: str
    level: int


class GetClassUseCase:
    """Use case for retrieving a specific class"""
    
    def __init__(
        self,
        class_repository: ClassRepository,
        event_publisher: EventPublisher
    ):
        self.class_repository = class_repository
        self.event_publisher = event_publisher
    
    async def execute(self, query: GetClassQuery) -> UseCaseResult:
        """Get class by ID"""
        try:
            class_id = ClassId(query.class_id)
            dnd_class = await self.class_repository.find_by_id(class_id)
            
            if not dnd_class:
                return UseCaseResult(
                    success=False,
                    message=f"Class {query.class_id} not found"
                )
            
            # Publish view event for analytics
            from .entities import ClassViewed
            import uuid
            from datetime import datetime
            
            event = ClassViewed(
                event_id=str(uuid.uuid4()),
                timestamp=datetime.utcnow().isoformat(),
                aggregate_id=dnd_class.id.value,
                class_name=dnd_class.name
            )
            await self.event_publisher.publish(event)
            
            return UseCaseResult(
                success=True,
                message="Class found",
                data=self._format_class_for_display(dnd_class)
            )
            
        except ValueError as e:
            return UseCaseResult(
                success=False,
                message=f"Invalid class ID: {e}"
            )
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Failed to retrieve class: {e}"
            )
    
    def _format_class_for_display(self, dnd_class: DndClass) -> Dict[str, Any]:
        """Format class data for UI display"""
        return {
            **dnd_class.to_dict(),
            "core_features": [
                {
                    "name": f.name,
                    "level": f.level.value,
                    "summary": f.get_summary()
                }
                for f in dnd_class.get_core_features()
            ],
            "is_spellcaster": dnd_class.is_spellcaster(),
            "is_full_caster": dnd_class.is_full_caster(),
            "spell_levels_available": [
                dnd_class.get_max_spell_level_at(Level(level))
                for level in range(1, 21)
            ] if dnd_class.is_spellcaster() else []
        }


class SearchClassesUseCase:
    """Use case for searching classes"""
    
    def __init__(self, class_repository: ClassRepository):
        self.class_repository = class_repository
    
    async def execute(self, query: SearchClassesQuery) -> UseCaseResult:
        """Search classes with filters and pagination"""
        try:
            classes = await self.class_repository.search(
                query.search_term or "",
                query.filters or {}
            )
            
            # Apply pagination
            start_idx = (query.page - 1) * query.page_size
            end_idx = start_idx + query.page_size
            paginated_classes = classes[start_idx:end_idx]
            
            return UseCaseResult(
                success=True,
                message=f"Found {len(classes)} classes",
                data={
                    "classes": [self._format_class_summary(c) for c in paginated_classes],
                    "total_count": len(classes),
                    "page": query.page,
                    "page_size": query.page_size,
                    "has_more": end_idx < len(classes)
                }
            )
            
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Search failed: {e}"
            )
    
    def _format_class_summary(self, dnd_class: DndClass) -> Dict[str, Any]:
        """Format class for search results"""
        return {
            "id": dnd_class.id.value,
            "name": dnd_class.name,
            "primary_ability": dnd_class.primary_ability.value,
            "hit_die": dnd_class.hit_die,
            "is_spellcaster": dnd_class.is_spellcaster(),
            "feature_count": len(dnd_class.features),
            "subclass_count": len(dnd_class.subclasses)
        }


class GetClassStatsUseCase:
    """Use case for getting class statistics at specific level"""
    
    def __init__(self, class_repository: ClassRepository):
        self.class_repository = class_repository
    
    async def execute(self, query: GetClassStatsQuery) -> UseCaseResult:
        """Get class statistics at specific level"""
        try:
            class_id = ClassId(query.class_id)
            dnd_class = await self.class_repository.find_by_id(class_id)
            
            if not dnd_class:
                return UseCaseResult(
                    success=False,
                    message=f"Class {query.class_id} not found"
                )
            
            level = Level(query.level)
            stats = self._calculate_class_stats(dnd_class, level)
            
            return UseCaseResult(
                success=True,
                message="Statistics calculated",
                data=stats
            )
            
        except ValueError as e:
            return UseCaseResult(
                success=False,
                message=f"Invalid parameters: {e}"
            )
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Failed to calculate stats: {e}"
            )
    
    def _calculate_class_stats(self, dnd_class: DndClass, level: Level) -> Dict[str, Any]:
        """Calculate class statistics at level"""
        features_at_level = dnd_class.get_features_at_level(level)
        
        stats = {
            "level": level.value,
            "class_name": dnd_class.name,
            "features_available": len(features_at_level),
            "proficiency_bonus": 2 + ((level.value - 1) // 4),
            "features": [
                {
                    "name": f.name,
                    "level_gained": f.level.value,
                    "is_new": f.level.value == level.value
                }
                for f in features_at_level
            ]
        }
        
        # Add spellcasting stats if applicable
        if dnd_class.is_spellcaster() and dnd_class.spell_progression:
            stats["spellcasting"] = {
                "cantrips_known": dnd_class.spell_progression.get_cantrips_at_level(level),
                "max_spell_level": dnd_class.get_max_spell_level_at(level),
                "spell_slots": dnd_class.spell_progression.spell_slots_by_level.get(level.value, [])
            }
        
        return stats


# Use Case Factory for Dependency Injection
class UseCaseFactory:
    """Factory for creating use cases with proper dependencies"""
    
    def __init__(
        self,
        class_repository: ClassRepository,
        spell_repository: SpellRepository,
        event_publisher: EventPublisher
    ):
        self.class_repository = class_repository
        self.spell_repository = spell_repository
        self.event_publisher = event_publisher
    
    def create_parse_class_use_case(self) -> ParseClassUseCase:
        return ParseClassUseCase(self.class_repository, self.event_publisher)
    
    def create_get_class_use_case(self) -> GetClassUseCase:
        return GetClassUseCase(self.class_repository, self.event_publisher)
    
    def create_search_classes_use_case(self) -> SearchClassesUseCase:
        return SearchClassesUseCase(self.class_repository)
    
    def create_get_class_stats_use_case(self) -> GetClassStatsUseCase:
        return GetClassStatsUseCase(self.class_repository)