"""
Connection Pool Management for Database Performance
"""
import asyncio
import logging
from typing import Any, Dict, Optional, List, Set, Callable
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from motor.motor_asyncio import AsyncIOMotorClient

logger = logging.getLogger(__name__)


class ConnectionState(Enum):
    """Connection states in the pool"""
    AVAILABLE = "available"
    IN_USE = "in_use"
    TESTING = "testing"
    FAILED = "failed"
    EXPIRED = "expired"


@dataclass
class PoolStats:
    """Connection pool statistics"""
    total_connections: int = 0
    available_connections: int = 0
    in_use_connections: int = 0
    failed_connections: int = 0
    total_requests: int = 0
    successful_requests: int = 0
    failed_requests: int = 0
    avg_wait_time_ms: float = 0.0
    max_wait_time_ms: float = 0.0
    pool_exhaustion_count: int = 0
    
    @property
    def success_rate(self) -> float:
        """Calculate request success rate"""
        return (self.successful_requests / self.total_requests) if self.total_requests > 0 else 0.0
    
    @property
    def utilization_rate(self) -> float:
        """Calculate pool utilization rate"""
        return (self.in_use_connections / self.total_connections) if self.total_connections > 0 else 0.0


@dataclass
class PooledConnection:
    """Wrapper for pooled database connections"""
    connection_id: str
    client: AsyncIOMotorClient
    created_at: datetime
    last_used: datetime
    use_count: int = 0
    state: ConnectionState = ConnectionState.AVAILABLE
    health_check_count: int = 0
    failure_count: int = 0
    
    def mark_used(self) -> None:
        """Mark connection as used"""
        self.last_used = datetime.now()
        self.use_count += 1
        self.state = ConnectionState.IN_USE
    
    def mark_available(self) -> None:
        """Mark connection as available"""
        self.state = ConnectionState.AVAILABLE
    
    def mark_failed(self) -> None:
        """Mark connection as failed"""
        self.state = ConnectionState.FAILED
        self.failure_count += 1
    
    def is_expired(self, max_age_seconds: int) -> bool:
        """Check if connection has expired"""
        age = (datetime.now() - self.created_at).total_seconds()
        return age > max_age_seconds
    
    def is_idle_too_long(self, max_idle_seconds: int) -> bool:
        """Check if connection has been idle too long"""
        idle_time = (datetime.now() - self.last_used).total_seconds()
        return idle_time > max_idle_seconds


