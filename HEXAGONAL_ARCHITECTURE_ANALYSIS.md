# Hexagonal Architecture Implementation Analysis

## Current State Assessment

### ✅ Parser Service (Already Hexagonal-Ready)
```
srd_parser/
├── domain/                 # ✅ Domain layer implemented
│   ├── entities.py        # Aggregates & entities
│   ├── value_objects.py   # Immutable value objects
│   └── services.py        # Domain services
├── parsers/               # Application layer
│   └── classes_improved.py
└── adapters/              # Infrastructure (partial)
    └── persistence/
```

### ⚠️ Editor Service (Traditional Layered)
```
editor/
├── routers/               # Presentation layer
│   └── pages.py          # Tightly coupled to infrastructure
├── core/                  # Mixed concerns
│   ├── database.py       # Direct MongoDB coupling
│   └── config.py         # Configuration
└── services/              # Service layer (thin)
    └── content_service.py
```

## Proposed Shared Hexagonal Architecture

### Shared Domain (New)
```
shared_domain/
├── entities.py            # ✅ Core domain entities
├── use_cases.py           # ✅ Application use cases
├── ports.py              # Repository & service interfaces
└── events.py             # Domain events
```

### Parser Service (Write-Side)
```
srd_parser/
├── application/           # Application layer
│   ├── handlers/         # Command handlers
│   ├── services/         # Application services
│   └── use_cases/        # Parse operations
├── adapters/             # Infrastructure adapters
│   ├── persistence/      # MongoDB adapter
│   ├── parsers/          # Markdown parsing adapter
│   └── events/           # Event publishing adapter
└── main.py               # Composition root
```

### Editor Service (Read-Side)  
```
editor/
├── application/           # Application layer
│   ├── handlers/         # Query handlers
│   ├── services/         # Application services
│   └── use_cases/        # View operations
├── adapters/             # Infrastructure adapters
│   ├── persistence/      # MongoDB read adapter
│   ├── web/              # FastAPI/HTMX adapter
│   └── templates/        # Template adapter
└── main.py               # Composition root
```

## Domain Sharing Strategy

### ✅ **What Should Be Shared**
- **Core Entities**: DndClass, Subclass, ClassFeature, Spell
- **Value Objects**: Level, Ability, ClassId, EntityId
- **Domain Services**: ClassValidationService, SpellCalculationService
- **Repository Interfaces**: ClassRepository, SpellRepository
- **Domain Events**: ClassParsed, ClassViewed, DataUpdated

### ❌ **What Should NOT Be Shared**
- **Use Cases**: Parser has write operations, Editor has read operations
- **Infrastructure**: Different databases, different frameworks
- **Application Services**: Different business workflows
- **Adapters**: Different external system integrations

## Implementation Benefits

### 🎯 **Separation of Concerns**
```python
# Parser: Write-optimized use cases
class ParseClassUseCase:
    async def execute(self, command: ParseClassCommand) -> UseCaseResult:
        # Complex validation, parsing, saving logic
        
# Editor: Read-optimized use cases  
class GetClassUseCase:
    async def execute(self, query: GetClassQuery) -> UseCaseResult:
        # Simple retrieval, formatting, display logic
```

### 🔄 **CQRS Pattern Natural Fit**
- **Parser**: Command side (Create, Update, Delete)
- **Editor**: Query side (Read, Search, Display)
- **Shared Events**: Keep read/write sides synchronized

### 🧪 **Testability Improvements**
```python
# Easy to mock repositories for testing
async def test_parse_class_use_case():
    mock_repo = MockClassRepository()
    mock_publisher = MockEventPublisher()
    
    use_case = ParseClassUseCase(mock_repo, mock_publisher)
    result = await use_case.execute(ParseClassCommand(...))
    
    assert result.success
    assert mock_repo.save_called
    assert mock_publisher.publish_called
```

## Migration Strategy

### Phase 1: Extract Shared Domain ✅ COMPLETED
- [x] Create `shared_domain/` with entities and value objects
- [x] Define repository interfaces (ports)
- [x] Implement domain services and validation
- [x] Create application use cases for both sides

