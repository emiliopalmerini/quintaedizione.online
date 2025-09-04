"""
Content-related Commands for CQRS Write Side
Commands for content management operations
"""
from typing import Any, Dict, List, Optional
from dataclasses import dataclass
from .base_commands import Command


@dataclass
class CacheContentCommand(Command):
    """Command to cache document content for performance"""
    collection: str = ""
    document_slug: str = ""
    content_data: Dict[str, Any] = None
    cache_ttl_seconds: int = 3600
    cache_tags: List[str] = None
    
    def __post_init__(self):
        super().__post_init__()
        if self.content_data is None:
            self.content_data = {}
        if self.cache_tags is None:
            self.cache_tags = []
        
        self.command_data.update({
            "collection": self.collection,
            "document_slug": self.document_slug,
            "content_data": self.content_data,
            "cache_ttl_seconds": self.cache_ttl_seconds,
            "cache_tags": self.cache_tags
        })


@dataclass
class InvalidateCacheCommand(Command):
    """Command to invalidate cached content"""
    cache_keys: List[str] = None
    cache_patterns: List[str] = None
    collection: str = ""
    invalidate_all: bool = False
    
    def __post_init__(self):
        super().__post_init__()
        if self.cache_keys is None:
            self.cache_keys = []
        if self.cache_patterns is None:
            self.cache_patterns = []
        
        self.command_data.update({
            "cache_keys": self.cache_keys,
            "cache_patterns": self.cache_patterns,
            "collection": self.collection,
            "invalidate_all": self.invalidate_all
        })


@dataclass
class PreloadContentCommand(Command):
    """Command to preload frequently accessed content"""
    collections: List[str] = None
    priority_documents: List[Dict[str, str]] = None  # [{"collection": "x", "slug": "y"}]
    preload_count: int = 50
    include_navigation: bool = True
    
    def __post_init__(self):
        super().__post_init__()
        if self.collections is None:
            self.collections = []
        if self.priority_documents is None:
            self.priority_documents = []
        
        self.command_data.update({
            "collections": self.collections,
            "priority_documents": self.priority_documents,
            "preload_count": self.preload_count,
            "include_navigation": self.include_navigation
        })


@dataclass
class OptimizeSearchCommand(Command):
    """Command to optimize search indexes and performance"""
    collection: str = ""
    rebuild_indexes: bool = False
    optimize_queries: bool = True
    analyze_performance: bool = True
    target_response_time_ms: float = 100.0
    
    def __post_init__(self):
        super().__post_init__()
        self.command_data.update({
            "collection": self.collection,
            "rebuild_indexes": self.rebuild_indexes,
            "optimize_queries": self.optimize_queries,
            "analyze_performance": self.analyze_performance,
            "target_response_time_ms": self.target_response_time_ms
        })


@dataclass
class RecordAnalyticsCommand(Command):
    """Command to record analytics data"""
    event_type: str = ""
    event_data: Dict[str, Any] = None
    user_session: str = ""
    timestamp: str = ""
    collection: str = ""
    document_slug: str = ""
    
    def __post_init__(self):
        super().__post_init__()
        if self.event_data is None:
            self.event_data = {}
        
        self.command_data.update({
            "event_type": self.event_type,
            "event_data": self.event_data,
            "user_session": self.user_session,
            "timestamp": self.timestamp,
            "collection": self.collection,
            "document_slug": self.document_slug
        })