"""Simplified content service for D&D 5e SRD operations."""

from typing import Dict, List, Any, Optional, Tuple
from core.database import get_database
from core.repository import SimpleRepository
from core.query_builder import build_text_search, build_collection_filters, build_sort_criteria
from core.config import COLLECTIONS
from utils.markdown import render_md


class ContentService:
    """Unified service for content operations."""
    
    def __init__(self):
        self.repo: Optional[SimpleRepository] = None
    
    async def _get_repo(self) -> SimpleRepository:
        """Get repository instance."""
        if not self.repo:
            db = await get_database()
            self.repo = SimpleRepository(db)
        return self.repo
    
    async def get_collection_counts(self) -> Dict[str, int]:
        """Get document counts for all collections."""
        repo = await self._get_repo()
        counts = {}
        
        for collection in COLLECTIONS:
            try:
                count = await repo.count_documents(collection)
                counts[collection] = count
            except Exception:
                counts[collection] = 0
                
        return counts
    
    async def list_documents(
        self,
        collection: str,
        query: Optional[str] = None,
        filters: Optional[Dict[str, str]] = None,
        page: int = 1,
        page_size: int = 20,
        sort_by: str = "alpha"
    ) -> Tuple[List[Dict[str, Any]], int, bool, bool]:
        """List documents with pagination, search, and filtering."""
        repo = await self._get_repo()
        
        # Build MongoDB filter
        mongo_filter = {}
        
        # Add text search
        if query:
            text_filter = build_text_search(query, [
                "name", "nome", "title", "titolo", 
                "description", "descrizione", "content"
            ])
            if text_filter:
                mongo_filter.update(text_filter)
        
        # Add collection-specific filters
        if filters:
            collection_filter = build_collection_filters(collection, filters)
            if collection_filter:
                # Merge filters properly
                if "$and" in mongo_filter and "$and" in collection_filter:
                    mongo_filter["$and"].extend(collection_filter["$and"])
                elif "$and" in collection_filter:
                    mongo_filter["$and"] = collection_filter["$and"]
                else:
                    mongo_filter.update(collection_filter)
        
        # Get total count for pagination
        total = await repo.count_documents(collection, mongo_filter)
        
        # Calculate pagination
        skip = (page - 1) * page_size
        has_prev = page > 1
        has_next = skip + page_size < total
        
        # Get documents
        sort_criteria = build_sort_criteria(sort_by)
        documents = await repo.find_all(
            collection,
            filter_query=mongo_filter,
            sort_by=sort_criteria,
            skip=skip,
            limit=page_size
        )
        
        return documents, total, has_prev, has_next
    
    async def get_document(self, collection: str, slug: str) -> Optional[Dict[str, Any]]:
        """Get single document by slug."""
        repo = await self._get_repo()
        
        # Try different slug fields
        for slug_field in ["slug", "name", "nome", "title", "titolo"]:
            doc = await repo.find_one(collection, {slug_field: slug})
            if doc:
                return doc
        
        return None
    
    async def get_navigation_context(
        self, 
        collection: str, 
        current_slug: str,
        filters: Optional[Dict[str, str]] = None
    ) -> Tuple[Optional[str], Optional[str]]:
        """Get previous and next document slugs for navigation."""
        repo = await self._get_repo()
        
        # Build the same filter as used in listing
        mongo_filter = {}
        if filters:
            collection_filter = build_collection_filters(collection, filters)
            mongo_filter.update(collection_filter)
        
        # Get all documents with same filter/sort as list view
        sort_criteria = build_sort_criteria("alpha")
        all_docs = await repo.find_all(
            collection,
            filter_query=mongo_filter,
            sort_by=sort_criteria
        )
        
        # Find current position and get prev/next
        prev_slug = None
        next_slug = None
        
        for i, doc in enumerate(all_docs):
            doc_slug = doc.get("slug") or doc.get("name") or doc.get("nome")
            if doc_slug == current_slug:
                if i > 0:
                    prev_doc = all_docs[i - 1]
                    prev_slug = prev_doc.get("slug") or prev_doc.get("name") or prev_doc.get("nome")
                if i < len(all_docs) - 1:
                    next_doc = all_docs[i + 1]
                    next_slug = next_doc.get("slug") or next_doc.get("name") or next_doc.get("nome")
                break
        
        return prev_slug, next_slug
    
    async def render_document_content(self, doc: Dict[str, Any]) -> Tuple[Optional[str], Optional[str]]:
        """Render document content as HTML."""
        # Find content field
        content_field = None
        raw_content = None
        
        for field in ["description_md", "descrizione_md", "content", "description", "descrizione"]:
            if field in doc and doc[field]:
                content_field = field
                raw_content = doc[field]
                break
        
        if not raw_content:
            return None, None
        
        # Render markdown if it's a markdown field
        if content_field and content_field.endswith("_md"):
            html_content = render_md(raw_content)
            return html_content, raw_content
        else:
            return raw_content, raw_content


# Global service instance
_content_service: Optional[ContentService] = None


async def get_content_service() -> ContentService:
    """Get content service instance."""
    global _content_service
    if not _content_service:
        _content_service = ContentService()
    return _content_service