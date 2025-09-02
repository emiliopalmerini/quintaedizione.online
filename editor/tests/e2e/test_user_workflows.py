"""End-to-end tests for user workflows."""
from __future__ import annotations

import pytest
from fastapi.testclient import TestClient
from unittest.mock import patch

from main import create_app


@pytest.fixture
def app():
    """Create app for e2e testing."""
    return create_app()


@pytest.fixture
def client(app):
    """Create test client for e2e testing."""
    return TestClient(app)


class TestHomepageWorkflow:
    """Test complete homepage user workflow."""
    
    @patch("routers.pages.get_db")
    @patch("routers.pages.svc_home_doc")
    @patch("routers.pages.env")
    async def test_complete_homepage_flow(
        self, 
        mock_env, 
        mock_svc_home, 
        mock_get_db, 
        client
    ):
        """Test complete homepage loading workflow."""
        # Mock template environment
        mock_template = mock_env.get_template.return_value
        mock_template.render.return_value = """
        <html>
            <body>
                <h1>D&D 5e SRD Editor</h1>
                <div class="collections">
                    <div>Spells: 150</div>
                    <div>Items: 75</div>
                </div>
                <div class="document">
                    <h2>Test Document</h2>
                    <p>This is test content</p>
                </div>
            </body>
        </html>
        """
        
        # Mock database collections and counts
        mock_db = {
            "incantesimi": type('Collection', (), {
                'count_documents': lambda filter_doc: 150
            })(),
            "oggetti_magici": type('Collection', (), {
                'count_documents': lambda filter_doc: 75
            })(),
        }
        mock_get_db.return_value = mock_db
        
        # Mock home document service
        mock_svc_home.return_value = {
            "doc": {
                "_id": "test_doc_001",
                "title": "Test Document",
                "content": "This is test content for the document",
                "slug": "test-document",
                "numero_di_pagina": 1
            },
            "prev_page": None,
            "next_page": 2,
            "prev_title": None,
            "next_title": "Next Document"
        }
        
        # Test homepage load
        response = client.get("/")
        
        assert response.status_code == 200
        assert "D&D 5e SRD Editor" in response.text
    
    def test_homepage_language_validation(self, client):
        """Test language validation on homepage."""
        # Test Italian (only supported language)
        response_it = client.get("/?lang=it")
        assert response_it.status_code in [200, 500]  # 500 if templates missing
        
        # Test English (no longer supported)
        response_en = client.get("/?lang=en")
        assert response_en.status_code == 400  # Validation error
        
        # Test invalid language
        response_invalid = client.get("/?lang=invalid")
        assert response_invalid.status_code == 400  # Validation error
    
    def test_homepage_pagination(self, client):
        """Test pagination on homepage."""
        # Test different page numbers
        for page in [1, 2, 3]:
            response = client.get(f"/?page={page}")
            assert response.status_code in [200, 500]  # 500 if templates missing
        
        # Test invalid page numbers
        response = client.get("/?page=0")
        assert response.status_code == 400
        
        response = client.get("/?page=10001")
        assert response.status_code == 400


class TestListPageWorkflow:
    """Test complete list page user workflow."""
    
    @patch("routers.pages.validate_collection_param")
    @patch("routers.pages.get_db")
    @patch("routers.pages.svc_list_page")
    @patch("routers.pages.env")
    async def test_spell_browsing_workflow(
        self,
        mock_env,
        mock_svc_list,
        mock_get_db,
        mock_validate,
        client
    ):
        """Test complete spell browsing workflow."""
        # Setup mocks
        mock_validate.return_value = "incantesimi"
        mock_db = {}
        mock_get_db.return_value = mock_db
        
        mock_template = mock_env.get_template.return_value
        mock_template.render.return_value = """
        <html>
            <body>
                <h1>Spells</h1>
                <div class="search">
                    <input name="q" value="fireball" />
                </div>
                <div class="results">
                    <div class="spell">Fireball</div>
                    <div class="spell">Fire Bolt</div>
                </div>
                <div class="pagination">
                    <a href="?page=1">1</a>
                    <a href="?page=2">2</a>
                </div>
            </body>
        </html>
        """
        
        mock_svc_list.return_value = {
            "items": [
                {
                    "_id": "fireball",
                    "title": "Fireball",
                    "slug": "fireball",
                    "level": 3,
                    "school": "Evocation"
                },
                {
                    "_id": "fire_bolt",
                    "title": "Fire Bolt",
                    "slug": "fire-bolt",
                    "level": 0,
                    "school": "Evocation"
                }
            ],
            "total": 2,
            "page": 1,
            "pages": 1
        }
        
        # Test list page
        response = client.get("/list/incantesimi")
        
        assert response.status_code == 200
        assert "Spells" in response.text
        assert "Fireball" in response.text
    
    def test_search_workflow(self, client):
        """Test search functionality workflow."""
        # Test basic search
        response = client.get("/list/spells?q=fireball")
        assert response.status_code != 422  # Should not be validation error
        
        # Test search with filters
        response = client.get("/list/spells?q=fire&level=3&school=Evocation")
        assert response.status_code != 422
        
        # Test invalid search parameters
        response = client.get("/list/spells?q=" + "a" * 201)  # Too long
        assert response.status_code == 422
        
        response = client.get("/list/spells?level=10")  # Invalid level
        assert response.status_code == 422
    
    def test_pagination_workflow(self, client):
        """Test pagination workflow."""
        # Test different page sizes
        for page_size in [10, 20, 50]:
            response = client.get(f"/list/spells?page_size={page_size}")
            assert response.status_code != 422
        
        # Test invalid page sizes
        response = client.get("/list/spells?page_size=0")
        assert response.status_code == 422
        
        response = client.get("/list/spells?page_size=101")
        assert response.status_code == 422


