from typing import Optional
from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorDatabase
from .config import MONGO_URI, DB_NAME

_client: Optional[AsyncIOMotorClient] = None
_db: Optional[AsyncIOMotorDatabase] = None


async def init_db() -> None:
    """Initialize Mongo connection (only once)."""
    global _client, _db
    if _client is None:
        _client = AsyncIOMotorClient(MONGO_URI)
        _db = _client[DB_NAME]


async def close_db() -> None:
    """Close Mongo connection and reset globals."""
    global _client, _db
    if _client:
        _client.close()
    _client = None
    _db = None


async def get_db() -> AsyncIOMotorDatabase:
    """Return the database, initializing if needed."""
    global _db
    if _db is None:
        await init_db()
    return _db

