# Domain Simplification Plan: From 18 Entities to Single Document

**Created:** 2025-01-11
**Branch:** `refactor/simplify-domain-to-single-document`

## Goal

Replace the complex domain layer (18+ entity types, 16+ repositories) with a **unified Document model** containing only 4 fields: ID, Title, Filters, Content (HTML).

## Current State Analysis

### Domain Layer Complexity
- **18+ entity files:** animale.go, arma.go, armatura.go, background.go, cavalcatureVeicoli.go, classe.go, classe_builder.go, documento.go, equipaggiamento.go, incantesimo.go, mostro.go, oggettoMagico.go, regola.go, servizio.go, specie.go, strumento.go, talento.go, scelta.go
- **16+ repository interfaces:** One per entity type
- **Numerous value objects:** Costo, Peso, Velocita, Dadi, Caratteristica, Azione, Tratto, PuntiFerita, Attacco, Danno, Taglia, etc.
- **Type aliases:** AnimaleSlug, ArmaSlug, ArmaturaSlug, etc.
- **Estimated LOC:** ~2000-3000 lines

### Target State
- **1 entity:** Document
- **4 fields:** ID, Title, Filters, Content
- **1 repository:** DocumentRepository
- **3 value objects:** DocumentID, DocumentFilters, HTMLContent
- **Estimated LOC:** ~400-500 lines
- **Reduction:** ~85% less code

---

## Implementation Phases

### Phase 0: Preparation & Documentation ✓

**0.1 Create feature branch:**
```bash
git checkout -b refactor/simplify-domain-to-single-document
```

**0.2 Create plan documentation:**
- File: `docs/refactoring/domain-simplification-plan.md`
- Purpose: Reference during implementation, rollback guide

**0.3 Commit plan:**
```bash
git add docs/refactoring/domain-simplification-plan.md
git commit -m "docs(refactoring): add domain simplification plan"
```

---

### Phase 1: Create New Unified Domain Model ✓

**Goal:** Introduce new Document entity alongside existing entities.

**1.1 Create `internal/domain/document.go`:**
```go
package domain

// Document represents a unified D&D 5e SRD content entity
type Document struct {
    ID      DocumentID      `json:"id"      bson:"_id"`
    Title   string          `json:"title"   bson:"title"`
    Filters DocumentFilters `json:"filters" bson:"filters"`
    Content HTMLContent     `json:"content" bson:"content"`
}

// NewDocument creates a new Document
func NewDocument(id DocumentID, title string, filters DocumentFilters, content HTMLContent) *Document {
    return &Document{
        ID:      id,
        Title:   title,
        Filters: filters,
        Content: content,
    }
}

// EntityType implements ParsedEntity interface
func (d *Document) EntityType() string {
    if collection, ok := d.Filters["collection"].(string); ok {
        return collection
    }
    return "document"
}
```

**1.2 Create `internal/domain/document_id.go`:**
```go
package domain

// DocumentID is a unique identifier for documents (slug-based)
type DocumentID string

// NewDocumentID creates a DocumentID from a name using slug conversion
func NewDocumentID(name string) (DocumentID, error) {
    slug, err := NewSlug(name)
    if err != nil {
        return "", err
    }
    return DocumentID(slug), nil
}

// String returns the string representation
func (d DocumentID) String() string {
    return string(d)
}
```

**1.3 Create `internal/domain/document_filters.go`:**
```go
package domain

// DocumentFilters contains metadata for querying and filtering documents
// Common filter keys:
// - "collection": collection name (animali, armi, mostri, etc.)
// - "type": document type (Bestia, Arma Semplice, etc.)
// - "rarity": for magic items (Comune, Non Comune, Raro, etc.)
// - "level": for spells (0-9)
// - "cr": challenge rating for monsters
// - "category": category/subcategory
// - "source_file": original markdown file
// - "locale": content locale (always "ita")
type DocumentFilters map[string]any

// NewDocumentFilters creates a new DocumentFilters
func NewDocumentFilters() DocumentFilters {
    return make(DocumentFilters)
}

// Set adds or updates a filter
func (f DocumentFilters) Set(key string, value any) {
    f[key] = value
}

// Get retrieves a filter value
func (f DocumentFilters) Get(key string) (any, bool) {
    val, ok := f[key]
    return val, ok
}

// GetString retrieves a string filter value
func (f DocumentFilters) GetString(key string) string {
    if val, ok := f[key].(string); ok {
        return val
    }
    return ""
}

// GetInt retrieves an int filter value
func (f DocumentFilters) GetInt(key string) int {
    if val, ok := f[key].(int); ok {
        return val
    }
    return 0
}
```

