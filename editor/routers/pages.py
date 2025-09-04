"""Simplified router for D&D 5e SRD Editor with consistent error handling."""

from typing import Optional
from urllib.parse import urlencode

from fastapi import APIRouter, HTTPException, Query, Request
from fastapi.responses import HTMLResponse
from pydantic import ValidationError as PydanticValidationError

from core.config import COLLECTIONS, get_collection_label, is_valid_collection
from core.database import health_check
from services.content_service import ContentService
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
        service = ContentService()
        
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


@router.get("/debug/markdown", response_class=HTMLResponse)
async def debug_markdown():
    """Debug page for markdown rendering issues."""
    try:
        template = env.get_template("debug_markdown.html")
        content = template.render()
        return HTMLResponse(content)
    except Exception as e:
        return handle_error(e)


@router.get("/{collection}", response_class=HTMLResponse)
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
        
        service = ContentService()
        
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
        start_item = (page - 1) * page_size + 1
        end_item = min(page * page_size, total)
        
        # Build query string for navigation - only include non-empty parameters
        params = {k: v for k, v in request.query_params.items() if k != "page" and v and v.strip()}
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
            start_item=start_item,
            end_item=end_item,
            qs=query_string,
            request=request
        )
        
        return HTMLResponse(content)
        
    except AppError:
        raise
    except Exception as e:
        return handle_error(e)


@router.get("/{collection}/rows", response_class=HTMLResponse)
async def view_collection_htmx(
    collection: str,
    request: Request,
    q: Optional[str] = Query(default=None),
    page: int = Query(default=1, ge=1),
    page_size: int = Query(default=20, ge=5, le=100)
):
    """HTMX endpoint for dynamic collection loading."""
    # Debug to file
    try:
        with open('/tmp/router_debug.log', 'a') as f:
            f.write(f"ROUTER view_collection_htmx called for collection: {collection}\n")
    except Exception:
        pass
    try:
        # Validate collection
        if not is_valid_collection(collection):
            raise AppError(
                f"Collection '{collection}' not found",
                status_code=404,
                user_message=f"Collezione '{collection}' non trovata."
            )
        
        service = ContentService()
        
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
        print(f"ROUTER view_collection_htmx: Got {len(documents)} documents for {collection}")
        
        # Calculate pagination
        total_pages = (total + page_size - 1) // page_size
        start_item = (page - 1) * page_size + 1
        end_item = min(page * page_size, total)
        
        # Build query string - only include non-empty parameters
        params = {k: v for k, v in request.query_params.items() if k != "page" and v and v.strip()}
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


@router.get("/{collection}/{document_id}", response_class=HTMLResponse)
async def show_document(
    collection: str,
    document_id: str,
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
        
        service = ContentService()
        
        # Extract user context for events
        user_agent = request.headers.get("user-agent", "")
        referrer = request.headers.get("referer", "")
        
        # Get document
        doc = await service.get_document(collection, document_id)
        if not doc:
            raise AppError(
                f"Document '{document_id}' not found in collection '{collection}'",
                status_code=404,
                user_message=f"Documento '{document_id}' non trovato."
            )
        
        # Get navigation context (prev/next)
        query = request.query_params.get("q")
        filters = {}
        for key, value in request.query_params.items():
            if key not in ["q", "page", "page_size"] and value:
                filters[key] = value
        
        # Get navigation context 
        prev_slug, next_slug = await service.get_navigation_context(collection, document_id, query, filters)
        
        # Render document content
        body_html, body_raw = await service.render_document_content(doc)
        
        # Get document title
        doc_title = (
            doc.get("name") or 
            doc.get("nome") or 
            doc.get("title") or 
            doc.get("titolo") or 
            document_id
        )
        
        # Build query string for navigation - only include non-empty parameters
        filtered_params = {k: v for k, v in request.query_params.items() if v and v.strip()}
        query_string = urlencode(filtered_params) if filtered_params else ""
        
        template = env.get_template("show.html")
        content = template.render(
            collection=collection,
            collection_label=get_collection_label(collection),
            doc_obj=doc,
            doc_id=document_id,
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


@router.get("/api/filter-options/{collection}/{field}")
async def get_filter_field_options(collection: str, field: str, request: Request):
    """API endpoint to get dropdown options for a specific field."""
    try:
        # Validate collection
        if not is_valid_collection(collection):
            return HTMLResponse("<option value=''>Errore: collezione non trovata</option>")
        
        service = ContentService()
        # TODO: get_distinct_values not implemented in old service
        values = []
        
        # Get current selected value from request
        current_value = request.query_params.get(field, "")
        
        # Generate HTML options
        options_html = '<option value="">Tutte</option>'
        for value in values[:50]:  # Limit to 50 options for performance
            if value:
                selected = 'selected' if str(value) == current_value else ''
                options_html += f'<option value="{value}" {selected}>{value}</option>'
        
        return HTMLResponse(options_html)
        
    except Exception as e:
        logger.error(f"Failed to get options for {collection}.{field}: {e}")
        return HTMLResponse("<option value=''>Errore nel caricamento</option>")


@router.get("/test")
async def test_endpoint():
    """Test endpoint to verify routing works."""
    return {"message": "Test endpoint working"}


@router.get("/quicksearch/{collection}", response_class=HTMLResponse)
async def quicksearch_collection(
    collection: str,
    q: Optional[str] = Query(default=None, max_length=200)
):
    """HTMX endpoint for quick search in collection."""
    try:
        # Validate collection
        if not is_valid_collection(collection):
            raise AppError(
                f"Collection '{collection}' not found",
                status_code=404,
                user_message=f"Collezione '{collection}' non trovata."
            )
        
        items = []
        if q and q.strip():
            service = ContentService()
            # Get first 10 results for quick search
            documents, _, _, _ = await service.list_documents(
                collection=collection,
                query=q.strip(),
                page=1,
                page_size=10
            )
            items = documents
        
        template = env.get_template("_quicksearch.html")
        content = template.render(
            collection=collection,
            q=q,
            items=items
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