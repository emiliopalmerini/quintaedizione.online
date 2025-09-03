# Hexagonal Architecture Implementation

This document describes the complete hexagonal architecture implementation for the D&D 5e SRD system, featuring both Parser and Editor services following Domain-Driven Design principles.

## Overview

The system has been refactored from a traditional layered architecture to a hexagonal (ports and adapters) architecture with clear separation of concerns and CQRS pattern implementation.

### Architecture Diagram

```
┌─────────────────┐       ┌─────────────────┐
│   Parser Web    │       │   Editor Web    │
│   (FastAPI)     │       │   (FastAPI)     │
└─────────────────┘       └─────────────────┘
         │                         │
         ▼                         ▼
┌─────────────────┐       ┌─────────────────┐
│ Parser Commands │       │ Editor Queries  │
│   (Write-side)  │       │   (Read-side)   │
└─────────────────┘       └─────────────────┘
         │                         │
         ▼                         ▼
┌─────────────────┐       ┌─────────────────┐
│  Application    │       │  Application    │
│    Layer        │       │    Layer        │
└─────────────────┘       └─────────────────┘
         │                         │
         ▼                         ▼
┌─────────────────────────────────────────────┐
│           Shared Domain Layer               │
│    (Entities, Value Objects, Services)     │
└─────────────────────────────────────────────┘
         │                         │
         ▼                         ▼
┌─────────────────┐       ┌─────────────────┐
│ Write Repository│       │ Query Repository│
│   (MongoDB)     │       │   (MongoDB)     │
└─────────────────┘       └─────────────────┘
         │                         │
         ▼                         ▼
┌─────────────────────────────────────────────┐
│              MongoDB Database               │
└─────────────────────────────────────────────┘
```

## Key Components

### 1. Shared Domain Layer (`/shared_domain/`)

**Purpose**: Contains the core business logic shared between both services.

**Components**:
- **Entities** (`entities.py`): Core domain entities (DndClass, Subclass, etc.)
- **Value Objects** (`value_objects.py`): Immutable value types (Level, HitDie, etc.)
- **Use Cases** (`use_cases.py`): Application use cases for both read and write operations
- **Query Models** (`query_models.py`): CQRS read models and query results

**Key Features**:
- Rich domain model with business logic encapsulation
- Validation rules enforced at entity level
- Repository abstractions for both command and query sides
- Event-driven architecture support

### 2. Parser Service (Write-Side)

**Purpose**: Handles parsing SRD content and writing to the database.

**Structure**:
```
srd_parser/
├── adapters/           # Infrastructure adapters
│   ├── persistence/    # MongoDB write repository
│   └── events/         # Event publishing
├── application/        # Application layer
│   └── command_handlers.py
├── infrastructure/     # DI container
├── domain/            # Parser-specific domain services
└── web_hexagonal.py   # Updated web interface
```

**Key Features**:
- **MongoDBClassRepository**: Write-optimized with upsert operations
- **InMemoryEventPublisher**: Domain event publishing with subscriber pattern
- **Command Handlers**: Parse multiple classes, validate data
- **Dependency Injection**: Clean configuration and lifecycle management

### 3. Editor Service (Read-Side)

**Purpose**: Provides optimized read operations for the web interface.

**Structure**:
```
editor/
├── adapters/           # Infrastructure adapters
│   └── persistence/    # MongoDB query repository
├── application/        # Application layer
│   └── query_handlers.py
├── infrastructure/     # DI container
├── routers/           # Web layer
│   └── hexagonal_pages.py  # Hexagonal demo routes
└── templates/         # UI templates
```

**Key Features**:
- **MongoDBClassQueryRepository**: Read-optimized with projections and indexes
- **Query Handlers**: Search classes, get details, filter by criteria
- **CQRS Query Models**: Optimized for read operations (ClassSummary, ClassDetail)
- **Advanced Search**: Text search, filtering, pagination

## Implementation Details

### Domain-Driven Design Elements

1. **Entities**: 
   - `DndClass`: Rich domain entity with business logic
   - `Subclass`: Class specialization entity
   - Identity through strongly-typed IDs (`ClassId`, `EntityId`)

2. **Value Objects**:
   - `Level`: 1-20 validation
   - `HitDie`: d6-d12 validation  
   - `Ability`: Enum for six core abilities

3. **Domain Services**:
   - `ClassValidationService`: Business rule validation
   - `ClassParsingService`: Domain logic for parsing

### CQRS Implementation