**1.4 Create `internal/domain/html_content.go`:**
```go
package domain

// HTMLContent represents rendered HTML content
type HTMLContent string

// NewHTMLContent creates HTMLContent from a string
func NewHTMLContent(html string) HTMLContent {
    return HTMLContent(html)
}

// String returns the HTML string
func (h HTMLContent) String() string {
    return string(h)
}

// IsEmpty checks if content is empty
func (h HTMLContent) IsEmpty() bool {
    return len(h) == 0
}
```

**Files to create:**
- `internal/domain/document.go`
- `internal/domain/document_id.go`
- `internal/domain/document_filters.go`
- `internal/domain/html_content.go`

**Commit:**
```bash
git add internal/domain/document*.go internal/domain/html_content.go
git commit -m "feat(domain): add unified Document entity with value objects"
```

---

### Phase 2: Replace Repository Layer ✓

**Goal:** Create single DocumentRepository to replace all entity-specific repositories (including ContentRepository).

**2.1 Create `internal/domain/repositories/document_repository.go`:**
```go
package repositories

import (
    "context"
    "github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// DocumentRepository handles all document CRUD operations
type DocumentRepository interface {
    // Create inserts a new document
    Create(ctx context.Context, doc *domain.Document, collection string) error

    // Update updates an existing document
    Update(ctx context.Context, doc *domain.Document, collection string) error

    // Delete removes a document by ID
    Delete(ctx context.Context, id domain.DocumentID, collection string) error

    // FindByID retrieves a document by ID
    FindByID(ctx context.Context, id domain.DocumentID, collection string) (*domain.Document, error)

    // FindAll retrieves all documents in a collection
    FindAll(ctx context.Context, collection string, limit int) ([]*domain.Document, error)

    // FindByFilters retrieves documents matching filters
    FindByFilters(ctx context.Context, collection string, filters map[string]any, limit int) ([]*domain.Document, error)

    // Count returns the number of documents in a collection
    Count(ctx context.Context, collection string) (int64, error)

    // UpsertMany performs bulk upsert operations
    UpsertMany(ctx context.Context, collection string, documents []*domain.Document) (int, error)

    // GetCollectionStats returns statistics for a collection
    GetCollectionStats(ctx context.Context, collection string) (map[string]any, error)
}
```

**2.2 Create `internal/adapters/repositories/mongodb/document_mongo_repository.go`:**
```go
package mongodb

import (
    "context"
    "fmt"

    "github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
    "github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/repositories"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type documentMongoRepository struct {
    client *mongo.Client
    dbName string
}

// NewDocumentRepository creates a MongoDB implementation of DocumentRepository
func NewDocumentRepository(client *mongo.Client, dbName string) repositories.DocumentRepository {
    return &documentMongoRepository{
        client: client,
        dbName: dbName,
    }
}

func (r *documentMongoRepository) getCollection(collection string) *mongo.Collection {
    return r.client.Database(r.dbName).Collection(collection)
}

func (r *documentMongoRepository) Create(ctx context.Context, doc *domain.Document, collection string) error {
    coll := r.getCollection(collection)
    _, err := coll.InsertOne(ctx, doc)
    return err
}

func (r *documentMongoRepository) Update(ctx context.Context, doc *domain.Document, collection string) error {
    coll := r.getCollection(collection)
    filter := bson.M{"_id": doc.ID}
    _, err := coll.ReplaceOne(ctx, filter, doc)
    return err
}

func (r *documentMongoRepository) Delete(ctx context.Context, id domain.DocumentID, collection string) error {
    coll := r.getCollection(collection)
    filter := bson.M{"_id": id}
    _, err := coll.DeleteOne(ctx, filter)
    return err
}

func (r *documentMongoRepository) FindByID(ctx context.Context, id domain.DocumentID, collection string) (*domain.Document, error) {
    coll := r.getCollection(collection)
    filter := bson.M{"_id": id}

    var doc domain.Document
    err := coll.FindOne(ctx, filter).Decode(&doc)
    if err != nil {
        return nil, err
    }
    return &doc, nil
}

func (r *documentMongoRepository) FindAll(ctx context.Context, collection string, limit int) ([]*domain.Document, error) {
    coll := r.getCollection(collection)
    opts := options.Find()
    if limit > 0 {
        opts.SetLimit(int64(limit))
    }

    cursor, err := coll.Find(ctx, bson.M{}, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var docs []*domain.Document
    if err := cursor.All(ctx, &docs); err != nil {
        return nil, err
    }
    return docs, nil
}

func (r *documentMongoRepository) FindByFilters(ctx context.Context, collection string, filters map[string]any, limit int) ([]*domain.Document, error) {
    coll := r.getCollection(collection)

    // Build filter query
    filter := bson.M{}
    for key, value := range filters {
        filter[fmt.Sprintf("filters.%s", key)] = value
    }

    opts := options.Find()
    if limit > 0 {
        opts.SetLimit(int64(limit))
    }

    cursor, err := coll.Find(ctx, filter, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var docs []*domain.Document
    if err := cursor.All(ctx, &docs); err != nil {
        return nil, err
    }
    return docs, nil
}

func (r *documentMongoRepository) Count(ctx context.Context, collection string) (int64, error) {
    coll := r.getCollection(collection)
    return coll.CountDocuments(ctx, bson.M{})
}

func (r *documentMongoRepository) UpsertMany(ctx context.Context, collection string, documents []*domain.Document) (int, error) {
    if len(documents) == 0 {
        return 0, nil
    }

    coll := r.getCollection(collection)

    var models []mongo.WriteModel
    for _, doc := range documents {
        model := mongo.NewReplaceOneModel().
            SetFilter(bson.M{"_id": doc.ID}).
            SetReplacement(doc).
            SetUpsert(true)
        models = append(models, model)
    }

    result, err := coll.BulkWrite(ctx, models)
    if err != nil {
        return 0, err
    }

    return int(result.UpsertedCount + result.ModifiedCount), nil
}

func (r *documentMongoRepository) GetCollectionStats(ctx context.Context, collection string) (map[string]any, error) {
    count, err := r.Count(ctx, collection)
    if err != nil {
        return nil, err
    }

    return map[string]any{
        "collection": collection,
        "count":      count,
    }, nil
}
```

