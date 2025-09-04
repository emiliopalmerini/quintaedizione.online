"""
Command System Setup and Initialization
Configure command handlers and bus during application startup
"""
import logging
from domain.commands import (
    get_command_bus,
    CacheContentCommand,
    InvalidateCacheCommand,
    PreloadContentCommand,
    OptimizeSearchCommand,
    RecordAnalyticsCommand
)
from infrastructure.command_handlers import (
    CacheContentCommandHandler,
    InvalidateCacheCommandHandler,
    PreloadContentCommandHandler,
    OptimizeSearchCommandHandler,
    RecordAnalyticsCommandHandler
)

logger = logging.getLogger(__name__)


# Global handler instances
_cache_handler: CacheContentCommandHandler = None
_invalidate_handler: InvalidateCacheCommandHandler = None
_preload_handler: PreloadContentCommandHandler = None
_optimize_handler: OptimizeSearchCommandHandler = None
_analytics_handler: RecordAnalyticsCommandHandler = None


async def setup_command_system() -> None:
    """Initialize command system with handlers"""
    try:
        global _cache_handler, _invalidate_handler, _preload_handler, _optimize_handler, _analytics_handler
        
        command_bus = get_command_bus()
        
        # Create handler instances
        _cache_handler = CacheContentCommandHandler()
        _invalidate_handler = InvalidateCacheCommandHandler(_cache_handler)
        _preload_handler = PreloadContentCommandHandler(_cache_handler)
        _optimize_handler = OptimizeSearchCommandHandler()
        _analytics_handler = RecordAnalyticsCommandHandler()
        
        # Register handlers with command bus
        command_bus.register_handler(CacheContentCommand, _cache_handler)
        command_bus.register_handler(InvalidateCacheCommand, _invalidate_handler)
        command_bus.register_handler(PreloadContentCommand, _preload_handler)
        command_bus.register_handler(OptimizeSearchCommand, _optimize_handler)
        command_bus.register_handler(RecordAnalyticsCommand, _analytics_handler)
        
        logger.info("Command system initialized with handlers: cache, invalidation, preload, optimize, analytics")
        
    except Exception as e:
        logger.error(f"Failed to setup command system: {e}")
        raise


def get_cache_handler() -> CacheContentCommandHandler:
    """Get cache command handler instance"""
    global _cache_handler
    if _cache_handler is None:
        _cache_handler = CacheContentCommandHandler()
    return _cache_handler


def get_analytics_handler() -> RecordAnalyticsCommandHandler:
    """Get analytics command handler instance"""
    global _analytics_handler
    if _analytics_handler is None:
        _analytics_handler = RecordAnalyticsCommandHandler()
    return _analytics_handler


def get_optimization_handler() -> OptimizeSearchCommandHandler:
    """Get optimization command handler instance"""
    global _optimize_handler
    if _optimize_handler is None:
        _optimize_handler = OptimizeSearchCommandHandler()
    return _optimize_handler


async def execute_cache_preload(collections: list = None) -> dict:
    """Execute cache preloading for specified collections"""
    try:
        command_bus = get_command_bus()
        
        if collections is None:
            collections = ["incantesimi", "mostri", "classi", "armi"]
        
        command = PreloadContentCommand(
            collections=collections,
            preload_count=20,
            include_navigation=True
        )
        
        result = await command_bus.execute(command)
        return {
            "success": result.success,
            "message": result.message,
            "data": result.data,
            "execution_time_ms": result.execution_time_ms
        }
        
    except Exception as e:
        logger.error(f"Error executing cache preload: {e}")
        return {
            "success": False,
            "message": f"Error executing cache preload: {e}",
            "data": None,
            "execution_time_ms": 0.0
        }


async def execute_search_optimization(collection: str = "") -> dict:
    """Execute search optimization for specified collection"""
    try:
        command_bus = get_command_bus()
        
        command = OptimizeSearchCommand(
            collection=collection,
            rebuild_indexes=False,
            optimize_queries=True,
            analyze_performance=True,
            target_response_time_ms=100.0
        )
        
        result = await command_bus.execute(command)
        return {
            "success": result.success,
            "message": result.message,
            "data": result.data,
            "execution_time_ms": result.execution_time_ms
        }
        
    except Exception as e:
        logger.error(f"Error executing search optimization: {e}")
        return {
            "success": False,
            "message": f"Error executing search optimization: {e}",
            "data": None,
            "execution_time_ms": 0.0
        }