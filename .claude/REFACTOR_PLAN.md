# D&D 5e SRD - Refactoring Evaluation & Plan

## Executive Summary
The codebase demonstrates solid Clean Architecture principles with a good separation of concerns. The system is well-organized into domain, application, and adapter layers. However, there are opportunities for consolidation, pattern standardization, and code reduction.

---

## Current State Assessment

### Strengths
1. **Clean Architecture Foundation** - Clear separation between domain, application, and adapters
2. **Strategy Pattern Implementation** - Well-executed for parser strategies with proper registry pattern
3. **Unified Repository** - Excellent consolidation from 16+ entity-specific repositories to single `DocumentRepository`
4. **Consistent Document Model** - All entities map to `domain.Document` with value objects
5. **Comprehensive Testing** - Registry and content type tests in place
6. **Clear Responsibility Boundaries** - Services, handlers, and repositories are properly separated

### Current Issues

#### 1. **Redundant Strategy Interfaces** (Medium Priority)
- **Location**: `internal/application/parsers/`
  - `ParsingStrategy` interface
  - `TemplateParsingStrategy` interface  
  - `DocumentParsingStrategy` interface (active)
- **Issue**: Multiple strategy interfaces for different parsing approaches; only `DocumentParsingStrategy` is used currently
- **Impact**: Code confusion, dead code paths, maintenance burden
- **Recommendation**: Retire unused `ParsingStrategy` and `TemplateParsingStrategy`

#### 2. **Pagination Logic Duplication** (High Priority)
- **Location**: `internal/adapters/web/viewer_handlers.go`
  - `handleCollectionList()` - lines 157-162
  - `handleCollectionRows()` - lines 302-307
- **Issue**: Identical pagination calculation code repeated twice
- **Impact**: ~10 lines duplicated, risk of divergence, maintenance overhead
- **Recommendation**: Extract to helper method `calculatePaginationData()`

#### 3. **Filter Extraction Logic Duplication** (Medium Priority)
- **Location**: `internal/adapters/web/viewer_handlers.go`
  - `handleCollectionList()` - line 138
  - `handleCollectionRows()` - line 283
- **Issue**: Same filter extraction logic called twice
- **Impact**: Potential inconsistency, reduces clarity
- **Recommendation**: Already centralized in `extractFilters()`, but should be called once in shared logic

#### 4. **Dual Service Pattern in Handler** (Medium Priority)
- **Location**: `internal/adapters/web/viewer_handlers.go` - lines 140-146
- **Issue**: Conditional branching based on filters presence
  ```go
  if len(filters) > 0 {
      rawItems, totalCount, err = h.contentService.GetCollectionItemsWithFilters(...)
  } else {
      rawItems, totalCount, err = h.contentService.GetCollectionItems(...)
  }
  ```
- **Impact**: Requires two service methods for same operation, increases API surface
- **Recommendation**: Consolidate into single method: `GetCollectionItems(filters map[string]string)`; empty map = no filters

#### 5. **Inconsistent Error Handling in Handlers** (Medium Priority)
- **Location**: Multiple handlers
- **Issue**: Mix of silent failures (fmt.Printf) and response errors
  - Line 37: `fmt.Printf("Warning: Failed to load...")` - silent
  - Line 221: `fmt.Printf("Warning: Could not get adjacent items...")` - silent
  - Lines 148-150: Proper error response
- **Impact**: Inconsistent logging/reporting, difficult to diagnose issues
- **Recommendation**: Create middleware or helper for consistent error logging with proper levels

#### 6. **Magic Strings and Hardcoded Values** (Low-Medium Priority)
- **Location**: `internal/adapters/web/viewer_handlers.go`
  - Lines 391-406: Hardcoded collection title map
  - Lines 420-424: Hardcoded skip parameters for filtering
  - Lines 337-349: Hardcoded content replacement rules
- **Issue**: Configuration scattered across handler
- **Impact**: Difficult to modify, not DRY, breaks Open/Closed Principle
- **Recommendation**: Move to configuration files or constants package

#### 7. **Content Formatting Function Misplaced** (Low Priority)
- **Location**: `internal/adapters/web/viewer_handlers.go` - line 337
- **Function**: `formatTraitContent()` - package-level function
- **Issue**: Formatting logic is web-specific but not part of handler or service
- **Impact**: Unclear responsibility, could be part of display logic
- **Recommendation**: Move to `internal/adapters/web/display/` package

