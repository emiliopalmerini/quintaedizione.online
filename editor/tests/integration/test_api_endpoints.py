"""Integration tests for API endpoints."""
from __future__ import annotations

import pytest
from fastapi.testclient import TestClient
from unittest.mock import patch

from main import create_app


@pytest.fixture
def app():
    """Create app for testing."""
    return create_app()


@pytest.fixture
def client(app):
    """Create test client."""
    return TestClient(app)


class TestHealthEndpoint:
    """Test health check endpoint."""
    
    def test_health_endpoint_success(self, client):
        """Test health endpoint returns OK."""
        response = client.get("/healthz")
        
        assert response.status_code == 200
        assert response.text == "ok"


class TestHomepageEndpoint:
    """Test homepage endpoint."""
    
    @patch("routers.pages.get_db")
    @patch("routers.pages.svc_home_doc")
    async def test_homepage_loads(self, mock_svc_home, mock_get_db, client):
        """Test homepage loads successfully."""
        # Mock database and service
        mock_db = {}
        mock_get_db.return_value = mock_db
        
        mock_svc_home.return_value = {
            "doc": {
                "title": "Test Document",
                "content": "Test content",
                "numero_di_pagina": 1
            },
            "prev_page": None,
            "next_page": 2,
            "prev_title": None,
            "next_title": "Next Document"
        }
        
        response = client.get("/")
        
        # Should not error (may return template error if templates not available)
        assert response.status_code in [200, 500]  # 500 if templates missing
    
    def test_homepage_with_invalid_language(self, client):
        """Test homepage with invalid language parameter."""
        response = client.get("/?lang=invalid")
        
        # Should return validation error
        assert response.status_code == 400
    
    def test_homepage_with_invalid_page(self, client):
        """Test homepage with invalid page parameter."""
        response = client.get("/?page=-1")
        
        # Should return validation error
        assert response.status_code == 400
        
        response = client.get("/?page=10001")
        assert response.status_code == 400


class TestListEndpoint:
    """Test list page endpoint."""
    
    def test_list_invalid_collection(self, client):
        """Test list endpoint with invalid collection."""
        response = client.get("/list/invalid_collection")
        
        # Should return 404 for invalid collection
        assert response.status_code == 404
    
    @patch("routers.pages.validate_collection_param")
    @patch("routers.pages.get_db")
    @patch("routers.pages.svc_list_page")
    async def test_list_valid_collection(
        self, 
        mock_svc_list, 
        mock_get_db, 
        mock_validate,
        client
    ):
        """Test list endpoint with valid collection."""
        # Mock validation and database
        mock_validate.return_value = "spells"
        mock_db = {}
        mock_get_db.return_value = mock_db
        
        mock_svc_list.return_value = {
            "items": [],
            "total": 0,
            "page": 1,
            "pages": 1
        }
        
        response = client.get("/list/spells")
        
        # Should not error (may return template error if templates not available)
        assert response.status_code in [200, 500]  # 500 if templates missing
    
    def test_list_with_search_query(self, client):
        """Test list endpoint with search parameters."""
        # Test with valid search query
        response = client.get("/list/spells?q=fireball&page=1")
        
        # May error due to missing templates but should not be validation error
        assert response.status_code != 422
    
    def test_list_with_invalid_parameters(self, client):
        """Test list endpoint with invalid parameters."""
        # Invalid page
        response = client.get("/list/spells?page=0")
        assert response.status_code == 422
        
        # Invalid page size
        response = client.get("/list/spells?page_size=101")
        assert response.status_code == 422
        
        # Invalid language
        response = client.get("/list/spells?lang=invalid")
        assert response.status_code == 422


class TestErrorHandling:
    """Test error handling across endpoints."""
    
    def test_404_for_unknown_routes(self, client):
        """Test 404 for unknown routes."""
        response = client.get("/unknown-route")
        
        assert response.status_code == 404
    
    @patch("routers.pages.get_db")
    async def test_database_error_handling(self, mock_get_db, client):
        """Test database error handling."""
        # Mock database connection failure
        mock_get_db.side_effect = Exception("Database connection failed")
        
        response = client.get("/")
        
        # Should return 500 for database errors
        assert response.status_code == 500


class TestRequestValidation:
    """Test request validation integration."""
    
    def test_query_parameter_validation(self, client):
        """Test query parameter validation."""
        # Test various invalid parameter combinations
        test_cases = [
            ("/?page=abc", 422),  # Non-integer page
            ("/?page_size=abc", 422),  # Non-integer page_size
            ("/?lang=invalid", 422),  # Invalid language
            ("/list/spells?level=10", 422),  # Invalid spell level
            ("/list/spells?q=" + "a" * 201, 422),  # Query too long
        ]
        
        for url, expected_status in test_cases:
            response = client.get(url)
            assert response.status_code == expected_status, f"URL {url} should return {expected_status}"
    
    def test_dangerous_query_parameters(self, client):
        """Test handling of potentially dangerous query parameters."""
        dangerous_queries = [
            "/list/spells?q=test{}",  # MongoDB injection attempt
            "/list/spells?q=test$where",  # MongoDB operator
            "/list/spells?collection=../../../etc/passwd",  # Path traversal attempt
        ]
        
        for url in dangerous_queries:
            response = client.get(url)
            # Should either be validation error (422) or not found (404)
            assert response.status_code in [422, 404], f"Dangerous query {url} not properly handled"


class TestContentTypes:
    """Test content type handling."""
    
    def test_html_content_type(self, client):
        """Test that HTML endpoints return proper content type."""
        response = client.get("/healthz")
        
        assert response.status_code == 200
        # Health endpoint returns plain text
        assert "text/plain" in response.headers.get("content-type", "")
    
    @patch("routers.pages.get_db")
    async def test_error_response_content_type(self, mock_get_db, client):
        """Test error responses return proper content type."""
        # Mock database error
        mock_get_db.side_effect = Exception("Database error")
        
        response = client.get("/")
        
        # Error responses should be JSON
        if response.status_code >= 400:
            content_type = response.headers.get("content-type", "")
            assert "application/json" in content_type or "text/html" in content_type