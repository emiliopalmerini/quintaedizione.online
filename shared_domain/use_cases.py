"""
Application Use Cases for both Editor and Parser
Implements hexagonal architecture application layer
"""
from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Dict, List, Optional, Any, Protocol
from .entities import DndClass, ClassId, ClassRepository, SpellRepository, Level, DomainEvent
from .spell_entities import Spell, SpellId, SpellQueryRepository
from .monster_entities import Monster, MonsterId, MonsterQueryRepository
from .document_entities import Document, DocumentId, DocumentQueryRepository
from .equipment_entities import EquipmentQueryRepository
from .background_entities import BackgroundQueryRepository, FeatQueryRepository

# Import advanced use cases if available
try:
    from .advanced_use_cases import (
        AdvancedNavigationUseCase,
        ContentDiscoveryUseCase,
        SearchSuggestionUseCase,
        NavigationQuery,
        ContentDiscoveryQuery,
        SearchSuggestionQuery,
        NavigationContext,
        ContentDiscovery,
        SearchSuggestions
    )
    ADVANCED_USE_CASES_AVAILABLE = True
except ImportError:
    ADVANCED_USE_CASES_AVAILABLE = False


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


# Additional Use Cases for Editor Service

# Spell Use Cases
@dataclass
class SearchSpellsQuery:
    """Query to search spells with various filters"""
    text_query: Optional[str] = None
    level: Optional[int] = None
    school: Optional[str] = None
    character_class: Optional[str] = None
    ritual: Optional[bool] = None
    concentration: Optional[bool] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


class SearchSpellsUseCase:
    """Use case for searching spells"""
    
    def __init__(self, spell_query_repository: SpellQueryRepository):
        self.spell_query_repository = spell_query_repository
    
    async def handle(self, query: SearchSpellsQuery) -> UseCaseResult:
        """Execute spell search"""
        try:
            from .query_models import SpellSearchQuery
            
            # Convert to domain query
            spell_query = SpellSearchQuery(
                text_query=query.text_query,
                level=query.level,
                school=query.school,
                character_class=query.character_class,
                ritual=query.ritual,
                concentration=query.concentration,
                sort_by=query.sort_by,
                limit=query.limit,
                offset=query.offset
            )
            
            spells = await self.spell_query_repository.search_spells(spell_query)
            
            return UseCaseResult(
                success=True,
                message="Spells retrieved successfully",
                data=spells
            )
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Error searching spells: {str(e)}",
                errors=[str(e)]
            )


@dataclass
class GetSpellQuery:
    """Query to get single spell by ID"""
    spell_id: str


class GetSpellUseCase:
    """Use case for getting single spell"""
    
    def __init__(self, spell_query_repository: SpellQueryRepository):
        self.spell_query_repository = spell_query_repository
    
    async def handle(self, query: GetSpellQuery) -> UseCaseResult:
        """Get spell by ID"""
        try:
            spell = await self.spell_query_repository.find_by_id(SpellId(query.spell_id))
            
            if not spell:
                return UseCaseResult(
                    success=False,
                    message="Spell not found"
                )
            
            return UseCaseResult(
                success=True,
                message="Spell retrieved successfully",
                data=spell
            )
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Error getting spell: {str(e)}",
                errors=[str(e)]
            )


# Monster Use Cases
@dataclass
class SearchMonstersQuery:
    """Query to search monsters with various filters"""
    text_query: Optional[str] = None
    creature_type: Optional[str] = None
    min_challenge_rating: Optional[float] = None
    max_challenge_rating: Optional[float] = None
    size: Optional[str] = None
    alignment: Optional[str] = None
    environment: Optional[str] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


