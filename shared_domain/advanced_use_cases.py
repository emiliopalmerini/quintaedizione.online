"""
Advanced Use Cases for Navigation and Content Discovery
Sophisticated business operations that combine multiple domain services
"""
from typing import Dict, List, Any, Optional, Tuple
from dataclasses import dataclass
import logging

from .use_cases import UseCaseResult
from .query_models import SpellSearchQuery, MonsterSearchQuery, DocumentSearchQuery

logger = logging.getLogger(__name__)


@dataclass
class NavigationQuery:
    """Query for advanced navigation operations"""
    collection: str
    current_document_slug: str
    filters: Dict[str, Any] = None
    include_breadcrumbs: bool = True
    include_related: bool = True
    context_size: int = 5


@dataclass 
class ContentDiscoveryQuery:
    """Query for content discovery and recommendations"""
    collection: str
    reference_document_slug: str = ""
    user_preferences: Dict[str, Any] = None
    discovery_strategy: str = "similar"  # "similar", "related", "popular", "recent"
    max_results: int = 10


@dataclass
class SearchSuggestionQuery:
    """Query for search suggestions and autocomplete"""
    partial_query: str
    collection: str = ""
    max_suggestions: int = 10
    include_facets: bool = True


@dataclass
class NavigationContext:
    """Rich navigation context with multiple dimensions"""
    current_document: Dict[str, Any]
    previous_document: Optional[Dict[str, Any]] = None
    next_document: Optional[Dict[str, Any]] = None
    breadcrumbs: List[Dict[str, Any]] = None
    related_documents: List[Dict[str, Any]] = None
    position_info: Dict[str, Any] = None
    
    def __post_init__(self):
        if self.breadcrumbs is None:
            self.breadcrumbs = []
        if self.related_documents is None:
            self.related_documents = []
        if self.position_info is None:
            self.position_info = {}


@dataclass
class ContentDiscovery:
    """Content discovery results with recommendations"""
    recommended_documents: List[Dict[str, Any]]
    discovery_reason: str
    confidence_score: float
    facets: Dict[str, List[str]] = None
    
    def __post_init__(self):
        if self.facets is None:
            self.facets = {}


@dataclass
class SearchSuggestions:
    """Search suggestions with metadata"""
    query_suggestions: List[str]
    document_suggestions: List[Dict[str, Any]]
    facet_suggestions: Dict[str, List[str]]
    correction: Optional[str] = None


