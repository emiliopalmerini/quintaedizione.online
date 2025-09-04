"""
Domain Events for Content Operations
"""
from typing import Any, Dict, Optional
from dataclasses import dataclass
from .base_events import DomainEvent


@dataclass
class DocumentViewedEvent(DomainEvent):
    """Event fired when a document is viewed"""
    collection: str = ""
    document_slug: str = ""
    document_name: str = ""
    user_agent: str = ""
    referrer: str = ""
    
    def __post_init__(self):
        super().__post_init__()
        self.event_data.update({
            "collection": self.collection,
            "document_slug": self.document_slug,
            "document_name": self.document_name,
            "user_agent": self.user_agent,
            "referrer": self.referrer
        })


@dataclass
class SearchPerformedEvent(DomainEvent):
    """Event fired when a search is performed"""
    collection: str = ""
    query: str = ""
    filters: Dict[str, Any] = None
    results_count: int = 0
    search_time_ms: float = 0.0
    
    def __post_init__(self):
        super().__post_init__()
        if self.filters is None:
            self.filters = {}
        self.event_data.update({
            "collection": self.collection,
            "query": self.query,
            "filters": self.filters,
            "results_count": self.results_count,
            "search_time_ms": self.search_time_ms
        })


@dataclass
class FilterAppliedEvent(DomainEvent):
    """Event fired when filters are applied"""
    collection: str = ""
    filters: Dict[str, Any] = None
    filter_combination_valid: bool = True
    results_count: int = 0
    
    def __post_init__(self):
        super().__post_init__()
        if self.filters is None:
            self.filters = {}
        self.event_data.update({
            "collection": self.collection,
            "filters": self.filters,
            "filter_combination_valid": self.filter_combination_valid,
            "results_count": self.results_count
        })


@dataclass
class NavigationPerformedEvent(DomainEvent):
    """Event fired when navigation occurs between documents"""
    collection: str = ""
    from_document: str = ""
    to_document: str = ""
    direction: str = ""  # "next", "previous", "direct"
    navigation_context: Dict[str, Any] = None
    
    def __post_init__(self):
        super().__post_init__()
        if self.navigation_context is None:
            self.navigation_context = {}
        self.event_data.update({
            "collection": self.collection,
            "from_document": self.from_document,
            "to_document": self.to_document,
            "direction": self.direction,
            "navigation_context": self.navigation_context
        })


@dataclass
class ContentRenderedEvent(DomainEvent):
    """Event fired when content is rendered"""
    collection: str = ""
    document_slug: str = ""
    content_format: str = ""  # "markdown", "html", "plain_text"
    content_length: int = 0
    has_markdown_syntax: bool = False
    rendering_time_ms: float = 0.0
    
    def __post_init__(self):
        super().__post_init__()
        self.event_data.update({
            "collection": self.collection,
            "document_slug": self.document_slug,
            "content_format": self.content_format,
            "content_length": self.content_length,
            "has_markdown_syntax": self.has_markdown_syntax,
            "rendering_time_ms": self.rendering_time_ms
        })