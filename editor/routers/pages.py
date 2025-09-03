"""Simplified router for D&D 5e SRD Editor with consistent error handling."""

from typing import Optional
from urllib.parse import urlencode

from fastapi import APIRouter, HTTPException, Query, Request
from fastapi.responses import HTMLResponse
from pydantic import ValidationError as PydanticValidationError

from core.config import COLLECTIONS, get_collection_label, is_valid_collection
from core.database import health_check
from services.content_service import get_content_service
from core.templates import env
import logging

router = APIRouter()
logger = logging.getLogger(__name__)


class AppError(Exception):
    """Base application error."""
    def __init__(self, message: str, status_code: int = 500, user_message: str = None):
        self.message = message
        self.status_code = status_code
        self.user_message = user_message or message
        super().__init__(message)


def handle_error(error: Exception, template: str = "error.html") -> HTMLResponse:
    """Consistent error handling."""
    if isinstance(error, AppError):
        status_code = error.status_code
        user_message = error.user_message
        logger.warning(f"App error: {error.message}")
    else:
        status_code = 500
        user_message = "Si è verificato un errore imprevisto."
        logger.error(f"Unexpected error: {str(error)}", exc_info=True)
    
    try:
        error_template = env.get_template(template)
        content = error_template.render(
            error_message=user_message,
            status_code=status_code
        )
        return HTMLResponse(content, status_code=status_code)
    except Exception:
        # Fallback to simple error response
        return HTMLResponse(
            f"<h1>Errore {status_code}</h1><p>{user_message}</p>",
            status_code=status_code
        )


@router.get("/", response_class=HTMLResponse)
async def homepage(page: Optional[int] = Query(default=1, ge=1, le=1000)):
    """Homepage with collection overview."""
    try:
        service = await get_content_service()
        
        # Get collection counts
        counts = await service.get_collection_counts()
        
        # Prepare collections data
        collections_data = []
        for collection in COLLECTIONS:
            collections_data.append({
                "name": collection,
                "label": get_collection_label(collection),
                "count": counts.get(collection, 0)
            })
        
        # Sort by label
        collections_data.sort(key=lambda x: x["label"])
        
        # Calculate total count
        total_count = sum(counts.values())
        
        template = env.get_template("index.html")
        content = template.render(
            collections=collections_data,
            total=total_count,
            page=page
        )
        
        return HTMLResponse(content)
        
    except Exception as e:
        return handle_error(e)


@router.get("/list/{collection}", response_class=HTMLResponse)
async def list_collection(
    collection: str,
    request: Request,
    q: Optional[str] = Query(default=None, max_length=200),
    page: int = Query(default=1, ge=1, le=1000),
    page_size: int = Query(default=20, ge=5, le=100)
):
    """List documents in collection with search and filtering."""
    try:
        # Handle case where collection might be a serialized dictionary
        if collection.startswith("{") and "name" in collection:
            try:
                # Attempt to extract collection name from dictionary string
                import re
                match = re.search(r"'name':\s*'([^']+)'", collection)
                if match:
                    collection = match.group(1)
                    logger.warning(f"Extracted collection name '{collection}' from malformed URL parameter")
            except Exception as e:
                logger.error(f"Failed to parse collection parameter: {e}")
        
        # Validate collection
        if not is_valid_collection(collection):
            raise AppError(
                f"Collection '{collection}' not found",
                status_code=404,
                user_message=f"Collezione '{collection}' non trovata."
            )
        
        service = await get_content_service()
        
        # Extract filters from query parameters
        filters = {}
        for key, value in request.query_params.items():
            if key not in ["q", "page", "page_size"] and value:
                filters[key] = value
        
        # Get documents
        documents, total, has_prev, has_next = await service.list_documents(
            collection=collection,
            query=q,
            filters=filters,
            page=page,
            page_size=page_size
        )
        
        # Calculate pagination info
        total_pages = (total + page_size - 1) // page_size
        
        # Build query string for navigation
        params = dict(request.query_params)
        params.pop("page", None)  # Remove page from query string
        query_string = urlencode(params) if params else ""
        
        template = env.get_template("list.html")
        content = template.render(
            collection=collection,
            collection_label=get_collection_label(collection),
            documents=documents,
            q=q,
            page=page,
            page_size=page_size,
            total=total,
            total_pages=total_pages,
            has_prev=has_prev,
            has_next=has_next,
            qs=query_string,
            request=request
        )
        
        return HTMLResponse(content)
        
    except AppError:
        raise
    except Exception as e:
        return handle_error(e)


