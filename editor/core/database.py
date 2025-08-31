"""Advanced MongoDB connection manager for D&D 5e SRD Editor."""
from __future__ import annotations

import asyncio
import time
from contextlib import asynccontextmanager
from dataclasses import dataclass
from typing import AsyncGenerator, Dict, List, Optional, Any
from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorDatabase, AsyncIOMotorCollection
from pymongo.errors import ServerSelectionTimeoutError, ConnectionFailure, ConfigurationError
from pymongo.monitoring import CommandListener, CommandStartedEvent, CommandSucceededEvent, CommandFailedEvent

from core.config import MONGO_URI, DB_NAME
from core.errors import DatabaseError, ErrorCode
from core.logging_config import get_logger

logger = get_logger(__name__)


@dataclass
class ConnectionConfig:
    """MongoDB connection configuration."""
    
    uri: str = MONGO_URI
    db_name: str = DB_NAME
    max_pool_size: int = 50
    min_pool_size: int = 10
    max_idle_time_ms: int = 30000
    server_selection_timeout_ms: int = 5000
    connect_timeout_ms: int = 10000
    socket_timeout_ms: int = 30000
    retry_writes: bool = True
    retry_reads: bool = True
    max_retries: int = 3
    
    # Index configuration
    ensure_indexes_on_startup: bool = True
    index_creation_timeout: int = 30


@dataclass
class ConnectionStats:
    """MongoDB connection statistics."""
    
    total_connections: int = 0
    active_connections: int = 0
    available_connections: int = 0
    total_commands: int = 0
    failed_commands: int = 0
    avg_command_duration_ms: float = 0.0
    last_error: Optional[str] = None
    last_error_time: Optional[float] = None
    uptime_seconds: float = 0.0
    connected_since: Optional[float] = None


class DatabaseMonitor(CommandListener):
    """MongoDB command monitoring for statistics and logging."""
    
    def __init__(self):
        self.stats = ConnectionStats()
        self._command_start_times: Dict[int, float] = {}
        self._command_durations: List[float] = []
        self._lock = asyncio.Lock()
    
    def started(self, event: CommandStartedEvent) -> None:
        """Handle command started event."""
        self._command_start_times[event.request_id] = time.time()
        logger.debug(
            f"MongoDB command started: {event.command_name}",
            extra={
                "command_name": event.command_name,
                "database_name": event.database_name,
                "request_id": event.request_id,
            }
        )
    
    def succeeded(self, event: CommandSucceededEvent) -> None:
        """Handle command succeeded event."""
        start_time = self._command_start_times.pop(event.request_id, None)
        if start_time:
            duration = (time.time() - start_time) * 1000  # Convert to ms
            self._command_durations.append(duration)
            
            # Keep only last 1000 durations for averaging
            if len(self._command_durations) > 1000:
                self._command_durations = self._command_durations[-1000:]
            
            self.stats.total_commands += 1
            self.stats.avg_command_duration_ms = sum(self._command_durations) / len(self._command_durations)
            
            logger.debug(
                f"MongoDB command succeeded: {event.command_name} ({duration:.2f}ms)",
                extra={
                    "command_name": event.command_name,
                    "duration_ms": duration,
                    "request_id": event.request_id,
                }
            )
    
    def failed(self, event: CommandFailedEvent) -> None:
        """Handle command failed event."""
        self._command_start_times.pop(event.request_id, None)
        self.stats.failed_commands += 1
        self.stats.last_error = str(event.failure)
        self.stats.last_error_time = time.time()
        
        logger.error(
            f"MongoDB command failed: {event.command_name}",
            extra={
                "command_name": event.command_name,
                "error": str(event.failure),
                "request_id": event.request_id,
            }
        )


