"""Simplified main application for D&D 5e SRD Editor."""

import asyncio
import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from fastapi.middleware.cors import CORSMiddleware

from core.database import close_database
from routers.pages import router as pages_router
# from routers.hexagonal_pages import router as hex_router
# from infrastructure.container import reset_container

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan management."""
    # Startup
    logger.info("Starting D&D 5e SRD Editor...")
    
    yield
    
    # Shutdown
    logger.info("Shutting down D&D 5e SRD Editor...")
    await close_database()
    # await reset_container()


# Create FastAPI application
app = FastAPI(
    title="D&D 5e SRD Editor",
    description="Editor per il System Reference Document di Dungeons & Dragons 5e in italiano",
    version="2.0.0",
    lifespan=lifespan
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure as needed
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Static files
app.mount("/static", StaticFiles(directory="static"), name="static")

# Include routers
app.include_router(pages_router)
# app.include_router(hex_router)  # Hexagonal architecture demo


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main_simple:app",
        host="0.0.0.0",
        port=8000,
        reload=True,
        log_level="info"
    )