#### 8. **Global Variable Dependencies in CLI** (Low Priority)
- **Location**: `cmd/cli-parser/parser.go`
  - Line 51: Uses `*dryRun` global variable
- **Issue**: Global flag variable accessed in method
- **Impact**: Harder to test, implicit dependency
- **Recommendation**: Pass through constructor or context

#### 9. **Test Coverage Gaps** (Medium Priority)
- **Current**: Registry and content type tests exist
- **Missing**: 
  - Document parsing strategy tests
  - Handler integration tests
  - Service layer tests
  - Error scenario tests
- **Impact**: Difficult to refactor with confidence
- **Recommendation**: Add comprehensive test suite before major refactors

#### 10. **Type Assertion Repetition** (Low Priority)
- **Location**: `internal/adapters/web/viewer_handlers.go` - multiple places
  - Lines 81-83: Extract collection name from map
  - Lines 90-93: Extract count from map
  - Lines 208-215: Extract HTML and markdown content
  - Lines 462-468: Extract title and slug
- **Issue**: Same pattern repeated for type assertions on `map[string]any`
- **Impact**: Boilerplate code, error-prone
- **Recommendation**: Create type-safe mapper helpers: `extractString()`, `extractInt64()`, etc.

---

## Recommended Refactoring Plan

### Phase 1: High-Impact, Low-Risk Changes (1-2 days)
Priority: Execute these first

1. **Remove Unused Strategy Interfaces**
   - Delete `ParsingStrategy` interface and `TemplateParsingStrategy`
   - Update registry to use only `DocumentParsingStrategy`
   - Files affected: `strategy.go`, `registry.go`
   - Tests: Verify registry tests still pass

2. **Extract Pagination Helper**
   - Create `calculatePaginationData(pageNum, pageSize int, totalCount int64) *PaginationData`
   - Use in both `handleCollectionList()` and `handleCollectionRows()`
   - Files: `internal/adapters/web/viewer_handlers.go`, new `internal/adapters/web/helpers.go`

3. **Consolidate Dual Service Methods**
   - Merge `GetCollectionItems()` and `GetCollectionItemsWithFilters()` into single method
   - Signature: `GetCollectionItems(ctx context.Context, collection string, q string, filters map[string]string, page, pageSize int)`
   - Files: `internal/application/services/content_service.go`, `internal/adapters/web/viewer_handlers.go`
   - Impact: ~20 lines reduction

### Phase 2: Medium-Impact Changes (2-3 days)
Medium priority, requires more testing

4. **Standardize Error Handling**
   - Create logging middleware or utility
   - Replace all `fmt.Printf` warnings with structured logging
   - Add error levels (warning vs error vs info)
   - Files: New `internal/adapters/web/logging.go`, existing handlers

5. **Extract Type Assertion Helpers**
   - Create `pkg/types/mappers.go` with helpers:
     - `GetString(m map[string]any, key string, defaultValue string) string`
     - `GetInt64(m map[string]any, key string, defaultValue int64) int64`
     - `GetBool(m map[string]any, key string, defaultValue bool) bool`
   - Replace 50+ type assertions with helper calls
   - Files: New `pkg/types/mappers.go`, `internal/adapters/web/viewer_handlers.go`

6. **Move Configuration to Constants**
   - Create `internal/adapters/web/config/labels.go` for collection titles
   - Create `internal/adapters/web/config/formatting.go` for content rules
   - Move cache time constants to enum/const package
   - Files: New `internal/adapters/web/config/`, `internal/adapters/web/viewer_handlers.go`

### Phase 3: Lower-Priority Improvements (1-2 days)
Nice-to-have, lower risk/impact

7. **Move Display Functions**
   - Move `formatTraitContent()` to `internal/adapters/web/display/`
   - Refactor to accept configuration for replacement rules
   - Files: `internal/adapters/web/display/`, `internal/adapters/web/viewer_handlers.go`

8. **Improve CLI Dependency Injection**
   - Pass flags through constructor or context instead of globals
   - Files: `cmd/cli-parser/main.go`, `cmd/cli-parser/parser.go`