class DatabaseConnectionManager:
    """Advanced MongoDB connection manager with pooling, monitoring, and health checks."""
    
    def __init__(self, config: Optional[ConnectionConfig] = None):
        self.config = config or ConnectionConfig()
        self.monitor = DatabaseMonitor()
        self._client: Optional[AsyncIOMotorClient] = None
        self._db: Optional[AsyncIOMotorDatabase] = None
        self._connection_lock = asyncio.Lock()
        self._health_check_task: Optional[asyncio.Task] = None
        self._is_healthy = False
        
    async def connect(self) -> None:
        """Initialize database connection with advanced configuration."""
        async with self._connection_lock:
            if self._client is not None:
                logger.info("Database already connected")
                return
            
            logger.info("Initializing database connection", extra={"uri": self._mask_uri(self.config.uri)})
            
            try:
                # Configure client with advanced options
                self._client = AsyncIOMotorClient(
                    self.config.uri,
                    maxPoolSize=self.config.max_pool_size,
                    minPoolSize=self.config.min_pool_size,
                    maxIdleTimeMS=self.config.max_idle_time_ms,
                    serverSelectionTimeoutMS=self.config.server_selection_timeout_ms,
                    connectTimeoutMS=self.config.connect_timeout_ms,
                    socketTimeoutMS=self.config.socket_timeout_ms,
                    retryWrites=self.config.retry_writes,
                    retryReads=self.config.retry_reads,
                    event_listeners=[self.monitor],
                    appname="dnd-5e-srd-editor",
                )
                
                self._db = self._client[self.config.db_name]
                
                # Test connection
                await self._test_connection()
                
                # Set connection stats
                self.monitor.stats.connected_since = time.time()
                self.monitor.stats.uptime_seconds = 0.0
                
                # Ensure indexes if configured
                if self.config.ensure_indexes_on_startup:
                    await self._ensure_indexes()
                
                # Start health check task
                await self._start_health_monitoring()
                
                self._is_healthy = True
                logger.info("Database connection initialized successfully")
                
            except Exception as e:
                logger.error("Failed to initialize database connection", exc_info=e)
                await self._cleanup()
                raise DatabaseError(
                    f"Failed to connect to database: {str(e)}",
                    ErrorCode.DATABASE_CONNECTION_FAILED,
                    context={"uri": self._mask_uri(self.config.uri)}
                )
    
    async def disconnect(self) -> None:
        """Close database connection and cleanup resources."""
        async with self._connection_lock:
            logger.info("Closing database connection")
            await self._cleanup()
            logger.info("Database connection closed")
    
    async def _cleanup(self) -> None:
        """Internal cleanup method."""
        if self._health_check_task:
            self._health_check_task.cancel()
            try:
                await self._health_check_task
            except asyncio.CancelledError:
                pass
            self._health_check_task = None
        
        if self._client:
            self._client.close()
            self._client = None
            self._db = None
        
        self._is_healthy = False
    
    async def get_database(self) -> AsyncIOMotorDatabase:
        """Get database instance, connecting if necessary."""
        if self._db is None:
            await self.connect()
        
        if not self._is_healthy:
            raise DatabaseError(
                "Database connection is not healthy",
                ErrorCode.DATABASE_CONNECTION_FAILED
            )
        
        return self._db
    
    async def get_collection(self, name: str) -> AsyncIOMotorCollection:
        """Get collection instance."""
        db = await self.get_database()
        return db[name]
    
    async def _test_connection(self) -> None:
        """Test database connection."""
        if self._db is None:
            raise DatabaseError("Database not initialized", ErrorCode.DATABASE_CONNECTION_FAILED)
        
        try:
            # Use ping command to test connection
            await asyncio.wait_for(
                self._db.command("ping"),
                timeout=self.config.server_selection_timeout_ms / 1000
            )
            logger.debug("Database connection test successful")
            
        except (ServerSelectionTimeoutError, ConnectionFailure, asyncio.TimeoutError) as e:
            raise DatabaseError(
                f"Database connection test failed: {str(e)}",
                ErrorCode.DATABASE_CONNECTION_FAILED
            )
    
    async def _ensure_indexes(self) -> None:
        """Ensure required indexes exist."""
        if self._db is None:
            return
        
        logger.info("Ensuring database indexes")
        
        indexes = [
            ("documenti", [("numero_di_pagina", 1)], "idx_page"),
            ("documenti_en", [("numero_di_pagina", 1)], "idx_page_en"),
            ("documenti", [("_sortkey_alpha", 1)], "idx_sort"),
            ("documenti_en", [("_sortkey_alpha", 1)], "idx_sort_en"),
            ("documenti", [("slug", 1)], "idx_slug"),
            ("documenti_en", [("slug", 1)], "idx_slug_en"),
            ("incantesimi", [("level", 1), ("school", 1)], "idx_level_school"),
            ("incantesimi_en", [("level", 1), ("school", 1)], "idx_level_school_en"),
            ("incantesimi", [("classes", 1)], "idx_classes"),
            ("incantesimi_en", [("classes", 1)], "idx_classes_en"),
            ("oggetti_magici", [("type", 1), ("rarity", 1)], "idx_type_rarity"),
            ("oggetti_magici_en", [("type", 1), ("rarity", 1)], "idx_type_rarity_en"),
        ]
        
        index_tasks = []
        for collection_name, index_spec, index_name in indexes:
            task = self._create_index_safe(collection_name, index_spec, index_name)
            index_tasks.append(task)
        
        try:
            # Create indexes concurrently with timeout
            await asyncio.wait_for(
                asyncio.gather(*index_tasks, return_exceptions=True),
                timeout=self.config.index_creation_timeout
            )
            logger.info("Database indexes ensured successfully")
            
        except asyncio.TimeoutError:
            logger.warning("Index creation timed out, continuing anyway")
    
    async def _create_index_safe(self, collection_name: str, index_spec: List, index_name: str) -> None:
        """Safely create an index with error handling."""
        try:
            collection = self._db[collection_name]
            await collection.create_index(
                index_spec,
                name=index_name,
                background=True
            )
            logger.debug(f"Index '{index_name}' ensured on collection '{collection_name}'")
            
        except Exception as e:
            # Log but don't fail - indexes are performance optimizations
            logger.warning(
                f"Failed to create index '{index_name}' on collection '{collection_name}': {str(e)}"
            )
    
    async def _start_health_monitoring(self) -> None:
        """Start background health monitoring task."""
        if self._health_check_task is None:
            self._health_check_task = asyncio.create_task(self._health_monitor_loop())
    
    async def _health_monitor_loop(self) -> None:
        """Background health monitoring loop."""
        while True:
            try:
                await asyncio.sleep(30)  # Check every 30 seconds
                
                if self._db is not None:
                    # Update uptime
                    if self.monitor.stats.connected_since:
                        self.monitor.stats.uptime_seconds = time.time() - self.monitor.stats.connected_since
                    
                    # Test connection health
                    try:
                        await asyncio.wait_for(self._db.command("ping"), timeout=5.0)
                        if not self._is_healthy:
                            logger.info("Database connection restored")
                            self._is_healthy = True
                            
                    except Exception as e:
                        if self._is_healthy:
                            logger.error(f"Database health check failed: {str(e)}")
                            self._is_healthy = False
                            self.monitor.stats.last_error = str(e)
                            self.monitor.stats.last_error_time = time.time()
                
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Error in health monitoring: {str(e)}")
    
    async def get_health_status(self) -> Dict[str, Any]:
        """Get comprehensive database health status."""
        if self._db is None:
            return {
                "status": "disconnected",
                "healthy": False,
                "error": "Database not initialized"
            }
        
        try:
            # Test connection
            start_time = time.time()
            await asyncio.wait_for(self._db.command("ping"), timeout=5.0)
            ping_time = (time.time() - start_time) * 1000
            
            # Get server status
            server_info = await self._db.command("hello")
            
            # Update connection stats
            client_info = self._client.topology_description
            
            return {
                "status": "connected",
                "healthy": self._is_healthy,
                "ping_ms": round(ping_time, 2),
                "server_version": server_info.get("version", "unknown"),
                "database_name": self.config.db_name,
                "uptime_seconds": round(self.monitor.stats.uptime_seconds, 2),
                "total_commands": self.monitor.stats.total_commands,
                "failed_commands": self.monitor.stats.failed_commands,
                "avg_command_duration_ms": round(self.monitor.stats.avg_command_duration_ms, 2),
                "connection_pool": {
                    "max_pool_size": self.config.max_pool_size,
                    "min_pool_size": self.config.min_pool_size,
                },
                "last_error": self.monitor.stats.last_error,
                "last_error_time": self.monitor.stats.last_error_time,
            }
            
        except Exception as e:
            return {
                "status": "error",
                "healthy": False,
                "error": str(e),
                "last_successful_ping": self.monitor.stats.connected_since
            }
    
    @asynccontextmanager
    async def transaction(self) -> AsyncGenerator[None, None]:
        """Context manager for database transactions."""
        if self._client is None:
            raise DatabaseError("Database not connected", ErrorCode.DATABASE_CONNECTION_FAILED)
        
        async with await self._client.start_session() as session:
            async with session.start_transaction():
                logger.debug("Started database transaction")
                try:
                    yield
                    logger.debug("Committed database transaction")
                except Exception:
                    logger.debug("Aborted database transaction")
                    raise
    
    def _mask_uri(self, uri: str) -> str:
        """Mask sensitive information in URI for logging."""
        try:
            from urllib.parse import urlparse, urlunparse
            parsed = urlparse(uri)
            if parsed.password:
                netloc = f"{parsed.username}:***@{parsed.hostname}"
                if parsed.port:
                    netloc += f":{parsed.port}"
                masked = parsed._replace(netloc=netloc)
                return urlunparse(masked)
            return uri
        except Exception:
            return "mongodb://***:***@***"


# Global connection manager instance
_connection_manager: Optional[DatabaseConnectionManager] = None


async def get_connection_manager() -> DatabaseConnectionManager:
    """Get global connection manager instance."""
    global _connection_manager
    if _connection_manager is None:
        _connection_manager = DatabaseConnectionManager()
        await _connection_manager.connect()
    return _connection_manager


async def get_database() -> AsyncIOMotorDatabase:
    """Get database instance from connection manager."""
    manager = await get_connection_manager()
    return await manager.get_database()


async def get_collection(name: str) -> AsyncIOMotorCollection:
    """Get collection instance from connection manager."""
    manager = await get_connection_manager()
    return await manager.get_collection(name)


async def close_connection_manager() -> None:
    """Close global connection manager."""
    global _connection_manager
    if _connection_manager:
        await _connection_manager.disconnect()
        _connection_manager = None


# Compatibility functions for existing code
async def init_db() -> None:
    """Initialize database connection (compatibility function)."""
    await get_connection_manager()


async def close_db() -> None:
    """Close database connection (compatibility function)."""
    await close_connection_manager()


async def get_db() -> AsyncIOMotorDatabase:
    """Get database instance (compatibility function)."""
    return await get_database()
