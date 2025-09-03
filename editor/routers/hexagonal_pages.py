"""
Hexagonal architecture router for D&D 5e SRD Editor
Demonstrates clean integration with application layer
"""
from typing import Optional
from urllib.parse import urlencode

from fastapi import APIRouter, HTTPException, Query, Request
from fastapi.responses import HTMLResponse
from pydantic import ValidationError as PydanticValidationError

from editor.infrastructure.container import get_container
from editor.application.query_handlers import (
    SearchClassesQuery,
    GetClassDetailQuery,
    GetClassesByAbilityQuery,
    GetSpellcastingClassesQuery
)
from core.templates import env
from core.logging_config import get_logger

router = APIRouter(prefix="/hex")
logger = get_logger(__name__)


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
        return HTMLResponse(
            f"<h1>Errore {status_code}</h1><p>{user_message}</p>",
            status_code=status_code
        )


@router.get("/", response_class=HTMLResponse)
async def hex_index(request: Request):
    """Hexagonal architecture demo homepage"""
    try:
        template = env.get_template("hex_index.html")
        content = template.render(
            request=request,
            title="D&D 5e SRD - Hexagonal Architecture Demo"
        )
        return HTMLResponse(content)
    except Exception as error:
        return handle_error(error)


@router.get("/classes", response_class=HTMLResponse)
async def hex_classes_list(
    request: Request,
    q: Optional[str] = Query(None, description="Search query"),
    ability: Optional[str] = Query(None, description="Primary ability filter"),
    min_hit_die: Optional[int] = Query(None, description="Minimum hit die"),
    max_hit_die: Optional[int] = Query(None, description="Maximum hit die"),
    spellcaster: Optional[bool] = Query(None, description="Filter spellcasters"),
    page: int = Query(1, ge=1, description="Page number"),
    per_page: int = Query(20, ge=1, le=100, description="Items per page")
):
    """Search and list classes using hexagonal architecture"""
    try:
        container = get_container()
        handler = container.get_search_classes_handler()
        
        # Calculate offset
        offset = (page - 1) * per_page
        
        # Build query
        search_query = SearchClassesQuery(
            text_query=q,
            primary_ability=ability,
            min_hit_die=min_hit_die,
            max_hit_die=max_hit_die,
            is_spellcaster=spellcaster,
            sort_by="name",
            limit=per_page,
            offset=offset
        )
        
        # Execute query
        result = await handler.handle(search_query)
        
        if not result.success:
            raise AppError(result.error, 500, "Errore durante la ricerca delle classi")
        
        classes = result.data or []
        
        # Build pagination info
        has_next = len(classes) == per_page
        has_prev = page > 1
        
        # Build filter query string for pagination
        filter_params = {}
        if q:
            filter_params['q'] = q
        if ability:
            filter_params['ability'] = ability
        if min_hit_die:
            filter_params['min_hit_die'] = min_hit_die
        if max_hit_die:
            filter_params['max_hit_die'] = max_hit_die
        if spellcaster is not None:
            filter_params['spellcaster'] = spellcaster
        
        template = env.get_template("hex_classes_list.html")
        content = template.render(
            request=request,
            classes=classes,
            search_query=q or "",
            ability_filter=ability,
            min_hit_die=min_hit_die,
            max_hit_die=max_hit_die,
            spellcaster_filter=spellcaster,
            page=page,
            per_page=per_page,
            has_next=has_next,
            has_prev=has_prev,
            next_page_url=f"/hex/classes?{urlencode({**filter_params, 'page': page + 1, 'per_page': per_page})}" if has_next else None,
            prev_page_url=f"/hex/classes?{urlencode({**filter_params, 'page': page - 1, 'per_page': per_page})}" if has_prev else None,
            total_results=len(classes),
            metadata=result.metadata
        )
        return HTMLResponse(content)
        
    except AppError:
        raise
    except Exception as error:
        return handle_error(error)


@router.get("/classes/{class_id}", response_class=HTMLResponse) 
async def hex_class_detail(request: Request, class_id: str):
    """Get detailed class information using hexagonal architecture"""
    try:
        container = get_container()
        handler = container.get_class_detail_handler()
        
        query = GetClassDetailQuery(class_id=class_id)
        result = await handler.handle(query)
        
        if not result.success:
            raise AppError(result.error, 500, "Errore durante il caricamento della classe")
        
        if result.data is None:
            raise AppError(f"Class not found: {class_id}", 404, "Classe non trovata")
        
        class_detail = result.data
        
        template = env.get_template("hex_class_detail.html")
        content = template.render(
            request=request,
            class_detail=class_detail,
            metadata=result.metadata
        )
        return HTMLResponse(content)
        
    except AppError:
        raise
    except Exception as error:
        return handle_error(error)


@router.get("/classes-by-ability/{ability}", response_class=HTMLResponse)
async def hex_classes_by_ability(request: Request, ability: str):
    """Get classes filtered by primary ability using hexagonal architecture"""
    try:
        container = get_container()
        handler = container.get_classes_by_ability_handler()
        
        query = GetClassesByAbilityQuery(primary_ability=ability)
        result = await handler.handle(query)
        
        if not result.success:
            raise AppError(result.error, 500, "Errore durante la ricerca per abilità")
        
        classes = result.data or []
        
        template = env.get_template("hex_classes_by_ability.html")
        content = template.render(
            request=request,
            classes=classes,
            ability=ability,
            metadata=result.metadata
        )
        return HTMLResponse(content)
        
    except AppError:
        raise
    except Exception as error:
        return handle_error(error)


@router.get("/spellcasting-classes", response_class=HTMLResponse)
async def hex_spellcasting_classes(request: Request):
    """Get all spellcasting classes using hexagonal architecture"""
    try:
        container = get_container()
        handler = container.get_spellcasting_classes_handler()
        
        query = GetSpellcastingClassesQuery()
        result = await handler.handle(query)
        
        if not result.success:
            raise AppError(result.error, 500, "Errore durante la ricerca degli incantatori")
        
        classes = result.data or []
        
        template = env.get_template("hex_spellcasting_classes.html")
        content = template.render(
            request=request,
            classes=classes,
            metadata=result.metadata
        )
        return HTMLResponse(content)
        
    except AppError:
        raise
    except Exception as error:
        return handle_error(error)