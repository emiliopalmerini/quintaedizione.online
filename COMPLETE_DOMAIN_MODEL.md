# Complete D&D 5e SRD Domain Model

This document provides an overview of the complete domain model implementation for the D&D 5e SRD system, following the ADR data model specification.

## Architecture Overview

The domain model follows **Hexagonal Architecture** with **Domain-Driven Design** principles and implements **CQRS** (Command Query Responsibility Segregation) pattern for optimal read/write operations.

```
┌─────────────────────────────────────────────────┐
│              Shared Domain Layer                │
│  (All D&D 5e entities with business logic)     │
└─────────────────────────────────────────────────┘
           │                         │
           ▼                         ▼
┌─────────────────┐         ┌─────────────────┐
│ Write Repositories      │  │ Query Repositories      │
│ (Parser Service)        │  │ (Editor Service)        │
└─────────────────┘         └─────────────────┘
```

## Entity Types Implementation

Based on the ADR data model, all major D&D 5e entity types have been implemented:

### ✅ **Classi** (Classes)
- **Entity**: `DndClass`, `Subclass`, `ClassFeature`, `SpellProgression`
- **Location**: `/shared_domain/entities.py`
- **Features**: Complete class progression, multiclassing rules, spellcasting, subclasses
- **Validation**: Class feature distribution, spellcasting consistency, ability requirements

### ✅ **Incantesimi** (Spells)
- **Entity**: `Spell`, `SpellCasting`, `SpellComponent`
- **Location**: `/shared_domain/spell_entities.py`
- **Features**: All schools, levels 0-9, components, concentration, rituals
- **Validation**: Component requirements, concentration rules, class availability

### ✅ **Mostri** (Monsters)
- **Entity**: `Monster`, `MonsterAction`, `MonsterTrait`, `AbilityScores`
- **Location**: `/shared_domain/monster_entities.py`
- **Features**: All sizes/types, challenge ratings, legendary actions, spellcasting
- **Validation**: CR/XP consistency, legendary creature rules, ability score ranges

### ✅ **Animali** (Animals)
- **Entity**: Uses same `Monster` entity (different collection)
- **Features**: Same as monsters but for animal stat blocks

### ✅ **Armi** (Weapons)
- **Entity**: `Weapon`, `DamageInfo`, `WeaponRange`
- **Location**: `/shared_domain/equipment_entities.py`
- **Features**: All categories, properties, mastery system, damage types
- **Validation**: Property combinations, range requirements, damage consistency

### ✅ **Armature** (Armor)
- **Entity**: `Armor`
- **Features**: All categories, AC calculations, strength requirements, stealth
- **Validation**: Category rules, stealth disadvantage logic

### ✅ **Strumenti** (Tools)
- **Entity**: `Tool`
- **Features**: Tool categories, associated abilities, proficiency bonuses
- **Validation**: Category consistency, ability associations

### ✅ **Equipaggiamento** (Adventuring Gear)
- **Entity**: `AdventuringGear`
- **Features**: General equipment, costs, weights, categories
- **Validation**: Cost/weight reasonableness

### ✅ **Oggetti Magici** (Magic Items)
- **Entity**: `MagicItem`
- **Features**: All rarities, attunement, item types, estimated costs
- **Validation**: Rarity/attunement rules, cost estimation

### ✅ **Backgrounds** (Backgrounds)
- **Entity**: `Background`, `EquipmentOption`
- **Location**: `/shared_domain/background_entities.py`
- **Features**: Ability scores, skills, tools, equipment options, feats
- **Validation**: Ability score distribution, skill counts, equipment balance

### ✅ **Talenti** (Feats)
- **Entity**: `Feat`, `FeatBenefit`, `AbilityScoreIncrease`
- **Features**: All categories, prerequisites, ability increases, class restrictions
- **Validation**: Prerequisite logic, ability increase limits, category rules

### ✅ **Servizi** (Services)
- **Entity**: `Service`
- **Features**: Service categories, costs, availability
- **Validation**: Cost reasonableness

### ✅ **Documenti** (Documents)
- **Entity**: `Document`, `DocumentParagraph`, `DocumentSubparagraph`
- **Location**: `/shared_domain/document_entities.py`
- **Features**: Nested paragraph structure, page references, search, TOC
- **Validation**: Paragraph numbering, content length, structure consistency

## Domain Model Features

### 🏗️ **Rich Domain Entities**
- Business logic encapsulation within entities
- Value objects with validation (Level, HitDie, SpellLevel, etc.)
- Immutable value objects prevent invalid state
- Domain services for complex business rules

