"""
Infrastructure Event Handlers
Event handlers that perform infrastructure-specific operations
"""
import logging
from typing import Dict, Any
from datetime import datetime
from domain.events import (
    DomainEvent, EventHandler,
    DocumentViewedEvent, SearchPerformedEvent,
    FilterAppliedEvent, NavigationPerformedEvent,
    ContentRenderedEvent
)

logger = logging.getLogger(__name__)


class LoggingEventHandler:
    """Event handler that logs all events for debugging and monitoring"""
    
    async def handle(self, event: DomainEvent) -> None:
        """Log event details"""
        logger.info(
            f"Event: {event.event_type} | "
            f"ID: {event.event_id} | "
            f"Time: {event.occurred_at.isoformat()} | "
            f"Data: {event.event_data}"
        )


class MetricsEventHandler:
    """Event handler that collects metrics and analytics"""
    
    def __init__(self):
        self._metrics: Dict[str, Any] = {
            "total_events": 0,
            "events_by_type": {},
            "document_views": {},
            "search_metrics": {
                "total_searches": 0,
                "total_results": 0,
                "avg_search_time": 0.0
            },
            "navigation_metrics": {
                "total_navigations": 0,
                "directions": {}
            },
            "content_metrics": {
                "total_renders": 0,
                "formats": {},
                "avg_render_time": 0.0
            }
        }
    
    async def handle(self, event: DomainEvent) -> None:
        """Collect metrics from events"""
        try:
            self._metrics["total_events"] += 1
            
            event_type = event.event_type
            if event_type not in self._metrics["events_by_type"]:
                self._metrics["events_by_type"][event_type] = 0
            self._metrics["events_by_type"][event_type] += 1
            
            # Handle specific event types
            if isinstance(event, DocumentViewedEvent):
                await self._handle_document_viewed(event)
            elif isinstance(event, SearchPerformedEvent):
                await self._handle_search_performed(event)
            elif isinstance(event, NavigationPerformedEvent):
                await self._handle_navigation_performed(event)
            elif isinstance(event, ContentRenderedEvent):
                await self._handle_content_rendered(event)
                
        except Exception as e:
            logger.error(f"Error in MetricsEventHandler: {e}")
    
    async def _handle_document_viewed(self, event: DocumentViewedEvent) -> None:
        """Handle document viewed metrics"""
        key = f"{event.collection}:{event.document_slug}"
        if key not in self._metrics["document_views"]:
            self._metrics["document_views"][key] = {
                "count": 0,
                "collection": event.collection,
                "slug": event.document_slug,
                "name": event.document_name
            }
        self._metrics["document_views"][key]["count"] += 1
    
    async def _handle_search_performed(self, event: SearchPerformedEvent) -> None:
        """Handle search metrics"""
        metrics = self._metrics["search_metrics"]
        metrics["total_searches"] += 1
        metrics["total_results"] += event.results_count
        
        # Update average search time
        if metrics["total_searches"] > 1:
            metrics["avg_search_time"] = (
                (metrics["avg_search_time"] * (metrics["total_searches"] - 1) + 
                 event.search_time_ms) / metrics["total_searches"]
            )
        else:
            metrics["avg_search_time"] = event.search_time_ms
    
    async def _handle_navigation_performed(self, event: NavigationPerformedEvent) -> None:
        """Handle navigation metrics"""
        metrics = self._metrics["navigation_metrics"]
        metrics["total_navigations"] += 1
        
        direction = event.direction
        if direction not in metrics["directions"]:
            metrics["directions"][direction] = 0
        metrics["directions"][direction] += 1
    
    async def _handle_content_rendered(self, event: ContentRenderedEvent) -> None:
        """Handle content rendering metrics"""
        metrics = self._metrics["content_metrics"]
        metrics["total_renders"] += 1
        
        content_format = event.content_format
        if content_format not in metrics["formats"]:
            metrics["formats"][content_format] = 0
        metrics["formats"][content_format] += 1
        
        # Update average render time
        if metrics["total_renders"] > 1:
            metrics["avg_render_time"] = (
                (metrics["avg_render_time"] * (metrics["total_renders"] - 1) + 
                 event.rendering_time_ms) / metrics["total_renders"]
            )
        else:
            metrics["avg_render_time"] = event.rendering_time_ms
    
    def get_metrics(self) -> Dict[str, Any]:
        """Get current metrics snapshot"""
        return self._metrics.copy()
    
    def reset_metrics(self) -> None:
        """Reset all metrics"""
        self._metrics = {
            "total_events": 0,
            "events_by_type": {},
            "document_views": {},
            "search_metrics": {
                "total_searches": 0,
                "total_results": 0,
                "avg_search_time": 0.0
            },
            "navigation_metrics": {
                "total_navigations": 0,
                "directions": {}
            },
            "content_metrics": {
                "total_renders": 0,
                "formats": {},
                "avg_render_time": 0.0
            }
        }


class CacheInvalidationEventHandler:
    """Event handler that manages cache invalidation based on events"""
    
    def __init__(self):
        self._cache_keys_to_invalidate = set()
    
    async def handle(self, event: DomainEvent) -> None:
        """Handle cache invalidation based on event type"""
        try:
            if isinstance(event, SearchPerformedEvent):
                # Invalidate search-related caches
                self._cache_keys_to_invalidate.add(f"search_results:{event.collection}")
                self._cache_keys_to_invalidate.add(f"filter_options:{event.collection}")
            
            elif isinstance(event, DocumentViewedEvent):
                # Invalidate document-specific caches
                self._cache_keys_to_invalidate.add(f"document:{event.collection}:{event.document_slug}")
                self._cache_keys_to_invalidate.add(f"navigation:{event.collection}:{event.document_slug}")
            
            elif isinstance(event, ContentRenderedEvent):
                # Invalidate content caches
                self._cache_keys_to_invalidate.add(f"content:{event.collection}:{event.document_slug}")
                
            logger.debug(f"Cache invalidation scheduled for keys: {list(self._cache_keys_to_invalidate)}")
            
        except Exception as e:
            logger.error(f"Error in CacheInvalidationEventHandler: {e}")
    
    def get_keys_to_invalidate(self) -> set:
        """Get and clear cache keys that need invalidation"""
        keys = self._cache_keys_to_invalidate.copy()
        self._cache_keys_to_invalidate.clear()
        return keys


# Global event handler instances
_logging_handler: LoggingEventHandler = None
_metrics_handler: MetricsEventHandler = None
_cache_handler: CacheInvalidationEventHandler = None


def get_logging_handler() -> LoggingEventHandler:
    """Get global logging event handler"""
    global _logging_handler
    if _logging_handler is None:
        _logging_handler = LoggingEventHandler()
    return _logging_handler


def get_metrics_handler() -> MetricsEventHandler:
    """Get global metrics event handler"""
    global _metrics_handler
    if _metrics_handler is None:
        _metrics_handler = MetricsEventHandler()
    return _metrics_handler


def get_cache_handler() -> CacheInvalidationEventHandler:
    """Get global cache invalidation handler"""
    global _cache_handler
    if _cache_handler is None:
        _cache_handler = CacheInvalidationEventHandler()
    return _cache_handler