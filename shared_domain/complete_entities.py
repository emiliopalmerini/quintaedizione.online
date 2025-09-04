"""
Complete D&D 5e SRD Domain Model
Aggregates all entities following the ADR data model specification

This module provides the complete domain model for the D&D 5e SRD system,
including all entity types defined in the ADR with their business logic,
validation rules, and repository interfaces.
"""
from __future__ import annotations

# Core entities and base types
from .entities import (
    # Base types
    Ability, EntityId, ClassId, Level, HitDie,
    
    # Domain entities
    DndClass, Subclass, ClassFeature, SpellProgression,
    
    # Repository interfaces
    ClassRepository, ClassQueryRepository,
    
    
    # Validation services
    ClassValidationService,
    
    # Events
    DomainEvent, EventPublisher
)

# Spell entities
from .spell_entities import (
    # Spell enums and value objects
    SpellSchool, CastingTime, SpellRange, SpellDuration,
    SpellId, SpellLevel, SpellComponent, SpellCasting,
    
    # Spell entity
    Spell,
    
    # Repository interfaces
    SpellRepository, SpellQueryRepository,
    
    # Validation service
    SpellValidationService
)

# Monster entities  
from .monster_entities import (
    # Monster enums and value objects
    MonsterSize, MonsterType, Alignment, DamageType, ConditionType,
    MonsterId, ChallengeRating, AbilityScores, DamageResistance, Speed,
    MonsterAttack, MonsterAction, MonsterTrait,
    
    # Monster entity
    Monster,
    
    # Repository interfaces
    MonsterRepository, MonsterQueryRepository,
    
    # Validation service
    MonsterValidationService
)

# Equipment entities
from .equipment_entities import (
    # Equipment enums and value objects
    WeaponCategory, WeaponProperty, ArmorCategory, MagicItemRarity,
    EquipmentId, Currency, Weight, DamageInfo, WeaponRange,
    
    # Equipment entities
    Weapon, Armor, Tool, AdventuringGear, MagicItem,
    
    # Repository interfaces
    EquipmentRepository, EquipmentQueryRepository,
    
    # Validation service
    EquipmentValidationService
)

# Background and feat entities
from .background_entities import (
    # Background and feat enums
    AbilityName, FeatCategory,
    BackgroundId, FeatId,
    
    # Value objects
    EquipmentOption, FeatBenefit, AbilityScoreIncrease,
    
    # Entities
    Background, Feat, Service,
    
    # Repository interfaces
    BackgroundRepository, FeatRepository,
    BackgroundQueryRepository, FeatQueryRepository,
    
    # Validation services
    BackgroundValidationService, FeatValidationService
)

# Document entities
from .document_entities import (
    # Document value objects
    DocumentId, DocumentSubparagraph, DocumentParagraphBody, DocumentParagraph,
    
    # Document entity
    Document,
    
    # Repository interfaces
    DocumentRepository, DocumentQueryRepository,
    
    # Validation service
    DocumentValidationService
)

# Query models (CQRS read models)
from .query_models import (
    # Generic query result
    QueryResult,
    
    # Class query models
    ClassSearchQuery, ClassSummary, ClassDetail,
    ClassFeatureDetail, SpellSlotProgression, SubclassDetail,
    
    # Spell query models
    SpellSearchQuery, SpellSummary, SpellDetail,
    
    # Monster query models
    MonsterSearchQuery, MonsterSummary, MonsterDetail,
    
    # Equipment query models
    WeaponSearchQuery, ArmorSearchQuery, MagicItemSearchQuery,
    WeaponSummary, ArmorSummary, MagicItemSummary,
    
    # Background and feat query models
    BackgroundSearchQuery, FeatSearchQuery,
    BackgroundSummary, BackgroundDetail, FeatSummary, FeatDetail,
    
    # Document query models
    DocumentSearchQuery, DocumentSummary, DocumentDetail, DocumentContentMatch
)


# Export all public interfaces
__all__ = [
    # === CORE ENTITIES ===
    "Ability", "EntityId", "ClassId", "Level", "HitDie",
    "DndClass", "Subclass", "ClassFeature", "SpellProgression",
    "ClassRepository", "ClassQueryRepository",
    "ClassValidationService",
    "DomainEvent", "EventPublisher",
    
    # === SPELL ENTITIES ===
    "SpellSchool", "CastingTime", "SpellRange", "SpellDuration",
    "SpellId", "SpellLevel", "SpellComponent", "SpellCasting",
    "Spell", "SpellRepository", "SpellQueryRepository", "SpellValidationService",
    
    # === MONSTER ENTITIES ===
    "MonsterSize", "MonsterType", "Alignment", "DamageType", "ConditionType",
    "MonsterId", "ChallengeRating", "AbilityScores", "DamageResistance", "Speed",
    "MonsterAttack", "MonsterAction", "MonsterTrait",
    "Monster", "MonsterRepository", "MonsterQueryRepository", "MonsterValidationService",
    
    # === EQUIPMENT ENTITIES ===
    "WeaponCategory", "WeaponProperty", "ArmorCategory", "MagicItemRarity",
    "EquipmentId", "Currency", "Weight", "DamageInfo", "WeaponRange",
    "Weapon", "Armor", "Tool", "AdventuringGear", "MagicItem",
    "EquipmentRepository", "EquipmentQueryRepository", "EquipmentValidationService",
    
    # === BACKGROUND & FEAT ENTITIES ===
    "AbilityName", "FeatCategory", "BackgroundId", "FeatId",
    "EquipmentOption", "FeatBenefit", "AbilityScoreIncrease",
    "Background", "Feat", "Service",
    "BackgroundRepository", "FeatRepository", "BackgroundQueryRepository", "FeatQueryRepository",
    "BackgroundValidationService", "FeatValidationService",
    
    # === DOCUMENT ENTITIES ===
    "DocumentId", "DocumentSubparagraph", "DocumentParagraphBody", "DocumentParagraph",
    "Document", "DocumentRepository", "DocumentQueryRepository", "DocumentValidationService",
    
    # === QUERY MODELS (CQRS) ===
    "QueryResult",
    "ClassSearchQuery", "ClassSummary", "ClassDetail", "ClassFeatureDetail", "SpellSlotProgression", "SubclassDetail",
    "SpellSearchQuery", "SpellSummary", "SpellDetail",
    "MonsterSearchQuery", "MonsterSummary", "MonsterDetail", 
    "WeaponSearchQuery", "ArmorSearchQuery", "MagicItemSearchQuery",
    "WeaponSummary", "ArmorSummary", "MagicItemSummary",
    "BackgroundSearchQuery", "FeatSearchQuery", "BackgroundSummary", "BackgroundDetail", "FeatSummary", "FeatDetail",
    "DocumentSearchQuery", "DocumentSummary", "DocumentDetail", "DocumentContentMatch"
]


