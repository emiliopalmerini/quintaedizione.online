"""
Search Domain Service
Pure business logic for search functionality without infrastructure dependencies
"""
from typing import Dict, Any, Optional, List, Tuple
from dataclasses import dataclass
from enum import Enum
import re


class SearchScope(Enum):
    """Search scope options"""
    NAME_ONLY = "name_only"
    CONTENT_ONLY = "content_only"
    ALL_FIELDS = "all_fields"
    METADATA_ONLY = "metadata_only"


class SortStrategy(Enum):
    """Sorting strategy options"""
    ALPHABETICAL = "alpha"
    RELEVANCE = "relevance"
    DATE = "date"
    CUSTOM = "custom"


@dataclass
class SearchQuery:
    """Normalized search query"""
    original_query: str
    normalized_query: str
    terms: List[str]
    scope: SearchScope
    filters: Dict[str, Any]
    sort_strategy: SortStrategy


@dataclass
class SearchResult:
    """Search result with relevance information"""
    document: Dict[str, Any]
    relevance_score: float
    matched_fields: List[str]
    highlighted_snippets: Dict[str, str]


class SearchQueryService:
    """Domain service for search query processing"""
    
    @staticmethod
    def normalize_search_query(query: str) -> str:
        """Normalize search query for better matching"""
        if not query:
            return ""
        
        # Convert to lowercase
        normalized = query.lower().strip()
        
        # Remove special characters except spaces and basic punctuation
        normalized = re.sub(r'[^\w\s\-\'\"àèéìîíòóùúç]', ' ', normalized)
        
        # Collapse multiple spaces
        normalized = re.sub(r'\s+', ' ', normalized).strip()
        
        return normalized
    
    @staticmethod
    def extract_search_terms(query: str) -> List[str]:
        """Extract individual search terms from query"""
        if not query:
            return []
        
        normalized = SearchQueryService.normalize_search_query(query)
        
        # Split on spaces but preserve quoted phrases
        terms = []
        in_quotes = False
        current_term = ""
        
        words = normalized.split()
        for word in words:
            if word.startswith('"') and word.endswith('"') and len(word) > 1:
                # Complete quoted phrase in single word
                terms.append(word[1:-1])
            elif word.startswith('"'):
                # Start of quoted phrase
                in_quotes = True
                current_term = word[1:]
            elif word.endswith('"') and in_quotes:
                # End of quoted phrase
                current_term += " " + word[:-1]
                terms.append(current_term)
                current_term = ""
                in_quotes = False
            elif in_quotes:
                # Middle of quoted phrase
                current_term += " " + word
            else:
                # Regular word
                terms.append(word)
        
        # Handle unclosed quotes
        if in_quotes and current_term:
            terms.append(current_term)
        
        # Filter out empty terms and stopwords
        stopwords = {"di", "da", "in", "con", "su", "per", "tra", "fra", "a", "il", "la", "lo", "gli", "le", "un", "una", "uno"}
        terms = [term for term in terms if term and term not in stopwords and len(term) >= 2]
        
        return terms
    
    @staticmethod
    def build_search_query(
        query: str,
        scope: SearchScope = SearchScope.ALL_FIELDS,
        filters: Optional[Dict[str, Any]] = None,
        sort_strategy: SortStrategy = SortStrategy.RELEVANCE
    ) -> SearchQuery:
        """Build normalized search query object"""
        normalized = SearchQueryService.normalize_search_query(query)
        terms = SearchQueryService.extract_search_terms(query)
        
        return SearchQuery(
            original_query=query,
            normalized_query=normalized,
            terms=terms,
            scope=scope,
            filters=filters or {},
            sort_strategy=sort_strategy
        )


