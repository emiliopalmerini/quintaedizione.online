"""FastAPI dependencies for request validation and processing."""
from __future__ import annotations

from typing import Any, Dict, Optional
from fastapi import HTTPException, Request
from pydantic import ValidationError as PydanticValidationError

from core.errors import ValidationError, ErrorCode
import logging

logger = logging.getLogger(__name__)


def handle_validation_error(e: PydanticValidationError, context: str = "request validation") -> None:
    """Convert Pydantic validation error to ApplicationError.
    
    Args:
        e: Pydantic validation error
        context: Context description for logging
        
    Raises:
        ValidationError: Converted application error
    """
    errors = []
    for error in e.errors():
        field = ".".join(str(x) for x in error["loc"])
        message = error["msg"]
        errors.append(f"{field}: {message}")
    
    error_message = f"Validation failed for {context}: {'; '.join(errors)}"
    
    logger.warning(
        error_message,
        extra={
            "context": context,
            "validation_errors": e.errors()
        }
    )
    
    raise ValidationError(
        error_message,
        details={"validation_errors": e.errors()},
        user_message="I dati forniti non sono validi. Controlla i parametri e riprova."
    )


def safe_query_param_parser(
    model_class, 
    request: Request, 
    context: str = "query parameters"
):
    """Safely parse query parameters using Pydantic model.
    
    Args:
        model_class: Pydantic model class to validate with
        request: FastAPI request object
        context: Context for error messages
        
    Returns:
        Validated model instance
        
    Raises:
        ValidationError: If validation fails
    """
    try:
        query_params = dict(request.query_params)
        return model_class(**query_params)
    except PydanticValidationError as e:
        handle_validation_error(e, context)


def safe_path_param_parser(
    model_class,
    path_params: Dict[str, Any],
    context: str = "path parameters"
):
    """Safely parse path parameters using Pydantic model.
    
    Args:
        model_class: Pydantic model class to validate with
        path_params: Dictionary of path parameters
        context: Context for error messages
        
    Returns:
        Validated model instance
        
    Raises:
        ValidationError: If validation fails
    """
    try:
        return model_class(**path_params)
    except PydanticValidationError as e:
        handle_validation_error(e, context)


def safe_body_parser(
    model_class,
    body_data: Dict[str, Any],
    context: str = "request body"
):
    """Safely parse request body using Pydantic model.
    
    Args:
        model_class: Pydantic model class to validate with
        body_data: Dictionary of body data
        context: Context for error messages
        
    Returns:
        Validated model instance
        
    Raises:
        ValidationError: If validation fails
    """
    try:
        return model_class(**body_data)
    except PydanticValidationError as e:
        handle_validation_error(e, context)