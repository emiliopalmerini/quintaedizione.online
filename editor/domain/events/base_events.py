"""
Base Event Infrastructure for Domain Events
"""
from typing import Any, Dict, List, Optional, Type, Protocol
from dataclasses import dataclass, field
from datetime import datetime
from abc import ABC, abstractmethod
import asyncio
import logging

logger = logging.getLogger(__name__)


@dataclass
class DomainEvent:
    """Base class for all domain events"""
    event_id: str = field(default_factory=lambda: str(id(object())))
    occurred_at: datetime = field(default_factory=datetime.now)
    event_type: str = field(default="")
    event_data: Dict[str, Any] = field(default_factory=dict)
    
    def __post_init__(self):
        if not self.event_type:
            self.event_type = self.__class__.__name__


class EventHandler(Protocol):
    """Protocol for event handlers"""
    
    async def handle(self, event: DomainEvent) -> None:
        """Handle a domain event"""
        ...


class EventBus:
    """Simple event bus for domain events"""
    
    def __init__(self):
        self._handlers: Dict[Type[DomainEvent], List[EventHandler]] = {}
        self._global_handlers: List[EventHandler] = []
    
    def subscribe(self, event_type: Type[DomainEvent], handler: EventHandler) -> None:
        """Subscribe handler to specific event type"""
        if event_type not in self._handlers:
            self._handlers[event_type] = []
        self._handlers[event_type].append(handler)
    
    def subscribe_to_all(self, handler: EventHandler) -> None:
        """Subscribe handler to all events"""
        self._global_handlers.append(handler)
    
    async def publish(self, event: DomainEvent) -> None:
        """Publish event to all subscribed handlers"""
        try:
            # Get specific handlers for this event type
            specific_handlers = self._handlers.get(type(event), [])
            
            # Combine with global handlers
            all_handlers = specific_handlers + self._global_handlers
            
            if not all_handlers:
                logger.debug(f"No handlers registered for event {type(event).__name__}")
                return
            
            # Execute all handlers concurrently
            tasks = [handler.handle(event) for handler in all_handlers]
            await asyncio.gather(*tasks, return_exceptions=True)
            
            logger.debug(f"Published event {event.event_type} to {len(all_handlers)} handlers")
            
        except Exception as e:
            logger.error(f"Error publishing event {event.event_type}: {e}")


# Global event bus instance
_event_bus: Optional[EventBus] = None


def get_event_bus() -> EventBus:
    """Get global event bus instance"""
    global _event_bus
    if _event_bus is None:
        _event_bus = EventBus()
    return _event_bus