**2.3 Update `internal/adapters/repositories/factory.go`:**
Add method to create DocumentRepository while keeping existing factories for now (incremental migration).

**Files to create/modify:**
- Create: `internal/domain/repositories/document_repository.go`
- Create: `internal/adapters/repositories/mongodb/document_mongo_repository.go`
- Modify: `internal/adapters/repositories/factory.go`

**Commit:**
```bash
git add internal/domain/repositories/document_repository.go internal/adapters/repositories/mongodb/document_mongo_repository.go internal/adapters/repositories/factory.go
git commit -m "feat(repositories): add unified DocumentRepository implementation"
```

---

### Phase 3: Update Parser Layer ✓

**Goal:** Modify parsers to output HTML and return Document instead of entity-specific types.

**3.1 Add HTML rendering dependency:**
- Use existing markdown parser + HTML renderer (e.g., goldmark, blackfriday)

**3.2 Update `ParsingStrategy` interface:**
```go
type ParsingStrategy interface {
    Parse(content string, metadata map[string]string) (*domain.Document, error)
    ContentType() string
}
```

**3.3 Update base parser to include HTML rendering:**
- Extract markdown content as before
- Convert markdown → HTML using renderer
- Populate DocumentFilters with metadata
- Return Document

**3.4 Update all strategy implementations:**
- Modify each *_strategy.go file to return Document
- Populate filters with collection-specific metadata
- Render content as HTML

**Files to modify:**
- `internal/application/parsers/strategy.go`
- `internal/application/parsers/base_parser.go`
- All `internal/application/parsers/*_strategy.go` files

**Commit:**
```bash
git add internal/application/parsers/
git commit -m "feat(parsers): update strategies to output HTML and Document entities"
```

---

### Phase 4: Update Application Services ✓

**Goal:** Update services and handlers to work with Document.

**4.1 Update ContentService:**
- Accept collection name + filters
- Return Document instances
- Remove entity-specific type handling

**4.2 Update web handlers:**
- Render HTML content directly (no markdown processing)
- Use filters for search/filtering
- Update templates to work with Document structure

**Files to modify:**
- `internal/application/services/content_service.go`
- `internal/adapters/web/handlers/*.go`
- Templates in `pkg/templates/`

**Commit:**
```bash
git add internal/application/services/ internal/adapters/web/handlers/ pkg/templates/
git commit -m "refactor(services): migrate to unified Document model"
```

**Note:** After initial implementation, decided to remove ContentRepository entirely and consolidate to DocumentRepository only. This simplifies the architecture by:
- Eliminating duplicate repository code
- Using single repository with both type-safe (Document) and flexible (map) interfaces
- Adding map-based methods to DocumentRepository for viewer compatibility

---

### Phase 5: Cleanup Old Domain Files ✓

**Goal:** Remove obsolete entity files and value objects.

**5.1 Remove 18 entity files:**
```bash
rm internal/domain/animale.go
rm internal/domain/arma.go
rm internal/domain/armatura.go
rm internal/domain/background.go
rm internal/domain/cavalcatureVeicoli.go
rm internal/domain/classe.go
rm internal/domain/classe_builder.go
rm internal/domain/documento.go
rm internal/domain/equipaggiamento.go
rm internal/domain/incantesimo.go
rm internal/domain/mostro.go
rm internal/domain/oggettoMagico.go
rm internal/domain/regola.go
rm internal/domain/servizio.go
rm internal/domain/specie.go
rm internal/domain/strumento.go
rm internal/domain/talento.go
rm internal/domain/scelta.go
```

