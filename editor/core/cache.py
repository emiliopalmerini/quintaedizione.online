"""Redis caching system for D&D 5e SRD Editor."""
from __future__ import annotations

import asyncio
import json
import hashlib
import time
from contextlib import asynccontextmanager
from dataclasses import dataclass, asdict
from typing import Any, Dict, List, Optional, Union, Callable, AsyncGenerator
from functools import wraps

try:
    import redis.asyncio as redis
    REDIS_AVAILABLE = True
except ImportError:
    redis = None
    REDIS_AVAILABLE = False

from core.errors import DatabaseError, ErrorCode
from core.logging_config import get_logger

logger = get_logger(__name__)


@dataclass
class CacheConfig:
    """Redis cache configuration."""
    
    redis_url: str = "redis://localhost:6379"
    default_ttl: int = 300  # 5 minutes
    max_retries: int = 3
    retry_delay: float = 1.0
    connection_pool_size: int = 10
    socket_timeout: int = 30
    socket_connect_timeout: int = 10
    health_check_interval: int = 60
    
    # Cache key prefixes
    key_prefix: str = "dnd5e:"
    collection_counts_prefix: str = "counts:"
    document_prefix: str = "doc:"
    search_results_prefix: str = "search:"
    list_results_prefix: str = "list:"


@dataclass 
class CacheStats:
    """Cache statistics."""
    
    hits: int = 0
    misses: int = 0
    sets: int = 0
    deletes: int = 0
    errors: int = 0
    total_operations: int = 0
    avg_response_time_ms: float = 0.0
    last_error: Optional[str] = None
    last_error_time: Optional[float] = None
    connected_since: Optional[float] = None


