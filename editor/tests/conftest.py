"""Pytest configuration and fixtures for D&D 5e SRD Editor tests."""
from __future__ import annotations

import asyncio
import pytest
import pytest_asyncio
from typing import AsyncGenerator, Dict, Any
from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorDatabase
from fastapi.testclient import TestClient
from unittest.mock import AsyncMock, MagicMock

from main import create_app
from core.db import init_db, close_db, get_db
from adapters.persistence.mongo_repository import MongoRepository


# Test database configuration
TEST_MONGO_URI = "mongodb://admin:password@localhost:27017/?authSource=admin"
TEST_DB_NAME = "test_dnd_srd"


@pytest.fixture(scope="session")
def event_loop():
    """Create an instance of the default event loop for the test session."""
    loop = asyncio.new_event_loop()
    yield loop
    loop.close()


@pytest_asyncio.fixture(scope="function")
async def test_db() -> AsyncGenerator[AsyncIOMotorDatabase, None]:
    """Provide a clean test database for each test."""
    client = AsyncIOMotorClient(TEST_MONGO_URI)
    db = client[TEST_DB_NAME]
    
    # Clean up any existing test data
    collections = await db.list_collection_names()
    for collection_name in collections:
        await db[collection_name].delete_many({})
    
    yield db
    
    # Cleanup after test
    await client.drop_database(TEST_DB_NAME)
    client.close()


@pytest_asyncio.fixture
async def test_repo(test_db: AsyncIOMotorDatabase) -> MongoRepository:
    """Provide a repository instance for testing."""
    return MongoRepository(test_db)


@pytest.fixture
def mock_db():
    """Provide a mock database for unit tests."""
    mock_db = AsyncMock()
    mock_collection = AsyncMock()
    mock_db.__getitem__ = MagicMock(return_value=mock_collection)
    mock_db.list_collection_names = AsyncMock(return_value=[])
    return mock_db


@pytest.fixture
def app():
    """Create FastAPI app instance for testing."""
    return create_app()


@pytest.fixture
def client(app):
    """Create test client for FastAPI app."""
    with TestClient(app) as test_client:
        yield test_client


@pytest_asyncio.fixture
async def app_with_test_db(test_db):
    """Create FastAPI app with test database."""
    app = create_app()
    
    # Override database dependency
    async def override_get_db():
        return test_db
    
    app.dependency_overrides[get_db] = override_get_db
    yield app
    app.dependency_overrides.clear()


@pytest.fixture
def sample_spell_data() -> Dict[str, Any]:
    """Sample spell document for testing."""
    return {
        "_id": "test_spell_001",
        "title": "Magic Missile",
        "slug": "magic-missile",
        "content": "You create three glowing darts of magical force...",
        "level": 1,
        "school": "Evocation",
        "ritual": False,
        "concentration": False,
        "casting_time": "1 action",
        "range": "120 feet",
        "components": "V, S",
        "duration": "Instantaneous",
        "classes": ["Mago", "Stregone"],
        "numero_di_pagina": 257,
        "_sortkey_alpha": "magic missile",
        "modified": False,
        "translated": True,
    }


@pytest.fixture
def sample_item_data() -> Dict[str, Any]:
    """Sample magic item document for testing."""
    return {
        "_id": "test_item_001", 
        "title": "Potion of Healing",
        "slug": "potion-of-healing",
        "content": "You regain hit points when you drink this potion...",
        "type": "Potion",
        "rarity": "Common",
        "attunement": False,
        "numero_di_pagina": 187,
        "_sortkey_alpha": "potion of healing",
        "modified": False,
        "translated": True,
    }


@pytest.fixture
def sample_documents_list(sample_spell_data, sample_item_data) -> list[Dict[str, Any]]:
    """List of sample documents for testing."""
    return [sample_spell_data, sample_item_data]


@pytest_asyncio.fixture
async def populated_test_db(test_db, sample_documents_list):
    """Test database populated with sample data."""
    
    # Insert spells
    if sample_documents_list:
        spell_docs = [doc for doc in sample_documents_list if doc.get("level") is not None]
        if spell_docs:
            await test_db["incantesimi"].insert_many(spell_docs)
        
        # Insert items
        item_docs = [doc for doc in sample_documents_list if doc.get("type") is not None]
        if item_docs:
            await test_db["oggetti_magici"].insert_many(item_docs)
    
    yield test_db


# Mock fixtures for isolated unit tests
@pytest.fixture
def mock_logger():
    """Mock logger for testing."""
    return MagicMock()


@pytest.fixture
def mock_template_env():
    """Mock Jinja2 environment for testing."""
    mock_env = MagicMock()
    mock_template = MagicMock()
    mock_template.render.return_value = "<html>Mock Template</html>"
    mock_env.get_template.return_value = mock_template
    return mock_env


# Parametrized fixtures for testing different scenarios
@pytest.fixture(params=["it", "en"])
def language_param(request):
    """Parametrized fixture for testing different languages."""
    return request.param


@pytest.fixture(params=[1, 2, 10])
def page_param(request):
    """Parametrized fixture for testing different page numbers."""
    return request.param


@pytest.fixture(params=[10, 20, 50])
def page_size_param(request):
    """Parametrized fixture for testing different page sizes.""" 
    return request.param


# Test data builders
class DocumentBuilder:
    """Builder for creating test documents."""
    
    def __init__(self):
        self.data = {
            "title": "Test Document",
            "slug": "test-document",
            "content": "Test content",
            "numero_di_pagina": 1,
            "_sortkey_alpha": "test document",
            "modified": False,
            "translated": False,
        }
    
    def with_title(self, title: str):
        self.data["title"] = title
        self.data["slug"] = title.lower().replace(" ", "-")
        self.data["_sortkey_alpha"] = title.lower()
        return self
    
    def with_spell_data(self, level: int = 1, school: str = "Evocation"):
        self.data.update({
            "level": level,
            "school": school,
            "ritual": False,
            "concentration": False,
            "casting_time": "1 action",
            "range": "60 feet",
            "components": "V, S",
            "duration": "Instantaneous",
            "classes": ["Mago"],
        })
        return self
    
    def with_item_data(self, item_type: str = "Potion", rarity: str = "Common"):
        self.data.update({
            "type": item_type,
            "rarity": rarity,
            "attunement": False,
        })
        return self
    
    def as_translated(self):
        self.data["translated"] = True
        return self
    
    def as_modified(self):
        self.data["modified"] = True
        return self
    
    def build(self) -> Dict[str, Any]:
        return self.data.copy()


@pytest.fixture
def document_builder():
    """Provide document builder for tests."""
    return DocumentBuilder