class SearchMonstersUseCase:
    """Use case for searching monsters"""
    
    def __init__(self, monster_query_repository: MonsterQueryRepository):
        self.monster_query_repository = monster_query_repository
    
    async def handle(self, query: SearchMonstersQuery) -> UseCaseResult:
        """Execute monster search"""
        try:
            from .query_models import MonsterSearchQuery
            
            # Convert to domain query
            monster_query = MonsterSearchQuery(
                text_query=query.text_query,
                creature_type=query.creature_type,
                min_challenge_rating=query.min_challenge_rating,
                max_challenge_rating=query.max_challenge_rating,
                size=query.size,
                alignment=query.alignment,
                environment=query.environment,
                sort_by=query.sort_by,
                limit=query.limit,
                offset=query.offset
            )
            
            monsters = await self.monster_query_repository.search_monsters(monster_query)
            
            return UseCaseResult(
                success=True,
                message="Monsters retrieved successfully",
                data=monsters
            )
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Error searching monsters: {str(e)}",
                errors=[str(e)]
            )


# Generic Document Use Cases
@dataclass
class SearchDocumentsQuery:
    """Query to search documents of any type"""
    document_type: str
    text_query: Optional[str] = None
    category: Optional[str] = None
    item_type: Optional[str] = None
    rarity: Optional[str] = None
    filters: Optional[Dict[str, Any]] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


class SearchDocumentsUseCase:
    """Use case for searching generic documents"""
    
    def __init__(self, document_query_repository: DocumentQueryRepository):
        self.document_query_repository = document_query_repository
    
    async def handle(self, query: SearchDocumentsQuery) -> UseCaseResult:
        """Execute document search"""
        try:
            from .query_models import DocumentSearchQuery
            
            # Convert to domain query
            doc_query = DocumentSearchQuery(
                document_type=query.document_type,
                text_query=query.text_query,
                category=query.category,
                item_type=query.item_type,
                rarity=query.rarity,
                filters=query.filters or {},
                sort_by=query.sort_by,
                limit=query.limit,
                offset=query.offset
            )
            
            documents = await self.document_query_repository.search_documents(doc_query)
            
            return UseCaseResult(
                success=True,
                message="Documents retrieved successfully",
                data=documents
            )
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Error searching documents: {str(e)}",
                errors=[str(e)]
            )


@dataclass
class GetDocumentQuery:
    """Query to get single document by ID and type"""
    document_id: str
    document_type: str


class GetDocumentUseCase:
    """Use case for getting single document"""
    
    def __init__(self, document_query_repository: DocumentQueryRepository):
        self.document_query_repository = document_query_repository
    
    async def handle(self, query: GetDocumentQuery) -> UseCaseResult:
        """Get document by ID"""
        try:
            document = await self.document_query_repository.find_by_id(
                DocumentId(query.document_id), 
                query.document_type
            )
            
            if not document:
                return UseCaseResult(
                    success=False,
                    message="Document not found"
                )
            
            return UseCaseResult(
                success=True,
                message="Document retrieved successfully",
                data=document
            )
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Error getting document: {str(e)}",
                errors=[str(e)]
            )


@dataclass
class GetFilterOptionsQuery:
    """Query to get filter options for a document type"""
    document_type: str


class GetFilterOptionsUseCase:
    """Use case for getting filter options for a document type"""
    
    def __init__(self, document_query_repository: DocumentQueryRepository):
        self.document_query_repository = document_query_repository
    
    async def handle(self, query: GetFilterOptionsQuery) -> UseCaseResult:
        """Get available filter options"""
        try:
            # Define filter fields by document type
            filter_configs = {
                "incantesimi": ["classi", "scuola"],
                "oggetti_magici": ["tipo", "rarita"],
                "mostri": ["tipo", "allineamento", "taglia"],
                "armi": ["categoria", "maestria", "proprieta"],
                "strumenti": ["caratteristica", "categoria"],
                "servizi": ["categoria", "disponibilita"],
                "equipaggiamento": ["categoria", "peso"],
                "backgrounds": ["competenze_abilita", "competenze_strumenti"],
                "specie": ["tipo_creatura", "velocita_movimento"],
                "talenti": ["aumento_caratteristica"],
                "classi": ["competenze_armi"],
            }
            
            fields = filter_configs.get(query.document_type, [])
            options = {}
            
            for field in fields:
                values = await self.document_query_repository.get_distinct_values(
                    query.document_type, 
                    field
                )
                options[field] = values
            
            return UseCaseResult(
                success=True,
                message="Filter options retrieved successfully",
                data=options
            )
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Error getting filter options: {str(e)}",
                errors=[str(e)]
            )