### Phase 2: Refactor Parser (Week 1)
```python
# Before: Tightly coupled
def parse_classes(md_lines: List[str]) -> List[Dict]:
    # Direct database calls, mixed concerns

# After: Clean hexagonal
class ParseClassUseCase:
    def __init__(self, repo: ClassRepository, publisher: EventPublisher):
        self.repo = repo
        self.publisher = publisher
        
    async def execute(self, command: ParseClassCommand) -> UseCaseResult:
        # Pure business logic, testable
```

### Phase 3: Refactor Editor (Week 2)
```python  
# Before: Router calls database directly
@router.get("/classes/{class_id}")
async def get_class(class_id: str):
    db = await get_database()
    result = await db.classes.find_one({"_id": class_id})

# After: Router uses use cases
@router.get("/classes/{class_id}")  
async def get_class(class_id: str, use_case: GetClassUseCase = Depends()):
    result = await use_case.execute(GetClassQuery(class_id))
    return format_response(result)
```

### Phase 4: Add Event-Driven Communication (Week 3)
```python
# Parser publishes events when data changes
await self.event_publisher.publish(ClassParsed(
    class_name="Barbaro",
    version="2.1.0"
))

# Editor subscribes to events for cache invalidation
class ClassCacheInvalidator:
    async def handle_class_parsed(self, event: ClassParsed):
        await self.cache.invalidate(f"class:{event.class_name}")
```

## Architecture Validation

### ✅ **Hexagonal Principles Met**
- **Domain Independence**: Business logic isolated from framework concerns
- **Dependency Inversion**: High-level modules don't depend on low-level modules
- **Port & Adapter**: Clear interfaces for external systems
- **Testability**: All components mockable and testable

### ✅ **DDD Principles Met**  
- **Ubiquitous Language**: Consistent terminology across both services
- **Bounded Context**: Clear boundaries between parsing and viewing contexts
- **Aggregate Consistency**: DndClass aggregate maintains invariants
- **Domain Events**: Communication between bounded contexts

### ✅ **Clean Architecture Benefits**
- **Independence of Frameworks**: Can switch from FastAPI to Django easily
- **Independence of Database**: Can switch from MongoDB to PostgreSQL
- **Independence of UI**: Can add GraphQL API alongside REST
- **Testable**: Business rules testable without external dependencies

## Code Quality Improvements

### Before Refactoring
```python
# Tightly coupled, hard to test
async def get_class_data(class_id: str):
    db = await get_database()  # Infrastructure coupling
    data = await db.classes.find_one({"_id": class_id})  # Database coupling
    if not data:
        raise HTTPException(404)  # Framework coupling
    return transform_data(data)  # Business logic mixed
```

### After Refactoring  
```python
# Clean, testable, framework-independent
class GetClassUseCase:
    def __init__(self, repository: ClassRepository):
        self.repository = repository  # Dependency injection
        
    async def execute(self, query: GetClassQuery) -> UseCaseResult:
        dnd_class = await self.repository.find_by_id(ClassId(query.class_id))
        if not dnd_class:
            return UseCaseResult(success=False, message="Class not found")
        
        return UseCaseResult(
            success=True,
            data=self._format_for_display(dnd_class)
        )  # Pure business logic
```

## Performance Considerations

### Read/Write Optimization
- **Parser**: Optimized for complex writes with validation
- **Editor**: Optimized for fast reads with caching
- **Shared Events**: Eventual consistency between sides

### Caching Strategy
```python
# Editor can implement aggressive caching
class CachedClassRepository:
    async def find_by_id(self, class_id: ClassId) -> Optional[DndClass]:
        cached = await self.cache.get(f"class:{class_id.value}")
        if cached:
            return cached
        
        result = await self.database_repo.find_by_id(class_id)
        await self.cache.set(f"class:{class_id.value}", result, ttl=3600)
        return result
```

## Conclusion

✅ **Feasibility**: **HIGHLY FEASIBLE** - Domain sharing works perfectly for this use case

✅ **Benefits**: 
- Consistent domain model across services
- Clear separation of concerns (CQRS)
- Improved testability and maintainability
- Better scalability (read/write optimization)

✅ **Implementation**: **LOW RISK** - Can be done incrementally without breaking existing functionality

The shared domain approach is ideal for this D&D SRD project because both services operate on the same core entities but with different responsibilities and performance characteristics.