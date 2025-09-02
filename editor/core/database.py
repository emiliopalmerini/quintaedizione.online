"""Simplified database connection for D&D 5e SRD Editor."""

import os
from typing import Optional
from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorDatabase
import asyncio

# Global connection
_client: Optional[AsyncIOMotorClient] = None
_database: Optional[AsyncIOMotorDatabase] = None

MONGO_URI = os.getenv("MONGO_URI", "mongodb://localhost:27017")
DB_NAME = os.getenv("DB_NAME", "dnd")


async def get_database() -> AsyncIOMotorDatabase:
    """Get database connection with simple connection management."""
    global _client, _database
    
    if _database is None:
        _client = AsyncIOMotorClient(MONGO_URI)
        _database = _client[DB_NAME]
        
        # Simple connection test
        try:
            await _database.command("ping")
        except Exception as e:
            raise ConnectionError(f"Database connection failed: {e}")
    
    return _database


async def close_database():
    """Close database connection."""
    global _client, _database
    
    if _client:
        _client.close()
        _client = None
        _database = None


async def health_check() -> bool:
    """Simple health check - just ping database."""
    try:
        db = await get_database()
        await db.command("ping")
        return True
    except Exception:
        return False


# Backwards compatibility
get_db = get_database