@router.get("/show/{collection}/{slug}", response_class=HTMLResponse)
async def show_document(
    collection: str,
    slug: str,
    request: Request
):
    """Show single document."""
    try:
        # Validate collection
        if not is_valid_collection(collection):
            raise AppError(
                f"Collection '{collection}' not found",
                status_code=404,
                user_message=f"Collezione '{collection}' non trovata."
            )
        
        service = await get_content_service()
        
        # Get document
        doc = await service.get_document(collection, slug)
        if not doc:
            raise AppError(
                f"Document '{slug}' not found in collection '{collection}'",
                status_code=404,
                user_message=f"Documento '{slug}' non trovato."
            )
        
        # Get navigation context (prev/next)
        filters = {}
        for key, value in request.query_params.items():
            if value:
                filters[key] = value
        
        prev_slug, next_slug = await service.get_navigation_context(
            collection, slug, filters
        )
        
        # Render document content
        body_html, body_raw = await service.render_document_content(doc)
        
        # Get document title
        doc_title = (
            doc.get("name") or 
            doc.get("nome") or 
            doc.get("title") or 
            doc.get("titolo") or 
            slug
        )
        
        # Build query string for navigation
        query_string = urlencode(request.query_params) if request.query_params else ""
        
        template = env.get_template("show.html")
        content = template.render(
            collection=collection,
            collection_label=get_collection_label(collection),
            doc_obj=doc,
            doc_id=slug,
            doc_title=doc_title,
            body_html=body_html,
            body_raw=body_raw,
            prev_id=prev_slug,
            next_id=next_slug,
            qs=query_string
        )
        
        return HTMLResponse(content)
        
    except AppError:
        raise
    except Exception as e:
        return handle_error(e)


@router.get("/view/{collection}", response_class=HTMLResponse)
async def view_collection_htmx(
    collection: str,
    request: Request,
    q: Optional[str] = Query(default=None),
    page: int = Query(default=1, ge=1),
    page_size: int = Query(default=20, ge=5, le=100)
):
    """HTMX endpoint for dynamic collection loading."""
    try:
        # Validate collection
        if not is_valid_collection(collection):
            raise AppError(
                f"Collection '{collection}' not found",
                status_code=404,
                user_message=f"Collezione '{collection}' non trovata."
            )
        
        service = await get_content_service()
        
        # Extract filters
        filters = {}
        for key, value in request.query_params.items():
            if key not in ["q", "page", "page_size"] and value:
                filters[key] = value
        
        # Get documents
        documents, total, has_prev, has_next = await service.list_documents(
            collection=collection,
            query=q,
            filters=filters,
            page=page,
            page_size=page_size
        )
        
        # Calculate pagination
        total_pages = (total + page_size - 1) // page_size
        start_item = (page - 1) * page_size + 1
        end_item = min(page * page_size, total)
        
        # Build query string
        params = dict(request.query_params)
        params.pop("page", None)
        query_string = urlencode(params) if params else ""
        
        template = env.get_template("_rows.html")
        content = template.render(
            collection=collection,
            documents=documents,
            page=page,
            page_size=page_size,
            total=total,
            total_pages=total_pages,
            has_prev=has_prev,
            has_next=has_next,
            start_item=start_item,
            end_item=end_item,
            qs=query_string
        )
        
        return HTMLResponse(content)
        
    except AppError:
        raise
    except Exception as e:
        return handle_error(e, "_error.html")


@router.get("/health")
async def health():
    """Simple health check endpoint."""
    try:
        db_healthy = await health_check()
        
        if db_healthy:
            return {"status": "healthy", "database": True}
        else:
            return {"status": "unhealthy", "database": False}
            
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        return {"status": "unhealthy", "database": False, "error": str(e)}