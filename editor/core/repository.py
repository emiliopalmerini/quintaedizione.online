"""Simplified repository for database operations."""

from typing import Dict, List, Any, Optional
from motor.motor_asyncio import AsyncIOMotorDatabase, AsyncIOMotorCollection
from core.config import DB_COLLECTIONS_IT


class SimpleRepository:
    """Simple repository for basic CRUD operations."""
    
    def __init__(self, database: AsyncIOMotorDatabase):
        self.db = database
    
    def get_collection(self, collection_name: str) -> AsyncIOMotorCollection:
        """Get MongoDB collection by logical name."""
        db_name = DB_COLLECTIONS_IT.get(collection_name, collection_name)
        return self.db[db_name]
    
    async def find_all(
        self, 
        collection_name: str, 
        filter_query: Optional[Dict[str, Any]] = None,
        sort_by: Optional[List[tuple]] = None,
        limit: Optional[int] = None,
        skip: Optional[int] = None
    ) -> List[Dict[str, Any]]:
        """Find documents with optional filtering, sorting, and pagination."""
        collection = self.get_collection(collection_name)
        
        cursor = collection.find(filter_query or {})
        
        if sort_by:
            cursor = cursor.sort(sort_by)
        if skip:
            cursor = cursor.skip(skip)
        if limit:
            cursor = cursor.limit(limit)
            
        return await cursor.to_list(length=None)
    
    async def find_one(
        self, 
        collection_name: str, 
        filter_query: Dict[str, Any]
    ) -> Optional[Dict[str, Any]]:
        """Find single document."""
        collection = self.get_collection(collection_name)
        return await collection.find_one(filter_query)
    
    async def count_documents(
        self, 
        collection_name: str, 
        filter_query: Optional[Dict[str, Any]] = None
    ) -> int:
        """Count documents matching filter."""
        collection = self.get_collection(collection_name)
        return await collection.count_documents(filter_query or {})
    
    async def get_collection_stats(self, collection_name: str) -> Dict[str, Any]:
        """Get basic collection statistics."""
        collection = self.get_collection(collection_name)
        stats = await collection.aggregate([
            {"$group": {"_id": None, "count": {"$sum": 1}}}
        ]).to_list(1)
        
        return {
            "count": stats[0]["count"] if stats else 0,
            "name": collection_name
        }
    
    async def get_distinct_values(self, collection_name: str, field: str) -> List[Any]:
        """Get distinct values for a field in a collection."""
        collection = self.get_collection(collection_name)
        return await collection.distinct(field)