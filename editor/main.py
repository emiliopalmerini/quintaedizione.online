"""D&D 5e SRD Editor with Hexagonal Architecture."""

import asyncio
import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from fastapi.middleware.cors import CORSMiddleware

from core.database import close_database
from routers.pages import router as pages_router

# Hexagonal Architecture Components
try:
    from infrastructure.event_setup import setup_event_system
    from infrastructure.command_setup import setup_command_system
    from infrastructure.performance import get_cache_manager, get_metrics_collector
    HEXAGONAL_AVAILABLE = True
except ImportError as e:
    logging.warning(f"Hexagonal architecture components not available: {e}")
    HEXAGONAL_AVAILABLE = False

logger = logging.getLogger(__name__)


async def setup_hexagonal_systems():
    """Initialize hexagonal architecture systems"""
    if not HEXAGONAL_AVAILABLE:
        logger.info("Hexagonal systems not available, using legacy mode")
        return
    
    try:
        # 1. Setup event system
        await setup_event_system()
        logger.info("✅ Event system initialized")
        
        # 2. Setup command system
        await setup_command_system()
        logger.info("✅ Command system initialized")
        
        # 3. Setup performance systems
        cache_manager = get_cache_manager()
        metrics_collector = get_metrics_collector()
        await metrics_collector.start_collection()
        logger.info("✅ Performance systems initialized")
        
        # 4. Execute initial cache preload (background)
        from infrastructure.command_setup import execute_cache_preload
        asyncio.create_task(_background_cache_preload())
        
        logger.info("🚀 Hexagonal architecture fully initialized")
        
    except Exception as e:
        logger.error(f"❌ Failed to initialize hexagonal systems: {e}")
        # Continue with legacy mode
        pass


async def _background_cache_preload():
    """Background task for cache preloading"""
    try:
        await asyncio.sleep(5)  # Wait for app to be fully started
        from infrastructure.command_setup import execute_cache_preload
        result = await execute_cache_preload(["incantesimi", "mostri", "classi"])
        logger.info(f"Cache preload completed: {result}")
    except Exception as e:
        logger.warning(f"Cache preload failed (non-critical): {e}")


async def shutdown_hexagonal_systems():
    """Shutdown hexagonal architecture systems"""
    if not HEXAGONAL_AVAILABLE:
        return
    
    try:
        # Stop metrics collection
        metrics_collector = get_metrics_collector()
        await metrics_collector.stop_collection()
        
        # Stop cache manager
        cache_manager = get_cache_manager()
        cache_manager.stop()
        
        logger.info("Hexagonal systems shutdown completed")
        
    except Exception as e:
        logger.error(f"Error during hexagonal systems shutdown: {e}")


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan management with hexagonal architecture."""
    # Startup
    logger.info("🚀 Starting D&D 5e SRD Editor with Hexagonal Architecture...")
    
    # Initialize hexagonal systems
    await setup_hexagonal_systems()
    
    yield
    
    # Shutdown
    logger.info("🛑 Shutting down D&D 5e SRD Editor...")
    await shutdown_hexagonal_systems()
    await close_database()


# Create FastAPI application
app = FastAPI(
    title="D&D 5e SRD Editor",
    description="Editor per il System Reference Document di Dungeons & Dragons 5e in italiano con Architettura Esagonale",
    version="3.0.0",  # Updated for hexagonal architecture
    lifespan=lifespan
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure as needed
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Static files
app.mount("/static", StaticFiles(directory="static"), name="static")

# Include routers
app.include_router(pages_router)

# Include admin router if available
try:
    from routers.admin import router as admin_router
    app.include_router(admin_router)
    logger.info("✅ Admin endpoints enabled")
except ImportError as e:
    logger.warning(f"Admin endpoints not available: {e}")

# Health check endpoint
@app.get("/health")
async def health_check():
    """Health check endpoint with system status"""
    status = {
        "status": "healthy",
        "version": "3.0.0",
        "architecture": "hexagonal" if HEXAGONAL_AVAILABLE else "legacy"
    }
    
    if HEXAGONAL_AVAILABLE:
        try:
            # Get system metrics
            metrics_collector = get_metrics_collector()
            cache_manager = get_cache_manager()
            
            cache_stats = await cache_manager.get_stats()
            current_metrics = await metrics_collector.get_current_metrics()
            
            status.update({
                "cache": {
                    "hit_ratio": cache_stats.hit_ratio,
                    "total_entries": cache_stats.total_entries,
                    "total_size_mb": cache_stats.total_size_bytes / 1024 / 1024
                },
                "system": {
                    "cpu_percent": current_metrics.cpu_percent,
                    "memory_percent": current_metrics.memory_percent,
                    "active_connections": current_metrics.active_connections
                }
            })
        except Exception as e:
            status["warning"] = f"Could not get detailed metrics: {e}"
    
    return status


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main_simple:app",
        host="0.0.0.0",
        port=8000,
        reload=True,
        log_level="info"
    )