"""
Infrastructure Command Handlers
Handlers that perform infrastructure operations based on commands
"""
import logging
import asyncio
from typing import Dict, Any, List
from datetime import datetime, timedelta

from domain.commands import (
    Command, CommandResult,
    CacheContentCommand,
    InvalidateCacheCommand,
    PreloadContentCommand,
    OptimizeSearchCommand,
    RecordAnalyticsCommand
)

logger = logging.getLogger(__name__)


class CacheContentCommandHandler:
    """Handler for content caching commands"""
    
    def __init__(self):
        self._cache: Dict[str, Dict[str, Any]] = {}
        self._cache_metadata: Dict[str, Dict[str, Any]] = {}
    
    async def handle(self, command: CacheContentCommand) -> CommandResult:
        """Handle content caching"""
        try:
            cache_key = f"{command.collection}:{command.document_slug}"
            
            # Store content in cache
            self._cache[cache_key] = {
                "content": command.content_data,
                "cached_at": datetime.now(),
                "ttl_seconds": command.cache_ttl_seconds,
                "tags": command.cache_tags
            }
            
            # Store metadata
            self._cache_metadata[cache_key] = {
                "collection": command.collection,
                "document_slug": command.document_slug,
                "size_bytes": len(str(command.content_data)),
                "tags": command.cache_tags,
                "hits": 0
            }
            
            logger.debug(f"Cached content for {cache_key}")
            return CommandResult.success_result(
                f"Content cached for {cache_key}",
                {"cache_key": cache_key}
            )
            
        except Exception as e:
            logger.error(f"Failed to cache content: {e}")
            return CommandResult.failure_result(f"Failed to cache content: {e}")
    
    def get_cached_content(self, cache_key: str) -> Dict[str, Any]:
        """Get content from cache if valid"""
        if cache_key not in self._cache:
            return None
        
        cached_item = self._cache[cache_key]
        cached_at = cached_item["cached_at"]
        ttl_seconds = cached_item["ttl_seconds"]
        
        # Check if expired
        if datetime.now() > cached_at + timedelta(seconds=ttl_seconds):
            del self._cache[cache_key]
            if cache_key in self._cache_metadata:
                del self._cache_metadata[cache_key]
            return None
        
        # Update hit count
        if cache_key in self._cache_metadata:
            self._cache_metadata[cache_key]["hits"] += 1
        
        return cached_item["content"]
    
    def get_cache_stats(self) -> Dict[str, Any]:
        """Get cache statistics"""
        total_items = len(self._cache)
        total_size = sum(
            metadata.get("size_bytes", 0) 
            for metadata in self._cache_metadata.values()
        )
        total_hits = sum(
            metadata.get("hits", 0) 
            for metadata in self._cache_metadata.values()
        )
        
        return {
            "total_items": total_items,
            "total_size_bytes": total_size,
            "total_hits": total_hits,
            "cache_keys": list(self._cache.keys())
        }


class InvalidateCacheCommandHandler:
    """Handler for cache invalidation commands"""
    
    def __init__(self, cache_handler: CacheContentCommandHandler):
        self.cache_handler = cache_handler
    
    async def handle(self, command: InvalidateCacheCommand) -> CommandResult:
        """Handle cache invalidation"""
        try:
            invalidated_keys = []
            
            if command.invalidate_all:
                # Clear all cache
                invalidated_keys = list(self.cache_handler._cache.keys())
                self.cache_handler._cache.clear()
                self.cache_handler._cache_metadata.clear()
            else:
                # Invalidate specific keys
                for key in command.cache_keys:
                    if key in self.cache_handler._cache:
                        del self.cache_handler._cache[key]
                        invalidated_keys.append(key)
                    if key in self.cache_handler._cache_metadata:
                        del self.cache_handler._cache_metadata[key]
                
                # Invalidate pattern matches
                for pattern in command.cache_patterns:
                    matching_keys = [
                        key for key in self.cache_handler._cache.keys() 
                        if pattern in key
                    ]
                    for key in matching_keys:
                        del self.cache_handler._cache[key]
                        if key in self.cache_handler._cache_metadata:
                            del self.cache_handler._cache_metadata[key]
                        invalidated_keys.append(key)
            
            logger.debug(f"Invalidated {len(invalidated_keys)} cache entries")
            return CommandResult.success_result(
                f"Invalidated {len(invalidated_keys)} cache entries",
                {"invalidated_keys": invalidated_keys}
            )
            
        except Exception as e:
            logger.error(f"Failed to invalidate cache: {e}")
            return CommandResult.failure_result(f"Failed to invalidate cache: {e}")


