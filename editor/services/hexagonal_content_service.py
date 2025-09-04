"""
Hexagonal Architecture Content Service
Replaces the traditional ContentService with use cases from shared_domain
"""
from typing import Dict, List, Any, Optional, Tuple
import logging
import time

from shared_domain.use_cases import (
    SearchSpellsQuery, SearchMonstersQuery, SearchDocumentsQuery,
    GetDocumentQuery, GetFilterOptionsQuery, UseCaseResult
)
from infrastructure.container import get_container
from core.config import COLLECTIONS
from utils.markdown import render_md
from domain.services import (
    ContentRenderingService,
    NavigationService,
    SearchRelevanceService,
    FilterService,
    SearchQueryService
)
from domain.events import (
    get_event_bus,
    DocumentViewedEvent,
    SearchPerformedEvent,
    FilterAppliedEvent,
    NavigationPerformedEvent,
    ContentRenderedEvent
)

logger = logging.getLogger(__name__)


class HexagonalContentService:
    """Content service using hexagonal architecture with shared domain use cases"""
    
    def __init__(self):
        self.container = get_container()
        self.event_bus = get_event_bus()
    
    async def get_collection_counts(self) -> Dict[str, int]:
        """Get document counts for all collections using direct MongoDB queries"""
        # For now, keep the same implementation as the old service
        # This could be converted to use cases later
        counts = {}
        
        for collection in COLLECTIONS:
            try:
                if collection == "incantesimi":
                    search_use_case = self.container.get_search_spells_use_case()
                    result = await search_use_case.handle(SearchSpellsQuery(limit=1))
                    # We can't get count easily with current use case, so use repository directly
                    repo = self.container.get_spell_query_repository()
                    count = len(await repo.search_spells(
                        self._create_empty_spell_search_query()
                    ))
                    counts[collection] = count
                elif collection == "mostri":
                    # Similar for monsters
                    repo = self.container.get_monster_query_repository()
                    count = len(await repo.search_monsters(
                        self._create_empty_monster_search_query()
                    ))
                    counts[collection] = count
                else:
                    # For other document types
                    repo = self.container.get_document_query_repository()
                    count = len(await repo.search_documents(
                        self._create_empty_document_search_query(collection)
                    ))
                    counts[collection] = count
            except Exception as e:
                logger.error(f"Error getting count for {collection}: {e}")
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
        """List documents with pagination, search, and filtering using use cases and domain services"""
        logger.info(f"HexagonalContentService.list_documents called for collection: {collection}")
        start_time = time.time()
        
        try:
            # Use SearchQueryService to normalize the query if provided
            normalized_query = query
            if query:
                normalized_query = SearchQueryService.normalize_search_query(query)
            
            # Use FilterService to validate and normalize filters
            normalized_filters = filters
            filter_combination_valid = True
            if filters:
                normalized_filters = FilterService.normalize_filter_values(filters)
                
                # Validate filter combinations
                warnings = FilterService.validate_filter_combination(normalized_filters)
                if warnings:
                    logger.warning(f"Filter validation warnings: {warnings}")
                    filter_combination_valid = False
            
            # Calculate offset
            offset = (page - 1) * page_size
            
            # Route to appropriate use case based on collection type
            logger.info(f"Routing collection '{collection}' to appropriate search method")
            if collection == "incantesimi":
                result = await self._search_spells(normalized_query, normalized_filters, sort_by, page_size, offset)
            elif collection == "mostri":
                result = await self._search_monsters(normalized_query, normalized_filters, sort_by, page_size, offset)
            elif collection == "classi":
                logger.info(f"Searching classi collection with query: {normalized_query}, filters: {normalized_filters}")
                result = await self._search_classes(normalized_query, normalized_filters, sort_by, page_size, offset)
            else:
                logger.info(f"Using generic document search for collection '{collection}'")
                result = await self._search_documents(collection, normalized_query, normalized_filters, sort_by, page_size, offset)
            
            if not result.success:
                logger.error(f"Search failed: {result.message}")
                return [], 0, False, False
            
            documents = result.data or []
            
            # Convert domain objects to dictionaries for template compatibility
            doc_dicts = [self._domain_to_dict(doc) for doc in documents]
            
            # Calculate pagination (simplified for now)
            total = len(doc_dicts)  # This is not ideal - we need proper count queries
            has_prev = page > 1
            has_next = len(doc_dicts) == page_size  # Rough estimation
            
            # Publish events
            search_time_ms = (time.time() - start_time) * 1000
            
            if normalized_query:
                await self.event_bus.publish(SearchPerformedEvent(
                    collection=collection,
                    query=normalized_query,
                    filters=normalized_filters or {},
                    results_count=len(doc_dicts),
                    search_time_ms=search_time_ms
                ))
            
            if normalized_filters:
                await self.event_bus.publish(FilterAppliedEvent(
                    collection=collection,
                    filters=normalized_filters,
                    filter_combination_valid=filter_combination_valid,
                    results_count=len(doc_dicts)
                ))
            
            return doc_dicts, total, has_prev, has_next
            
        except Exception as e:
            logger.error(f"Error in list_documents: {e}")
            return [], 0, False, False
    
    async def get_document(self, collection: str, slug: str, user_agent: str = "", referrer: str = "") -> Optional[Dict[str, Any]]:
        """Get single document by slug using use cases and NavigationService"""
        try:
            # Use NavigationService to normalize the slug for search
            normalized_slug = NavigationService.normalize_slug(slug)
            
            # For now, we need to do a search since our use cases use IDs
            # In a full implementation, we'd have a proper "get by slug" use case
            
            if collection == "incantesimi":
                # Search spells by name
                search_use_case = self.container.get_search_spells_use_case()
                result = await search_use_case.handle(SearchSpellsQuery(
                    text_query=slug,
                    limit=10
                ))
            elif collection == "mostri":
                # Search monsters by name
                search_use_case = self.container.get_search_monsters_use_case()
                result = await search_use_case.handle(SearchMonstersQuery(
                    text_query=slug,
                    limit=10
                ))
            elif collection == "classi":
                # Search classes by name
                result = await self._search_classes(slug, {}, "alpha", 10, 0)
            else:
                # Search documents by name
                search_use_case = self.container.get_search_documents_use_case()
                result = await search_use_case.handle(SearchDocumentsQuery(
                    document_type=collection,
                    text_query=slug,
                    limit=10
                ))
            
            if not result.success or not result.data:
                return None
            
            # Convert to dicts for NavigationService compatibility
            candidate_docs = [self._domain_to_dict(item) for item in result.data]
            
            # Find exact match using NavigationService slug extraction logic
            found_doc = None
            for doc in candidate_docs:
                doc_slug = NavigationService.extract_document_slug(doc)
                if doc_slug and NavigationService.normalize_slug(doc_slug) == normalized_slug:
                    found_doc = doc
                    break
            
            # If no exact match, use first result
            if not found_doc and candidate_docs:
                found_doc = candidate_docs[0]
            
            # Publish document viewed event
            if found_doc:
                document_name = found_doc.get('nome', found_doc.get('name', slug))
                await self.event_bus.publish(DocumentViewedEvent(
                    collection=collection,
                    document_slug=slug,
                    document_name=document_name,
                    user_agent=user_agent,
                    referrer=referrer
                ))
            
            return found_doc
            
        except Exception as e:
            logger.error(f"Error getting document {slug} from {collection}: {e}")
            return None
    
    async def get_filter_options(self, collection: str) -> Dict[str, List[str]]:
        """Get filter options for a collection using FilterService and use cases"""
        try:
            filter_use_case = self.container.get_filter_options_use_case()
            result = await filter_use_case.handle(GetFilterOptionsQuery(
                document_type=collection
            ))
            
            if result.success and result.data:
                # Use FilterService to normalize the filter values
                normalized_filters = FilterService.normalize_filter_values(result.data)
                return normalized_filters
            else:
                logger.error(f"Failed to get filter options: {result.message}")
                return {}
                
        except Exception as e:
            logger.error(f"Error getting filter options for {collection}: {e}")
            return {}
    
    async def render_document_content(self, doc: Dict[str, Any], collection: str = "") -> Tuple[Optional[str], Optional[str]]:
        """Render document content as HTML using ContentRenderingService"""
        start_time = time.time()
        
        content_service = ContentRenderingService()
        result = content_service.render_document_content(doc, markdown_renderer=render_md)
        
        if result:
            # Publish content rendered event
            render_time_ms = (time.time() - start_time) * 1000
            document_slug = NavigationService.extract_document_slug(doc) or ""
            
            await self.event_bus.publish(ContentRenderedEvent(
                collection=collection,
                document_slug=document_slug,
                content_format=result.format_used.value,
                content_length=result.metadata.content_length,
                has_markdown_syntax=result.metadata.has_markdown_syntax,
                rendering_time_ms=render_time_ms
            ))
            
            return result.html_content, result.raw_content
        else:
            return None, None
    
    async def get_navigation_context(
        self, 
        collection: str, 
        current_slug: str,
        filters: Optional[Dict[str, str]] = None,
        from_document: str = ""
    ) -> Tuple[Optional[str], Optional[str]]:
        """Get previous and next document slugs for navigation using NavigationService"""
        try:
            # First get all documents in the collection (filtered)
            documents, _, _, _ = await self.list_documents(
                collection=collection, 
                filters=filters,
                page_size=1000  # Get all for navigation
            )
            
            if not documents:
                return None, None
            
            # Use NavigationService to calculate context
            nav_context = NavigationService.calculate_navigation_context(
                documents=documents,
                current_slug=current_slug,
                collection=collection,
                filters=filters
            )
            
            if nav_context:
                # Publish navigation event if we're actually navigating (not just getting context)
                if from_document and from_document != current_slug:
                    await self.event_bus.publish(NavigationPerformedEvent(
                        collection=collection,
                        from_document=from_document,
                        to_document=current_slug,
                        direction="direct",
                        navigation_context={
                            "current_position": nav_context.current_position,
                            "total_items": nav_context.total_items,
                            "filters_applied": nav_context.filters_applied
                        }
                    ))
                
                return nav_context.previous_slug, nav_context.next_slug
            else:
                return None, None
                
        except Exception as e:
            logger.error(f"Error getting navigation context: {e}")
            return None, None
    
    # Helper methods
    
    def _create_empty_spell_search_query(self):
        """Create empty spell search query for counting"""
        from shared_domain.query_models import SpellSearchQuery
        return SpellSearchQuery()
    
    def _create_empty_monster_search_query(self):
        """Create empty monster search query for counting"""
        from shared_domain.query_models import MonsterSearchQuery
        return MonsterSearchQuery()
    
    def _create_empty_document_search_query(self, document_type: str):
        """Create empty document search query for counting"""
        from shared_domain.query_models import DocumentSearchQuery
        return DocumentSearchQuery(document_type=document_type)
    
    async def _search_spells(self, query: Optional[str], filters: Optional[Dict[str, str]], 
                           sort_by: str, limit: int, offset: int) -> UseCaseResult:
        """Search spells using spell use case"""
        search_use_case = self.container.get_search_spells_use_case()
        
        # Convert filters
        character_class = filters.get("classi") if filters else None
        school = filters.get("scuola") if filters else None
        
        return await search_use_case.handle(SearchSpellsQuery(
            text_query=query,
            character_class=character_class,
            school=school,
            sort_by=sort_by,
            limit=limit,
            offset=offset
        ))
    
    async def _search_monsters(self, query: Optional[str], filters: Optional[Dict[str, str]], 
                             sort_by: str, limit: int, offset: int) -> UseCaseResult:
        """Search monsters using monster use case"""
        search_use_case = self.container.get_search_monsters_use_case()
        
        # Convert filters
        creature_type = filters.get("tipo") if filters else None
        size = filters.get("taglia") if filters else None
        alignment = filters.get("allineamento") if filters else None
        
        return await search_use_case.handle(SearchMonstersQuery(
            text_query=query,
            creature_type=creature_type,
            size=size,
            alignment=alignment,
            sort_by=sort_by,
            limit=limit,
            offset=offset
        ))
    
    async def _search_documents(self, collection: str, query: Optional[str], 
                              filters: Optional[Dict[str, str]], sort_by: str, 
                              limit: int, offset: int) -> UseCaseResult:
        """Search generic documents using document use case"""
        search_use_case = self.container.get_search_documents_use_case()
        
        # Convert filters
        category = filters.get("categoria") if filters else None
        item_type = filters.get("tipo") if filters else None
        rarity = filters.get("rarita") if filters else None
        
        return await search_use_case.handle(SearchDocumentsQuery(
            document_type=collection,
            text_query=query,
            category=category,
            item_type=item_type,
            rarity=rarity,
            filters=filters or {},
            sort_by=sort_by,
            limit=limit,
            offset=offset
        ))
    
    async def _search_classes(self, query: Optional[str], filters: Optional[Dict[str, str]], 
                           sort_by: str, limit: int, offset: int) -> UseCaseResult:
        """Search classes using class use case"""
        logger.info(f"_search_classes called with query='{query}', filters={filters}")
        
        # Convert filters for class search
        primary_ability = filters.get("caratteristica_primaria") if filters else None
        
        # Handle dado_vita as string (e.g., "d12", "d8") not integer
        hit_die = filters.get("dado_vita") if filters else None
        
        spellcaster = None
        if filters and filters.get("ha_incantesimi"):
            spellcaster = filters.get("ha_incantesimi").lower() == "true"
        
        # Use repository directly since the use case pattern isn't fully implemented
        class_repo = self.container.get_class_query_repository()
        from shared_domain.query_models import ClassSearchQuery
        
        class_query = ClassSearchQuery(
            text_query=query,
            primary_ability=primary_ability,
            hit_die=hit_die,  # Pass as string, not int
            is_spellcaster=spellcaster,
            sort_by=sort_by,
            limit=limit,
            offset=offset
        )
        
        # Create a mock UseCaseResult since we're calling the repository directly
        try:
            logger.info(f"About to call class_repo.search_classes with query: {class_query}")
            classes = await class_repo.search_classes(class_query)
            logger.info(f"_search_classes got {len(classes)} classes from repository")
            
            # Mock successful result
            from shared_domain.use_cases import UseCaseResult
            return UseCaseResult(
                success=True,
                data=classes,
                message="Classes retrieved successfully"
            )
        except Exception as e:
            logger.error(f"Error in _search_classes: {e}")
            return UseCaseResult(
                success=False,
                data=[],
                message=f"Error searching classes: {str(e)}"
            )
    
    def _domain_to_dict(self, domain_obj: Any) -> Dict[str, Any]:
        """Convert domain object to dictionary for template compatibility"""
        # Handle raw MongoDB documents (already dictionaries)
        if isinstance(domain_obj, dict):
            result = domain_obj.copy()
            # Convert ObjectId to string for template compatibility
            if '_id' in result:
                result['id'] = str(result['_id'])
            return result
        elif hasattr(domain_obj, 'to_dict'):
            return domain_obj.to_dict()
        elif hasattr(domain_obj, '__dict__'):
            # Convert dataclass or object to dict
            result = {}
            for key, value in domain_obj.__dict__.items():
                if not key.startswith('_'):
                    result[key] = value
            
            # Special handling for ClassSummary objects
            from shared_domain.query_models import ClassSummary
            if isinstance(domain_obj, ClassSummary):
                # Ensure proper field mappings for templates
                result['slug'] = domain_obj.id  # Use id as slug for URL generation
                result['nome'] = domain_obj.name  # Italian name for display
                result['caratteristica_primaria'] = domain_obj.primary_ability
                result['dado_vita'] = domain_obj.hit_die
                # Keep English fields for backward compatibility
                result['name'] = domain_obj.name
                result['primary_ability'] = domain_obj.primary_ability
                result['hit_die'] = domain_obj.hit_die
                return result
            
            # Add compatibility fields for templates
            if hasattr(domain_obj, 'italian_name'):
                result['nome'] = domain_obj.italian_name
            if hasattr(domain_obj, 'name'):
                result['name'] = domain_obj.name
                if 'nome' not in result:
                    result['nome'] = domain_obj.name
            
            # Add slug compatibility - for templates to generate URLs
            if hasattr(domain_obj, 'titolo') and domain_obj.titolo:
                # For DocumentSummary, use titolo as slug for URL generation
                result['slug'] = domain_obj.titolo.lower().replace(' ', '-')
                if 'name' not in result:
                    result['name'] = domain_obj.titolo
                if 'nome' not in result:
                    result['nome'] = domain_obj.titolo
            elif hasattr(domain_obj, 'id') and not result.get('slug'):
                # Fallback: use id as slug if no other slug is available
                result['slug'] = domain_obj.id
            
            return result
        else:
            # Fallback for primitive types
            return {"value": domain_obj}


# Global service instance
_hexagonal_content_service: Optional[HexagonalContentService] = None


async def get_hexagonal_content_service() -> HexagonalContentService:
    """Get hexagonal content service instance"""
    global _hexagonal_content_service
    if not _hexagonal_content_service:
        print("CREATING NEW HEXAGONAL CONTENT SERVICE")
        _hexagonal_content_service = HexagonalContentService()
    print("RETURNING HEXAGONAL CONTENT SERVICE")
    return _hexagonal_content_service