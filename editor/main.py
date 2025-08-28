from fastapi import FastAPI
from editor_app.core.db import init_db, close_db
from editor_app.routers.pages import router as pages_router
from fastapi.responses import PlainTextResponse

def create_app() -> FastAPI:
    app = FastAPI(title="D&D SRD Editor")
    app.include_router(pages_router)
    
    @app.get("/healthz", response_class=PlainTextResponse)
    async def healthz():
        return PlainTextResponse("ok")

    @app.on_event("startup")
    async def _startup() -> None:
        await init_db()

    @app.on_event("shutdown")
    async def _shutdown() -> None:
        await close_db()

    return app

app = create_app()
