"""Advanced health check system for D&D 5e SRD Editor."""
from __future__ import annotations

import asyncio
import time
from dataclasses import dataclass, asdict
from enum import Enum
from typing import Any, Dict, List, Optional, Callable, Awaitable
from contextlib import asynccontextmanager

from core.database import get_connection_manager, DatabaseConnectionManager
from core.cache import get_cache_manager, CacheManager
from core.logging_config import get_logger
from core.errors import ApplicationError, ErrorCode

logger = get_logger(__name__)


class HealthStatus(Enum):
    """Health check status values."""
    
    HEALTHY = "healthy"
    UNHEALTHY = "unhealthy"
    DEGRADED = "degraded"
    UNKNOWN = "unknown"


@dataclass
class HealthCheckResult:
    """Result of an individual health check."""
    
    name: str
    status: HealthStatus
    message: str
    duration_ms: float
    timestamp: float
    details: Optional[Dict[str, Any]] = None
    error: Optional[str] = None


@dataclass 
class SystemHealthStatus:
    """Overall system health status."""
    
    status: HealthStatus
    timestamp: float
    checks: List[HealthCheckResult]
    summary: Dict[str, Any]
    uptime_seconds: float
    version: str = "1.0.0"


class HealthChecker:
    """Base class for health checkers."""
    
    def __init__(self, name: str, timeout: float = 10.0):
        self.name = name
        self.timeout = timeout
    
    async def check(self) -> HealthCheckResult:
        """Perform health check."""
        start_time = time.time()
        timestamp = start_time
        
        try:
            async with asyncio.timeout(self.timeout):
                await self.execute_check()
            
            duration = (time.time() - start_time) * 1000
            
            return HealthCheckResult(
                name=self.name,
                status=HealthStatus.HEALTHY,
                message=f"{self.name} is healthy",
                duration_ms=duration,
                timestamp=timestamp
            )
            
        except asyncio.TimeoutError:
            duration = (time.time() - start_time) * 1000
            return HealthCheckResult(
                name=self.name,
                status=HealthStatus.UNHEALTHY,
                message=f"{self.name} check timed out after {self.timeout}s",
                duration_ms=duration,
                timestamp=timestamp,
                error="timeout"
            )
            
        except Exception as e:
            duration = (time.time() - start_time) * 1000
            return HealthCheckResult(
                name=self.name,
                status=HealthStatus.UNHEALTHY,
                message=f"{self.name} check failed: {str(e)}",
                duration_ms=duration,
                timestamp=timestamp,
                error=str(e)
            )
    
    async def execute_check(self) -> None:
        """Override this method to implement specific health check logic."""
        raise NotImplementedError


class DatabaseHealthChecker(HealthChecker):
    """Health checker for MongoDB database."""
    
    def __init__(self, timeout: float = 10.0):
        super().__init__("database", timeout)
        self.connection_manager: Optional[DatabaseConnectionManager] = None
    
    async def execute_check(self) -> None:
        """Check database health."""
        try:
            # Get connection manager
            self.connection_manager = await get_connection_manager()
            
            # Get database instance
            db = await self.connection_manager.get_database()
            
            # Test basic connectivity
            await db.command("ping")
            
            # Test a simple query on a known collection
            collections = await db.list_collection_names()
            if not collections:
                raise Exception("No collections found in database")
            
            # Test query performance on documenti collection
            start_query = time.time()
            count = await db.documenti.estimated_document_count()
            query_time = (time.time() - start_query) * 1000
            
            logger.debug(f"Database health check passed: {count} documents, query time: {query_time:.2f}ms")
            
        except Exception as e:
            logger.error(f"Database health check failed: {e}")
            raise
    
    async def get_detailed_status(self) -> Dict[str, Any]:
        """Get detailed database status."""
        if not self.connection_manager:
            return {"error": "Connection manager not initialized"}
        
        return await self.connection_manager.get_health_status()


class CacheHealthChecker(HealthChecker):
    """Health checker for Redis cache."""
    
    def __init__(self, timeout: float = 5.0):
        super().__init__("cache", timeout)
        self.cache_manager: Optional[CacheManager] = None
    
    async def execute_check(self) -> None:
        """Check cache health."""
        try:
            # Get cache manager
            self.cache_manager = await get_cache_manager()
            
            # Test cache operations
            test_key = "health_check_test"
            test_value = {"timestamp": time.time(), "test": True}
            
            # Test set
            set_result = await self.cache_manager.set(test_key, test_value, ttl=60)
            if not set_result:
                raise Exception("Failed to set test value in cache")
            
            # Test get
            retrieved_value = await self.cache_manager.get(test_key)
            if retrieved_value != test_value:
                raise Exception("Retrieved value doesn't match set value")
            
            # Test delete
            delete_result = await self.cache_manager.delete(test_key)
            if not delete_result:
                logger.warning("Failed to delete test key from cache (non-critical)")
            
            logger.debug("Cache health check passed")
            
        except Exception as e:
            logger.error(f"Cache health check failed: {e}")
            raise
    
    async def get_detailed_status(self) -> Dict[str, Any]:
        """Get detailed cache status."""
        if not self.cache_manager:
            return {"error": "Cache manager not initialized"}
        
        return await self.cache_manager.get_health_status()


