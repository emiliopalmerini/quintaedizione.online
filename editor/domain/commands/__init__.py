"""
CQRS Commands for Editor Domain
Command-side operations for write scenarios
"""
from .base_commands import Command, CommandHandler, CommandBus, CommandResult
from .content_commands import (
    CacheContentCommand,
    InvalidateCacheCommand,
    PreloadContentCommand,
    OptimizeSearchCommand,
    RecordAnalyticsCommand
)

__all__ = [
    # Command infrastructure
    "Command",
    "CommandHandler",
    "CommandBus", 
    "CommandResult",
    
    # Content commands
    "CacheContentCommand",
    "InvalidateCacheCommand",
    "PreloadContentCommand",
    "OptimizeSearchCommand",
    "RecordAnalyticsCommand"
]