class CacheManager:
    """Redis-based cache manager with fallback to in-memory caching."""
    
    def __init__(self, config: Optional[CacheConfig] = None):
        self.config = config or CacheConfig()
        self.stats = CacheStats()
        self._redis: Optional[redis.Redis] = None
        self._fallback_cache: Dict[str, Any] = {}
        self._fallback_ttl: Dict[str, float] = {}
        self._connection_lock = asyncio.Lock()
        self._health_check_task: Optional[asyncio.Task] = None
        self._is_healthy = False
        self._response_times: List[float] = []
        
        if not REDIS_AVAILABLE:
            logger.warning("Redis not available, using in-memory fallback cache")
    
    async def connect(self) -> None:
        """Initialize Redis connection."""
        if not REDIS_AVAILABLE:
            logger.info("Redis not available, using in-memory cache only")
            self._is_healthy = True
            self.stats.connected_since = time.time()
            return
        
        async with self._connection_lock:
            if self._redis is not None:
                logger.info("Cache already connected")
                return
            
            try:
                logger.info("Connecting to Redis cache", extra={"url": self._mask_url(self.config.redis_url)})
                
                self._redis = redis.from_url(
                    self.config.redis_url,
                    max_connections=self.config.connection_pool_size,
                    socket_timeout=self.config.socket_timeout,
                    socket_connect_timeout=self.config.socket_connect_timeout,
                    retry_on_timeout=True,
                    health_check_interval=self.config.health_check_interval,
                    decode_responses=True,
                )
                
                # Test connection
                await self._test_connection()
                
                self._is_healthy = True
                self.stats.connected_since = time.time()
                
                # Start health monitoring
                await self._start_health_monitoring()
                
                logger.info("Redis cache connected successfully")
                
            except Exception as e:
                logger.warning(f"Failed to connect to Redis, using fallback cache: {str(e)}")
                self._redis = None
                self._is_healthy = True  # Fallback cache is always "healthy"
                self.stats.connected_since = time.time()
    
    async def disconnect(self) -> None:
        """Close Redis connection."""
        async with self._connection_lock:
            if self._health_check_task:
                self._health_check_task.cancel()
                try:
                    await self._health_check_task
                except asyncio.CancelledError:
                    pass
                self._health_check_task = None
            
            if self._redis:
                try:
                    await self._redis.aclose()
                except Exception as e:
                    logger.warning(f"Error closing Redis connection: {str(e)}")
                self._redis = None
            
            # Clear fallback cache
            self._fallback_cache.clear()
            self._fallback_ttl.clear()
            self._is_healthy = False
            
            logger.info("Cache disconnected")
    
    async def _test_connection(self) -> None:
        """Test Redis connection."""
        if self._redis:
            try:
                await asyncio.wait_for(self._redis.ping(), timeout=5.0)
                logger.debug("Redis connection test successful")
            except Exception as e:
                raise DatabaseError(f"Redis connection test failed: {str(e)}", ErrorCode.DATABASE_CONNECTION_FAILED)
    
    async def _start_health_monitoring(self) -> None:
        """Start background health monitoring."""
        if self._redis and self._health_check_task is None:
            self._health_check_task = asyncio.create_task(self._health_monitor_loop())
    
    async def _health_monitor_loop(self) -> None:
        """Background health monitoring loop."""
        while True:
            try:
                await asyncio.sleep(self.config.health_check_interval)
                
                if self._redis:
                    try:
                        await asyncio.wait_for(self._redis.ping(), timeout=5.0)
                        if not self._is_healthy:
                            logger.info("Redis connection restored")
                            self._is_healthy = True
                    except Exception as e:
                        if self._is_healthy:
                            logger.error(f"Redis health check failed, falling back to in-memory cache: {str(e)}")
                            self._is_healthy = False
                            self.stats.last_error = str(e)
                            self.stats.last_error_time = time.time()
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Error in cache health monitoring: {str(e)}")
    
    def _make_key(self, *parts: str) -> str:
        """Create cache key from parts."""
        return self.config.key_prefix + ":".join(str(p) for p in parts)
    
    def _hash_key(self, data: Any) -> str:
        """Create hash-based cache key from complex data."""
        serialized = json.dumps(data, sort_keys=True, ensure_ascii=False)
        return hashlib.md5(serialized.encode('utf-8')).hexdigest()
    
    async def get(self, key: str) -> Optional[Any]:
        """Get value from cache."""
        start_time = time.time()
        
        try:
            if self._redis and self._is_healthy:
                # Try Redis first
                try:
                    value = await self._redis.get(key)
                    if value is not None:
                        self.stats.hits += 1
                        result = json.loads(value)
                        self._update_response_time(start_time)
                        logger.debug(f"Cache hit (Redis): {key}")
                        return result
                except Exception as e:
                    logger.warning(f"Redis get failed, trying fallback: {str(e)}")
                    self.stats.errors += 1
            
            # Fallback to in-memory cache
            if key in self._fallback_cache:
                # Check TTL
                if key in self._fallback_ttl and time.time() > self._fallback_ttl[key]:
                    del self._fallback_cache[key]
                    del self._fallback_ttl[key]
                else:
                    self.stats.hits += 1
                    self._update_response_time(start_time)
                    logger.debug(f"Cache hit (fallback): {key}")
                    return self._fallback_cache[key]
            
            # Cache miss
            self.stats.misses += 1
            logger.debug(f"Cache miss: {key}")
            return None
            
        finally:
            self.stats.total_operations += 1
    
    async def set(self, key: str, value: Any, ttl: Optional[int] = None) -> bool:
        """Set value in cache."""
        start_time = time.time()
        ttl = ttl or self.config.default_ttl
        
        try:
            serialized = json.dumps(value, ensure_ascii=False)
            
            if self._redis and self._is_healthy:
                # Try Redis first
                try:
                    await self._redis.setex(key, ttl, serialized)
                    self.stats.sets += 1
                    self._update_response_time(start_time)
                    logger.debug(f"Cache set (Redis): {key}, TTL: {ttl}s")
                    return True
                except Exception as e:
                    logger.warning(f"Redis set failed, using fallback: {str(e)}")
                    self.stats.errors += 1
            
            # Fallback to in-memory cache
            self._fallback_cache[key] = value
            self._fallback_ttl[key] = time.time() + ttl
            self.stats.sets += 1
            self._update_response_time(start_time)
            logger.debug(f"Cache set (fallback): {key}, TTL: {ttl}s")
            
            # Limit fallback cache size
            if len(self._fallback_cache) > 1000:
                self._cleanup_fallback_cache()
            
            return True
            
        except Exception as e:
            logger.error(f"Failed to set cache key {key}: {str(e)}")
            self.stats.errors += 1
            return False
        finally:
            self.stats.total_operations += 1
    
    async def delete(self, key: str) -> bool:
        """Delete value from cache."""
        start_time = time.time()
        
        try:
            deleted = False
            
            if self._redis and self._is_healthy:
                try:
                    result = await self._redis.delete(key)
                    deleted = result > 0
                except Exception as e:
                    logger.warning(f"Redis delete failed: {str(e)}")
                    self.stats.errors += 1
            
            # Also delete from fallback cache
            if key in self._fallback_cache:
                del self._fallback_cache[key]
                if key in self._fallback_ttl:
                    del self._fallback_ttl[key]
                deleted = True
            
            if deleted:
                self.stats.deletes += 1
                logger.debug(f"Cache delete: {key}")
            
            self._update_response_time(start_time)
            return deleted
            
        except Exception as e:
            logger.error(f"Failed to delete cache key {key}: {str(e)}")
            self.stats.errors += 1
            return False
        finally:
            self.stats.total_operations += 1
    
    async def delete_pattern(self, pattern: str) -> int:
        """Delete keys matching pattern."""
        try:
            deleted_count = 0
            
            if self._redis and self._is_healthy:
                try:
                    keys = await self._redis.keys(pattern)
                    if keys:
                        deleted_count = await self._redis.delete(*keys)
                        logger.debug(f"Cache pattern delete (Redis): {pattern} ({deleted_count} keys)")
                except Exception as e:
                    logger.warning(f"Redis pattern delete failed: {str(e)}")
                    self.stats.errors += 1
            
            # Fallback cache pattern delete
            import fnmatch
            fallback_keys = [k for k in self._fallback_cache.keys() if fnmatch.fnmatch(k, pattern)]
            for key in fallback_keys:
                del self._fallback_cache[key]
                if key in self._fallback_ttl:
                    del self._fallback_ttl[key]
                deleted_count += 1
            
            if fallback_keys:
                logger.debug(f"Cache pattern delete (fallback): {pattern} ({len(fallback_keys)} keys)")
            
            self.stats.deletes += deleted_count
            return deleted_count
            
        except Exception as e:
            logger.error(f"Failed to delete cache pattern {pattern}: {str(e)}")
            self.stats.errors += 1
            return 0
    
    def _cleanup_fallback_cache(self) -> None:
        """Clean up expired entries from fallback cache."""
        current_time = time.time()
        expired_keys = [
            key for key, expire_time in self._fallback_ttl.items()
            if current_time > expire_time
        ]
        
        for key in expired_keys:
            del self._fallback_cache[key]
            del self._fallback_ttl[key]
        
        logger.debug(f"Cleaned up {len(expired_keys)} expired cache entries")
    
    def _update_response_time(self, start_time: float) -> None:
        """Update average response time statistics."""
        response_time = (time.time() - start_time) * 1000  # Convert to ms
        self._response_times.append(response_time)
        
        # Keep only last 1000 response times
        if len(self._response_times) > 1000:
            self._response_times = self._response_times[-1000:]
        
        self.stats.avg_response_time_ms = sum(self._response_times) / len(self._response_times)
    
    async def get_health_status(self) -> Dict[str, Any]:
        """Get cache health status."""
        health = {
            "status": "healthy" if self._is_healthy else "unhealthy",
            "type": "redis" if (self._redis and self._is_healthy) else "fallback",
            "stats": asdict(self.stats),
            "config": {
                "redis_available": REDIS_AVAILABLE,
                "default_ttl": self.config.default_ttl,
                "connection_pool_size": self.config.connection_pool_size,
            }
        }
        
        if self._redis and self._is_healthy:
            try:
                info = await self._redis.info()
                health["redis_info"] = {
                    "version": info.get("redis_version"),
                    "memory_used": info.get("used_memory_human"),
                    "connected_clients": info.get("connected_clients"),
                    "total_connections_received": info.get("total_connections_received"),
                    "keyspace_hits": info.get("keyspace_hits", 0),
                    "keyspace_misses": info.get("keyspace_misses", 0),
                }
            except Exception as e:
                health["redis_error"] = str(e)
        
        if not self._redis or not self._is_healthy:
            health["fallback_stats"] = {
                "entries": len(self._fallback_cache),
                "ttl_entries": len(self._fallback_ttl),
            }
        
        return health
    
    def _mask_url(self, url: str) -> str:
        """Mask sensitive information in URL for logging."""
        try:
            from urllib.parse import urlparse, urlunparse
            parsed = urlparse(url)
            if parsed.password:
                netloc = f"{parsed.username}:***@{parsed.hostname}"
                if parsed.port:
                    netloc += f":{parsed.port}"
                masked = parsed._replace(netloc=netloc)
                return urlunparse(masked)
            return url
        except Exception:
            return "redis://***"


