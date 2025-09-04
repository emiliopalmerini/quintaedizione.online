"""
Comprehensive Metrics Collection and Performance Monitoring
"""
import asyncio
import time
import logging

try:
    import psutil
    PSUTIL_AVAILABLE = True
except ImportError:
    PSUTIL_AVAILABLE = False
    logging.warning("psutil not available - using basic system metrics")
from typing import Any, Dict, List, Optional, Callable, Set
from dataclasses import dataclass, field, asdict
from datetime import datetime, timedelta
from collections import deque, defaultdict
from enum import Enum

logger = logging.getLogger(__name__)


class MetricType(Enum):
    """Types of metrics collected"""
    COUNTER = "counter"
    GAUGE = "gauge"
    HISTOGRAM = "histogram"
    TIMER = "timer"


@dataclass
class PerformanceMetrics:
    """System and application performance metrics"""
    timestamp: datetime
    
    # System metrics
    cpu_percent: float = 0.0
    memory_percent: float = 0.0
    memory_used_mb: float = 0.0
    memory_available_mb: float = 0.0
    disk_usage_percent: float = 0.0
    disk_free_gb: float = 0.0
    
    # Application metrics
    active_connections: int = 0
    request_rate_per_second: float = 0.0
    avg_response_time_ms: float = 0.0
    error_rate_percent: float = 0.0
    
    # Cache metrics
    cache_hit_rate_percent: float = 0.0
    cache_size_mb: float = 0.0
    cache_entries: int = 0
    
    # Database metrics
    db_query_count: int = 0
    avg_db_response_time_ms: float = 0.0
    db_connection_pool_usage: float = 0.0
    
    # Custom application metrics
    custom_metrics: Dict[str, Any] = field(default_factory=dict)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary"""
        result = asdict(self)
        result['timestamp'] = self.timestamp.isoformat()
        return result


@dataclass
class LatencyBucket:
    """Histogram bucket for latency measurements"""
    upper_bound_ms: float
    count: int = 0
    
    def increment(self) -> None:
        self.count += 1


class LatencyTracker:
    """Track latency distribution with histogram buckets"""
    
    def __init__(self, buckets: List[float] = None):
        # Default buckets in milliseconds: 1ms, 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, +Inf
        if buckets is None:
            buckets = [1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, float('inf')]
        
        self.buckets = [LatencyBucket(upper_bound_ms=bound) for bound in buckets]
        self.total_count = 0
        self.total_time_ms = 0.0
        self.min_time_ms = float('inf')
        self.max_time_ms = 0.0
        
        # Keep recent samples for percentile calculation
        self._recent_samples = deque(maxlen=1000)
    
    def record(self, latency_ms: float) -> None:
        """Record a latency measurement"""
        self.total_count += 1
        self.total_time_ms += latency_ms
        self.min_time_ms = min(self.min_time_ms, latency_ms)
        self.max_time_ms = max(self.max_time_ms, latency_ms)
        
        # Add to recent samples
        self._recent_samples.append(latency_ms)
        
        # Update histogram buckets
        for bucket in self.buckets:
            if latency_ms <= bucket.upper_bound_ms:
                bucket.increment()
                break
    
    def get_percentile(self, percentile: float) -> float:
        """Get percentile from recent samples"""
        if not self._recent_samples:
            return 0.0
        
        sorted_samples = sorted(self._recent_samples)
        index = int((percentile / 100.0) * len(sorted_samples))
        index = min(index, len(sorted_samples) - 1)
        return sorted_samples[index]
    
    def get_stats(self) -> Dict[str, Any]:
        """Get latency statistics"""
        if self.total_count == 0:
            return {
                "count": 0,
                "avg_ms": 0.0,
                "min_ms": 0.0,
                "max_ms": 0.0,
                "p50_ms": 0.0,
                "p90_ms": 0.0,
                "p95_ms": 0.0,
                "p99_ms": 0.0,
                "histogram": []
            }
        
        return {
            "count": self.total_count,
            "avg_ms": self.total_time_ms / self.total_count,
            "min_ms": self.min_time_ms if self.min_time_ms != float('inf') else 0.0,
            "max_ms": self.max_time_ms,
            "p50_ms": self.get_percentile(50),
            "p90_ms": self.get_percentile(90),
            "p95_ms": self.get_percentile(95),
            "p99_ms": self.get_percentile(99),
            "histogram": [
                {
                    "upper_bound_ms": bucket.upper_bound_ms,
                    "count": bucket.count
                }
                for bucket in self.buckets
                if bucket.count > 0
            ]
        }
    
    def reset(self) -> None:
        """Reset all statistics"""
        for bucket in self.buckets:
            bucket.count = 0
        self.total_count = 0
        self.total_time_ms = 0.0
        self.min_time_ms = float('inf')
        self.max_time_ms = 0.0
        self._recent_samples.clear()


class MetricsCollector:
    """Comprehensive metrics collection system"""
    
    def __init__(
        self,
        collection_interval_seconds: int = 10,
        retention_hours: int = 24,
        max_metrics_points: int = 8640  # 24 hours * 60 minutes * 6 (10-second intervals)
    ):
        self.collection_interval_seconds = collection_interval_seconds
        self.retention_hours = retention_hours
        self.max_metrics_points = max_metrics_points
        
        # Metrics storage
        self._metrics_history: deque[PerformanceMetrics] = deque(maxlen=max_metrics_points)
        self._counters: Dict[str, int] = defaultdict(int)
        self._gauges: Dict[str, float] = defaultdict(float)
        self._timers: Dict[str, LatencyTracker] = defaultdict(LatencyTracker)
        
        # Collection state
        self._collection_task: Optional[asyncio.Task] = None
        self._is_collecting = False
        self._custom_collectors: List[Callable[[], Dict[str, Any]]] = []
        
        # Alerting thresholds
        self._alert_thresholds: Dict[str, Dict[str, float]] = {
            "cpu_percent": {"warning": 70.0, "critical": 90.0},
            "memory_percent": {"warning": 80.0, "critical": 95.0},
            "disk_usage_percent": {"warning": 85.0, "critical": 95.0},
            "error_rate_percent": {"warning": 5.0, "critical": 10.0},
            "avg_response_time_ms": {"warning": 500.0, "critical": 1000.0}
        }
        
        self._alert_callbacks: List[Callable[[str, str, float], None]] = []
    
    async def start_collection(self) -> None:
        """Start metrics collection"""
        if self._is_collecting:
            return
        
        self._is_collecting = True
        self._collection_task = asyncio.create_task(self._collection_loop())
        logger.info(f"Started metrics collection (interval: {self.collection_interval_seconds}s)")
    
    async def stop_collection(self) -> None:
        """Stop metrics collection"""
        self._is_collecting = False
        
        if self._collection_task:
            self._collection_task.cancel()
            try:
                await self._collection_task
            except asyncio.CancelledError:
                pass
        
        logger.info("Stopped metrics collection")
    
    def increment_counter(self, name: str, value: int = 1) -> None:
        """Increment a counter metric"""
        self._counters[name] += value
    
    def set_gauge(self, name: str, value: float) -> None:
        """Set a gauge metric value"""
        self._gauges[name] = value
    
    def record_timer(self, name: str, value_ms: float) -> None:
        """Record a timer/latency metric"""
        self._timers[name].record(value_ms)
    
    def add_custom_collector(self, collector: Callable[[], Dict[str, Any]]) -> None:
        """Add custom metrics collector function"""
        self._custom_collectors.append(collector)
    
    def add_alert_callback(self, callback: Callable[[str, str, float], None]) -> None:
        """Add alert callback function"""
        self._alert_callbacks.append(callback)
    
    def set_alert_threshold(self, metric_name: str, warning: float, critical: float) -> None:
        """Set alert thresholds for a metric"""
        self._alert_thresholds[metric_name] = {
            "warning": warning,
            "critical": critical
        }
    
    async def get_current_metrics(self) -> PerformanceMetrics:
        """Get current metrics snapshot"""
        return await self._collect_metrics()
    
    async def get_metrics_history(
        self,
        hours_back: int = 1,
        downsample_minutes: int = 1
    ) -> List[PerformanceMetrics]:
        """Get historical metrics with optional downsampling"""
        cutoff_time = datetime.now() - timedelta(hours=hours_back)
        
        # Filter metrics by time
        filtered_metrics = [
            metrics for metrics in self._metrics_history
            if metrics.timestamp >= cutoff_time
        ]
        
        if downsample_minutes <= 1:
            return list(filtered_metrics)
        
        # Downsample by averaging metrics within time buckets
        if not filtered_metrics:
            return []
        
        bucket_size = timedelta(minutes=downsample_minutes)
        downsampled = []
        
        current_bucket_start = filtered_metrics[0].timestamp
        current_bucket_metrics = []
        
        for metrics in filtered_metrics:
            if metrics.timestamp >= current_bucket_start + bucket_size:
                # Process current bucket
                if current_bucket_metrics:
                    downsampled.append(self._average_metrics(current_bucket_metrics))
                
                # Start new bucket
                current_bucket_start = metrics.timestamp
                current_bucket_metrics = []
            
            current_bucket_metrics.append(metrics)
        
        # Process last bucket
        if current_bucket_metrics:
            downsampled.append(self._average_metrics(current_bucket_metrics))
        
        return downsampled
    
    async def get_metrics_summary(self) -> Dict[str, Any]:
        """Get comprehensive metrics summary"""
        if not self._metrics_history:
            return {"status": "no_data"}
        
        latest_metrics = self._metrics_history[-1]
        
        # Calculate rates and trends over last hour
        hourly_metrics = await self.get_metrics_history(hours_back=1)
        
        if len(hourly_metrics) < 2:
            trends = {}
        else:
            first_metrics = hourly_metrics[0]
            trends = {
                "cpu_trend": latest_metrics.cpu_percent - first_metrics.cpu_percent,
                "memory_trend": latest_metrics.memory_percent - first_metrics.memory_percent,
                "response_time_trend": latest_metrics.avg_response_time_ms - first_metrics.avg_response_time_ms,
                "error_rate_trend": latest_metrics.error_rate_percent - first_metrics.error_rate_percent
            }
        
        # Timer statistics
        timer_stats = {}
        for name, tracker in self._timers.items():
            timer_stats[name] = tracker.get_stats()
        
        # Active alerts
        active_alerts = self._check_alerts(latest_metrics)
        
        return {
            "timestamp": latest_metrics.timestamp.isoformat(),
            "system": {
                "cpu_percent": latest_metrics.cpu_percent,
                "memory_percent": latest_metrics.memory_percent,
                "memory_used_mb": latest_metrics.memory_used_mb,
                "memory_available_mb": latest_metrics.memory_available_mb,
                "disk_usage_percent": latest_metrics.disk_usage_percent,
                "disk_free_gb": latest_metrics.disk_free_gb
            },
            "application": {
                "active_connections": latest_metrics.active_connections,
                "request_rate_per_second": latest_metrics.request_rate_per_second,
                "avg_response_time_ms": latest_metrics.avg_response_time_ms,
                "error_rate_percent": latest_metrics.error_rate_percent
            },
            "cache": {
                "hit_rate_percent": latest_metrics.cache_hit_rate_percent,
                "size_mb": latest_metrics.cache_size_mb,
                "entries": latest_metrics.cache_entries
            },
            "database": {
                "query_count": latest_metrics.db_query_count,
                "avg_response_time_ms": latest_metrics.avg_db_response_time_ms,
                "connection_pool_usage": latest_metrics.db_connection_pool_usage
            },
            "counters": dict(self._counters),
            "gauges": dict(self._gauges),
            "timers": timer_stats,
            "trends": trends,
            "alerts": active_alerts,
            "custom_metrics": latest_metrics.custom_metrics,
            "collection_info": {
                "is_collecting": self._is_collecting,
                "interval_seconds": self.collection_interval_seconds,
                "history_points": len(self._metrics_history),
                "max_points": self.max_metrics_points
            }
        }
    
    async def reset_metrics(self) -> None:
        """Reset all collected metrics"""
        self._metrics_history.clear()
        self._counters.clear()
        self._gauges.clear()
        
        for timer in self._timers.values():
            timer.reset()
        
        logger.info("Reset all metrics")
    
    async def _collection_loop(self) -> None:
        """Main metrics collection loop"""
        while self._is_collecting:
            try:
                metrics = await self._collect_metrics()
                self._metrics_history.append(metrics)
                
                # Check for alerts
                alerts = self._check_alerts(metrics)
                for alert in alerts:
                    self._trigger_alert(alert["metric"], alert["level"], alert["value"])
                
                await asyncio.sleep(self.collection_interval_seconds)
                
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Error in metrics collection loop: {e}")
                await asyncio.sleep(self.collection_interval_seconds)
    
    async def _collect_metrics(self) -> PerformanceMetrics:
        """Collect current system and application metrics"""
        timestamp = datetime.now()
        
        # System metrics (with fallback if psutil not available)
        if PSUTIL_AVAILABLE:
            cpu_percent = psutil.cpu_percent(interval=None)
            memory = psutil.virtual_memory()
            disk = psutil.disk_usage('/')
        else:
            # Basic fallback metrics
            cpu_percent = 0.0
            class MockMemory:
                percent = 0.0
                used = 0
                available = 1024*1024*1024  # 1GB
            class MockDisk:
                percent = 0.0
                free = 10*1024*1024*1024  # 10GB
            memory = MockMemory()
            disk = MockDisk()
        
        # Calculate request rate (simplified)
        request_count = self._counters.get('http_requests_total', 0)
        error_count = self._counters.get('http_errors_total', 0)
        
        # Calculate rates per second (based on collection interval)
        request_rate = request_count / max(self.collection_interval_seconds, 1)
        error_rate = (error_count / max(request_count, 1)) * 100 if request_count > 0 else 0
        
        # Collect custom metrics
        custom_metrics = {}
        for collector in self._custom_collectors:
            try:
                custom_data = collector()
                custom_metrics.update(custom_data)
            except Exception as e:
                logger.warning(f"Error in custom metrics collector: {e}")
        
        metrics = PerformanceMetrics(
            timestamp=timestamp,
            
            # System metrics
            cpu_percent=cpu_percent,
            memory_percent=memory.percent,
            memory_used_mb=memory.used / 1024 / 1024,
            memory_available_mb=memory.available / 1024 / 1024,
            disk_usage_percent=disk.percent,
            disk_free_gb=disk.free / 1024 / 1024 / 1024,
            
            # Application metrics
            active_connections=self._gauges.get('active_connections', 0),
            request_rate_per_second=request_rate,
            avg_response_time_ms=self._gauges.get('avg_response_time_ms', 0),
            error_rate_percent=error_rate,
            
            # Cache metrics
            cache_hit_rate_percent=self._gauges.get('cache_hit_rate_percent', 0),
            cache_size_mb=self._gauges.get('cache_size_mb', 0),
            cache_entries=self._gauges.get('cache_entries', 0),
            
            # Database metrics
            db_query_count=self._counters.get('db_queries_total', 0),
            avg_db_response_time_ms=self._gauges.get('avg_db_response_time_ms', 0),
            db_connection_pool_usage=self._gauges.get('db_connection_pool_usage', 0),
            
            # Custom metrics
            custom_metrics=custom_metrics
        )
        
        return metrics
    
    def _check_alerts(self, metrics: PerformanceMetrics) -> List[Dict[str, Any]]:
        """Check for alert conditions"""
        alerts = []
        
        # Check each metric against thresholds
        for metric_name, thresholds in self._alert_thresholds.items():
            value = getattr(metrics, metric_name, None)
            
            if value is None:
                continue
            
            level = None
            if value >= thresholds.get("critical", float('inf')):
                level = "critical"
            elif value >= thresholds.get("warning", float('inf')):
                level = "warning"
            
            if level:
                alerts.append({
                    "metric": metric_name,
                    "level": level,
                    "value": value,
                    "threshold": thresholds[level],
                    "timestamp": metrics.timestamp.isoformat()
                })
        
        return alerts
    
    def _trigger_alert(self, metric_name: str, level: str, value: float) -> None:
        """Trigger alert callbacks"""
        for callback in self._alert_callbacks:
            try:
                callback(metric_name, level, value)
            except Exception as e:
                logger.error(f"Error in alert callback: {e}")
    
    def _average_metrics(self, metrics_list: List[PerformanceMetrics]) -> PerformanceMetrics:
        """Average a list of metrics for downsampling"""
        if not metrics_list:
            return None
        
        if len(metrics_list) == 1:
            return metrics_list[0]
        
        # Calculate averages for numeric fields
        avg_metrics = PerformanceMetrics(timestamp=metrics_list[-1].timestamp)
        
        numeric_fields = [
            'cpu_percent', 'memory_percent', 'memory_used_mb', 'memory_available_mb',
            'disk_usage_percent', 'disk_free_gb', 'active_connections',
            'request_rate_per_second', 'avg_response_time_ms', 'error_rate_percent',
            'cache_hit_rate_percent', 'cache_size_mb', 'cache_entries',
            'db_query_count', 'avg_db_response_time_ms', 'db_connection_pool_usage'
        ]
        
        for field in numeric_fields:
            values = [getattr(m, field, 0) for m in metrics_list]
            setattr(avg_metrics, field, sum(values) / len(values))
        
        # Merge custom metrics (take last)
        avg_metrics.custom_metrics = metrics_list[-1].custom_metrics
        
        return avg_metrics


# Global metrics collector instance
_metrics_collector: Optional[MetricsCollector] = None


def get_metrics_collector() -> MetricsCollector:
    """Get global metrics collector instance"""
    global _metrics_collector
    if _metrics_collector is None:
        _metrics_collector = MetricsCollector()
    return _metrics_collector