class SRDDomainModel:
    """
    Facade for the complete SRD domain model
    
    This class provides easy access to all entity types and their associated
    repositories, validation services, and query models. It serves as the
    main entry point for interacting with the domain model.
    """
    
    # Entity types by collection (matching ADR specification)
    ENTITY_TYPES = {
        "classi": DndClass,
        "incantesimi": Spell,
        "mostri": Monster,
        "animali": Monster,  # Animals use same entity as monsters
        "armi": Weapon,
        "armature": Armor,
        "strumenti": Tool,
        "equipaggiamento": AdventuringGear,
        "oggetti_magici": MagicItem,
        "backgrounds": Background,
        "talenti": Feat,
        "servizi": Service,
        "documenti": Document,
    }
    
    # Repository types for write operations
    WRITE_REPOSITORIES = {
        "classi": ClassRepository,
        "incantesimi": SpellRepository,
        "mostri": MonsterRepository,
        "animali": MonsterRepository,
        "armi": EquipmentRepository,  # Handles multiple equipment types
        "armature": EquipmentRepository,
        "strumenti": EquipmentRepository,
        "equipaggiamento": EquipmentRepository,
        "oggetti_magici": EquipmentRepository,
        "backgrounds": BackgroundRepository,
        "talenti": FeatRepository,
        "documenti": DocumentRepository,
    }
    
    # Repository types for read operations (CQRS)
    QUERY_REPOSITORIES = {
        "classi": ClassQueryRepository,
        "incantesimi": SpellQueryRepository,
        "mostri": MonsterQueryRepository,
        "animali": MonsterQueryRepository,
        "armi": EquipmentQueryRepository,
        "armature": EquipmentQueryRepository,
        "strumenti": EquipmentQueryRepository,
        "equipaggiamento": EquipmentQueryRepository,
        "oggetti_magici": EquipmentQueryRepository,
        "backgrounds": BackgroundQueryRepository,
        "talenti": FeatQueryRepository,
        "documenti": DocumentQueryRepository,
    }
    
    # Validation services
    VALIDATION_SERVICES = {
        "classi": ClassValidationService,
        "incantesimi": SpellValidationService,
        "mostri": MonsterValidationService,
        "animali": MonsterValidationService,
        "armi": EquipmentValidationService,
        "armature": EquipmentValidationService,
        "strumenti": EquipmentValidationService,
        "equipaggiamento": EquipmentValidationService,
        "oggetti_magici": EquipmentValidationService,
        "backgrounds": BackgroundValidationService,
        "talenti": FeatValidationService,
        "documenti": DocumentValidationService,
    }
    
    @classmethod
    def get_entity_type(cls, collection_name: str):
        """Get entity type for collection"""
        return cls.ENTITY_TYPES.get(collection_name)
    
    @classmethod
    def get_write_repository_type(cls, collection_name: str):
        """Get write repository interface for collection"""
        return cls.WRITE_REPOSITORIES.get(collection_name)
    
    @classmethod
    def get_query_repository_type(cls, collection_name: str):
        """Get query repository interface for collection"""
        return cls.QUERY_REPOSITORIES.get(collection_name)
    
    @classmethod
    def get_validation_service(cls, collection_name: str):
        """Get validation service for collection"""
        return cls.VALIDATION_SERVICES.get(collection_name)
    
    @classmethod
    def get_all_collections(cls) -> list[str]:
        """Get all supported collection names"""
        return list(cls.ENTITY_TYPES.keys())
    
    @classmethod
    def is_supported_collection(cls, collection_name: str) -> bool:
        """Check if collection is supported by domain model"""
        return collection_name in cls.ENTITY_TYPES
    
    @classmethod
    def get_domain_info(cls) -> dict[str, any]:
        """Get information about the domain model"""
        return {
            "total_entity_types": len(cls.ENTITY_TYPES),
            "supported_collections": cls.get_all_collections(),
            "has_cqrs_separation": True,
            "has_validation_services": True,
            "has_event_publishing": True,
            "architecture_pattern": "Hexagonal Architecture with DDD",
            "design_patterns": [
                "Repository Pattern",
                "CQRS (Command Query Responsibility Segregation)",
                "Domain-Driven Design",
                "Event-Driven Architecture",
                "Value Object Pattern",
                "Aggregate Pattern"
            ]
        }