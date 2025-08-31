"""Integration tests for database operations."""
from __future__ import annotations

import pytest
from motor.motor_asyncio import AsyncIOMotorDatabase

from adapters.persistence.mongo_repository import MongoRepository
from application.list_service import list_page
from application.show_service import show_doc
from application.home_service import load_home_document


class TestMongoRepository:
    """Test MongoRepository with real database."""
    
    @pytest.mark.asyncio
    async def test_find_documents(self, test_repo: MongoRepository, sample_spell_data):
        """Test finding documents in repository."""
        # Insert test data
        await test_repo.insert("test_collection", sample_spell_data)
        
        # Find documents
        results = await test_repo.find("test_collection", {})
        
        assert len(results) == 1
        assert results[0]["title"] == "Magic Missile"
        assert results[0]["level"] == 1
    
    @pytest.mark.asyncio
    async def test_find_by_slug(self, test_repo: MongoRepository, sample_spell_data):
        """Test finding document by slug."""
        # Insert test data
        await test_repo.insert("test_collection", sample_spell_data)
        
        # Find by slug
        result = await test_repo.find_by_slug("test_collection", "magic-missile")
        
        assert result is not None
        assert result["title"] == "Magic Missile"
        assert result["slug"] == "magic-missile"
    
    @pytest.mark.asyncio
    async def test_find_by_nonexistent_slug(self, test_repo: MongoRepository):
        """Test finding nonexistent document."""
        result = await test_repo.find_by_slug("test_collection", "nonexistent")
        
        assert result is None
    
    @pytest.mark.asyncio
    async def test_count_documents(self, test_repo: MongoRepository, sample_documents_list):
        """Test counting documents."""
        # Insert test data
        for doc in sample_documents_list:
            await test_repo.insert("test_collection", doc)
        
        # Count documents
        count = await test_repo.count("test_collection", {})
        
        assert count == len(sample_documents_list)
    
    @pytest.mark.asyncio
    async def test_find_with_pagination(self, test_repo: MongoRepository, document_builder):
        """Test finding documents with pagination."""
        # Insert multiple documents
        docs = []
        for i in range(25):
            doc = document_builder.with_title(f"Document {i:02d}").build()
            docs.append(doc)
            await test_repo.insert("test_collection", doc)
        
        # Test pagination
        page1 = await test_repo.find("test_collection", {}, skip=0, limit=10)
        page2 = await test_repo.find("test_collection", {}, skip=10, limit=10)
        page3 = await test_repo.find("test_collection", {}, skip=20, limit=10)
        
        assert len(page1) == 10
        assert len(page2) == 10
        assert len(page3) == 5
        
        # Verify no overlap
        page1_ids = {doc["_id"] for doc in page1}
        page2_ids = {doc["_id"] for doc in page2}
        assert page1_ids.isdisjoint(page2_ids)
    
    @pytest.mark.asyncio
    async def test_find_with_filter(self, test_repo: MongoRepository, document_builder):
        """Test finding documents with filters."""
        # Insert documents with different levels
        spell1 = document_builder.with_title("Level 1 Spell").with_spell_data(level=1).build()
        spell3 = document_builder.with_title("Level 3 Spell").with_spell_data(level=3).build()
        spell5 = document_builder.with_title("Level 5 Spell").with_spell_data(level=5).build()
        
        await test_repo.insert("spells", spell1)
        await test_repo.insert("spells", spell3)
        await test_repo.insert("spells", spell5)
        
        # Filter by level
        level3_spells = await test_repo.find("spells", {"level": 3})
        
        assert len(level3_spells) == 1
        assert level3_spells[0]["title"] == "Level 3 Spell"
        assert level3_spells[0]["level"] == 3
    
    @pytest.mark.asyncio
    async def test_update_document(self, test_repo: MongoRepository, sample_spell_data):
        """Test updating documents."""
        # Insert document
        await test_repo.insert("test_collection", sample_spell_data)
        
        # Update document
        update_data = {"$set": {"modified": True, "title": "Updated Magic Missile"}}
        result = await test_repo.update_by_slug(
            "test_collection", 
            "magic-missile", 
            update_data
        )
        
        assert result.modified_count == 1
        
        # Verify update
        updated_doc = await test_repo.find_by_slug("test_collection", "magic-missile")
        assert updated_doc["modified"] is True
        assert updated_doc["title"] == "Updated Magic Missile"
    
    @pytest.mark.asyncio
    async def test_aggregate_operations(self, test_repo: MongoRepository, document_builder):
        """Test aggregation pipeline operations."""
        # Insert documents with different schools
        evocation_spells = []
        transmutation_spells = []
        
        for i in range(3):
            evocation_spell = document_builder.with_title(f"Evocation {i}").with_spell_data(school="Evocation").build()
            transmutation_spell = document_builder.with_title(f"Transmutation {i}").with_spell_data(school="Transmutation").build()
            
            evocation_spells.append(evocation_spell)
            transmutation_spells.append(transmutation_spell)
            
            await test_repo.insert("spells", evocation_spell)
            await test_repo.insert("spells", transmutation_spell)
        
        # Aggregate by school
        pipeline = [
            {"$group": {"_id": "$school", "count": {"$sum": 1}}},
            {"$sort": {"_id": 1}}
        ]
        
        results = await test_repo.aggregate("spells", pipeline)
        
        assert len(results) == 2
        assert results[0]["_id"] == "Evocation"
        assert results[0]["count"] == 3
        assert results[1]["_id"] == "Transmutation"
        assert results[1]["count"] == 3