# Extended Use Case Factory
class ExtendedUseCaseFactory(UseCaseFactory):
    """Extended factory with all use cases for Editor service"""
    
    def __init__(
        self,
        class_repository: ClassRepository,
        spell_repository: SpellRepository,
        event_publisher: EventPublisher,
        spell_query_repository: Optional[SpellQueryRepository] = None,
        monster_query_repository: Optional[MonsterQueryRepository] = None,
        document_query_repository: Optional[DocumentQueryRepository] = None
    ):
        super().__init__(class_repository, spell_repository, event_publisher)
        self.spell_query_repository = spell_query_repository
        self.monster_query_repository = monster_query_repository
        self.document_query_repository = document_query_repository
    
    # Spell use cases
    def create_search_spells_use_case(self) -> SearchSpellsUseCase:
        if not self.spell_query_repository:
            raise ValueError("SpellQueryRepository required for SearchSpellsUseCase")
        return SearchSpellsUseCase(self.spell_query_repository)
    
    def create_get_spell_use_case(self) -> GetSpellUseCase:
        if not self.spell_query_repository:
            raise ValueError("SpellQueryRepository required for GetSpellUseCase")
        return GetSpellUseCase(self.spell_query_repository)
    
    # Monster use cases
    def create_search_monsters_use_case(self) -> SearchMonstersUseCase:
        if not self.monster_query_repository:
            raise ValueError("MonsterQueryRepository required for SearchMonstersUseCase")
        return SearchMonstersUseCase(self.monster_query_repository)
    
    # Document use cases
    def create_search_documents_use_case(self) -> SearchDocumentsUseCase:
        if not self.document_query_repository:
            raise ValueError("DocumentQueryRepository required for SearchDocumentsUseCase")
        return SearchDocumentsUseCase(self.document_query_repository)
    
    def create_get_document_use_case(self) -> GetDocumentUseCase:
        if not self.document_query_repository:
            raise ValueError("DocumentQueryRepository required for GetDocumentUseCase")
        return GetDocumentUseCase(self.document_query_repository)
    
    def create_get_filter_options_use_case(self) -> GetFilterOptionsUseCase:
        if not self.document_query_repository:
            raise ValueError("DocumentQueryRepository required for GetFilterOptionsUseCase")
        return GetFilterOptionsUseCase(self.document_query_repository)


# Equipment Use Cases
@dataclass
class SearchWeaponsQuery:
    """Query to search weapons with various filters"""
    text_query: Optional[str] = None
    category: Optional[str] = None
    weapon_type: Optional[str] = None
    properties: Optional[List[str]] = None
    proficiency: Optional[str] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


@dataclass
class SearchArmorQuery:
    """Query to search armor with various filters"""
    text_query: Optional[str] = None
    category: Optional[str] = None
    armor_type: Optional[str] = None
    min_armor_class: Optional[int] = None
    max_armor_class: Optional[int] = None
    stealth_disadvantage: Optional[bool] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


@dataclass
class SearchMagicItemsQuery:
    """Query to search magic items with various filters"""
    text_query: Optional[str] = None
    item_type: Optional[str] = None
    rarity: Optional[str] = None
    requires_attunement: Optional[bool] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


