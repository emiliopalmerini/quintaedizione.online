"""Structured logging configuration for D&D 5e SRD Editor."""
from __future__ import annotations

import logging
import logging.config
import sys
import time
from typing import Any, Dict
import json
from pathlib import Path


class StructuredFormatter(logging.Formatter):
    """JSON formatter for structured logging."""
    
    def format(self, record: logging.LogRecord) -> str:
        log_entry = {
            "timestamp": time.time(),
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
            "line": record.lineno,
        }
        
        # Add exception info if present
        if record.exc_info:
            log_entry["exception"] = self.formatException(record.exc_info)
        
        # Add extra fields from record
        extra_fields = {
            key: value 
            for key, value in record.__dict__.items() 
            if key not in {
                'name', 'msg', 'args', 'levelname', 'levelno', 'pathname', 
                'filename', 'module', 'exc_info', 'exc_text', 'stack_info',
                'lineno', 'funcName', 'created', 'msecs', 'relativeCreated',
                'thread', 'threadName', 'processName', 'process', 'message'
            }
        }
        
        if extra_fields:
            log_entry["extra"] = extra_fields
            
        return json.dumps(log_entry, ensure_ascii=False)


class RequestContextFilter(logging.Filter):
    """Add request context to log records."""
    
    def filter(self, record: logging.LogRecord) -> bool:
        # Try to get request context from contextvars if available
        try:
            from contextvars import ContextVar
            request_id = getattr(record, 'request_id', None)
            if request_id:
                record.request_id = request_id
        except (ImportError, AttributeError):
            pass
        return True


def setup_logging(
    log_level: str = "INFO",
    log_file: str | None = None,
    structured: bool = True
) -> None:
    """Configure application logging.
    
    Args:
        log_level: Logging level (DEBUG, INFO, WARNING, ERROR, CRITICAL)
        log_file: Optional log file path
        structured: Whether to use structured JSON logging
    """
    
    level = getattr(logging, log_level.upper(), logging.INFO)
    
    # Create logs directory if needed
    if log_file:
        Path(log_file).parent.mkdir(parents=True, exist_ok=True)
    
    config = {
        "version": 1,
        "disable_existing_loggers": False,
        "formatters": {
            "structured": {
                "()": StructuredFormatter,
            },
            "simple": {
                "format": "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
                "datefmt": "%Y-%m-%d %H:%M:%S",
            },
        },
        "filters": {
            "request_context": {
                "()": RequestContextFilter,
            },
        },
        "handlers": {
            "console": {
                "class": "logging.StreamHandler",
                "stream": sys.stdout,
                "formatter": "structured" if structured else "simple",
                "filters": ["request_context"],
                "level": level,
            },
        },
        "root": {
            "level": level,
            "handlers": ["console"],
        },
        "loggers": {
            # FastAPI and uvicorn loggers
            "uvicorn": {"level": "INFO"},
            "uvicorn.error": {"level": "INFO"},
            "uvicorn.access": {"level": "INFO"},
            "fastapi": {"level": "INFO"},
            
            # MongoDB driver
            "motor": {"level": "WARNING"},
            "pymongo": {"level": "WARNING"},
            
            # Application loggers
            "editor": {"level": level, "propagate": True},
            "editor.routers": {"level": level, "propagate": True},
            "editor.application": {"level": level, "propagate": True},
            "editor.adapters": {"level": level, "propagate": True},
            "editor.core": {"level": level, "propagate": True},
        },
    }
    
    # Add file handler if specified
    if log_file:
        config["handlers"]["file"] = {
            "class": "logging.handlers.RotatingFileHandler",
            "filename": log_file,
            "maxBytes": 10485760,  # 10MB
            "backupCount": 5,
            "formatter": "structured" if structured else "simple",
            "filters": ["request_context"],
            "level": level,
        }
        config["root"]["handlers"].append("file")
    
    logging.config.dictConfig(config)


def get_logger(name: str) -> logging.Logger:
    """Get a logger instance with the given name.
    
    Args:
        name: Logger name, typically __name__
        
    Returns:
        Logger instance
    """
    return logging.getLogger(name)


# Convenience functions for common logging patterns
def log_database_operation(
    logger: logging.Logger,
    operation: str,
    collection: str,
    filter_doc: Dict[str, Any] | None = None,
    duration_ms: float | None = None
) -> None:
    """Log database operations with consistent format."""
    logger.info(
        f"Database operation: {operation}",
        extra={
            "operation": operation,
            "collection": collection,
            "filter": filter_doc,
            "duration_ms": duration_ms,
        }
    )


def log_http_request(
    logger: logging.Logger,
    method: str,
    path: str,
    status_code: int,
    duration_ms: float,
    user_id: str | None = None
) -> None:
    """Log HTTP requests with consistent format."""
    logger.info(
        f"HTTP {method} {path} -> {status_code}",
        extra={
            "http_method": method,
            "http_path": path,
            "http_status": status_code,
            "duration_ms": duration_ms,
            "user_id": user_id,
        }
    )


def log_error_with_context(
    logger: logging.Logger,
    message: str,
    error: Exception,
    context: Dict[str, Any] | None = None
) -> None:
    """Log errors with full context and stack trace."""
    logger.error(
        message,
        exc_info=error,
        extra={"error_context": context or {}}
    )