### 🔄 **CQRS Implementation**
- **Write Side**: Optimized repositories for parsing/ingesting data
- **Read Side**: Optimized query models for web interface
- Separate concerns for different access patterns
- Performance optimization for each side

### 📦 **Repository Pattern**
- Abstract interfaces for data access
- Write repositories for command operations
- Query repositories for read operations
- Technology-agnostic domain layer

### ✅ **Comprehensive Validation**
- Domain-specific validation services
- Business rule enforcement
- Data consistency checks
- Suggestion systems for missing data

### 🎯 **Event-Driven Architecture**
- Domain events for entity changes
- Publisher/subscriber pattern
- Loose coupling between components
- Audit trail capabilities

## File Structure

```
shared_domain/
├── __init__.py                      # Main exports
├── complete_entities.py             # Aggregate module with facade
├── entities.py                      # Core class entities (original)
├── spell_entities.py               # Spell domain model
├── monster_entities.py             # Monster domain model  
├── equipment_entities.py           # Equipment domain model
├── background_entities.py          # Background & feat domain model
├── document_entities.py            # Document domain model
└── query_models.py                 # CQRS read models (extended)
```

## Usage Examples

### Accessing the Domain Model

```python
from shared_domain import SRDDomainModel

# Get entity type for collection
ClassEntity = SRDDomainModel.get_entity_type("classi")
SpellEntity = SRDDomainModel.get_entity_type("incantesimi")
MonsterEntity = SRDDomainModel.get_entity_type("mostri")

# Get repository interface for write operations
ClassWriteRepo = SRDDomainModel.get_write_repository_type("classi")

# Get repository interface for read operations  
SpellQueryRepo = SRDDomainModel.get_query_repository_type("incantesimi")

# Get validation service
MonsterValidator = SRDDomainModel.get_validation_service("mostri")

# Get domain info
info = SRDDomainModel.get_domain_info()
print(f"Total entity types: {info['total_entity_types']}")
```

### Creating Domain Entities

```python
from shared_domain import *

# Create a spell
spell = Spell(
    id=SpellId("fireball"),
    nome="Fireball",
    livello=SpellLevel(3),
    scuola=SpellSchool.EVOCATION,
    classi=["Mago", "Stregone"],
    lancio=SpellCasting(
        tempo=CastingTime.ACTION,
        gittata=SpellRange.FEET_150,
        componenti=[
            SpellComponent("V"),
            SpellComponent("S"),
            SpellComponent("M", "a tiny ball of bat guano and sulfur")
        ],
        durata=SpellDuration.INSTANTANEOUS
    ),
    descrizione="A bright streak flashes...",
    contenuto_markdown="### Fireball\n*3rd-level evocation*..."
)

# Validate spell
validation_service = SpellValidationService()
errors = validation_service.validate_spell(spell)
```

### Using CQRS Query Models

```python
# Search queries
spell_query = SpellSearchQuery(
    text_query="fire",
    class_name="Mago",
    level=3,
    sort_by="name"
)

monster_query = MonsterSearchQuery(
    monster_type="Drago",
    min_cr=5.0,
    max_cr=15.0
)

# Query results optimized for reading
spell_summary = SpellSummary(
    id="fireball",
    nome="Fireball", 
    livello=3,
    scuola="Invocazione",
    classi=["Mago", "Stregone"],
    is_ritual=False,
    requires_concentration=False
)
```

## Benefits of This Implementation

### 🎯 **Complete ADR Compliance**
- All entity types from ADR data model implemented
- Maintains original JSON structure requirements
- Supports all specified fields and relationships

### 🏗️ **Clean Architecture**
- Hexagonal architecture with clear boundaries
- Domain logic independent of infrastructure
- Testable and maintainable codebase

### ⚡ **Performance Optimized**
- CQRS for read/write optimization
- Query models designed for UI needs
- Repository abstraction allows caching

### 🔒 **Data Integrity**
- Comprehensive validation rules
- Business logic enforcement
- Type safety with value objects

### 🔄 **Extensible Design**
- Event-driven architecture for extensions
- Repository pattern for different data stores
- Easy to add new entity types or features

## Integration with Services

### Parser Service (Write Side)
- Uses write repositories and validation services
- Publishes domain events on entity changes
- Optimized for bulk insert/update operations

### Editor Service (Read Side)  
- Uses query repositories and read models
- Optimized for search and retrieval
- Supports complex filtering and aggregation

### Shared Between Both
- Domain entities with business logic
- Value objects and validation rules
- Event definitions and interfaces

This complete domain model provides a solid foundation for both current and future D&D 5e SRD applications, following best practices in domain-driven design and clean architecture.