class PreloadContentCommandHandler:
    """Handler for content preloading commands"""
    
    def __init__(self, cache_handler: CacheContentCommandHandler):
        self.cache_handler = cache_handler
    
    async def handle(self, command: PreloadContentCommand) -> CommandResult:
        """Handle content preloading"""
        try:
            preloaded_count = 0
            
            # This is a simplified implementation
            # In a real system, this would fetch popular content from the database
            # and populate the cache
            
            for collection in command.collections:
                # Simulate preloading popular documents
                for i in range(min(command.preload_count, 10)):
                    cache_key = f"{collection}:popular_doc_{i}"
                    
                    # Simulate content
                    mock_content = {
                        "name": f"Popular Document {i}",
                        "collection": collection,
                        "content": f"Preloaded content for {collection}"
                    }
                    
                    self.cache_handler._cache[cache_key] = {
                        "content": mock_content,
                        "cached_at": datetime.now(),
                        "ttl_seconds": 3600,
                        "tags": ["preloaded", collection]
                    }
                    preloaded_count += 1
            
            # Handle priority documents
            for priority_doc in command.priority_documents:
                collection = priority_doc.get("collection", "")
                slug = priority_doc.get("slug", "")
                cache_key = f"{collection}:{slug}"
                
                # Mock priority content
                mock_content = {
                    "name": slug,
                    "collection": collection,
                    "content": f"Priority content for {slug}"
                }
                
                self.cache_handler._cache[cache_key] = {
                    "content": mock_content,
                    "cached_at": datetime.now(),
                    "ttl_seconds": 7200,  # Longer TTL for priority content
                    "tags": ["priority", "preloaded", collection]
                }
                preloaded_count += 1
            
            logger.info(f"Preloaded {preloaded_count} documents")
            return CommandResult.success_result(
                f"Preloaded {preloaded_count} documents",
                {"preloaded_count": preloaded_count}
            )
            
        except Exception as e:
            logger.error(f"Failed to preload content: {e}")
            return CommandResult.failure_result(f"Failed to preload content: {e}")


class OptimizeSearchCommandHandler:
    """Handler for search optimization commands"""
    
    def __init__(self):
        self._optimization_stats: Dict[str, Any] = {}
    
    async def handle(self, command: OptimizeSearchCommand) -> CommandResult:
        """Handle search optimization"""
        try:
            optimizations_applied = []
            
            if command.rebuild_indexes:
                # Simulate index rebuilding
                await asyncio.sleep(0.1)  # Simulate work
                optimizations_applied.append("indexes_rebuilt")
            
            if command.optimize_queries:
                # Simulate query optimization
                await asyncio.sleep(0.05)
                optimizations_applied.append("queries_optimized")
            
            if command.analyze_performance:
                # Simulate performance analysis
                await asyncio.sleep(0.02)
                self._optimization_stats[command.collection] = {
                    "last_optimized": datetime.now().isoformat(),
                    "target_response_time_ms": command.target_response_time_ms,
                    "current_avg_response_time_ms": 85.0,  # Mock value
                    "improvement_percentage": 15.0
                }
                optimizations_applied.append("performance_analyzed")
            
            logger.info(f"Applied search optimizations: {optimizations_applied}")
            return CommandResult.success_result(
                f"Applied {len(optimizations_applied)} search optimizations",
                {
                    "optimizations": optimizations_applied,
                    "stats": self._optimization_stats.get(command.collection, {})
                }
            )
            
        except Exception as e:
            logger.error(f"Failed to optimize search: {e}")
            return CommandResult.failure_result(f"Failed to optimize search: {e}")
    
    def get_optimization_stats(self) -> Dict[str, Any]:
        """Get search optimization statistics"""
        return self._optimization_stats.copy()


class RecordAnalyticsCommandHandler:
    """Handler for analytics recording commands"""
    
    def __init__(self):
        self._analytics_store: List[Dict[str, Any]] = []
    
    async def handle(self, command: RecordAnalyticsCommand) -> CommandResult:
        """Handle analytics recording"""
        try:
            analytics_record = {
                "event_type": command.event_type,
                "event_data": command.event_data,
                "user_session": command.user_session,
                "timestamp": command.timestamp or datetime.now().isoformat(),
                "collection": command.collection,
                "document_slug": command.document_slug,
                "recorded_at": datetime.now().isoformat()
            }
            
            self._analytics_store.append(analytics_record)
            
            # Keep only last 1000 records (simple cleanup)
            if len(self._analytics_store) > 1000:
                self._analytics_store = self._analytics_store[-1000:]
            
            logger.debug(f"Recorded analytics event: {command.event_type}")
            return CommandResult.success_result(
                f"Recorded analytics event: {command.event_type}",
                {"record_id": len(self._analytics_store)}
            )
            
        except Exception as e:
            logger.error(f"Failed to record analytics: {e}")
            return CommandResult.failure_result(f"Failed to record analytics: {e}")
    
    def get_analytics_summary(self) -> Dict[str, Any]:
        """Get analytics summary"""
        if not self._analytics_store:
            return {"total_events": 0}
        
        event_types = {}
        collections = {}
        
        for record in self._analytics_store:
            event_type = record.get("event_type", "unknown")
            collection = record.get("collection", "unknown")
            
            event_types[event_type] = event_types.get(event_type, 0) + 1
            collections[collection] = collections.get(collection, 0) + 1
        
        return {
            "total_events": len(self._analytics_store),
            "event_types": event_types,
            "collections": collections,
            "latest_events": self._analytics_store[-10:]  # Last 10 events
        }