"""Unit tests for request validation models."""
from __future__ import annotations

import pytest
from pydantic import ValidationError

from models.request_models import (
    PaginationParams,
    LanguageParams,
    SearchQuery,
    FilterParams,
    ListPageParams,
    ShowPageParams,
    EditPageParams,
    DocumentUpdateData,
    CollectionParams,
)


class TestPaginationParams:
    """Test PaginationParams model."""
    
    def test_valid_pagination(self):
        """Test valid pagination parameters."""
        params = PaginationParams(page=1, page_size=20)
        assert params.page == 1
        assert params.page_size == 20
    
    def test_default_values(self):
        """Test default pagination values."""
        params = PaginationParams()
        assert params.page == 1
        assert params.page_size == 20
    
    def test_invalid_page(self):
        """Test invalid page numbers."""
        with pytest.raises(ValidationError) as exc_info:
            PaginationParams(page=0)
        assert "ensure this value is greater than or equal to 1" in str(exc_info.value)
        
        with pytest.raises(ValidationError) as exc_info:
            PaginationParams(page=10001)
        assert "ensure this value is less than or equal to 10000" in str(exc_info.value)
    
    def test_invalid_page_size(self):
        """Test invalid page sizes."""
        with pytest.raises(ValidationError) as exc_info:
            PaginationParams(page_size=0)
        assert "ensure this value is greater than or equal to 1" in str(exc_info.value)
        
        with pytest.raises(ValidationError) as exc_info:
            PaginationParams(page_size=101)
        assert "ensure this value is less than or equal to 100" in str(exc_info.value)
    
    def test_extra_fields_forbidden(self):
        """Test that extra fields are forbidden."""
        with pytest.raises(ValidationError) as exc_info:
            PaginationParams(page=1, page_size=20, extra_field="value")
        assert "extra fields not permitted" in str(exc_info.value)


class TestLanguageParams:
    """Test LanguageParams model."""
    
    def test_valid_languages(self):
        """Test valid language parameters."""
        params_it = LanguageParams(lang="it")
        assert params_it.lang == "it"
        
        params_en = LanguageParams(lang="en")
        assert params_en.lang == "en"
    
    def test_default_language(self):
        """Test default language."""
        params = LanguageParams()
        assert params.lang == "it"
    
    def test_invalid_language(self):
        """Test invalid language."""
        with pytest.raises(ValidationError) as exc_info:
            LanguageParams(lang="fr")
        assert "unexpected value" in str(exc_info.value)


class TestSearchQuery:
    """Test SearchQuery model."""
    
    def test_valid_query(self):
        """Test valid search query."""
        params = SearchQuery(q="fireball", collection="spells")
        assert params.q == "fireball"
        assert params.collection == "spells"
    
    def test_empty_query(self):
        """Test empty query handling."""
        params = SearchQuery(q="", collection="spells")
        assert params.q == ""
        
        params = SearchQuery(q=None, collection="spells")
        assert params.q is None
    
    def test_query_validation(self):
        """Test query string validation."""
        # Test dangerous characters
        with pytest.raises(ValidationError) as exc_info:
            SearchQuery(q="test{}")
        assert "invalid characters" in str(exc_info.value)
        
        with pytest.raises(ValidationError) as exc_info:
            SearchQuery(q="test$")
        assert "invalid characters" in str(exc_info.value)
    
    def test_query_length_validation(self):
        """Test query length limits."""
        # Too long query
        long_query = "a" * 201
        with pytest.raises(ValidationError) as exc_info:
            SearchQuery(q=long_query)
        assert "ensure this value has at most 200 characters" in str(exc_info.value)
    
    def test_collection_validation(self):
        """Test collection name validation."""
        # Valid collection
        params = SearchQuery(collection="valid_collection")
        assert params.collection == "valid_collection"
        
        # Invalid collection format
        with pytest.raises(ValidationError) as exc_info:
            SearchQuery(collection="123invalid")
        assert "Invalid collection name format" in str(exc_info.value)
        
        with pytest.raises(ValidationError) as exc_info:
            SearchQuery(collection="invalid-collection")
        assert "Invalid collection name format" in str(exc_info.value)
    
    def test_whitespace_handling(self):
        """Test whitespace handling in queries."""
        params = SearchQuery(q="  trimmed query  ")
        assert params.q == "trimmed query"
        
        # Only whitespace should be invalid
        with pytest.raises(ValidationError) as exc_info:
            SearchQuery(q="   ")
        assert "cannot be empty after cleaning" in str(exc_info.value)


class TestFilterParams:
    """Test FilterParams model."""
    
    def test_valid_filters(self):
        """Test valid filter parameters."""
        params = FilterParams(
            level=3,
            school="Evocation",
            class_name="Wizard",
            ritual=True,
            concentration=False,
            translated=True,
            modified=False
        )
        
        assert params.level == 3
        assert params.school == "Evocation"
        assert params.class_name == "Wizard"
        assert params.ritual is True
        assert params.concentration is False
        assert params.translated is True
        assert params.modified is False
    
    def test_level_validation(self):
        """Test spell level validation."""
        # Valid levels
        for level in range(10):
            params = FilterParams(level=level)
            assert params.level == level
        
        # Invalid levels
        with pytest.raises(ValidationError) as exc_info:
            FilterParams(level=-1)
        assert "ensure this value is greater than or equal to 0" in str(exc_info.value)
        
        with pytest.raises(ValidationError) as exc_info:
            FilterParams(level=10)
        assert "ensure this value is less than or equal to 9" in str(exc_info.value)
    
    def test_string_field_validation(self):
        """Test string field validation."""
        # Valid strings
        params = FilterParams(school="Evocation", class_name="Death Knight")
        assert params.school == "Evocation"
        assert params.class_name == "Death Knight"
        
        # Invalid characters
        with pytest.raises(ValidationError) as exc_info:
            FilterParams(school="Test123")
        assert "invalid characters" in str(exc_info.value)
        
        with pytest.raises(ValidationError) as exc_info:
            FilterParams(class_name="Test@Class")
        assert "invalid characters" in str(exc_info.value)
    
    def test_string_field_length(self):
        """Test string field length limits."""
        long_string = "a" * 51
        
        with pytest.raises(ValidationError) as exc_info:
            FilterParams(school=long_string)
        assert "ensure this value has at most 50 characters" in str(exc_info.value)


