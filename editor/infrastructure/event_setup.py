"""
Event System Setup and Initialization
Configure event handlers and subscribers during application startup
"""
import logging
from domain.events import get_event_bus
from infrastructure.event_handlers import (
    get_logging_handler,
    get_metrics_handler,
    get_cache_handler
)

logger = logging.getLogger(__name__)


async def setup_event_system() -> None:
    """Initialize event system with handlers"""
    try:
        event_bus = get_event_bus()
        
        # Get handler instances
        logging_handler = get_logging_handler()
        metrics_handler = get_metrics_handler()
        cache_handler = get_cache_handler()
        
        # Subscribe handlers to all events
        event_bus.subscribe_to_all(logging_handler)
        event_bus.subscribe_to_all(metrics_handler)
        event_bus.subscribe_to_all(cache_handler)
        
        logger.info("Event system initialized with handlers: logging, metrics, cache")
        
    except Exception as e:
        logger.error(f"Failed to setup event system: {e}")
        raise


def get_metrics_summary():
    """Get current metrics from the metrics handler"""
    try:
        metrics_handler = get_metrics_handler()
        return metrics_handler.get_metrics()
    except Exception as e:
        logger.error(f"Error getting metrics summary: {e}")
        return {}


def reset_metrics():
    """Reset all collected metrics"""
    try:
        metrics_handler = get_metrics_handler()
        metrics_handler.reset_metrics()
        logger.info("Metrics reset successfully")
    except Exception as e:
        logger.error(f"Error resetting metrics: {e}")


def get_cache_invalidation_keys():
    """Get cache keys that need invalidation"""
    try:
        cache_handler = get_cache_handler()
        return cache_handler.get_keys_to_invalidate()
    except Exception as e:
        logger.error(f"Error getting cache invalidation keys: {e}")
        return set()