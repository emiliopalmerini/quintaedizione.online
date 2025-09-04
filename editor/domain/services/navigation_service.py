"""
Navigation Domain Service  
Pure business logic for document navigation without infrastructure dependencies
"""
from typing import Dict, Any, Optional, Tuple, List
from dataclasses import dataclass
from enum import Enum


class NavigationDirection(Enum):
    """Navigation direction options"""
    PREVIOUS = "previous"
    NEXT = "next"
    FIRST = "first"
    LAST = "last"


@dataclass
class NavigationContext:
    """Navigation context with previous/next items"""
    previous_slug: Optional[str]
    next_slug: Optional[str]
    current_position: int
    total_items: int
    collection: str
    filters_applied: Dict[str, Any]


@dataclass
class NavigationItem:
    """Single navigation item"""
    slug: str
    name: str
    display_name: str
    position: int


class NavigationService:
    """Domain service for document navigation business logic"""
    
    @staticmethod
    def extract_document_slug(document: Dict[str, Any]) -> Optional[str]:
        """Extract slug from document following priority rules"""
        slug_field_priority = ["slug", "name", "nome", "title", "titolo"]
        
        for field_name in slug_field_priority:
            if field_name in document and document[field_name]:
                slug = str(document[field_name]).strip()
                if slug:
                    return slug
        
        return None
    
    @staticmethod
    def normalize_slug(slug: str) -> str:
        """Normalize slug for consistent comparison"""
        return slug.lower().strip()
    
    @staticmethod 
    def create_navigation_item(document: Dict[str, Any], position: int) -> Optional[NavigationItem]:
        """Create navigation item from document"""
        slug = NavigationService.extract_document_slug(document)
        if not slug:
            return None
        
        # Determine display name
        name = document.get("nome", document.get("name", ""))
        display_name = document.get("name", document.get("nome", slug))
        
        return NavigationItem(
            slug=slug,
            name=name,
            display_name=display_name,
            position=position
        )
    
    @staticmethod
    def find_document_position(
        documents: List[Dict[str, Any]], 
        target_slug: str
    ) -> Optional[int]:
        """Find position of document with matching slug"""
        normalized_target = NavigationService.normalize_slug(target_slug)
        
        for i, doc in enumerate(documents):
            doc_slug = NavigationService.extract_document_slug(doc)
            if doc_slug and NavigationService.normalize_slug(doc_slug) == normalized_target:
                return i
        
        return None
    
    @staticmethod
    def calculate_navigation_context(
        documents: List[Dict[str, Any]],
        current_slug: str,
        collection: str,
        filters: Optional[Dict[str, Any]] = None
    ) -> Optional[NavigationContext]:
        """Calculate navigation context for current document"""
        if not documents:
            return None
        
        current_position = NavigationService.find_document_position(documents, current_slug)
        if current_position is None:
            return None
        
        # Find previous document
        prev_slug = None
        if current_position > 0:
            prev_doc = documents[current_position - 1]
            prev_slug = NavigationService.extract_document_slug(prev_doc)
        
        # Find next document  
        next_slug = None
        if current_position < len(documents) - 1:
            next_doc = documents[current_position + 1]
            next_slug = NavigationService.extract_document_slug(next_doc)
        
        return NavigationContext(
            previous_slug=prev_slug,
            next_slug=next_slug,
            current_position=current_position + 1,  # 1-based for display
            total_items=len(documents),
            collection=collection,
            filters_applied=filters or {}
        )
    
    @staticmethod
    def build_navigation_query_params(filters: Dict[str, Any]) -> str:
        """Build query parameters for navigation links to maintain filtering context"""
        if not filters:
            return ""
        
        from urllib.parse import urlencode
        
        # Only include non-empty filters
        clean_filters = {k: v for k, v in filters.items() if v is not None and v != ""}
        if not clean_filters:
            return ""
        
        return "?" + urlencode(clean_filters)
    
    @staticmethod
    def should_enable_navigation(
        navigation_context: Optional[NavigationContext],
        min_items: int = 2
    ) -> bool:
        """Business rule: determine if navigation should be enabled"""
        if not navigation_context:
            return False
        
        # Enable navigation only if there are enough items
        return navigation_context.total_items >= min_items
    
    @staticmethod
    def get_navigation_summary(navigation_context: NavigationContext) -> str:
        """Generate human-readable navigation summary"""
        return (f"Elemento {navigation_context.current_position} "
                f"di {navigation_context.total_items}")


class CollectionNavigationService:
    """Domain service for collection-level navigation"""
    
    @staticmethod
    def calculate_pagination_info(
        total_items: int,
        current_page: int,
        items_per_page: int
    ) -> Dict[str, Any]:
        """Calculate pagination information"""
        if items_per_page <= 0:
            items_per_page = 20  # Default safe value
        
        total_pages = max(1, (total_items + items_per_page - 1) // items_per_page)
        current_page = max(1, min(current_page, total_pages))
        
        start_item = (current_page - 1) * items_per_page + 1
        end_item = min(current_page * items_per_page, total_items)
        
        return {
            "total_items": total_items,
            "total_pages": total_pages,
            "current_page": current_page,
            "items_per_page": items_per_page,
            "start_item": start_item,
            "end_item": end_item,
            "has_previous": current_page > 1,
            "has_next": current_page < total_pages,
            "previous_page": current_page - 1 if current_page > 1 else None,
            "next_page": current_page + 1 if current_page < total_pages else None
        }
    
    @staticmethod
    def generate_page_range(
        current_page: int, 
        total_pages: int, 
        max_visible_pages: int = 5
    ) -> List[int]:
        """Generate list of page numbers to display in pagination"""
        if total_pages <= max_visible_pages:
            return list(range(1, total_pages + 1))
        
        # Calculate range around current page
        half_range = max_visible_pages // 2
        start = max(1, current_page - half_range)
        end = min(total_pages, current_page + half_range)
        
        # Adjust if range is too small
        if end - start < max_visible_pages - 1:
            if start == 1:
                end = min(total_pages, start + max_visible_pages - 1)
            else:
                start = max(1, end - max_visible_pages + 1)
        
        return list(range(start, end + 1))
    
    @staticmethod
    def validate_pagination_params(
        page: int, 
        page_size: int,
        max_page_size: int = 100
    ) -> Tuple[int, int]:
        """Validate and normalize pagination parameters"""
        # Validate page number
        page = max(1, page)
        
        # Validate page size
        page_size = max(1, min(page_size, max_page_size))
        
        return page, page_size