class SearchEquipmentUseCase:
    """Use case for searching equipment (weapons, armor, magic items)"""
    
    def __init__(self, equipment_query_repository: EquipmentQueryRepository):
        self.equipment_query_repository = equipment_query_repository
    
    async def search_weapons(self, query: SearchWeaponsQuery) -> UseCaseResult:
        """Search weapons"""
        try:
            from .query_models import WeaponSearchQuery
            
            weapon_query = WeaponSearchQuery(
                text_query=query.text_query,
                category=query.category,
                weapon_type=query.weapon_type,
                properties=query.properties,
                proficiency=query.proficiency,
                sort_by=query.sort_by,
                limit=query.limit,
                offset=query.offset
            )
            
            weapons = await self.equipment_query_repository.search_weapons(weapon_query)
            
            return UseCaseResult(
                success=True,
                message="Weapons retrieved successfully",
                data=weapons
            )
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Error searching weapons: {str(e)}",
                errors=[str(e)]
            )
    
    async def search_armor(self, query: SearchArmorQuery) -> UseCaseResult:
        """Search armor"""
        try:
            from .query_models import ArmorSearchQuery
            
            armor_query = ArmorSearchQuery(
                text_query=query.text_query,
                category=query.category,
                armor_type=query.armor_type,
                min_armor_class=query.min_armor_class,
                max_armor_class=query.max_armor_class,
                stealth_disadvantage=query.stealth_disadvantage,
                sort_by=query.sort_by,
                limit=query.limit,
                offset=query.offset
            )
            
            armor = await self.equipment_query_repository.search_armor(armor_query)
            
            return UseCaseResult(
                success=True,
                message="Armor retrieved successfully",
                data=armor
            )
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Error searching armor: {str(e)}",
                errors=[str(e)]
            )
    
    async def search_magic_items(self, query: SearchMagicItemsQuery) -> UseCaseResult:
        """Search magic items"""
        try:
            from .query_models import MagicItemSearchQuery
            
            magic_item_query = MagicItemSearchQuery(
                text_query=query.text_query,
                item_type=query.item_type,
                rarity=query.rarity,
                requires_attunement=query.requires_attunement,
                sort_by=query.sort_by,
                limit=query.limit,
                offset=query.offset
            )
            
            magic_items = await self.equipment_query_repository.search_magic_items(magic_item_query)
            
            return UseCaseResult(
                success=True,
                message="Magic items retrieved successfully",
                data=magic_items
            )
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Error searching magic items: {str(e)}",
                errors=[str(e)]
            )


# Background and Feat Use Cases
@dataclass
class SearchBackgroundsQuery:
    """Query to search backgrounds with various filters"""
    text_query: Optional[str] = None
    skill_proficiencies: Optional[List[str]] = None
    tool_proficiencies: Optional[List[str]] = None
    languages: Optional[List[str]] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


@dataclass
class SearchFeatsQuery:
    """Query to search feats with various filters"""
    text_query: Optional[str] = None
    ability_score_increases: Optional[List[str]] = None
    has_prerequisites: Optional[bool] = None
    sort_by: str = "name"
    limit: Optional[int] = None
    offset: Optional[int] = None


class SearchBackgroundsUseCase:
    """Use case for searching backgrounds"""
    
    def __init__(self, background_query_repository: BackgroundQueryRepository):
        self.background_query_repository = background_query_repository
    
    async def handle(self, query: SearchBackgroundsQuery) -> UseCaseResult:
        """Search backgrounds"""
        try:
            from .query_models import BackgroundSearchQuery
            
            bg_query = BackgroundSearchQuery(
                text_query=query.text_query,
                skill_proficiencies=query.skill_proficiencies,
                tool_proficiencies=query.tool_proficiencies,
                languages=query.languages,
                sort_by=query.sort_by,
                limit=query.limit,
                offset=query.offset
            )
            
            backgrounds = await self.background_query_repository.search_backgrounds(bg_query)
            
            return UseCaseResult(
                success=True,
                message="Backgrounds retrieved successfully",
                data=backgrounds
            )
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Error searching backgrounds: {str(e)}",
                errors=[str(e)]
            )


