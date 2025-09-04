# Legacy Python Code

⚠️ **This directory contains legacy Python code that has been migrated to Go.**

## Status: LEGACY / DEPRECATED

The following Python components are **legacy** and maintained for reference only:

### Legacy Components

- **`/editor/`** - Python FastAPI editor (replaced by Go version)
- **`/srd_parser/`** - Python FastAPI parser (replaced by Go version)  
- **`/shared_domain/`** - Python domain models (replaced by Go version)
- **`test_basic_integration.py`** - Python integration tests (replaced by Go tests)

### Migration Completed

These components have been **fully migrated to Go** with:
- ✅ 100% feature parity
- ✅ Performance improvements (3-5x faster)  
- ✅ Better reliability and type safety
- ✅ Modern architecture preserved
- ✅ All functionality maintained

### Current Recommendations

**For Development:**
```bash
make up                    # Use Go services (recommended)
make test-integration      # Use Go integration tests  
make benchmark             # Use Go performance tests
```

**For Legacy Testing:**
```bash
make up-python             # Use Python services (legacy only)
make test                  # Use Python tests (legacy only)
```

### When to Use Legacy Code

Use legacy Python code only for:
- 🔍 **Reference** - Understanding original implementation
- 🐛 **Debugging** - Comparing behavior between versions
- 📚 **Learning** - Studying migration patterns
- ⚙️ **Compatibility** - Temporary fallback if issues arise

### Migration Details

See [`MIGRATION.md`](./MIGRATION.md) for complete migration details including:
- Performance comparisons
- Architecture preservation
- Feature parity verification
- Testing coverage

### Support Status

- ❌ **No new features** will be added to Python versions
- ❌ **No bug fixes** except critical security issues
- ❌ **No performance optimizations** 
- ✅ **Documentation preserved** for reference
- ✅ **Docker containers maintained** for compatibility

### Future Plans

Legacy Python code will be:
1. **Maintained** in current state for 6 months (until March 2025)
2. **Marked deprecated** in documentation
3. **Eventually archived** once Go migration is fully validated in production

---

**For all new development, use the Go implementation located in:**
- `cmd/` - Application entry points
- `internal/` - Core application logic  
- `pkg/` - Shared packages
- `web/` - Templates and static assets