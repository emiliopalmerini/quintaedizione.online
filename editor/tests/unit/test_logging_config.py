"""Unit tests for logging configuration."""
from __future__ import annotations

import json
import logging
import pytest
from unittest.mock import MagicMock, patch
from io import StringIO

from core.logging_config import (
    StructuredFormatter,
    RequestContextFilter,
    setup_logging,
    get_logger,
    log_database_operation,
    log_http_request,
    log_error_with_context,
)


class TestStructuredFormatter:
    """Test StructuredFormatter class."""
    
    def test_basic_formatting(self):
        """Test basic log record formatting."""
        formatter = StructuredFormatter()
        
        record = logging.LogRecord(
            name="test.logger",
            level=logging.INFO,
            pathname="/test/file.py",
            lineno=42,
            msg="Test message",
            args=(),
            exc_info=None
        )
        record.funcName = "test_function"
        record.module = "test_module"
        
        result = formatter.format(record)
        log_data = json.loads(result)
        
        assert log_data["level"] == "INFO"
        assert log_data["logger"] == "test.logger"
        assert log_data["message"] == "Test message"
        assert log_data["module"] == "test_module"
        assert log_data["function"] == "test_function"
        assert log_data["line"] == 42
        assert "timestamp" in log_data
    
    def test_formatting_with_exception(self):
        """Test formatting with exception information."""
        formatter = StructuredFormatter()
        
        try:
            raise ValueError("Test exception")
        except ValueError:
            import sys
            exc_info = sys.exc_info()
        
        record = logging.LogRecord(
            name="test.logger",
            level=logging.ERROR,
            pathname="/test/file.py",
            lineno=42,
            msg="Error occurred",
            args=(),
            exc_info=exc_info
        )
        record.funcName = "test_function"
        record.module = "test_module"
        
        result = formatter.format(record)
        log_data = json.loads(result)
        
        assert "exception" in log_data
        assert "ValueError: Test exception" in log_data["exception"]
    
    def test_formatting_with_extra_fields(self):
        """Test formatting with extra fields."""
        formatter = StructuredFormatter()
        
        record = logging.LogRecord(
            name="test.logger",
            level=logging.INFO,
            pathname="/test/file.py",
            lineno=42,
            msg="Test message",
            args=(),
            exc_info=None
        )
        record.funcName = "test_function"
        record.module = "test_module"
        
        # Add extra fields
        record.request_id = "req_123"
        record.user_id = "user_456"
        
        result = formatter.format(record)
        log_data = json.loads(result)
        
        assert "extra" in log_data
        assert log_data["extra"]["request_id"] == "req_123"
        assert log_data["extra"]["user_id"] == "user_456"


class TestRequestContextFilter:
    """Test RequestContextFilter class."""
    
    def test_filter_passes_records(self):
        """Test that filter always passes records."""
        filter_instance = RequestContextFilter()
        
        record = logging.LogRecord(
            name="test.logger",
            level=logging.INFO,
            pathname="/test/file.py",
            lineno=42,
            msg="Test message",
            args=(),
            exc_info=None
        )
        
        assert filter_instance.filter(record) is True
    
    def test_filter_preserves_request_id(self):
        """Test that filter preserves existing request_id."""
        filter_instance = RequestContextFilter()
        
        record = logging.LogRecord(
            name="test.logger",
            level=logging.INFO,
            pathname="/test/file.py",
            lineno=42,
            msg="Test message",
            args=(),
            exc_info=None
        )
        record.request_id = "test_request_123"
        
        result = filter_instance.filter(record)
        
        assert result is True
        assert record.request_id == "test_request_123"


