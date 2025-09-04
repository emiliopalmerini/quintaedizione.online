"""
Advanced Cache Management for Performance Optimization
"""
import asyncio
import time
import json
import logging
from typing import Any, Dict, Optional, List, Set, Tuple, Union
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum

logger = logging.getLogger(__name__)


class CacheStrategy(Enum):
    """Cache eviction strategies"""
    LRU = "lru"  # Least Recently Used
    LFU = "lfu"  # Least Frequently Used
    TTL = "ttl"  # Time To Live
    ADAPTIVE = "adaptive"  # Adaptive based on usage patterns


@dataclass
class CacheKey:
    """Structured cache key with metadata"""
    namespace: str
    identifier: str
    version: str = "v1"
    parameters: Dict[str, Any] = field(default_factory=dict)
    
    def __str__(self) -> str:
        """Generate string representation for cache key"""
        param_str = ""
        if self.parameters:
            # Sort parameters for consistent key generation
            sorted_params = sorted(self.parameters.items())
            param_str = ":" + ":".join(f"{k}={v}" for k, v in sorted_params)
        
        return f"{self.namespace}:{self.version}:{self.identifier}{param_str}"


@dataclass
class CacheEntry:
    """Cache entry with metadata"""
    value: Any
    created_at: datetime
    last_accessed: datetime
    access_count: int = 0
    ttl_seconds: Optional[int] = None
    tags: Set[str] = field(default_factory=set)
    size_bytes: int = 0
    
    def is_expired(self) -> bool:
        """Check if entry is expired based on TTL"""
        if self.ttl_seconds is None:
            return False
        
        expiry_time = self.created_at + timedelta(seconds=self.ttl_seconds)
        return datetime.now() > expiry_time
    
    def mark_accessed(self) -> None:
        """Update access metadata"""
        self.last_accessed = datetime.now()
        self.access_count += 1


@dataclass 
class CacheStats:
    """Cache performance statistics"""
    total_entries: int = 0
    total_size_bytes: int = 0
    hit_count: int = 0
    miss_count: int = 0
    eviction_count: int = 0
    expired_count: int = 0
    
    @property
    def hit_ratio(self) -> float:
        """Calculate cache hit ratio"""
        total_requests = self.hit_count + self.miss_count
        return (self.hit_count / total_requests) if total_requests > 0 else 0.0
    
    @property
    def miss_ratio(self) -> float:
        """Calculate cache miss ratio"""
        return 1.0 - self.hit_ratio


