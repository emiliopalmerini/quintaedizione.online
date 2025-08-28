from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.responses import PlainTextResponse
from core.db import init_db, close_db
from routers.pages import router as pages_router

def create_app() -> FastAPI:
    @asynccontextmanager
    async def lifespan(app: FastAPI):
        await init_db()
        try:
            yield
        finally:
            await close_db()

    app = FastAPI(title="D&D SRD Editor", lifespan=lifespan)
    app.include_router(pages_router)

    @app.get("/healthz", response_class=PlainTextResponse)
    async def healthz():
        return PlainTextResponse("ok")

    return app

app = create_app()