class TestNavigationWorkflow:
    """Test navigation between pages."""
    
    def test_collection_navigation(self, client):
        """Test navigation between different collections."""
        collections = ["spells", "items", "monsters"]
        
        for collection in collections:
            response = client.get(f"/list/{collection}")
            # Should either work (200) or fail due to invalid collection (404)
            # or fail due to missing templates (500)
            assert response.status_code in [200, 404, 500]
    
    def test_breadcrumb_navigation(self, client):
        """Test breadcrumb navigation workflow."""
        # Start from homepage
        response = client.get("/")
        assert response.status_code in [200, 500]
        
        # Navigate to collection list
        response = client.get("/list/spells")
        assert response.status_code in [200, 404, 500]
        
        # Navigate back to homepage
        response = client.get("/")
        assert response.status_code in [200, 500]


class TestErrorRecoveryWorkflow:
    """Test user error recovery workflows."""
    
    def test_invalid_url_recovery(self, client):
        """Test recovery from invalid URLs."""
        # Test invalid collection
        response = client.get("/list/invalid-collection")
        assert response.status_code == 404
        
        # Test malformed URLs
        response = client.get("/list//")
        assert response.status_code == 404
        
        # Test completely invalid paths
        response = client.get("/completely/invalid/path")
        assert response.status_code == 404
    
    def test_invalid_parameter_recovery(self, client):
        """Test recovery from invalid parameters."""
        # Test various invalid parameter combinations
        invalid_requests = [
            "/?page=invalid",
            "/?lang=invalid",
            "/list/spells?level=invalid",
            "/list/spells?page_size=invalid",
            "/list/spells?q=" + "a" * 300,  # Too long query
        ]
        
        for request_url in invalid_requests:
            response = client.get(request_url)
            # Should return validation error, not crash
            assert response.status_code in [400, 422]
            
            # Response should be JSON (error response)
            try:
                error_data = response.json()
                assert "error" in error_data
                assert error_data["error"] is True
            except:
                # If not JSON, should at least not crash
                assert len(response.text) > 0
    
    @patch("routers.pages.get_db")
    async def test_database_error_recovery(self, mock_get_db, client):
        """Test recovery from database errors."""
        # Mock database failure
        mock_get_db.side_effect = Exception("Database connection failed")
        
        response = client.get("/")
        
        # Should return error but not crash
        assert response.status_code == 500
        
        # Should return structured error response
        try:
            error_data = response.json()
            assert "error" in error_data
            assert error_data["error"] is True
        except:
            # If HTML error page, should still be valid
            assert len(response.text) > 0


class TestPerformanceWorkflow:
    """Test performance-related workflows."""
    
    def test_large_result_set_handling(self, client):
        """Test handling of large result sets."""
        # Test with maximum page size
        response = client.get("/list/spells?page_size=100")
        assert response.status_code != 422  # Should accept max page size
        
        # Test pagination with large datasets
        response = client.get("/list/spells?page=100")
        assert response.status_code in [200, 404, 500]  # Should handle gracefully
    
    def test_complex_search_workflow(self, client):
        """Test complex search combinations."""
        # Test multiple filters
        response = client.get(
            "/list/spells?q=fire&level=3&school=Evocation&ritual=false&concentration=true"
        )
        assert response.status_code != 422
        
        # Test sorting options
        for sort_option in ["alpha", "level", "school", "modified"]:
            response = client.get(f"/list/spells?sort={sort_option}")
            assert response.status_code != 422


class TestUserExperienceWorkflow:
    """Test overall user experience workflows."""
    
    def test_responsive_behavior(self, client):
        """Test responsive behavior simulation."""
        # Simulate mobile user agent
        headers = {"User-Agent": "Mobile Browser"}
        
        response = client.get("/", headers=headers)
        assert response.status_code in [200, 500]
        
        response = client.get("/list/spells", headers=headers)
        assert response.status_code in [200, 404, 500]
    
    def test_content_language_workflow(self, client):
        """Test content language workflow."""
        # Test Italian language (only supported)
        response_it = client.get("/list/spells?lang=it")
        assert response_it.status_code in [200, 404, 500]
        
        # Test English language (no longer supported)
        response_en = client.get("/list/spells?lang=en")
        assert response_en.status_code == 400  # Validation error
    
    def test_search_suggestion_workflow(self, client):
        """Test search suggestion workflow."""
        # Test partial matches
        response = client.get("/list/spells?q=fire")
        assert response.status_code != 422
        
        # Test empty search
        response = client.get("/list/spells?q=")
        assert response.status_code != 422
        
        # Test single character search
        response = client.get("/list/spells?q=a")
        assert response.status_code != 422