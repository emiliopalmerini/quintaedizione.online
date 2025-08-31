from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException
from fastapi.responses import PlainTextResponse, JSONResponse
from core.database import init_db, close_db, close_connection_manager
from core.cache import close_cache_manager
from core.logging_config import setup_logging, get_logger
from core.errors import (
    ApplicationError,
    application_error_handler,
    generic_exception_handler,
    http_exception_handler,
)
from core.health import get_health_manager, check_system_health, is_system_healthy
from routers.pages import router as pages_router
import os
import time

def create_app() -> FastAPI:
    # Initialize logging
    log_level = os.getenv("LOG_LEVEL", "INFO")
    log_file = os.getenv("LOG_FILE")
    structured_logging = os.getenv("STRUCTURED_LOGGING", "true").lower() == "true"
    
    setup_logging(log_level, log_file, structured_logging)
    logger = get_logger(__name__)
    logger.info("Starting D&D 5e SRD Editor application")
    
    @asynccontextmanager
    async def lifespan(app: FastAPI):
        logger.info("Application startup - initializing database")
        try:
            await init_db()
            logger.info("Database initialized successfully")
        except Exception as e:
            logger.error("Failed to initialize database", exc_info=e)
            raise
            
        try:
            yield
        finally:
            logger.info("Application shutdown - closing connections")
            await close_db()
            await close_connection_manager()
            await close_cache_manager()
            logger.info("Application shutdown complete")

    app = FastAPI(
        title="D&D SRD Editor", 
        description="D&D 5e System Reference Document Editor",
        version="1.0.0",
        lifespan=lifespan
    )
    
    # Add exception handlers
    app.add_exception_handler(ApplicationError, application_error_handler)
    app.add_exception_handler(HTTPException, http_exception_handler)
    app.add_exception_handler(Exception, generic_exception_handler)
    
    app.include_router(pages_router)

    @app.get("/healthz", response_class=PlainTextResponse)
    async def healthz():
        """Simple health check endpoint for load balancers."""
        try:
            is_healthy = await is_system_healthy()
            if is_healthy:
                return PlainTextResponse("ok", status_code=200)
            else:
                return PlainTextResponse("unhealthy", status_code=503)
        except Exception:
            return PlainTextResponse("error", status_code=503)
    
    @app.get("/health", response_class=JSONResponse)
    async def health():
        """Basic health check with JSON response."""
        try:
            health_manager = get_health_manager()
            status = await health_manager.get_quick_status()
            
            status_code = 200 if status["status"] in ["healthy", "degraded"] else 503
            return JSONResponse(content=status, status_code=status_code)
        except Exception as e:
            logger.error("Health check failed", exc_info=e)
            return JSONResponse(
                content={
                    "status": "error",
                    "error": str(e),
                    "timestamp": time.time()
                },
                status_code=503
            )
    
    @app.get("/health/detailed", response_class=JSONResponse)
    async def health_detailed():
        """Detailed health check with full diagnostics."""
        try:
            system_status = await check_system_health(include_details=True)
            
            response_data = {
                "status": system_status.status.value,
                "timestamp": system_status.timestamp,
                "uptime_seconds": system_status.uptime_seconds,
                "version": system_status.version,
                "checks": [
                    {
                        "name": check.name,
                        "status": check.status.value,
                        "message": check.message,
                        "duration_ms": check.duration_ms,
                        "error": check.error
                    }
                    for check in system_status.checks
                ],
                "summary": system_status.summary
            }
            
            status_code = 200 if system_status.status.value in ["healthy", "degraded"] else 503
            return JSONResponse(content=response_data, status_code=status_code)
            
        except Exception as e:
            logger.error("Detailed health check failed", exc_info=e)
            return JSONResponse(
                content={
                    "status": "error",
                    "error": str(e),
                    "timestamp": time.time()
                },
                status_code=503
            )

    return app

app = create_app()