class ApplicationHealthChecker(HealthChecker):
    """Health checker for application-specific functionality."""
    
    def __init__(self, timeout: float = 15.0):
        super().__init__("application", timeout)
    
    async def execute_check(self) -> None:
        """Check application health."""
        try:
            # Check if we can import core modules
            from core.config import COLLECTIONS, MONGO_URI, DB_NAME
            if not COLLECTIONS:
                raise Exception("No collections configured")
            
            if not MONGO_URI or not DB_NAME:
                raise Exception("Database configuration missing")
            
            # Test that we can create query service
            from core.database import get_database
            from core.optimized_queries import create_optimized_query_service
            
            db = await get_database()
            query_service = create_optimized_query_service(db)
            
            # Test batch collection counts (lightweight operation)
            sample_collections = COLLECTIONS[:3]  # Test first 3 collections
            counts = await query_service.get_collection_counts_batch(sample_collections, "it")
            
            if not isinstance(counts, dict):
                raise Exception("Collection count query returned unexpected result")
            
            logger.debug(f"Application health check passed: {len(counts)} collections checked")
            
        except Exception as e:
            logger.error(f"Application health check failed: {e}")
            raise


class PerformanceHealthChecker(HealthChecker):
    """Health checker for system performance metrics."""
    
    def __init__(self, timeout: float = 10.0):
        super().__init__("performance", timeout)
    
    async def execute_check(self) -> None:
        """Check performance metrics."""
        try:
            import psutil
            import os
            
            # CPU usage
            cpu_percent = psutil.cpu_percent(interval=1)
            if cpu_percent > 90:
                raise Exception(f"High CPU usage: {cpu_percent}%")
            
            # Memory usage
            memory = psutil.virtual_memory()
            if memory.percent > 90:
                raise Exception(f"High memory usage: {memory.percent}%")
            
            # Disk usage
            disk = psutil.disk_usage('/')
            if disk.percent > 90:
                raise Exception(f"High disk usage: {disk.percent}%")
            
            # Check if we can create files (disk write test)
            test_file = "/tmp/health_check_write_test"
            try:
                with open(test_file, 'w') as f:
                    f.write("health check test")
                os.remove(test_file)
            except Exception:
                raise Exception("Disk write test failed")
            
            logger.debug(f"Performance health check passed: CPU {cpu_percent}%, Memory {memory.percent}%, Disk {disk.percent}%")
            
        except ImportError:
            # psutil not available, skip performance checks
            logger.warning("psutil not available, skipping performance checks")
        except Exception as e:
            logger.error(f"Performance health check failed: {e}")
            raise


class DependencyHealthChecker(HealthChecker):
    """Health checker for external dependencies."""
    
    def __init__(self, timeout: float = 10.0):
        super().__init__("dependencies", timeout)
    
    async def execute_check(self) -> None:
        """Check external dependencies."""
        try:
            # Check required Python packages
            required_packages = [
                "fastapi", "motor", "pymongo", "jinja2", 
                "pydantic", "uvicorn"
            ]
            
            for package in required_packages:
                try:
                    __import__(package)
                except ImportError:
                    raise Exception(f"Required package '{package}' not available")
            
            # Check optional packages
            optional_packages = ["redis"]
            available_optional = []
            for package in optional_packages:
                try:
                    __import__(package)
                    available_optional.append(package)
                except ImportError:
                    pass
            
            logger.debug(f"Dependencies health check passed: {len(required_packages)} required, {len(available_optional)} optional packages available")
            
        except Exception as e:
            logger.error(f"Dependencies health check failed: {e}")
            raise