class CacheManager:
    """Advanced cache manager with multiple strategies and performance optimization"""
    
    def __init__(
        self,
        max_size_bytes: int = 100 * 1024 * 1024,  # 100MB default
        max_entries: int = 10000,
        default_ttl_seconds: int = 3600,  # 1 hour
        cleanup_interval_seconds: int = 300,  # 5 minutes
        strategy: CacheStrategy = CacheStrategy.ADAPTIVE
    ):
        self.max_size_bytes = max_size_bytes
        self.max_entries = max_entries
        self.default_ttl_seconds = default_ttl_seconds
        self.cleanup_interval_seconds = cleanup_interval_seconds
        self.strategy = strategy
        
        self._cache: Dict[str, CacheEntry] = {}
        self._stats = CacheStats()
        self._cleanup_task: Optional[asyncio.Task] = None
        self._lock = asyncio.Lock()
        
        # Strategy-specific data structures
        self._access_order: List[str] = []  # For LRU
        self._frequency_scores: Dict[str, float] = {}  # For LFU
        
        # Start background cleanup
        self._start_cleanup_task()
    
    async def get(self, key: Union[str, CacheKey]) -> Optional[Any]:
        """Get value from cache"""
        async with self._lock:
            key_str = str(key)
            
            if key_str not in self._cache:
                self._stats.miss_count += 1
                return None
            
            entry = self._cache[key_str]
            
            # Check expiration
            if entry.is_expired():
                await self._remove_entry(key_str)
                self._stats.miss_count += 1
                self._stats.expired_count += 1
                return None
            
            # Update access metadata
            entry.mark_accessed()
            self._update_strategy_metadata(key_str, entry)
            
            self._stats.hit_count += 1
            logger.debug(f"Cache hit for key: {key_str}")
            
            return entry.value
    
    async def set(
        self, 
        key: Union[str, CacheKey], 
        value: Any,
        ttl_seconds: Optional[int] = None,
        tags: Set[str] = None
    ) -> bool:
        """Set value in cache"""
        async with self._lock:
            key_str = str(key)
            
            # Calculate value size
            size_bytes = self._calculate_size(value)
            
            # Check if value is too large
            if size_bytes > self.max_size_bytes * 0.5:  # Max 50% of cache size
                logger.warning(f"Value too large for cache: {size_bytes} bytes")
                return False
            
            # Make space if needed
            await self._make_space(size_bytes)
            
            # Create cache entry
            entry = CacheEntry(
                value=value,
                created_at=datetime.now(),
                last_accessed=datetime.now(),
                access_count=1,
                ttl_seconds=ttl_seconds or self.default_ttl_seconds,
                tags=tags or set(),
                size_bytes=size_bytes
            )
            
            # Update existing entry stats
            if key_str in self._cache:
                old_entry = self._cache[key_str]
                self._stats.total_size_bytes -= old_entry.size_bytes
            else:
                self._stats.total_entries += 1
            
            # Store entry
            self._cache[key_str] = entry
            self._stats.total_size_bytes += size_bytes
            
            # Update strategy metadata
            self._update_strategy_metadata(key_str, entry)
            
            logger.debug(f"Cached value for key: {key_str} ({size_bytes} bytes)")
            return True
    
    async def delete(self, key: Union[str, CacheKey]) -> bool:
        """Delete value from cache"""
        async with self._lock:
            key_str = str(key)
            return await self._remove_entry(key_str)
    
    async def delete_by_tags(self, tags: Set[str]) -> int:
        """Delete all entries with any of the specified tags"""
        async with self._lock:
            keys_to_delete = []
            
            for key_str, entry in self._cache.items():
                if tags & entry.tags:  # Intersection check
                    keys_to_delete.append(key_str)
            
            deleted_count = 0
            for key_str in keys_to_delete:
                if await self._remove_entry(key_str):
                    deleted_count += 1
            
            logger.info(f"Deleted {deleted_count} entries by tags: {tags}")
            return deleted_count
    
    async def clear(self) -> None:
        """Clear all cache entries"""
        async with self._lock:
            self._cache.clear()
            self._access_order.clear()
            self._frequency_scores.clear()
            self._stats = CacheStats()
            logger.info("Cache cleared")
    
    async def get_stats(self) -> CacheStats:
        """Get cache statistics"""
        async with self._lock:
            return CacheStats(
                total_entries=len(self._cache),
                total_size_bytes=self._stats.total_size_bytes,
                hit_count=self._stats.hit_count,
                miss_count=self._stats.miss_count,
                eviction_count=self._stats.eviction_count,
                expired_count=self._stats.expired_count
            )
    
    async def get_cache_info(self) -> Dict[str, Any]:
        """Get detailed cache information"""
        async with self._lock:
            stats = await self.get_stats()
            
            # Calculate memory usage percentage
            memory_usage_percent = (stats.total_size_bytes / self.max_size_bytes) * 100
            
            # Calculate entries usage percentage
            entries_usage_percent = (stats.total_entries / self.max_entries) * 100
            
            # Get top accessed entries
            sorted_entries = sorted(
                self._cache.items(),
                key=lambda x: x[1].access_count,
                reverse=True
            )
            top_entries = [
                {
                    "key": key,
                    "access_count": entry.access_count,
                    "size_bytes": entry.size_bytes,
                    "created_at": entry.created_at.isoformat()
                }
                for key, entry in sorted_entries[:10]
            ]
            
            return {
                "stats": {
                    "total_entries": stats.total_entries,
                    "total_size_bytes": stats.total_size_bytes,
                    "hit_count": stats.hit_count,
                    "miss_count": stats.miss_count,
                    "hit_ratio": stats.hit_ratio,
                    "miss_ratio": stats.miss_ratio,
                    "eviction_count": stats.eviction_count,
                    "expired_count": stats.expired_count
                },
                "usage": {
                    "memory_usage_percent": memory_usage_percent,
                    "entries_usage_percent": entries_usage_percent,
                    "max_size_bytes": self.max_size_bytes,
                    "max_entries": self.max_entries
                },
                "configuration": {
                    "strategy": self.strategy.value,
                    "default_ttl_seconds": self.default_ttl_seconds,
                    "cleanup_interval_seconds": self.cleanup_interval_seconds
                },
                "top_entries": top_entries
            }
    
    async def _make_space(self, required_bytes: int) -> None:
        """Make space for new entry by evicting entries based on strategy"""
        current_size = self._stats.total_size_bytes
        current_entries = len(self._cache)
        
        # Check if we need to make space
        needs_space = (
            current_size + required_bytes > self.max_size_bytes or
            current_entries >= self.max_entries
        )
        
        if not needs_space:
            return
        
        # Calculate how much space we need to free
        target_size = self.max_size_bytes * 0.8  # Target 80% usage after cleanup
        bytes_to_free = max(0, current_size + required_bytes - target_size)
        
        entries_to_evict = []
        
        if self.strategy == CacheStrategy.LRU:
            entries_to_evict = await self._select_lru_evictions(bytes_to_free)
        elif self.strategy == CacheStrategy.LFU:
            entries_to_evict = await self._select_lfu_evictions(bytes_to_free)
        elif self.strategy == CacheStrategy.TTL:
            entries_to_evict = await self._select_ttl_evictions(bytes_to_free)
        elif self.strategy == CacheStrategy.ADAPTIVE:
            entries_to_evict = await self._select_adaptive_evictions(bytes_to_free)
        
        # Evict selected entries
        for key_str in entries_to_evict:
            await self._remove_entry(key_str)
            self._stats.eviction_count += 1
        
        if entries_to_evict:
            logger.info(f"Evicted {len(entries_to_evict)} entries to make space")
    
    async def _select_lru_evictions(self, bytes_needed: int) -> List[str]:
        """Select entries to evict using LRU strategy"""
        candidates = []
        bytes_selected = 0
        
        # Sort by last accessed time (oldest first)
        sorted_entries = sorted(
            self._cache.items(),
            key=lambda x: x[1].last_accessed
        )
        
        for key_str, entry in sorted_entries:
            candidates.append(key_str)
            bytes_selected += entry.size_bytes
            
            if bytes_selected >= bytes_needed:
                break
        
        return candidates
    
    async def _select_lfu_evictions(self, bytes_needed: int) -> List[str]:
        """Select entries to evict using LFU strategy"""
        candidates = []
        bytes_selected = 0
        
        # Sort by access count (least frequent first)
        sorted_entries = sorted(
            self._cache.items(),
            key=lambda x: x[1].access_count
        )
        
        for key_str, entry in sorted_entries:
            candidates.append(key_str)
            bytes_selected += entry.size_bytes
            
            if bytes_selected >= bytes_needed:
                break
        
        return candidates
    
    async def _select_ttl_evictions(self, bytes_needed: int) -> List[str]:
        """Select entries to evict using TTL strategy (soonest to expire first)"""
        candidates = []
        bytes_selected = 0
        
        # Sort by expiration time (soonest first)
        now = datetime.now()
        
        def expiry_time(item):
            key_str, entry = item
            if entry.ttl_seconds is None:
                return now + timedelta(days=365)  # Far future for no TTL
            return entry.created_at + timedelta(seconds=entry.ttl_seconds)
        
        sorted_entries = sorted(self._cache.items(), key=expiry_time)
        
        for key_str, entry in sorted_entries:
            candidates.append(key_str)
            bytes_selected += entry.size_bytes
            
            if bytes_selected >= bytes_needed:
                break
        
        return candidates
    
    async def _select_adaptive_evictions(self, bytes_needed: int) -> List[str]:
        """Select entries using adaptive strategy combining multiple factors"""
        candidates = []
        bytes_selected = 0
        
        # Score entries based on multiple factors
        scored_entries = []
        
        for key_str, entry in self._cache.items():
            # Calculate composite score (lower = more likely to evict)
            age_score = (datetime.now() - entry.last_accessed).total_seconds()
            frequency_score = 1.0 / (entry.access_count + 1)  # Invert for lower score
            size_score = entry.size_bytes / 1024  # KB, penalty for large items
            
            # Weighted composite score
            composite_score = (age_score * 0.5) + (frequency_score * 0.3) + (size_score * 0.2)
            
            scored_entries.append((key_str, entry, composite_score))
        
        # Sort by composite score (highest score = most suitable for eviction)
        scored_entries.sort(key=lambda x: x[2], reverse=True)
        
        for key_str, entry, score in scored_entries:
            candidates.append(key_str)
            bytes_selected += entry.size_bytes
            
            if bytes_selected >= bytes_needed:
                break
        
        return candidates
    
    async def _remove_entry(self, key_str: str) -> bool:
        """Remove entry from cache and update metadata"""
        if key_str not in self._cache:
            return False
        
        entry = self._cache[key_str]
        
        # Update statistics
        self._stats.total_entries -= 1
        self._stats.total_size_bytes -= entry.size_bytes
        
        # Remove from cache
        del self._cache[key_str]
        
        # Update strategy metadata
        if key_str in self._access_order:
            self._access_order.remove(key_str)
        if key_str in self._frequency_scores:
            del self._frequency_scores[key_str]
        
        return True
    
    def _update_strategy_metadata(self, key_str: str, entry: CacheEntry) -> None:
        """Update strategy-specific metadata"""
        if self.strategy in [CacheStrategy.LRU, CacheStrategy.ADAPTIVE]:
            # Update LRU order
            if key_str in self._access_order:
                self._access_order.remove(key_str)
            self._access_order.append(key_str)
        
        if self.strategy in [CacheStrategy.LFU, CacheStrategy.ADAPTIVE]:
            # Update frequency scores
            self._frequency_scores[key_str] = entry.access_count
    
    def _calculate_size(self, value: Any) -> int:
        """Calculate approximate size of value in bytes"""
        try:
            # Serialize to JSON to get approximate size
            serialized = json.dumps(value, default=str)
            return len(serialized.encode('utf-8'))
        except Exception:
            # Fallback size estimation
            return 1024  # 1KB default
    
    async def _cleanup_expired(self) -> int:
        """Clean up expired entries"""
        expired_keys = []
        
        async with self._lock:
            for key_str, entry in self._cache.items():
                if entry.is_expired():
                    expired_keys.append(key_str)
        
        expired_count = 0
        async with self._lock:
            for key_str in expired_keys:
                if await self._remove_entry(key_str):
                    expired_count += 1
                    self._stats.expired_count += 1
        
        if expired_count > 0:
            logger.debug(f"Cleaned up {expired_count} expired cache entries")
        
        return expired_count
    
    def _start_cleanup_task(self) -> None:
        """Start background cleanup task"""
        if self._cleanup_task is None or self._cleanup_task.done():
            self._cleanup_task = asyncio.create_task(self._cleanup_loop())
    
    async def _cleanup_loop(self) -> None:
        """Background cleanup loop"""
        while True:
            try:
                await asyncio.sleep(self.cleanup_interval_seconds)
                await self._cleanup_expired()
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Error in cache cleanup loop: {e}")
    
    def stop(self) -> None:
        """Stop background cleanup task"""
        if self._cleanup_task and not self._cleanup_task.done():
            self._cleanup_task.cancel()


# Global cache manager instance
_cache_manager: Optional[CacheManager] = None


def get_cache_manager() -> CacheManager:
    """Get global cache manager instance"""
    global _cache_manager
    if _cache_manager is None:
        _cache_manager = CacheManager()
    return _cache_manager