class TestSetupLogging:
    """Test setup_logging function."""
    
    @patch('logging.config.dictConfig')
    def test_basic_setup(self, mock_dict_config):
        """Test basic logging setup."""
        setup_logging()
        
        mock_dict_config.assert_called_once()
        config = mock_dict_config.call_args[0][0]
        
        assert config["version"] == 1
        assert "formatters" in config
        assert "handlers" in config
        assert "root" in config
        assert "loggers" in config
    
    @patch('logging.config.dictConfig')
    def test_setup_with_debug_level(self, mock_dict_config):
        """Test logging setup with DEBUG level."""
        setup_logging(log_level="DEBUG")
        
        config = mock_dict_config.call_args[0][0]
        assert config["handlers"]["console"]["level"] == logging.DEBUG
    
    @patch('logging.config.dictConfig')
    @patch('pathlib.Path.mkdir')
    def test_setup_with_log_file(self, mock_mkdir, mock_dict_config):
        """Test logging setup with file handler."""
        setup_logging(log_file="/tmp/test.log")
        
        config = mock_dict_config.call_args[0][0]
        
        assert "file" in config["handlers"]
        assert config["handlers"]["file"]["filename"] == "/tmp/test.log"
        assert "file" in config["root"]["handlers"]
        mock_mkdir.assert_called_once()
    
    @patch('logging.config.dictConfig')
    def test_setup_with_simple_formatting(self, mock_dict_config):
        """Test logging setup with simple (non-JSON) formatting."""
        setup_logging(structured=False)
        
        config = mock_dict_config.call_args[0][0]
        assert config["handlers"]["console"]["formatter"] == "simple"


class TestGetLogger:
    """Test get_logger function."""
    
    def test_get_logger_returns_logger(self):
        """Test that get_logger returns a logger instance."""
        logger = get_logger("test.module")
        
        assert isinstance(logger, logging.Logger)
        assert logger.name == "test.module"


class TestLoggingHelpers:
    """Test logging helper functions."""
    
    def test_log_database_operation(self):
        """Test log_database_operation helper."""
        mock_logger = MagicMock()
        
        log_database_operation(
            mock_logger,
            "find",
            "spells",
            {"level": 3},
            150.5
        )
        
        mock_logger.info.assert_called_once()
        call_args = mock_logger.info.call_args
        
        assert "Database operation: find" in call_args[0][0]
        assert call_args[1]["extra"]["operation"] == "find"
        assert call_args[1]["extra"]["collection"] == "spells"
        assert call_args[1]["extra"]["filter"] == {"level": 3}
        assert call_args[1]["extra"]["duration_ms"] == 150.5
    
    def test_log_http_request(self):
        """Test log_http_request helper."""
        mock_logger = MagicMock()
        
        log_http_request(
            mock_logger,
            "GET",
            "/api/spells",
            200,
            250.0,
            "user123"
        )
        
        mock_logger.info.assert_called_once()
        call_args = mock_logger.info.call_args
        
        assert "HTTP GET /api/spells -> 200" in call_args[0][0]
        assert call_args[1]["extra"]["http_method"] == "GET"
        assert call_args[1]["extra"]["http_path"] == "/api/spells"
        assert call_args[1]["extra"]["http_status"] == 200
        assert call_args[1]["extra"]["duration_ms"] == 250.0
        assert call_args[1]["extra"]["user_id"] == "user123"
    
    def test_log_error_with_context(self):
        """Test log_error_with_context helper."""
        mock_logger = MagicMock()
        
        test_error = ValueError("Test error")
        context = {"operation": "database_query", "collection": "spells"}
        
        log_error_with_context(
            mock_logger,
            "Database operation failed",
            test_error,
            context
        )
        
        mock_logger.error.assert_called_once()
        call_args = mock_logger.error.call_args
        
        assert call_args[0][0] == "Database operation failed"
        assert call_args[1]["exc_info"] == test_error
        assert call_args[1]["extra"]["error_context"] == context


class TestLoggingIntegration:
    """Integration tests for logging system."""
    
    def test_end_to_end_logging(self):
        """Test complete logging flow."""
        # Capture log output
        log_capture = StringIO()
        handler = logging.StreamHandler(log_capture)
        handler.setFormatter(StructuredFormatter())
        
        logger = logging.getLogger("test.integration")
        logger.addHandler(handler)
        logger.setLevel(logging.INFO)
        
        # Log a message with extra data
        logger.info(
            "Integration test message",
            extra={
                "request_id": "req_123",
                "operation": "test"
            }
        )
        
        # Verify output
        log_output = log_capture.getvalue()
        log_data = json.loads(log_output.strip())
        
        assert log_data["message"] == "Integration test message"
        assert log_data["logger"] == "test.integration"
        assert log_data["extra"]["request_id"] == "req_123"
        assert log_data["extra"]["operation"] == "test"