class TestListPageParams:
    """Test ListPageParams combined model."""
    
    def test_valid_combined_params(self):
        """Test valid combined parameters."""
        params = ListPageParams(
            page=2,
            page_size=50,
            lang="en",
            q="magic missile",
            level=1,
            sort="level"
        )
        
        assert params.page == 2
        assert params.page_size == 50
        assert params.lang == "en"
        assert params.q == "magic missile"
        assert params.level == 1
        assert params.sort == "level"
    
    def test_sort_validation(self):
        """Test sort parameter validation."""
        valid_sorts = ["alpha", "level", "school", "modified"]
        
        for sort_value in valid_sorts:
            params = ListPageParams(sort=sort_value)
            assert params.sort == sort_value
        
        # Invalid sort
        with pytest.raises(ValidationError) as exc_info:
            ListPageParams(sort="invalid_sort")
        assert "unexpected value" in str(exc_info.value)
    
    def test_default_sort(self):
        """Test default sort value."""
        params = ListPageParams()
        assert params.sort == "alpha"


class TestShowPageParams:
    """Test ShowPageParams model."""
    
    def test_valid_slug(self):
        """Test valid slug parameters."""
        params = ShowPageParams(slug="magic-missile", lang="it")
        assert params.slug == "magic-missile"
        assert params.lang == "it"
    
    def test_slug_validation(self):
        """Test slug format validation."""
        # Valid slugs
        valid_slugs = ["simple", "hyphen-separated", "under_score", "mixed-under_score123"]
        
        for slug in valid_slugs:
            params = ShowPageParams(slug=slug)
            assert params.slug == slug
        
        # Invalid slugs
        invalid_slugs = ["spaces not allowed", "special@chars", "dots.not.allowed"]
        
        for slug in invalid_slugs:
            with pytest.raises(ValidationError) as exc_info:
                ShowPageParams(slug=slug)
            assert "Invalid slug format" in str(exc_info.value)
    
    def test_slug_length(self):
        """Test slug length validation."""
        # Too short
        with pytest.raises(ValidationError) as exc_info:
            ShowPageParams(slug="")
        assert "ensure this value has at least 1 characters" in str(exc_info.value)
        
        # Too long
        long_slug = "a" * 201
        with pytest.raises(ValidationError) as exc_info:
            ShowPageParams(slug=long_slug)
        assert "ensure this value has at most 200 characters" in str(exc_info.value)


class TestDocumentUpdateData:
    """Test DocumentUpdateData model."""
    
    def test_valid_update_data(self):
        """Test valid document update."""
        data = DocumentUpdateData(
            title="Updated Title",
            content="Updated content",
            level=5,
            school="Transmutation"
        )
        
        assert data.title == "Updated Title"
        assert data.content == "Updated content"
        assert data.level == 5
        assert data.school == "Transmutation"
    
    def test_partial_updates(self):
        """Test partial document updates."""
        # Only title
        data = DocumentUpdateData(title="Just Title")
        assert data.title == "Just Title"
        assert data.content is None
        
        # Only spell fields
        data = DocumentUpdateData(level=3, ritual=True)
        assert data.level == 3
        assert data.ritual is True
        assert data.title is None
    
    def test_string_list_validation(self):
        """Test validation of string lists."""
        # Valid lists
        data = DocumentUpdateData(classes=["Wizard", "Sorcerer"])
        assert data.classes == ["Wizard", "Sorcerer"]
        
        data = DocumentUpdateData(tags=["damage", "single-target"])
        assert data.tags == ["damage", "single-target"]
        
        # Empty strings filtered out
        data = DocumentUpdateData(classes=["Wizard", "", "  ", "Sorcerer"])
        assert data.classes == ["Wizard", "Sorcerer"]
        
        # Non-strings rejected
        with pytest.raises(ValidationError) as exc_info:
            DocumentUpdateData(classes=["Wizard", 123])
        assert "must be strings" in str(exc_info.value)
    
    def test_at_least_one_field_validation(self):
        """Test that at least one field must be provided."""
        with pytest.raises(ValidationError) as exc_info:
            DocumentUpdateData()
        assert "At least one field must be provided" in str(exc_info.value)
    
    def test_text_field_trimming(self):
        """Test that text fields are trimmed."""
        data = DocumentUpdateData(title="  Trimmed Title  ")
        assert data.title == "Trimmed Title"


class TestCollectionParams:
    """Test CollectionParams model."""
    
    def test_valid_collection(self):
        """Test valid collection validation."""
        # This test requires mocking COLLECTIONS
        with pytest.MonkeyPatch().context() as m:
            m.setattr("models.request_models.COLLECTIONS", ["spells", "items"])
            
            params = CollectionParams(collection="spells", lang="it")
            assert params.collection == "spells"
    
    def test_invalid_collection(self):
        """Test invalid collection validation."""
        with pytest.MonkeyPatch().context() as m:
            m.setattr("models.request_models.COLLECTIONS", ["spells", "items"])
            
            with pytest.raises(ValidationError) as exc_info:
                CollectionParams(collection="invalid", lang="it")
            assert "Unknown collection" in str(exc_info.value)
            assert "Available:" in str(exc_info.value)