class SearchFeatsUseCase:
    """Use case for searching feats"""
    
    def __init__(self, feat_query_repository: FeatQueryRepository):
        self.feat_query_repository = feat_query_repository
    
    async def handle(self, query: SearchFeatsQuery) -> UseCaseResult:
        """Search feats"""
        try:
            from .query_models import FeatSearchQuery
            
            feat_query = FeatSearchQuery(
                text_query=query.text_query,
                ability_score_increases=query.ability_score_increases,
                has_prerequisites=query.has_prerequisites,
                sort_by=query.sort_by,
                limit=query.limit,
                offset=query.offset
            )
            
            feats = await self.feat_query_repository.search_feats(feat_query)
            
            return UseCaseResult(
                success=True,
                message="Feats retrieved successfully",
                data=feats
            )
        except Exception as e:
            return UseCaseResult(
                success=False,
                message=f"Error searching feats: {str(e)}",
                errors=[str(e)]
            )


# Extended Use Case Factory Updates
class CompleteUseCaseFactory(ExtendedUseCaseFactory):
    """Complete factory with all use cases for Editor service including equipment and backgrounds"""
    
    def __init__(
        self,
        class_repository: ClassRepository,
        spell_repository: SpellRepository,
        event_publisher: EventPublisher,
        spell_query_repository: Optional[SpellQueryRepository] = None,
        monster_query_repository: Optional[MonsterQueryRepository] = None,
        document_query_repository: Optional[DocumentQueryRepository] = None,
        equipment_query_repository: Optional[EquipmentQueryRepository] = None,
        background_query_repository: Optional[BackgroundQueryRepository] = None,
        feat_query_repository: Optional[FeatQueryRepository] = None
    ):
        super().__init__(
            class_repository, spell_repository, event_publisher,
            spell_query_repository, monster_query_repository, document_query_repository
        )
        self.equipment_query_repository = equipment_query_repository
        self.background_query_repository = background_query_repository
        self.feat_query_repository = feat_query_repository
    
    # Equipment use cases
    def create_search_equipment_use_case(self) -> SearchEquipmentUseCase:
        if not self.equipment_query_repository:
            raise ValueError("EquipmentQueryRepository required for SearchEquipmentUseCase")
        return SearchEquipmentUseCase(self.equipment_query_repository)
    
    # Background and feat use cases
    def create_search_backgrounds_use_case(self) -> SearchBackgroundsUseCase:
        if not self.background_query_repository:
            raise ValueError("BackgroundQueryRepository required for SearchBackgroundsUseCase")
        return SearchBackgroundsUseCase(self.background_query_repository)
    
    def create_search_feats_use_case(self) -> SearchFeatsUseCase:
        if not self.feat_query_repository:
            raise ValueError("FeatQueryRepository required for SearchFeatsUseCase")
        return SearchFeatsUseCase(self.feat_query_repository)
    
    # Advanced use cases
    def create_advanced_navigation_use_case(self) -> 'AdvancedNavigationUseCase':
        """Create advanced navigation use case if available"""
        if not ADVANCED_USE_CASES_AVAILABLE:
            raise ValueError("Advanced use cases not available")
        return AdvancedNavigationUseCase(
            self.spell_query_repository,
            self.monster_query_repository, 
            self.document_query_repository
        )
    
    def create_content_discovery_use_case(self) -> 'ContentDiscoveryUseCase':
        """Create content discovery use case if available"""
        if not ADVANCED_USE_CASES_AVAILABLE:
            raise ValueError("Advanced use cases not available")
        return ContentDiscoveryUseCase(
            self.spell_query_repository,
            self.monster_query_repository,
            self.document_query_repository
        )
    
    def create_search_suggestion_use_case(self) -> 'SearchSuggestionUseCase':
        """Create search suggestion use case if available"""  
        if not ADVANCED_USE_CASES_AVAILABLE:
            raise ValueError("Advanced use cases not available")
        return SearchSuggestionUseCase(
            self.spell_query_repository,
            self.monster_query_repository,
            self.document_query_repository
        )