**Command Side (Parser)**:
- Optimized for write operations
- Domain events for side effects
- Business logic validation
- Upsert-based persistence

**Query Side (Editor)**:
- Optimized for read operations
- Denormalized views
- Advanced search capabilities
- Caching-ready architecture

### Dependency Injection

Both services use container-based DI:

```python
# Parser Container
container = ParserContainer()
handler = container.get_parse_multiple_classes_handler()

# Editor Container  
container = EditorContainer()
handler = container.get_search_classes_handler()
```

### Event-Driven Architecture

Domain events are published when entities change:

```python
# Domain Event
@dataclass
class ClassParsedEvent(DomainEvent):
    class_name: str
    aggregate_id: str = field(init=False)
    
# Event Publishing
await event_publisher.publish(ClassParsedEvent(class_name="Fighter"))
```

## Usage Examples

### Parser Service (Write Operations)

```python
# Parse multiple classes
command = ParseMultipleClassesCommand(
    markdown_lines=file_lines,
    source="SRD",
    dry_run=False
)
result = await handler.handle(command)

# Validate data
validation_command = ValidateClassDataCommand(
    markdown_lines=file_lines
)
validation_result = await validator.handle(validation_command)
```

### Editor Service (Read Operations)

```python
# Search classes
query = SearchClassesQuery(
    text_query="fighter",
    primary_ability="Forza", 
    is_spellcaster=False,
    limit=20
)
result = await handler.handle(query)

# Get class details
detail_query = GetClassDetailQuery(class_id="guerriero")
detail = await handler.handle(detail_query)
```

## Benefits Achieved

### 1. **Separation of Concerns**
- Web layer only handles HTTP concerns
- Application layer orchestrates use cases
- Domain layer contains business logic
- Infrastructure layer handles external systems

### 2. **Testability**
- Each layer can be unit tested in isolation
- Repository interfaces allow test doubles
- Dependency injection enables test configuration

### 3. **Maintainability**
- Clear boundaries between components
- Single responsibility at each layer
- Easy to locate and modify functionality

### 4. **Scalability**
- CQRS allows independent scaling of read/write sides
- Repository pattern abstracts data access
- Event-driven architecture enables loose coupling

### 5. **Technology Independence**
- Business logic independent of frameworks
- Database abstraction through repositories
- Web framework abstraction through application layer

## Integration Points

### Web Layer Integration

Both services expose their functionality through FastAPI:

```python
# Parser Web Interface
@app.post("/run-hexagonal")
async def run_hexagonal(command: ParseCommand):
    container = get_container()
    handler = container.get_parse_multiple_classes_handler()
    return await handler.handle(command)

# Editor Web Interface  
@app.get("/hex/classes")
async def hex_classes_list(query: SearchQuery):
    container = get_container()
    handler = container.get_search_classes_handler()
    return await handler.handle(query)
```

### Database Integration

Both services share the same MongoDB database but use different repository implementations:

- **Parser**: Write-optimized repository with upserts
- **Editor**: Read-optimized repository with projections and aggregations

### Shared Domain Integration

Both services import from the shared domain:

```python
from shared_domain.entities import DndClass, ClassRepository
from shared_domain.use_cases import ParseClassUseCase  
from shared_domain.query_models import ClassSummary
```

## Demo URLs

The hexagonal architecture can be explored at:

- **Editor Demo**: `http://localhost:8000/hex/`
- **Classes List**: `http://localhost:8000/hex/classes`
- **Class Detail**: `http://localhost:8000/hex/classes/{class_id}`
- **Search**: `http://localhost:8000/hex/classes?q=fighter&ability=Forza`
- **Parser Demo**: `http://localhost:8100/` (with hexagonal endpoints)

## Migration from Legacy

The hexagonal architecture coexists with the existing layered architecture:

- Legacy endpoints remain at root paths (`/`, `/classi`)
- Hexagonal demo at `/hex/*` paths
- Both architectures share the same database
- Gradual migration path available

## Next Steps

1. **Complete Migration**: Move all legacy endpoints to hexagonal architecture
2. **Add Caching**: Implement query result caching on read-side  
3. **Event Store**: Add persistent event store for audit and replay
4. **API Layer**: Add REST API endpoints using the same application layer
5. **Testing**: Add comprehensive integration and unit tests
6. **Monitoring**: Add application metrics and health checks

This implementation demonstrates a production-ready hexagonal architecture with clear separation of concerns, testability, and maintainability while preserving all existing functionality.