class ConnectionPoolManager:
    """Advanced connection pool manager with health monitoring"""
    
    def __init__(
        self,
        mongo_uri: str,
        db_name: str,
        min_pool_size: int = 2,
        max_pool_size: int = 20,
        max_idle_time_seconds: int = 600,  # 10 minutes
        max_connection_age_seconds: int = 3600,  # 1 hour
        health_check_interval_seconds: int = 60,  # 1 minute
        connection_timeout_seconds: int = 10,
        max_wait_time_seconds: int = 30
    ):
        self.mongo_uri = mongo_uri
        self.db_name = db_name
        self.min_pool_size = min_pool_size
        self.max_pool_size = max_pool_size
        self.max_idle_time_seconds = max_idle_time_seconds
        self.max_connection_age_seconds = max_connection_age_seconds
        self.health_check_interval_seconds = health_check_interval_seconds
        self.connection_timeout_seconds = connection_timeout_seconds
        self.max_wait_time_seconds = max_wait_time_seconds
        
        self._pool: Dict[str, PooledConnection] = {}
        self._available_queue = asyncio.Queue()
        self._stats = PoolStats()
        self._lock = asyncio.Lock()
        self._health_check_task: Optional[asyncio.Task] = None
        self._cleanup_task: Optional[asyncio.Task] = None
        self._is_initialized = False
        
        # Callback for connection events
        self._connection_event_callbacks: List[Callable[[str, PooledConnection], None]] = []
    
    async def initialize(self) -> None:
        """Initialize the connection pool"""
        if self._is_initialized:
            return
        
        async with self._lock:
            if self._is_initialized:
                return
            
            try:
                # Create minimum number of connections
                for i in range(self.min_pool_size):
                    await self._create_connection()
                
                # Start background tasks
                self._health_check_task = asyncio.create_task(self._health_check_loop())
                self._cleanup_task = asyncio.create_task(self._cleanup_loop())
                
                self._is_initialized = True
                logger.info(f"Connection pool initialized with {self.min_pool_size} connections")
                
            except Exception as e:
                logger.error(f"Failed to initialize connection pool: {e}")
                raise
    
    async def get_connection(self) -> PooledConnection:
        """Get a connection from the pool"""
        if not self._is_initialized:
            await self.initialize()
        
        start_time = asyncio.get_event_loop().time()
        self._stats.total_requests += 1
        
        try:
            # Try to get available connection
            connection = await self._get_available_connection()
            
            if connection:
                wait_time_ms = (asyncio.get_event_loop().time() - start_time) * 1000
                self._update_wait_time_stats(wait_time_ms)
                self._stats.successful_requests += 1
                
                # Mark connection as in use
                connection.mark_used()
                self._stats.in_use_connections += 1
                self._stats.available_connections -= 1
                
                self._trigger_connection_event("acquired", connection)
                return connection
            
            # Pool is exhausted
            self._stats.pool_exhaustion_count += 1
            self._stats.failed_requests += 1
            raise ConnectionPoolExhausted("No connections available in pool")
            
        except Exception as e:
            self._stats.failed_requests += 1
            logger.error(f"Failed to get connection from pool: {e}")
            raise
    
    async def return_connection(self, connection: PooledConnection, error: Optional[Exception] = None) -> None:
        """Return a connection to the pool"""
        async with self._lock:
            if connection.connection_id not in self._pool:
                logger.warning(f"Attempting to return unknown connection: {connection.connection_id}")
                return
            
            if error:
                # Connection had an error, mark as failed
                connection.mark_failed()
                logger.warning(f"Connection {connection.connection_id} returned with error: {error}")
                
                # Remove failed connection if failure count is too high
                if connection.failure_count >= 3:
                    await self._remove_connection(connection.connection_id)
                    self._trigger_connection_event("removed_failed", connection)
                    return
            else:
                # Healthy connection, mark as available
                connection.mark_available()
            
            # Update pool statistics
            self._stats.in_use_connections -= 1
            self._stats.available_connections += 1
            
            # Put back in available queue
            await self._available_queue.put(connection)
            
            self._trigger_connection_event("returned", connection)
    
    async def close(self) -> None:
        """Close all connections and shutdown pool"""
        async with self._lock:
            # Cancel background tasks
            if self._health_check_task:
                self._health_check_task.cancel()
            if self._cleanup_task:
                self._cleanup_task.cancel()
            
            # Close all connections
            for connection_id in list(self._pool.keys()):
                await self._remove_connection(connection_id)
            
            self._is_initialized = False
            logger.info("Connection pool closed")
    
    async def get_pool_stats(self) -> PoolStats:
        """Get current pool statistics"""
        async with self._lock:
            # Update current counts
            self._stats.total_connections = len(self._pool)
            self._stats.available_connections = sum(
                1 for conn in self._pool.values() 
                if conn.state == ConnectionState.AVAILABLE
            )
            self._stats.in_use_connections = sum(
                1 for conn in self._pool.values() 
                if conn.state == ConnectionState.IN_USE
            )
            self._stats.failed_connections = sum(
                1 for conn in self._pool.values() 
                if conn.state == ConnectionState.FAILED
            )
            
            return PoolStats(
                total_connections=self._stats.total_connections,
                available_connections=self._stats.available_connections,
                in_use_connections=self._stats.in_use_connections,
                failed_connections=self._stats.failed_connections,
                total_requests=self._stats.total_requests,
                successful_requests=self._stats.successful_requests,
                failed_requests=self._stats.failed_requests,
                avg_wait_time_ms=self._stats.avg_wait_time_ms,
                max_wait_time_ms=self._stats.max_wait_time_ms,
                pool_exhaustion_count=self._stats.pool_exhaustion_count
            )
    
    async def get_pool_info(self) -> Dict[str, Any]:
        """Get detailed pool information"""
        stats = await self.get_pool_stats()
        
        connection_details = []
        async with self._lock:
            for conn in self._pool.values():
                connection_details.append({
                    "connection_id": conn.connection_id,
                    "state": conn.state.value,
                    "created_at": conn.created_at.isoformat(),
                    "last_used": conn.last_used.isoformat(),
                    "use_count": conn.use_count,
                    "failure_count": conn.failure_count,
                    "health_check_count": conn.health_check_count,
                    "age_seconds": (datetime.now() - conn.created_at).total_seconds(),
                    "idle_seconds": (datetime.now() - conn.last_used).total_seconds()
                })
        
        return {
            "statistics": {
                "total_connections": stats.total_connections,
                "available_connections": stats.available_connections,
                "in_use_connections": stats.in_use_connections,
                "failed_connections": stats.failed_connections,
                "success_rate": stats.success_rate,
                "utilization_rate": stats.utilization_rate,
                "total_requests": stats.total_requests,
                "avg_wait_time_ms": stats.avg_wait_time_ms,
                "max_wait_time_ms": stats.max_wait_time_ms,
                "pool_exhaustion_count": stats.pool_exhaustion_count
            },
            "configuration": {
                "min_pool_size": self.min_pool_size,
                "max_pool_size": self.max_pool_size,
                "max_idle_time_seconds": self.max_idle_time_seconds,
                "max_connection_age_seconds": self.max_connection_age_seconds,
                "health_check_interval_seconds": self.health_check_interval_seconds
            },
            "connections": connection_details,
            "generated_at": datetime.now().isoformat()
        }
    
    def add_connection_event_callback(self, callback: Callable[[str, PooledConnection], None]) -> None:
        """Add callback for connection events"""
        self._connection_event_callbacks.append(callback)
    
    async def _create_connection(self) -> PooledConnection:
        """Create a new connection"""
        connection_id = f"conn_{len(self._pool)}_{int(datetime.now().timestamp())}"
        
        try:
            # Create MongoDB client
            client = AsyncIOMotorClient(
                self.mongo_uri,
                serverSelectionTimeoutMS=self.connection_timeout_seconds * 1000,
                maxPoolSize=1  # Each PooledConnection manages one actual connection
            )
            
            # Test connection
            await client.admin.command('ping')
            
            # Create pooled connection
            connection = PooledConnection(
                connection_id=connection_id,
                client=client,
                created_at=datetime.now(),
                last_used=datetime.now()
            )
            
            # Add to pool
            self._pool[connection_id] = connection
            await self._available_queue.put(connection)
            
            self._stats.total_connections += 1
            self._stats.available_connections += 1
            
            self._trigger_connection_event("created", connection)
            logger.debug(f"Created new connection: {connection_id}")
            
            return connection
            
        except Exception as e:
            logger.error(f"Failed to create connection: {e}")
            raise
    
    async def _get_available_connection(self) -> Optional[PooledConnection]:
        """Get an available connection from the pool"""
        try:
            # Wait for available connection with timeout
            connection = await asyncio.wait_for(
                self._available_queue.get(),
                timeout=self.max_wait_time_seconds
            )
            
            # Verify connection is still healthy and in pool
            if (connection.connection_id in self._pool and 
                connection.state == ConnectionState.AVAILABLE and
                not connection.is_expired(self.max_connection_age_seconds)):
                return connection
            
            # Connection is invalid, try to create new one if under limit
            async with self._lock:
                if len(self._pool) < self.max_pool_size:
                    return await self._create_connection()
            
            return None
            
        except asyncio.TimeoutError:
            # Try to create new connection if under limit
            async with self._lock:
                if len(self._pool) < self.max_pool_size:
                    return await self._create_connection()
            
            return None
    
    async def _remove_connection(self, connection_id: str) -> bool:
        """Remove connection from pool"""
        if connection_id not in self._pool:
            return False
        
        connection = self._pool[connection_id]
        
        try:
            # Close the MongoDB client
            connection.client.close()
        except Exception as e:
            logger.warning(f"Error closing connection {connection_id}: {e}")
        
        # Remove from pool
        del self._pool[connection_id]
        
        # Update stats
        self._stats.total_connections -= 1
        if connection.state == ConnectionState.AVAILABLE:
            self._stats.available_connections -= 1
        elif connection.state == ConnectionState.IN_USE:
            self._stats.in_use_connections -= 1
        elif connection.state == ConnectionState.FAILED:
            self._stats.failed_connections -= 1
        
        logger.debug(f"Removed connection: {connection_id}")
        return True
    
    async def _health_check_connection(self, connection: PooledConnection) -> bool:
        """Check if connection is healthy"""
        try:
            connection.state = ConnectionState.TESTING
            
            # Simple ping test
            await asyncio.wait_for(
                connection.client.admin.command('ping'),
                timeout=5.0
            )
            
            connection.health_check_count += 1
            connection.mark_available()
            return True
            
        except Exception as e:
            logger.warning(f"Health check failed for connection {connection.connection_id}: {e}")
            connection.mark_failed()
            return False
    
    async def _health_check_loop(self) -> None:
        """Background health check loop"""
        while True:
            try:
                await asyncio.sleep(self.health_check_interval_seconds)
                await self._perform_health_checks()
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Error in health check loop: {e}")
    
    async def _perform_health_checks(self) -> None:
        """Perform health checks on all connections"""
        connections_to_check = []
        
        async with self._lock:
            # Get available connections for health check
            connections_to_check = [
                conn for conn in self._pool.values()
                if conn.state == ConnectionState.AVAILABLE
            ]
        
        # Check connections concurrently
        health_check_tasks = [
            self._health_check_connection(conn) 
            for conn in connections_to_check
        ]
        
        if health_check_tasks:
            results = await asyncio.gather(*health_check_tasks, return_exceptions=True)
            
            failed_count = sum(1 for result in results if result is False)
            if failed_count > 0:
                logger.info(f"Health check completed: {failed_count} connections failed")
    
    async def _cleanup_loop(self) -> None:
        """Background cleanup loop"""
        while True:
            try:
                await asyncio.sleep(self.health_check_interval_seconds * 2)  # Run less frequently
                await self._cleanup_expired_connections()
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Error in cleanup loop: {e}")
    
    async def _cleanup_expired_connections(self) -> None:
        """Clean up expired and idle connections"""
        connections_to_remove = []
        
        async with self._lock:
            for connection_id, connection in self._pool.items():
                # Remove expired connections
                if connection.is_expired(self.max_connection_age_seconds):
                    connections_to_remove.append(connection_id)
                    continue
                
                # Remove idle connections (but keep minimum pool size)
                if (len(self._pool) > self.min_pool_size and
                    connection.state == ConnectionState.AVAILABLE and
                    connection.is_idle_too_long(self.max_idle_time_seconds)):
                    connections_to_remove.append(connection_id)
                    continue
                
                # Remove failed connections
                if connection.state == ConnectionState.FAILED:
                    connections_to_remove.append(connection_id)
        
        # Remove identified connections
        for connection_id in connections_to_remove:
            await self._remove_connection(connection_id)
            
        if connections_to_remove:
            logger.info(f"Cleaned up {len(connections_to_remove)} connections")
        
        # Ensure minimum pool size
        async with self._lock:
            current_healthy_count = sum(
                1 for conn in self._pool.values()
                if conn.state in [ConnectionState.AVAILABLE, ConnectionState.IN_USE]
            )
            
            if current_healthy_count < self.min_pool_size:
                connections_to_create = self.min_pool_size - current_healthy_count
                for _ in range(connections_to_create):
                    try:
                        await self._create_connection()
                    except Exception as e:
                        logger.error(f"Failed to create connection during cleanup: {e}")
                        break
    
    def _update_wait_time_stats(self, wait_time_ms: float) -> None:
        """Update wait time statistics"""
        if self._stats.total_requests == 1:
            self._stats.avg_wait_time_ms = wait_time_ms
        else:
            # Running average
            self._stats.avg_wait_time_ms = (
                (self._stats.avg_wait_time_ms * (self._stats.total_requests - 1) + wait_time_ms) /
                self._stats.total_requests
            )
        
        self._stats.max_wait_time_ms = max(self._stats.max_wait_time_ms, wait_time_ms)
    
    def _trigger_connection_event(self, event_type: str, connection: PooledConnection) -> None:
        """Trigger connection event callbacks"""
        for callback in self._connection_event_callbacks:
            try:
                callback(event_type, connection)
            except Exception as e:
                logger.error(f"Error in connection event callback: {e}")


class ConnectionPoolExhausted(Exception):
    """Exception raised when connection pool is exhausted"""
    pass


# Global connection pool manager instance
_pool_manager: Optional[ConnectionPoolManager] = None


async def get_pool_manager(mongo_uri: str = None, db_name: str = None) -> ConnectionPoolManager:
    """Get global connection pool manager instance"""
    global _pool_manager
    if _pool_manager is None:
        if not mongo_uri or not db_name:
            raise ValueError("mongo_uri and db_name required for first-time initialization")
        _pool_manager = ConnectionPoolManager(mongo_uri, db_name)
        await _pool_manager.initialize()
    return _pool_manager