class AdvancedNavigationUseCase:
    """Advanced navigation with rich context"""
    
    def __init__(self, spell_repo, monster_repo, document_repo):
        self.spell_repo = spell_repo
        self.monster_repo = monster_repo
        self.document_repo = document_repo
    
    async def handle(self, query: NavigationQuery) -> UseCaseResult[NavigationContext]:
        """Handle advanced navigation request"""
        try:
            # Get current document
            current_doc = await self._get_document(query.collection, query.current_document_slug)
            if not current_doc:
                return UseCaseResult(
                    success=False,
                    message=f"Document not found: {query.current_document_slug}",
                    data=None
                )
            
            # Get all documents in collection for navigation
            all_docs = await self._get_filtered_documents(query.collection, query.filters or {})
            
            # Find current position
            current_position = self._find_document_position(all_docs, query.current_document_slug)
            if current_position is None:
                return UseCaseResult(
                    success=False,
                    message=f"Could not determine position for document: {query.current_document_slug}",
                    data=None
                )
            
            # Build navigation context
            nav_context = NavigationContext(current_document=current_doc)
            
            # Add previous/next documents
            if current_position > 0:
                nav_context.previous_document = all_docs[current_position - 1]
            if current_position < len(all_docs) - 1:
                nav_context.next_document = all_docs[current_position + 1]
            
            # Add breadcrumbs if requested
            if query.include_breadcrumbs:
                nav_context.breadcrumbs = await self._build_breadcrumbs(
                    query.collection, current_doc, query.filters or {}
                )
            
            # Add related documents if requested
            if query.include_related:
                nav_context.related_documents = await self._find_related_documents(
                    query.collection, current_doc, query.context_size
                )
            
            # Add position information
            nav_context.position_info = {
                "current_position": current_position + 1,  # 1-based
                "total_documents": len(all_docs),
                "percentage": round((current_position + 1) / len(all_docs) * 100, 1) if all_docs else 0
            }
            
            return UseCaseResult(
                success=True,
                message="Navigation context built successfully",
                data=nav_context
            )
            
        except Exception as e:
            logger.error(f"Error in advanced navigation: {e}")
            return UseCaseResult(
                success=False,
                message=f"Navigation failed: {str(e)}",
                data=None
            )
    
    async def _get_document(self, collection: str, slug: str) -> Optional[Dict[str, Any]]:
        """Get document by slug from appropriate repository"""
        if collection == "incantesimi":
            results = await self.spell_repo.search_spells(SpellSearchQuery(text_query=slug, limit=10))
            for spell in results:
                if getattr(spell, 'name', '') == slug or getattr(spell, 'italian_name', '') == slug:
                    return self._spell_to_dict(spell)
        elif collection == "mostri":
            results = await self.monster_repo.search_monsters(MonsterSearchQuery(text_query=slug, limit=10))
            for monster in results:
                if getattr(monster, 'name', '') == slug or getattr(monster, 'italian_name', '') == slug:
                    return self._monster_to_dict(monster)
        else:
            results = await self.document_repo.search_documents(DocumentSearchQuery(
                document_type=collection, text_query=slug, limit=10
            ))
            for doc in results:
                if getattr(doc, 'name', '') == slug:
                    return self._document_to_dict(doc)
        
        return None
    
    async def _get_filtered_documents(self, collection: str, filters: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Get all documents in collection with filters applied"""
        results = []
        
        if collection == "incantesimi":
            query = SpellSearchQuery(
                character_class=filters.get("classi"),
                school=filters.get("scuola"),
                limit=1000
            )
            spells = await self.spell_repo.search_spells(query)
            results = [self._spell_to_dict(spell) for spell in spells]
            
        elif collection == "mostri":
            query = MonsterSearchQuery(
                creature_type=filters.get("tipo"),
                size=filters.get("taglia"),
                alignment=filters.get("allineamento"),
                limit=1000
            )
            monsters = await self.monster_repo.search_monsters(query)
            results = [self._monster_to_dict(monster) for monster in monsters]
            
        else:
            query = DocumentSearchQuery(
                document_type=collection,
                category=filters.get("categoria"),
                item_type=filters.get("tipo"),
                rarity=filters.get("rarita"),
                limit=1000
            )
            docs = await self.document_repo.search_documents(query)
            results = [self._document_to_dict(doc) for doc in docs]
        
        return results
    
    def _find_document_position(self, documents: List[Dict[str, Any]], slug: str) -> Optional[int]:
        """Find position of document with matching slug"""
        for i, doc in enumerate(documents):
            doc_slug = doc.get('nome', doc.get('name', ''))
            if doc_slug and doc_slug == slug:
                return i
        return None
    
    async def _build_breadcrumbs(self, collection: str, current_doc: Dict[str, Any], filters: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Build breadcrumb navigation"""
        breadcrumbs = [
            {"name": "Home", "url": "/", "active": False},
            {"name": collection.title(), "url": f"/{collection}", "active": False}
        ]
        
        # Add filter-based breadcrumbs
        for filter_key, filter_value in filters.items():
            if filter_value:
                breadcrumbs.append({
                    "name": f"{filter_key.title()}: {filter_value}",
                    "url": f"/{collection}?{filter_key}={filter_value}",
                    "active": False
                })
        
        # Add current document
        doc_name = current_doc.get('nome', current_doc.get('name', 'Document'))
        breadcrumbs.append({
            "name": doc_name,
            "url": f"/{collection}/{doc_name}",
            "active": True
        })
        
        return breadcrumbs
    
    async def _find_related_documents(self, collection: str, current_doc: Dict[str, Any], context_size: int) -> List[Dict[str, Any]]:
        """Find documents related to current document"""
        related = []
        
        # Simple related document logic based on shared attributes
        if collection == "incantesimi":
            # Find spells of same school or class
            school = current_doc.get('scuola', '')
            character_class = current_doc.get('classi', [])
            
            if school:
                query = SpellSearchQuery(school=school, limit=context_size)
                spells = await self.spell_repo.search_spells(query)
                related.extend([self._spell_to_dict(spell) for spell in spells[:context_size]])
                
        elif collection == "mostri":
            # Find monsters of same type or CR
            creature_type = current_doc.get('tipo', '')
            if creature_type:
                query = MonsterSearchQuery(creature_type=creature_type, limit=context_size)
                monsters = await self.monster_repo.search_monsters(query)
                related.extend([self._monster_to_dict(monster) for monster in monsters[:context_size]])
        
        # Remove current document from related list
        current_slug = current_doc.get('nome', current_doc.get('name', ''))
        related = [doc for doc in related if doc.get('nome', doc.get('name', '')) != current_slug]
        
        return related[:context_size]
    
    def _spell_to_dict(self, spell) -> Dict[str, Any]:
        """Convert spell domain object to dict"""
        if hasattr(spell, 'to_dict'):
            return spell.to_dict()
        return {
            "nome": getattr(spell, 'italian_name', getattr(spell, 'name', '')),
            "name": getattr(spell, 'name', ''),
            "scuola": getattr(spell, 'school', ''),
            "classi": getattr(spell, 'character_classes', [])
        }
    
    def _monster_to_dict(self, monster) -> Dict[str, Any]:
        """Convert monster domain object to dict"""
        if hasattr(monster, 'to_dict'):
            return monster.to_dict()
        return {
            "nome": getattr(monster, 'italian_name', getattr(monster, 'name', '')),
            "name": getattr(monster, 'name', ''),
            "tipo": getattr(monster, 'creature_type', ''),
            "taglia": getattr(monster, 'size', '')
        }
    
    def _document_to_dict(self, doc) -> Dict[str, Any]:
        """Convert document domain object to dict"""
        if hasattr(doc, 'to_dict'):
            return doc.to_dict()
        return {
            "nome": getattr(doc, 'name', ''),
            "name": getattr(doc, 'name', ''),
            "categoria": getattr(doc, 'category', ''),
            "tipo": getattr(doc, 'item_type', '')
        }


class ContentDiscoveryUseCase:
    """Content discovery and recommendation engine"""
    
    def __init__(self, spell_repo, monster_repo, document_repo):
        self.spell_repo = spell_repo
        self.monster_repo = monster_repo
        self.document_repo = document_repo
    
    async def handle(self, query: ContentDiscoveryQuery) -> UseCaseResult[ContentDiscovery]:
        """Handle content discovery request"""
        try:
            recommendations = []
            discovery_reason = ""
            confidence_score = 0.0
            
            if query.discovery_strategy == "similar":
                recommendations, confidence_score = await self._find_similar_content(query)
                discovery_reason = "Based on content similarity"
            elif query.discovery_strategy == "related":
                recommendations, confidence_score = await self._find_related_content(query)
                discovery_reason = "Based on related topics"
            elif query.discovery_strategy == "popular":
                recommendations, confidence_score = await self._find_popular_content(query)
                discovery_reason = "Based on popularity"
            elif query.discovery_strategy == "recent":
                recommendations, confidence_score = await self._find_recent_content(query)
                discovery_reason = "Based on recent activity"
            
            # Generate facets
            facets = await self._generate_facets(query.collection, recommendations)
            
            discovery = ContentDiscovery(
                recommended_documents=recommendations,
                discovery_reason=discovery_reason,
                confidence_score=confidence_score,
                facets=facets
            )
            
            return UseCaseResult(
                success=True,
                message=f"Found {len(recommendations)} recommendations",
                data=discovery
            )
            
        except Exception as e:
            logger.error(f"Error in content discovery: {e}")
            return UseCaseResult(
                success=False,
                message=f"Content discovery failed: {str(e)}",
                data=None
            )
    
    async def _find_similar_content(self, query: ContentDiscoveryQuery) -> Tuple[List[Dict[str, Any]], float]:
        """Find content similar to reference document"""
        # Simplified similarity - in real implementation would use embeddings/ML
        results = []
        confidence = 0.7
        
        if query.collection == "incantesimi":
            # Find spells with similar characteristics
            search_query = SpellSearchQuery(limit=query.max_results)
            spells = await self.spell_repo.search_spells(search_query)
            results = [self._spell_to_dict(spell) for spell in spells[:query.max_results]]
        elif query.collection == "mostri":
            search_query = MonsterSearchQuery(limit=query.max_results)
            monsters = await self.monster_repo.search_monsters(search_query)
            results = [self._monster_to_dict(monster) for monster in monsters[:query.max_results]]
        else:
            search_query = DocumentSearchQuery(document_type=query.collection, limit=query.max_results)
            docs = await self.document_repo.search_documents(search_query)
            results = [self._document_to_dict(doc) for doc in docs[:query.max_results]]
        
        return results, confidence
    
    async def _find_related_content(self, query: ContentDiscoveryQuery) -> Tuple[List[Dict[str, Any]], float]:
        """Find content related to reference document"""
        # Placeholder implementation
        return await self._find_similar_content(query)
    
    async def _find_popular_content(self, query: ContentDiscoveryQuery) -> Tuple[List[Dict[str, Any]], float]:
        """Find popular content in collection"""
        # Would use analytics data in real implementation
        return await self._find_similar_content(query)
    
    async def _find_recent_content(self, query: ContentDiscoveryQuery) -> Tuple[List[Dict[str, Any]], float]:
        """Find recently accessed content"""
        # Would use activity logs in real implementation
        return await self._find_similar_content(query)
    
    async def _generate_facets(self, collection: str, documents: List[Dict[str, Any]]) -> Dict[str, List[str]]:
        """Generate facets from document set"""
        facets = {}
        
        for doc in documents:
            # Extract facet values from documents
            if collection == "incantesimi":
                school = doc.get('scuola', '')
                if school:
                    facets.setdefault('scuola', set()).add(school)
                
                classes = doc.get('classi', [])
                for cls in classes if isinstance(classes, list) else []:
                    facets.setdefault('classi', set()).add(cls)
                    
            elif collection == "mostri":
                creature_type = doc.get('tipo', '')
                if creature_type:
                    facets.setdefault('tipo', set()).add(creature_type)
                    
                size = doc.get('taglia', '')
                if size:
                    facets.setdefault('taglia', set()).add(size)
            
        # Convert sets to sorted lists
        return {key: sorted(list(values)) for key, values in facets.items()}
    
    def _spell_to_dict(self, spell) -> Dict[str, Any]:
        """Convert spell to dict"""
        if hasattr(spell, 'to_dict'):
            return spell.to_dict()
        return {"nome": getattr(spell, 'italian_name', ''), "scuola": getattr(spell, 'school', '')}
    
    def _monster_to_dict(self, monster) -> Dict[str, Any]:
        """Convert monster to dict"""
        if hasattr(monster, 'to_dict'):
            return monster.to_dict()
        return {"nome": getattr(monster, 'italian_name', ''), "tipo": getattr(monster, 'creature_type', '')}
    
    def _document_to_dict(self, doc) -> Dict[str, Any]:
        """Convert document to dict"""
        if hasattr(doc, 'to_dict'):
            return doc.to_dict()
        return {"nome": getattr(doc, 'name', '')}


class SearchSuggestionUseCase:
    """Search suggestions and autocomplete"""
    
    def __init__(self, spell_repo, monster_repo, document_repo):
        self.spell_repo = spell_repo
        self.monster_repo = monster_repo
        self.document_repo = document_repo
        self._suggestion_cache = {}
    
    async def handle(self, query: SearchSuggestionQuery) -> UseCaseResult[SearchSuggestions]:
        """Handle search suggestion request"""
        try:
            # Generate query suggestions
            query_suggestions = await self._generate_query_suggestions(query.partial_query)
            
            # Find matching documents
            document_suggestions = await self._find_document_suggestions(query)
            
            # Generate facet suggestions
            facet_suggestions = await self._generate_facet_suggestions(query)
            
            # Check for spelling correction
            correction = await self._suggest_correction(query.partial_query)
            
            suggestions = SearchSuggestions(
                query_suggestions=query_suggestions,
                document_suggestions=document_suggestions,
                facet_suggestions=facet_suggestions,
                correction=correction
            )
            
            return UseCaseResult(
                success=True,
                message=f"Generated {len(query_suggestions)} suggestions",
                data=suggestions
            )
            
        except Exception as e:
            logger.error(f"Error generating search suggestions: {e}")
            return UseCaseResult(
                success=False,
                message=f"Suggestion generation failed: {str(e)}",
                data=None
            )
    
    async def _generate_query_suggestions(self, partial: str) -> List[str]:
        """Generate query completion suggestions"""
        suggestions = []
        
        # Common D&D terms that start with the partial query
        common_terms = [
            "incantesimo", "incantatore", "iniziativa", "intelligenza",
            "mostro", "magia", "mago", "monaco", "movimento",
            "classe", "chierico", "combattimento", "carisma", "costituzione",
            "dado", "danno", "difesa", "destrezza", "druido"
        ]
        
        partial_lower = partial.lower()
        for term in common_terms:
            if term.startswith(partial_lower) and len(partial) >= 2:
                suggestions.append(term)
        
        return suggestions[:10]
    
    async def _find_document_suggestions(self, query: SearchSuggestionQuery) -> List[Dict[str, Any]]:
        """Find documents matching partial query"""
        results = []
        
        if not query.partial_query or len(query.partial_query) < 2:
            return results
        
        try:
            if not query.collection or query.collection == "incantesimi":
                spells = await self.spell_repo.search_spells(SpellSearchQuery(
                    text_query=query.partial_query,
                    limit=5
                ))
                for spell in spells:
                    results.append({
                        "nome": getattr(spell, 'italian_name', getattr(spell, 'name', '')),
                        "collection": "incantesimi",
                        "type": "spell"
                    })
            
            if not query.collection or query.collection == "mostri":
                monsters = await self.monster_repo.search_monsters(MonsterSearchQuery(
                    text_query=query.partial_query,
                    limit=5
                ))
                for monster in monsters:
                    results.append({
                        "nome": getattr(monster, 'italian_name', getattr(monster, 'name', '')),
                        "collection": "mostri",
                        "type": "monster"
                    })
                    
        except Exception as e:
            logger.warning(f"Error finding document suggestions: {e}")
        
        return results[:query.max_suggestions]
    
    async def _generate_facet_suggestions(self, query: SearchSuggestionQuery) -> Dict[str, List[str]]:
        """Generate facet-based suggestions"""
        facets = {}
        
        if query.collection == "incantesimi":
            facets = {
                "scuola": ["abiurazione", "ammagliamento", "divinazione", "evocazione"],
                "classi": ["mago", "chierico", "druido", "stregone"]
            }
        elif query.collection == "mostri":
            facets = {
                "tipo": ["aberrazione", "bestia", "celestiale", "demone"],
                "taglia": ["piccola", "media", "grande", "enorme"]
            }
        
        return facets
    
    async def _suggest_correction(self, query: str) -> Optional[str]:
        """Suggest spelling correction for query"""
        # Simple correction mapping
        corrections = {
            "incanteismo": "incantesimo",
            "mostros": "mostri",
            "classe": "classi"
        }
        
        query_lower = query.lower()
        return corrections.get(query_lower)