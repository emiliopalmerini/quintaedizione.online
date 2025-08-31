"""Unit tests for error handling system."""
from __future__ import annotations

import pytest
from unittest.mock import MagicMock

from core.errors import (
    ErrorCode,
    ApplicationError,
    DatabaseError,
    ValidationError,
    NotFoundError,
    safe_operation,
)


class TestErrorCode:
    """Test ErrorCode enum."""
    
    def test_error_codes_exist(self):
        """Test that essential error codes exist."""
        assert ErrorCode.DATABASE_CONNECTION_FAILED == "DB_001"
        assert ErrorCode.DOCUMENT_NOT_FOUND == "DOC_001"
        assert ErrorCode.INVALID_INPUT == "VAL_001"
        assert ErrorCode.INTERNAL_SERVER_ERROR == "SYS_001"


class TestApplicationError:
    """Test ApplicationError base class."""
    
    def test_basic_error_creation(self):
        """Test creating basic application error."""
        error = ApplicationError(
            "Test error message",
            ErrorCode.INTERNAL_SERVER_ERROR
        )
        
        assert error.message == "Test error message"
        assert error.error_code == ErrorCode.INTERNAL_SERVER_ERROR
        assert error.status_code == 500
        assert error.details == {}
        assert error.context == {}
        assert "Si è verificato un errore interno" in error.user_message
    
    def test_error_with_details_and_context(self):
        """Test error with additional details and context."""
        details = {"field": "test_field", "value": "invalid"}
        context = {"operation": "test_operation"}
        
        error = ApplicationError(
            "Validation failed",
            ErrorCode.INVALID_INPUT,
            details=details,
            context=context,
            status_code=400,
            user_message="Custom user message"
        )
        
        assert error.details == details
        assert error.context == context
        assert error.status_code == 400
        assert error.user_message == "Custom user message"
    
    def test_to_dict(self):
        """Test error serialization to dictionary."""
        error = ApplicationError(
            "Test error",
            ErrorCode.DATABASE_CONNECTION_FAILED,
            details={"connection": "failed"},
            context={"host": "localhost"}
        )
        
        result = error.to_dict()
        
        assert result["error"] is True
        assert result["error_code"] == "DB_001"
        assert result["message"] == error.user_message
        assert result["details"] == {"connection": "failed"}
        assert result["context"] == {"host": "localhost"}
    
    def test_default_user_messages(self):
        """Test default user-friendly messages for different error codes."""
        db_error = ApplicationError("DB failed", ErrorCode.DATABASE_CONNECTION_FAILED)
        assert "database" in db_error.user_message.lower()
        
        not_found_error = ApplicationError("Not found", ErrorCode.DOCUMENT_NOT_FOUND)
        assert "non trovato" in not_found_error.user_message.lower()
        
        validation_error = ApplicationError("Invalid", ErrorCode.INVALID_INPUT)
        assert "non validi" in validation_error.user_message.lower()


class TestSpecificErrors:
    """Test specific error subclasses."""
    
    def test_database_error(self):
        """Test DatabaseError creation."""
        error = DatabaseError(
            "Connection timeout",
            ErrorCode.DATABASE_TIMEOUT,
            operation="find_documents",
            collection="spells"
        )
        
        assert error.error_code == ErrorCode.DATABASE_TIMEOUT
        assert error.context["operation"] == "find_documents"
        assert error.context["collection"] == "spells"
    
    def test_validation_error(self):
        """Test ValidationError creation."""
        error = ValidationError(
            "Invalid field value",
            field="email",
            value="invalid-email"
        )
        
        assert error.error_code == ErrorCode.INVALID_INPUT
        assert error.status_code == 400
        assert error.context["field"] == "email"
        assert error.context["invalid_value"] == "invalid-email"
    
    def test_not_found_error(self):
        """Test NotFoundError creation."""
        error = NotFoundError(
            "Document not found",
            resource_type="spell",
            resource_id="fireball"
        )
        
        assert error.error_code == ErrorCode.DOCUMENT_NOT_FOUND
        assert error.status_code == 404
        assert error.context["resource_type"] == "spell"
        assert error.context["resource_id"] == "fireball"


class TestSafeOperation:
    """Test safe operation decorator."""
    
    @pytest.mark.asyncio
    async def test_successful_operation(self):
        """Test decorator with successful operation."""
        
        @safe_operation("test_operation", ErrorCode.INTERNAL_SERVER_ERROR, ApplicationError)
        async def successful_func():
            return "success"
        
        result = await successful_func()
        assert result == "success"
    
    @pytest.mark.asyncio
    async def test_operation_with_application_error(self):
        """Test decorator re-raises ApplicationError as-is."""
        
        @safe_operation("test_operation", ErrorCode.INTERNAL_SERVER_ERROR, ApplicationError)
        async def func_with_app_error():
            raise ValidationError("Validation failed")
        
        with pytest.raises(ValidationError) as exc_info:
            await func_with_app_error()
        
        assert exc_info.value.error_code == ErrorCode.INVALID_INPUT
    
    @pytest.mark.asyncio
    async def test_operation_with_generic_error(self):
        """Test decorator converts generic errors to ApplicationError."""
        
        @safe_operation("test_operation", ErrorCode.DATABASE_CONNECTION_FAILED, DatabaseError)
        async def func_with_generic_error():
            raise ValueError("Something went wrong")
        
        with pytest.raises(DatabaseError) as exc_info:
            await func_with_generic_error()
        
        assert exc_info.value.error_code == ErrorCode.DATABASE_CONNECTION_FAILED
        assert "test_operation" in exc_info.value.message
        assert "Something went wrong" in exc_info.value.context["original_error"]


class TestSafeDbOperation:
    """Test safe database operation utility."""
    
    @pytest.mark.asyncio
    async def test_successful_db_operation(self):
        """Test successful database operation."""
        from core.errors import safe_db_operation
        
        async def successful_operation():
            return {"result": "success"}
        
        result = await safe_db_operation(
            successful_operation,
            ErrorCode.DATABASE_OPERATION_FAILED,
            "test operation"
        )
        
        assert result == {"result": "success"}
    
    @pytest.mark.asyncio
    async def test_failed_db_operation(self):
        """Test failed database operation."""
        from core.errors import safe_db_operation
        
        async def failed_operation():
            raise ConnectionError("Database unavailable")
        
        with pytest.raises(DatabaseError) as exc_info:
            await safe_db_operation(
                failed_operation,
                ErrorCode.DATABASE_CONNECTION_FAILED,
                "test operation"
            )
        
        error = exc_info.value
        assert error.error_code == ErrorCode.DATABASE_CONNECTION_FAILED
        assert "test operation" in error.message
        assert "Database unavailable" in error.context["original_error"]