class SearchRelevanceService:
    """Domain service for calculating search relevance"""
    
    @staticmethod
    def calculate_field_relevance_score(
        field_content: str,
        search_terms: List[str],
        field_weight: float = 1.0
    ) -> float:
        """Calculate relevance score for a specific field"""
        if not field_content or not search_terms:
            return 0.0
        
        content_lower = field_content.lower()
        total_score = 0.0
        
        for term in search_terms:
            term_lower = term.lower()
            
            # Exact match bonus
            if term_lower == content_lower:
                total_score += 100.0 * field_weight
                continue
            
            # Full term matches
            full_matches = len(re.findall(r'\b' + re.escape(term_lower) + r'\b', content_lower))
            total_score += full_matches * 10.0 * field_weight
            
            # Partial matches (less weight)
            partial_matches = content_lower.count(term_lower) - full_matches
            total_score += partial_matches * 2.0 * field_weight
            
            # Position bonus (earlier matches score higher)
            first_position = content_lower.find(term_lower)
            if first_position >= 0:
                position_bonus = max(0, 5.0 - (first_position / 100.0))
                total_score += position_bonus * field_weight
        
        return total_score
    
    @staticmethod
    def calculate_document_relevance(
        document: Dict[str, Any],
        search_query: SearchQuery
    ) -> Tuple[float, List[str]]:
        """Calculate overall document relevance score"""
        if not search_query.terms:
            return 0.0, []
        
        total_score = 0.0
        matched_fields = []
        
        # Define field weights based on importance
        field_weights = {
            # Names have highest weight
            "name": 5.0,
            "nome": 5.0,
            "title": 4.0,
            "titolo": 4.0,
            # Content fields have medium weight
            "description": 2.0,
            "descrizione": 2.0,
            "content": 2.0,
            "description_md": 2.0,
            "descrizione_md": 2.0,
            # Metadata has lower weight
            "categoria": 1.0,
            "tipo": 1.0,
            "scuola": 1.5,
            "classi": 1.5,
        }
        
        # Calculate score for each field
        for field_name, weight in field_weights.items():
            if field_name in document and document[field_name]:
                field_content = str(document[field_name])
                field_score = SearchRelevanceService.calculate_field_relevance_score(
                    field_content, search_query.terms, weight
                )
                
                if field_score > 0:
                    total_score += field_score
                    matched_fields.append(field_name)
        
        return total_score, matched_fields
    
    @staticmethod
    def generate_highlighted_snippets(
        document: Dict[str, Any],
        search_terms: List[str],
        matched_fields: List[str],
        snippet_length: int = 150
    ) -> Dict[str, str]:
        """Generate highlighted snippets for matched fields"""
        snippets = {}
        
        for field_name in matched_fields[:3]:  # Limit to top 3 fields
            if field_name not in document:
                continue
                
            field_content = str(document[field_name])
            if len(field_content) <= snippet_length:
                snippets[field_name] = field_content
                continue
            
            # Find best snippet position
            best_position = 0
            best_score = 0
            
            for term in search_terms:
                term_lower = term.lower()
                position = field_content.lower().find(term_lower)
                if position >= 0:
                    # Score based on how many terms are near this position
                    snippet_start = max(0, position - snippet_length // 2)
                    snippet_end = min(len(field_content), snippet_start + snippet_length)
                    snippet = field_content[snippet_start:snippet_end]
                    
                    score = sum(1 for t in search_terms if t.lower() in snippet.lower())
                    if score > best_score:
                        best_score = score
                        best_position = snippet_start
            
            # Extract snippet
            snippet_start = best_position
            snippet_end = min(len(field_content), snippet_start + snippet_length)
            snippet = field_content[snippet_start:snippet_end]
            
            # Add ellipsis if truncated
            if snippet_start > 0:
                snippet = "..." + snippet
            if snippet_end < len(field_content):
                snippet = snippet + "..."
            
            snippets[field_name] = snippet
        
        return snippets


class FilterService:
    """Domain service for search filters processing"""
    
    @staticmethod
    def normalize_filter_values(filters: Dict[str, Any]) -> Dict[str, Any]:
        """Normalize filter values for consistent processing"""
        normalized = {}
        
        for key, value in filters.items():
            if value is None or value == "":
                continue
            
            if isinstance(value, str):
                # Trim and normalize string values
                normalized_value = value.strip()
                if normalized_value:
                    normalized[key] = normalized_value
            elif isinstance(value, list):
                # Filter out empty values from lists
                normalized_list = [v for v in value if v is not None and v != ""]
                if normalized_list:
                    normalized[key] = normalized_list
            else:
                normalized[key] = value
        
        return normalized
    
    @staticmethod
    def validate_filter_combination(filters: Dict[str, Any]) -> List[str]:
        """Validate filter combinations and return warnings"""
        warnings = []
        
        # Check for conflicting numeric filters
        numeric_conflicts = [
            ("min_level", "max_level"),
            ("min_armor_class", "max_armor_class"),
            ("min_challenge_rating", "max_challenge_rating")
        ]
        
        for min_field, max_field in numeric_conflicts:
            if min_field in filters and max_field in filters:
                try:
                    min_val = float(filters[min_field])
                    max_val = float(filters[max_field])
                    if min_val > max_val:
                        warnings.append(f"Minimum {min_field} cannot be greater than maximum {max_field}")
                except (ValueError, TypeError):
                    warnings.append(f"Invalid numeric values for {min_field}/{max_field}")
        
        return warnings
    
    @staticmethod
    def get_available_filter_values(
        documents: List[Dict[str, Any]],
        filter_field: str
    ) -> List[str]:
        """Extract available values for a filter field from documents"""
        values = set()
        
        for doc in documents:
            if filter_field in doc:
                value = doc[filter_field]
                if isinstance(value, list):
                    values.update(str(v) for v in value if v is not None)
                elif value is not None:
                    values.add(str(value))
        
        return sorted(list(values))