**5.2 Simplify `internal/domain/common.go`:**
- Keep only Slug (used by DocumentID)
- Remove all value objects: Costo, Peso, Velocita, Dadi, Caratteristica, etc.

**5.3 Remove entity-specific repositories:**
```bash
rm internal/domain/repositories/*_repository.go  # except document_repository.go and base_repository.go
rm internal/adapters/repositories/mongodb/*_mongo_repository.go  # except document_mongo_repository.go
```

**5.4 Simplify factory.go:**
- Keep only CreateDocumentRepository()
- Remove all entity-specific factory methods

**Files to delete/modify:**
- Delete: 18 entity files
- Delete: 16 entity-specific repository interfaces
- Delete: 16 MongoDB repository implementations
- Modify: `internal/domain/common.go`
- Modify: `internal/adapters/repositories/factory.go`

**Commit:**
```bash
git add -A
git commit -m "refactor(domain): remove obsolete entity files and repositories"
```

---

### Phase 6: Update Tests & Verification

**Goal:** Ensure all tests pass and system works end-to-end.

**6.1 Update unit tests:**
- Replace entity-specific tests with Document tests
- Update parser tests to verify HTML output
- Update repository tests

**6.2 Update integration tests:**
- Modify to work with unified Document model
- Test parser → database → viewer flow

**6.3 Run full test suite:**
```bash
make test
make test-integration
```

**6.4 Verify with local deployment:**
```bash
make down
make build
make up
# Parse data
docker compose exec viewer /app/cli-parser parse --type all --locale ita
# Check viewer at http://localhost:8000
```

**6.5 Database verification:**
- Check document structure in MongoDB
- Verify filters are populated correctly
- Confirm HTML rendering works

**Commit:**
```bash
git add .
git commit -m "test: update tests for Document model"
```

---

## Migration Strategies

### Option A: Low-Risk Incremental (Recommended)

1. **Phase 1-2:** Add Document + DocumentRepository alongside existing code
2. **Phase 3:** Update parsers to dual-write (old entities + new Documents)
3. **Phase 4:** Update services to read from Documents, keeping old reads as fallback
4. **Verification:** Test viewer with new format
5. **Phase 5:** Remove old entities once verified
6. **Phase 6:** Final cleanup and testing

**Pros:** Rollback-friendly, can test incrementally
**Cons:** Temporary code duplication

### Option B: Clean-Cut

1. Execute all phases sequentially in one go
2. Re-run parser to populate database
3. Test end-to-end

**Pros:** Faster, cleaner git history
**Cons:** Higher risk, harder to debug issues

---

## Expected Outcomes

### Code Metrics
- **Before:** ~2000-3000 LOC in domain layer
- **After:** ~400-500 LOC in domain layer
- **Reduction:** ~85% less code
- **Files removed:** 35+ files (entities + repositories)
- **Files added:** 5 files (Document + value objects + repo)

### Architecture Benefits
- Single source of truth for all D&D content
- Simplified repository pattern (1 instead of 16)
- Easier to add new content types (just metadata)
- Frontend receives pre-rendered HTML (performance)
- Reduced coupling between domain and parsers

### Trade-offs
- Loss of compile-time type safety for entity fields
- All queries use map[string]any filters
- HTML rendering logic moves to parser layer
- Must maintain filter schema documentation

---

## Rollback Plan

If issues arise, rollback is straightforward on the incremental approach:

1. **Git revert:** `git reset --hard origin/main`
2. **Database:** Collections unchanged (schema-less MongoDB)
3. **Re-parse if needed:** Run parser with old code

For clean-cut approach, keep a database backup:
```bash
make seed-dump  # Before starting refactor
make seed-restore FILE=backup.archive.gz  # If rollback needed
```

---

## Success Criteria

- [ ] All 18 entity files removed
- [ ] Single Document entity with 4 fields
- [ ] Single DocumentRepository replacing 16 repos
- [ ] Parsers output HTML content
- [ ] Viewer renders documents correctly
- [ ] All tests passing (unit + integration)
- [ ] Database contains correctly structured documents
- [ ] No regression in viewer functionality

---

## References

- **CLAUDE.md:** Project architecture documentation
- **Current domain files:** `/Users/emiliopalmerini/repos/due-draghi-5e-srd/internal/domain/`
- **Current repositories:** `/Users/emiliopalmerini/repos/due-draghi-5e-srd/internal/adapters/repositories/`
- **Parser strategies:** `/Users/emiliopalmerini/repos/due-draghi-5e-srd/internal/application/parsers/`

---

**End of Plan**
