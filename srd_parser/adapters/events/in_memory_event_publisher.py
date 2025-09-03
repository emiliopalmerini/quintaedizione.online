"""
Simple in-memory event publisher for parser service
Can be extended to use Redis, RabbitMQ, etc. in production
"""
import asyncio
import logging
from typing import Dict, List, Callable, Any
from collections import defaultdict

from shared_domain.entities import DomainEvent
from shared_domain.use_cases import EventPublisher

logger = logging.getLogger(__name__)


class InMemoryEventPublisher(EventPublisher):
    """Simple in-memory event publisher with subscriber pattern"""
    
    def __init__(self):
        self.subscribers: Dict[str, List[Callable]] = defaultdict(list)
        self.event_store: List[DomainEvent] = []  # For debugging/replay
    
    async def publish(self, event: DomainEvent) -> None:
        """Publish event to all subscribers"""
        try:
            # Store event for debugging
            self.event_store.append(event)
            
            # Log event
            logger.info(f"Publishing event: {event.__class__.__name__} for {event.aggregate_id}")
            
            # Get event type name
            event_type = event.__class__.__name__
            
            # Notify all subscribers for this event type
            subscribers = self.subscribers.get(event_type, [])
            if subscribers:
                # Run all subscribers concurrently
                tasks = [self._safe_call_subscriber(subscriber, event) for subscriber in subscribers]
                await asyncio.gather(*tasks, return_exceptions=True)
            else:
                logger.debug(f"No subscribers found for event type: {event_type}")
                
        except Exception as e:
            logger.error(f"Error publishing event {event.__class__.__name__}: {e}", exc_info=True)
    
    async def _safe_call_subscriber(self, subscriber: Callable, event: DomainEvent) -> None:
        """Safely call subscriber with error handling"""
        try:
            if asyncio.iscoroutinefunction(subscriber):
                await subscriber(event)
            else:
                subscriber(event)
        except Exception as e:
            logger.error(f"Error in event subscriber {subscriber.__name__}: {e}", exc_info=True)
    
    def subscribe(self, event_type: str, subscriber: Callable) -> None:
        """Subscribe to events of specific type"""
        self.subscribers[event_type].append(subscriber)
        logger.info(f"Subscribed {subscriber.__name__} to {event_type}")
    
    def unsubscribe(self, event_type: str, subscriber: Callable) -> None:
        """Unsubscribe from events"""
        if event_type in self.subscribers:
            try:
                self.subscribers[event_type].remove(subscriber)
                logger.info(f"Unsubscribed {subscriber.__name__} from {event_type}")
            except ValueError:
                logger.warning(f"Subscriber {subscriber.__name__} not found for {event_type}")
    
    def get_events_for_aggregate(self, aggregate_id: str) -> List[DomainEvent]:
        """Get all events for specific aggregate (useful for debugging)"""
        return [event for event in self.event_store if event.aggregate_id == aggregate_id]
    
    def get_all_events(self) -> List[DomainEvent]:
        """Get all published events (useful for debugging)"""
        return self.event_store.copy()
    
    def clear_events(self) -> None:
        """Clear event store (useful for testing)"""
        self.event_store.clear()
        logger.info("Event store cleared")


class LoggingEventPublisher(EventPublisher):
    """Event publisher that just logs events (useful for development)"""
    
    def __init__(self):
        self.logger = logging.getLogger(__name__)
    
    async def publish(self, event: DomainEvent) -> None:
        """Log the event instead of publishing"""
        self.logger.info(
            f"Event Published: {event.__class__.__name__} "
            f"| Aggregate: {event.aggregate_id} "
            f"| Timestamp: {event.timestamp}"
        )
        
        # Log event-specific data
        if hasattr(event, 'class_name'):
            self.logger.info(f"Class Name: {event.class_name}")
        if hasattr(event, 'version'):
            self.logger.info(f"Version: {event.version}")


class CompositeEventPublisher(EventPublisher):
    """Composite publisher that sends events to multiple publishers"""
    
    def __init__(self, publishers: List[EventPublisher]):
        self.publishers = publishers
    
    async def publish(self, event: DomainEvent) -> None:
        """Publish to all configured publishers"""
        tasks = [publisher.publish(event) for publisher in self.publishers]
        await asyncio.gather(*tasks, return_exceptions=True)