# Cache decorators
def cached(ttl: int = 300, key_func: Optional[Callable] = None):
    """Decorator to cache function results."""
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        async def wrapper(*args, **kwargs):
            cache = await get_cache_manager()
            
            # Generate cache key
            if key_func:
                cache_key = key_func(*args, **kwargs)
            else:
                key_parts = [func.__name__]
                key_parts.extend(str(arg) for arg in args)
                key_parts.extend(f"{k}:{v}" for k, v in sorted(kwargs.items()))
                cache_key = cache._make_key(*key_parts)
            
            # Try to get from cache
            cached_result = await cache.get(cache_key)
            if cached_result is not None:
                return cached_result
            
            # Execute function and cache result
            result = await func(*args, **kwargs)
            await cache.set(cache_key, result, ttl)
            return result
        
        return wrapper
    return decorator


def cache_key_for_collection_counts(lang: str) -> str:
    """Generate cache key for collection counts."""
    manager = _cache_manager  # Assume global manager exists
    if manager:
        return manager._make_key(manager.config.collection_counts_prefix, lang)
    return f"dnd5e:counts:{lang}"


def cache_key_for_document(collection: str, slug: str, lang: str = "it") -> str:
    """Generate cache key for document."""
    manager = _cache_manager
    if manager:
        return manager._make_key(manager.config.document_prefix, collection, lang, slug)
    return f"dnd5e:doc:{collection}:{lang}:{slug}"


def cache_key_for_search(collection: str, query: str, filters: Dict[str, Any], lang: str = "it") -> str:
    """Generate cache key for search results."""
    manager = _cache_manager
    if manager:
        filter_hash = manager._hash_key(filters)
        return manager._make_key(manager.config.search_results_prefix, collection, lang, query, filter_hash)
    filter_hash = hashlib.md5(json.dumps(filters, sort_keys=True).encode()).hexdigest()
    return f"dnd5e:search:{collection}:{lang}:{query}:{filter_hash}"


# Global cache manager
_cache_manager: Optional[CacheManager] = None


async def get_cache_manager() -> CacheManager:
    """Get global cache manager instance."""
    global _cache_manager
    if _cache_manager is None:
        _cache_manager = CacheManager()
        await _cache_manager.connect()
    return _cache_manager


async def close_cache_manager() -> None:
    """Close global cache manager."""
    global _cache_manager
    if _cache_manager:
        await _cache_manager.disconnect()
        _cache_manager = None