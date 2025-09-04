"""
Base Command Infrastructure for CQRS Pattern
"""
from typing import Any, Dict, List, Optional, Type, Protocol
from dataclasses import dataclass, field
from datetime import datetime
from abc import ABC, abstractmethod
import asyncio
import logging

logger = logging.getLogger(__name__)


@dataclass
class CommandResult:
    """Result of command execution"""
    success: bool
    message: str = ""
    data: Any = None
    errors: List[str] = field(default_factory=list)
    execution_time_ms: float = 0.0
    
    @classmethod
    def success_result(cls, message: str = "Command executed successfully", data: Any = None) -> "CommandResult":
        """Create successful command result"""
        return cls(success=True, message=message, data=data)
    
    @classmethod
    def failure_result(cls, message: str, errors: List[str] = None) -> "CommandResult":
        """Create failed command result"""
        return cls(
            success=False,
            message=message,
            errors=errors or []
        )


@dataclass
class Command:
    """Base class for all commands"""
    command_id: str = field(default_factory=lambda: str(id(object())))
    issued_at: datetime = field(default_factory=datetime.now)
    command_type: str = field(default="")
    command_data: Dict[str, Any] = field(default_factory=dict)
    
    def __post_init__(self):
        if not self.command_type:
            self.command_type = self.__class__.__name__


class CommandHandler(Protocol):
    """Protocol for command handlers"""
    
    async def handle(self, command: Command) -> CommandResult:
        """Handle a command and return result"""
        ...


class CommandBus:
    """Command bus for CQRS pattern"""
    
    def __init__(self):
        self._handlers: Dict[Type[Command], CommandHandler] = {}
        self._middleware: List[CommandHandler] = []
    
    def register_handler(self, command_type: Type[Command], handler: CommandHandler) -> None:
        """Register handler for specific command type"""
        self._handlers[command_type] = handler
        logger.debug(f"Registered handler for command type: {command_type.__name__}")
    
    def add_middleware(self, middleware: CommandHandler) -> None:
        """Add middleware that processes all commands"""
        self._middleware.append(middleware)
    
    async def execute(self, command: Command) -> CommandResult:
        """Execute command using registered handler"""
        start_time = asyncio.get_event_loop().time()
        
        try:
            # Apply middleware first
            for middleware in self._middleware:
                middleware_result = await middleware.handle(command)
                if not middleware_result.success:
                    logger.warning(f"Middleware failed for command {command.command_type}: {middleware_result.message}")
                    return middleware_result
            
            # Find and execute handler
            command_type = type(command)
            if command_type not in self._handlers:
                error_msg = f"No handler registered for command type: {command_type.__name__}"
                logger.error(error_msg)
                return CommandResult.failure_result(error_msg)
            
            handler = self._handlers[command_type]
            result = await handler.handle(command)
            
            # Add execution time
            execution_time = (asyncio.get_event_loop().time() - start_time) * 1000
            result.execution_time_ms = execution_time
            
            logger.debug(f"Executed command {command.command_type} in {execution_time:.2f}ms")
            return result
            
        except Exception as e:
            execution_time = (asyncio.get_event_loop().time() - start_time) * 1000
            error_msg = f"Error executing command {command.command_type}: {str(e)}"
            logger.error(error_msg)
            
            result = CommandResult.failure_result(error_msg, [str(e)])
            result.execution_time_ms = execution_time
            return result


# Global command bus instance
_command_bus: Optional[CommandBus] = None


def get_command_bus() -> CommandBus:
    """Get global command bus instance"""
    global _command_bus
    if _command_bus is None:
        _command_bus = CommandBus()
    return _command_bus