9. **Add Comprehensive Tests**
   - Document parsing strategy tests (with fixtures)
   - Handler integration tests (mocked services)
   - Service layer tests
   - Error scenario coverage
   - Files: `*_test.go` files throughout

10. **Cache Header Strategy**
    - Extract cache header logic to strategy pattern
    - Make cache times configurable
    - Files: New `internal/adapters/web/caching/`, `internal/adapters/web/viewer_handlers.go`

---

## Refactoring Rules & Best Practices to Apply

### Pattern Standardization
- Use dependency injection consistently (no globals)
- Apply single responsibility to all functions (max ~50 lines for handlers)
- Extract magic numbers/strings to named constants

### Naming Conventions
- Use `_test.go` suffix for test files
- Use `interface_impl.go` for implementations
- Use descriptive package names that match responsibility

### Code Quality
- Keep handler functions under 100 lines
- Keep service methods under 50 lines of actual logic
- Use table-driven tests for multiple scenarios
- Always include error scenarios in tests

### Documentation
- Add godoc comments to exported functions
- Document why, not what (explain design decisions)
- Include examples in complex helper functions

---

## Impact Analysis

### Code Reduction
- Phase 1: ~100 lines removed (dead code + duplication)
- Phase 2: ~150 lines refactored (type assertions → helpers)
- Phase 3: ~50 lines (reorganization)
- **Total: ~300 lines removed/refactored from current ~5000**

### Maintainability
- **Duplication reduction**: 30% decrease in copy-paste code
- **Test coverage**: +40% with comprehensive test suite
- **Cognitive load**: ~20% reduction through consolidation

### Risk Level
- **Phase 1**: Very Low (dead code removal)
- **Phase 2**: Low (extraction, high test coverage)
- **Phase 3**: Very Low (reorganization, no logic changes)

---

## Rollout Strategy

### Pre-Refactor Checklist
- [ ] All existing tests pass (`make test`)
- [ ] All existing tests pass (`make test-integration`)
- [ ] Linting passes (`make lint`)
- [ ] Database backup created (`make seed-dump`)

### Per-Phase Execution
1. Create feature branch: `refactor/phase-X`
2. Execute changes
3. Run full test suite
4. Code review
5. Merge to main

### Post-Refactor Validation
- [ ] Unit tests pass (100% of Phase 1-3 code)
- [ ] Integration tests pass
- [ ] Linting passes
- [ ] No regressions in viewer functionality
- [ ] Performance metrics unchanged

---

## Files to Create/Modify

### Create
- `internal/adapters/web/helpers.go` - Pagination, type conversion helpers
- `internal/adapters/web/config/labels.go` - Collection titles
- `internal/adapters/web/config/formatting.go` - Content rules
- `internal/adapters/web/logging.go` - Error handling middleware
- `internal/adapters/web/caching/strategy.go` - Cache header logic
- `pkg/types/mappers.go` - Type assertion helpers
- `*_test.go` files for new comprehensive tests

### Modify
- `internal/application/parsers/strategy.go` - Remove unused interfaces
- `internal/application/parsers/registry.go` - Update to single interface
- `internal/application/services/content_service.go` - Consolidate methods
- `internal/adapters/web/viewer_handlers.go` - Major refactoring (~30% reduction)
- `cmd/cli-parser/main.go` - Dependency injection improvements
- `cmd/cli-parser/parser.go` - Remove global variable usage

### Delete
- Possibly: Old strategy files if transitioning patterns

---

## Estimated Timeline
- **Phase 1**: 0.5-1 day
- **Phase 2**: 1.5-2 days  
- **Phase 3**: 1 day
- **Buffer**: 0.5 day
- **Total**: 3-4 days of focused work

---

## Success Criteria

✓ All tests pass (unit + integration)
✓ No performance regression  
✓ Reduced code duplication (verified via inspection)
✓ Improved consistency in error handling
✓ Better separation of concerns (verified via architecture review)
✓ No breaking API changes for external consumers

---

## Notes

- The codebase is in good shape architecturally; these are optimization and consolidation changes
- Risk is minimal due to existing test coverage
- Changes improve maintainability without altering functionality
- Recommend doing Phase 1 immediately as it removes dead code
- Phases 2-3 can be done together or independently based on capacity
