"""Error handling system for D&D 5e SRD Editor."""
from __future__ import annotations

import traceback
from enum import Enum
from typing import Any, Dict, Optional
from fastapi import HTTPException, Request
from fastapi.responses import HTMLResponse, JSONResponse
from core.logging_config import get_logger

logger = get_logger(__name__)


class ErrorCode(Enum):
    """Application error codes for structured error handling."""
    
    # Database errors (DB_xxx)
    DATABASE_CONNECTION_FAILED = "DB_001"
    DATABASE_OPERATION_FAILED = "DB_002" 
    DATABASE_TIMEOUT = "DB_003"
    INDEX_CREATION_FAILED = "DB_004"
    
    # Collection/Document errors (DOC_xxx)
    DOCUMENT_NOT_FOUND = "DOC_001"
    DOCUMENT_INVALID_FORMAT = "DOC_002"
    COLLECTION_NOT_FOUND = "COL_001"
    INVALID_COLLECTION_NAME = "COL_002"
    
    # Validation errors (VAL_xxx)
    INVALID_INPUT = "VAL_001"
    MISSING_REQUIRED_FIELD = "VAL_002"
    INVALID_QUERY_PARAMETER = "VAL_003"
    PAGINATION_ERROR = "VAL_004"
    
    # Template/Rendering errors (TPL_xxx)
    TEMPLATE_NOT_FOUND = "TPL_001"
    TEMPLATE_RENDER_ERROR = "TPL_002"
    
    # Language/Localization errors (LANG_xxx)
    INVALID_LANGUAGE = "LANG_001"
    TRANSLATION_NOT_FOUND = "LANG_002"
    
    # Search errors (SEARCH_xxx)
    SEARCH_QUERY_INVALID = "SEARCH_001"
    SEARCH_ENGINE_ERROR = "SEARCH_002"
    
    # General system errors (SYS_xxx)
    INTERNAL_SERVER_ERROR = "SYS_001"
    SERVICE_UNAVAILABLE = "SYS_002"
    CONFIGURATION_ERROR = "SYS_003"