class HealthCheckManager:
    """Manager for coordinating health checks."""
    
    def __init__(self):
        self.checkers: List[HealthChecker] = []
        self.startup_time = time.time()
        self._last_check_results: Optional[SystemHealthStatus] = None
        self._check_lock = asyncio.Lock()
    
    def register_checker(self, checker: HealthChecker) -> None:
        """Register a health checker."""
        self.checkers.append(checker)
        logger.debug(f"Registered health checker: {checker.name}")
    
    def register_default_checkers(self) -> None:
        """Register default health checkers."""
        self.register_checker(DatabaseHealthChecker())
        self.register_checker(CacheHealthChecker())
        self.register_checker(ApplicationHealthChecker())
        self.register_checker(PerformanceHealthChecker())
        self.register_checker(DependencyHealthChecker())
    
    async def check_health(self, include_details: bool = False) -> SystemHealthStatus:
        """Perform all health checks."""
        async with self._check_lock:
            logger.info(f"Starting health check with {len(self.checkers)} checkers")
            start_time = time.time()
            
            # Run all checks concurrently
            check_tasks = [checker.check() for checker in self.checkers]
            results = await asyncio.gather(*check_tasks, return_exceptions=True)
            
            # Process results
            check_results = []
            healthy_count = 0
            degraded_count = 0
            unhealthy_count = 0
            
            for i, result in enumerate(results):
                if isinstance(result, Exception):
                    # Create error result for failed checks
                    check_result = HealthCheckResult(
                        name=self.checkers[i].name,
                        status=HealthStatus.UNHEALTHY,
                        message=f"Health check failed: {str(result)}",
                        duration_ms=0.0,
                        timestamp=start_time,
                        error=str(result)
                    )
                else:
                    check_result = result
                
                check_results.append(check_result)
                
                # Count statuses
                if check_result.status == HealthStatus.HEALTHY:
                    healthy_count += 1
                elif check_result.status == HealthStatus.DEGRADED:
                    degraded_count += 1
                else:
                    unhealthy_count += 1
            
            # Determine overall system status
            if unhealthy_count > 0:
                overall_status = HealthStatus.UNHEALTHY
            elif degraded_count > 0:
                overall_status = HealthStatus.DEGRADED
            else:
                overall_status = HealthStatus.HEALTHY
            
            # Calculate uptime
            uptime_seconds = time.time() - self.startup_time
            
            # Create summary
            summary = {
                "total_checks": len(check_results),
                "healthy": healthy_count,
                "degraded": degraded_count,
                "unhealthy": unhealthy_count,
                "avg_duration_ms": sum(r.duration_ms for r in check_results) / len(check_results) if check_results else 0,
                "total_duration_ms": (time.time() - start_time) * 1000
            }
            
            # Add detailed information if requested
            if include_details:
                summary["detailed_status"] = {}
                for checker in self.checkers:
                    if hasattr(checker, 'get_detailed_status'):
                        try:
                            detailed = await checker.get_detailed_status()
                            summary["detailed_status"][checker.name] = detailed
                        except Exception as e:
                            summary["detailed_status"][checker.name] = {"error": str(e)}
            
            system_status = SystemHealthStatus(
                status=overall_status,
                timestamp=start_time,
                checks=check_results,
                summary=summary,
                uptime_seconds=uptime_seconds
            )
            
            self._last_check_results = system_status
            
            logger.info(f"Health check completed: {overall_status.value} "
                       f"({healthy_count} healthy, {degraded_count} degraded, {unhealthy_count} unhealthy)")
            
            return system_status
    
    async def get_quick_status(self) -> Dict[str, Any]:
        """Get quick health status without running full checks."""
        if self._last_check_results is None:
            # Run a quick check if no previous results
            status = await self.check_health()
        else:
            status = self._last_check_results
        
        return {
            "status": status.status.value,
            "timestamp": status.timestamp,
            "uptime_seconds": status.uptime_seconds,
            "summary": status.summary,
            "version": status.version
        }
    
    async def is_healthy(self) -> bool:
        """Quick boolean health check."""
        try:
            status = await self.get_quick_status()
            return status["status"] in ["healthy", "degraded"]
        except Exception:
            return False


# Global health check manager
_health_manager: Optional[HealthCheckManager] = None


def get_health_manager() -> HealthCheckManager:
    """Get global health check manager."""
    global _health_manager
    if _health_manager is None:
        _health_manager = HealthCheckManager()
        _health_manager.register_default_checkers()
    return _health_manager


async def check_system_health(include_details: bool = False) -> SystemHealthStatus:
    """Convenience function to check system health."""
    manager = get_health_manager()
    return await manager.check_health(include_details)


async def is_system_healthy() -> bool:
    """Convenience function to check if system is healthy."""
    manager = get_health_manager()
    return await manager.is_healthy()


# Context manager for health check monitoring
@asynccontextmanager
async def health_monitoring(check_interval: int = 60):
    """Context manager for periodic health monitoring."""
    manager = get_health_manager()
    monitoring_task = None
    
    async def periodic_health_check():
        while True:
            try:
                await asyncio.sleep(check_interval)
                await manager.check_health()
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Error in periodic health check: {e}")
    
    try:
        monitoring_task = asyncio.create_task(periodic_health_check())
        logger.info(f"Started health monitoring with {check_interval}s interval")
        yield manager
    finally:
        if monitoring_task:
            monitoring_task.cancel()
            try:
                await monitoring_task
            except asyncio.CancelledError:
                pass
            logger.info("Stopped health monitoring")