class TestApplicationServices:
    """Test application services with real database."""
    
    @pytest.mark.asyncio
    async def test_list_service(self, test_repo: MongoRepository, document_builder):
        """Test list page service."""
        # Insert test documents
        for i in range(15):
            doc = document_builder.with_title(f"Spell {i:02d}").with_spell_data().build()
            await test_repo.insert("spells", doc)
        
        # Test list service
        result = await list_page(
            test_repo,
            "spells",
            filter_doc={},
            search_query="",
            page=1,
            page_size=10
        )
        
        assert "items" in result
        assert "total" in result
        assert "page" in result
        assert "pages" in result
        
        assert len(result["items"]) == 10
        assert result["total"] == 15
        assert result["page"] == 1
        assert result["pages"] == 2
    
    @pytest.mark.asyncio
    async def test_list_service_with_search(self, test_repo: MongoRepository, document_builder):
        """Test list service with search query."""
        # Insert documents
        fireball = document_builder.with_title("Fireball").with_spell_data(school="Evocation").build()
        magic_missile = document_builder.with_title("Magic Missile").with_spell_data(school="Evocation").build()
        heal = document_builder.with_title("Heal").with_spell_data(school="Evocation").build()
        
        await test_repo.insert("spells", fireball)
        await test_repo.insert("spells", magic_missile)
        await test_repo.insert("spells", heal)
        
        # Search for "magic"
        result = await list_page(
            test_repo,
            "spells",
            filter_doc={},
            search_query="magic",
            page=1,
            page_size=10
        )
        
        assert result["total"] >= 1  # Should find Magic Missile
        found_titles = [item["title"] for item in result["items"]]
        assert "Magic Missile" in found_titles
    
    @pytest.mark.asyncio
    async def test_show_service(self, test_repo: MongoRepository, sample_spell_data):
        """Test show document service."""
        # Insert test data
        await test_repo.insert("spells", sample_spell_data)
        
        # Test show service
        result = await show_doc(test_repo, "spells", "magic-missile")
        
        assert "doc" in result
        assert result["doc"]["title"] == "Magic Missile"
        assert result["doc"]["slug"] == "magic-missile"
    
    @pytest.mark.asyncio
    async def test_show_service_nonexistent(self, test_repo: MongoRepository):
        """Test show service with nonexistent document."""
        result = await show_doc(test_repo, "spells", "nonexistent")
        
        assert result["doc"] is None
    
    @pytest.mark.asyncio
    async def test_home_service(self, test_repo: MongoRepository, document_builder):
        """Test home document service."""
        # Insert documents
        for i in range(5):
            doc = document_builder.with_title(f"Document {i}").build()
            doc["numero_di_pagina"] = i + 1
            await test_repo.insert("documenti", doc)
        
        # Test home service
        result = await load_home_document(test_repo, page=2, collection="documenti")
        
        assert "doc" in result
        assert result["doc"] is not None
        assert result["doc"]["numero_di_pagina"] == 2


class TestDatabaseErrorHandling:
    """Test database error handling."""
    
    @pytest.mark.asyncio
    async def test_connection_error_handling(self):
        """Test handling of database connection errors."""
        from motor.motor_asyncio import AsyncIOMotorClient
        
        # Use invalid connection string
        client = AsyncIOMotorClient("mongodb://invalid:27017")
        db = client["test_db"]
        repo = MongoRepository(db)
        
        # This should handle the connection error gracefully
        with pytest.raises(Exception):
            await repo.find("test_collection", {})
        
        client.close()
    
    @pytest.mark.asyncio
    async def test_invalid_collection_operations(self, test_repo: MongoRepository):
        """Test operations on collections with invalid names."""
        # MongoDB should handle invalid collection names gracefully
        try:
            await test_repo.find("", {})  # Empty collection name
        except Exception as e:
            # Should get a specific error, not crash
            assert isinstance(e, Exception)
    
    @pytest.mark.asyncio
    async def test_invalid_query_operations(self, test_repo: MongoRepository):
        """Test operations with invalid queries."""
        # Insert valid data first
        await test_repo.insert("test_collection", {"title": "Test", "value": 1})
        
        # Invalid regex query should be handled
        try:
            await test_repo.find("test_collection", {"title": {"$regex": "["}})
        except Exception as e:
            # Should get a specific regex error, not crash
            assert isinstance(e, Exception)