class ApplicationError(Exception):
    """Base application exception with structured error information."""
    
    def __init__(
        self,
        message: str,
        error_code: ErrorCode,
        details: Optional[Dict[str, Any]] = None,
        status_code: int = 500,
        user_message: Optional[str] = None,
        context: Optional[Dict[str, Any]] = None
    ):
        """Initialize application error.
        
        Args:
            message: Technical error message for logging
            error_code: Structured error code
            details: Additional error details
            status_code: HTTP status code
            user_message: User-friendly error message
            context: Additional context for debugging
        """
        self.message = message
        self.error_code = error_code
        self.details = details or {}
        self.status_code = status_code
        self.user_message = user_message or self._get_default_user_message()
        self.context = context or {}
        super().__init__(message)
    
    def _get_default_user_message(self) -> str:
        """Get default user-friendly message based on error code."""
        user_messages = {
            ErrorCode.DATABASE_CONNECTION_FAILED: "Impossibile connettersi al database. Riprova più tardi.",
            ErrorCode.DATABASE_OPERATION_FAILED: "Errore durante l'operazione sul database.",
            ErrorCode.DOCUMENT_NOT_FOUND: "Documento non trovato.",
            ErrorCode.COLLECTION_NOT_FOUND: "Collezione non trovata.",
            ErrorCode.INVALID_COLLECTION_NAME: "Nome collezione non valido.",
            ErrorCode.INVALID_INPUT: "Dati di input non validi.",
            ErrorCode.INVALID_QUERY_PARAMETER: "Parametri di ricerca non validi.",
            ErrorCode.TEMPLATE_NOT_FOUND: "Template non trovato.",
            ErrorCode.INVALID_LANGUAGE: "Lingua non supportata.",
            ErrorCode.SEARCH_QUERY_INVALID: "Query di ricerca non valida.",
        }
        return user_messages.get(self.error_code, "Si è verificato un errore interno.")
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert error to dictionary for JSON responses."""
        return {
            "error": True,
            "error_code": self.error_code.value,
            "message": self.user_message,
            "details": self.details,
            "context": self.context
        }


class DatabaseError(ApplicationError):
    """Database-related errors."""
    
    def __init__(
        self,
        message: str,
        error_code: ErrorCode = ErrorCode.DATABASE_OPERATION_FAILED,
        operation: Optional[str] = None,
        collection: Optional[str] = None,
        **kwargs
    ):
        context = kwargs.pop("context", {})
        if operation:
            context["operation"] = operation
        if collection:
            context["collection"] = collection
            
        super().__init__(message, error_code, context=context, **kwargs)


class ValidationError(ApplicationError):
    """Input validation errors."""
    
    def __init__(
        self,
        message: str,
        field: Optional[str] = None,
        value: Optional[Any] = None,
        **kwargs
    ):
        context = kwargs.pop("context", {})
        if field:
            context["field"] = field
        if value is not None:
            context["invalid_value"] = str(value)
            
        super().__init__(
            message,
            ErrorCode.INVALID_INPUT,
            status_code=400,
            context=context,
            **kwargs
        )


class NotFoundError(ApplicationError):
    """Resource not found errors."""
    
    def __init__(
        self,
        message: str,
        resource_type: Optional[str] = None,
        resource_id: Optional[str] = None,
        **kwargs
    ):
        context = kwargs.pop("context", {})
        if resource_type:
            context["resource_type"] = resource_type
        if resource_id:
            context["resource_id"] = resource_id
            
        super().__init__(
            message,
            ErrorCode.DOCUMENT_NOT_FOUND,
            status_code=404,
            context=context,
            **kwargs
        )


# Error handler functions
async def application_error_handler(request: Request, exc: ApplicationError) -> JSONResponse:
    """Handle ApplicationError exceptions."""
    
    # Log error with full context
    logger.error(
        f"Application error: {exc.message}",
        extra={
            "error_code": exc.error_code.value,
            "status_code": exc.status_code,
            "details": exc.details,
            "context": exc.context,
            "request_path": str(request.url.path),
            "request_method": request.method,
        },
        exc_info=exc
    )
    
    return JSONResponse(
        status_code=exc.status_code,
        content=exc.to_dict()
    )


async def generic_exception_handler(request: Request, exc: Exception) -> JSONResponse:
    """Handle generic exceptions."""
    
    error_id = f"error_{id(request)}"
    
    logger.error(
        f"Unhandled exception [{error_id}]: {str(exc)}",
        extra={
            "error_id": error_id,
            "error_type": type(exc).__name__,
            "request_path": str(request.url.path),
            "request_method": request.method,
            "traceback": traceback.format_exc(),
        },
        exc_info=exc
    )
    
    # Don't expose internal errors to users
    return JSONResponse(
        status_code=500,
        content={
            "error": True,
            "error_code": ErrorCode.INTERNAL_SERVER_ERROR.value,
            "message": "Si è verificato un errore interno del server.",
            "error_id": error_id,
        }
    )


async def http_exception_handler(request: Request, exc: HTTPException) -> JSONResponse:
    """Handle FastAPI HTTP exceptions."""
    
    logger.warning(
        f"HTTP exception: {exc.status_code} - {exc.detail}",
        extra={
            "status_code": exc.status_code,
            "detail": exc.detail,
            "request_path": str(request.url.path),
            "request_method": request.method,
        }
    )
    
    return JSONResponse(
        status_code=exc.status_code,
        content={
            "error": True,
            "error_code": f"HTTP_{exc.status_code}",
            "message": exc.detail,
        }
    )


def safe_operation(
    operation_name: str,
    error_code: ErrorCode = ErrorCode.INTERNAL_SERVER_ERROR,
    reraise_as: type[ApplicationError] = ApplicationError
):
    """Decorator to safely execute operations with error handling.
    
    Args:
        operation_name: Name of the operation for logging
        error_code: Error code to use if operation fails
        reraise_as: Exception class to reraise as
    """
    def decorator(func):
        async def wrapper(*args, **kwargs):
            try:
                logger.debug(f"Starting operation: {operation_name}")
                result = await func(*args, **kwargs)
                logger.debug(f"Completed operation: {operation_name}")
                return result
                
            except ApplicationError:
                # Re-raise application errors as-is
                raise
                
            except Exception as e:
                logger.error(
                    f"Operation failed: {operation_name}",
                    extra={
                        "operation": operation_name,
                        "error_type": type(e).__name__,
                        "args": str(args)[:200] if args else None,
                        "kwargs": str(kwargs)[:200] if kwargs else None,
                    },
                    exc_info=e
                )
                
                raise reraise_as(
                    f"Operation '{operation_name}' failed: {str(e)}",
                    error_code,
                    context={"operation": operation_name, "original_error": str(e)}
                )
        
        return wrapper
    return decorator


async def safe_db_operation(operation, error_code: ErrorCode, context: str):
    """Safely execute database operations with proper error handling.
    
    Args:
        operation: Async callable to execute
        error_code: Error code for failures
        context: Context description for logging
    
    Returns:
        Operation result
        
    Raises:
        DatabaseError: If operation fails
    """
    try:
        logger.debug(f"Executing database operation: {context}")
        result = await operation()
        logger.debug(f"Database operation completed: {context}")
        return result
        
    except Exception as e:
        logger.error(
            f"Database operation failed: {context}",
            extra={
                "context": context,
                "error_code": error_code.value,
                "error_type": type(e).__name__,
                "original_error": str(e),
            },
            exc_info=e
        )
        
        raise DatabaseError(
            f"Database operation failed: {context}",
            error_code,
            context={"operation_context": context, "original_error": str(e)}
        )