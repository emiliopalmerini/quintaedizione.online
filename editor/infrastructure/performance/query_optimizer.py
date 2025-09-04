"""
Query Optimization and Performance Analysis
"""
import asyncio
import time
import logging
from typing import Any, Dict, List, Optional, Set, Tuple, Callable
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from collections import defaultdict

logger = logging.getLogger(__name__)


class QueryType(Enum):
    """Types of database queries"""
    SEARCH = "search"
    GET_BY_ID = "get_by_id" 
    COUNT = "count"
    AGGREGATE = "aggregate"
    FILTER = "filter"


@dataclass
class OptimizationHint:
    """Optimization hint for query performance"""
    hint_type: str
    description: str
    impact_level: str  # "low", "medium", "high"
    suggested_action: str
    estimated_improvement_percent: float = 0.0


@dataclass
class QueryStats:
    """Statistics for query performance"""
    query_type: QueryType
    collection: str
    query_hash: str
    execution_count: int = 0
    total_time_ms: float = 0.0
    min_time_ms: float = float('inf')
    max_time_ms: float = 0.0
    avg_time_ms: float = 0.0
    last_execution: Optional[datetime] = None
    slow_query_count: int = 0
    error_count: int = 0
    optimization_hints: List[OptimizationHint] = field(default_factory=list)
    
    def update_timing(self, execution_time_ms: float) -> None:
        """Update timing statistics"""
        self.execution_count += 1
        self.total_time_ms += execution_time_ms
        self.min_time_ms = min(self.min_time_ms, execution_time_ms)
        self.max_time_ms = max(self.max_time_ms, execution_time_ms)
        self.avg_time_ms = self.total_time_ms / self.execution_count
        self.last_execution = datetime.now()
        
        # Track slow queries (>100ms threshold)
        if execution_time_ms > 100:
            self.slow_query_count += 1


class QueryOptimizer:
    """Query optimizer with performance analysis and suggestions"""
    
    def __init__(
        self,
        slow_query_threshold_ms: float = 100.0,
        analysis_window_hours: int = 24,
        max_stats_entries: int = 1000
    ):
        self.slow_query_threshold_ms = slow_query_threshold_ms
        self.analysis_window_hours = analysis_window_hours
        self.max_stats_entries = max_stats_entries
        
        self._query_stats: Dict[str, QueryStats] = {}
        self._optimization_rules: List[Callable[[QueryStats], List[OptimizationHint]]] = []
        self._register_optimization_rules()
    
    async def track_query(
        self,
        query_type: QueryType,
        collection: str,
        query_params: Dict[str, Any],
        execution_time_ms: float,
        error: Optional[Exception] = None
    ) -> str:
        """Track query execution for optimization analysis"""
        
        # Generate query hash for grouping similar queries
        query_hash = self._generate_query_hash(query_type, collection, query_params)
        
        # Get or create query stats
        if query_hash not in self._query_stats:
            self._query_stats[query_hash] = QueryStats(
                query_type=query_type,
                collection=collection,
                query_hash=query_hash
            )
        
        stats = self._query_stats[query_hash]
        
        # Update statistics
        stats.update_timing(execution_time_ms)
        
        if error:
            stats.error_count += 1
            logger.warning(f"Query error tracked: {query_hash} - {error}")
        
        # Generate optimization hints for slow queries
        if execution_time_ms > self.slow_query_threshold_ms:
            await self._analyze_and_suggest_optimizations(stats, query_params)
        
        # Cleanup old stats if needed
        await self._cleanup_old_stats()
        
        logger.debug(f"Tracked query: {query_hash} ({execution_time_ms:.2f}ms)")
        return query_hash
    
    async def get_performance_report(
        self,
        collection: Optional[str] = None,
        query_type: Optional[QueryType] = None
    ) -> Dict[str, Any]:
        """Generate comprehensive performance report"""
        
        # Filter stats based on criteria
        filtered_stats = self._filter_stats(collection, query_type)
        
        if not filtered_stats:
            return {
                "summary": {"total_queries": 0},
                "slow_queries": [],
                "optimization_suggestions": [],
                "collection_analysis": {}
            }
        
        # Generate summary statistics
        total_queries = sum(stats.execution_count for stats in filtered_stats)
        total_slow_queries = sum(stats.slow_query_count for stats in filtered_stats)
        total_errors = sum(stats.error_count for stats in filtered_stats)
        
        avg_response_time = sum(
            stats.avg_time_ms * stats.execution_count for stats in filtered_stats
        ) / total_queries if total_queries > 0 else 0
        
        # Identify slowest queries
        slow_queries = [
            {
                "query_hash": stats.query_hash,
                "collection": stats.collection,
                "query_type": stats.query_type.value,
                "avg_time_ms": stats.avg_time_ms,
                "max_time_ms": stats.max_time_ms,
                "execution_count": stats.execution_count,
                "slow_query_percentage": (stats.slow_query_count / stats.execution_count) * 100,
                "last_execution": stats.last_execution.isoformat() if stats.last_execution else None
            }
            for stats in sorted(filtered_stats, key=lambda x: x.avg_time_ms, reverse=True)[:10]
        ]
        
        # Collect all optimization suggestions
        all_suggestions = []
        for stats in filtered_stats:
            for hint in stats.optimization_hints:
                all_suggestions.append({
                    "query_hash": stats.query_hash,
                    "collection": stats.collection,
                    "hint_type": hint.hint_type,
                    "description": hint.description,
                    "impact_level": hint.impact_level,
                    "suggested_action": hint.suggested_action,
                    "estimated_improvement_percent": hint.estimated_improvement_percent
                })
        
        # Collection-level analysis
        collection_stats = defaultdict(lambda: {
            "total_queries": 0,
            "avg_response_time": 0.0,
            "slow_query_count": 0,
            "error_count": 0,
            "query_types": set()
        })
        
        for stats in filtered_stats:
            col_stats = collection_stats[stats.collection]
            col_stats["total_queries"] += stats.execution_count
            col_stats["slow_query_count"] += stats.slow_query_count
            col_stats["error_count"] += stats.error_count
            col_stats["query_types"].add(stats.query_type.value)
            
            # Weighted average for response time
            current_total = col_stats["avg_response_time"] * (col_stats["total_queries"] - stats.execution_count)
            new_total = current_total + (stats.avg_time_ms * stats.execution_count)
            col_stats["avg_response_time"] = new_total / col_stats["total_queries"]
        
        # Convert sets to lists for JSON serialization
        for col_stats in collection_stats.values():
            col_stats["query_types"] = list(col_stats["query_types"])
        
        return {
            "summary": {
                "total_queries": total_queries,
                "avg_response_time_ms": avg_response_time,
                "slow_queries": total_slow_queries,
                "slow_query_percentage": (total_slow_queries / total_queries) * 100 if total_queries > 0 else 0,
                "total_errors": total_errors,
                "error_rate": (total_errors / total_queries) * 100 if total_queries > 0 else 0,
                "collections_analyzed": len(set(stats.collection for stats in filtered_stats))
            },
            "slow_queries": slow_queries,
            "optimization_suggestions": all_suggestions,
            "collection_analysis": dict(collection_stats),
            "generated_at": datetime.now().isoformat()
        }
    
    async def get_optimization_suggestions(
        self,
        collection: str,
        limit: int = 10
    ) -> List[Dict[str, Any]]:
        """Get optimization suggestions for a specific collection"""
        
        collection_stats = [
            stats for stats in self._query_stats.values()
            if stats.collection == collection
        ]
        
        # Sort by impact and frequency
        prioritized_suggestions = []
        
        for stats in collection_stats:
            for hint in stats.optimization_hints:
                priority_score = self._calculate_suggestion_priority(stats, hint)
                prioritized_suggestions.append({
                    "query_hash": stats.query_hash,
                    "priority_score": priority_score,
                    "hint": hint,
                    "stats": stats
                })
        
        # Sort by priority score and return top suggestions
        prioritized_suggestions.sort(key=lambda x: x["priority_score"], reverse=True)
        
        return [
            {
                "query_hash": item["query_hash"],
                "collection": item["stats"].collection,
                "query_type": item["stats"].query_type.value,
                "execution_count": item["stats"].execution_count,
                "avg_time_ms": item["stats"].avg_time_ms,
                "hint_type": item["hint"].hint_type,
                "description": item["hint"].description,
                "impact_level": item["hint"].impact_level,
                "suggested_action": item["hint"].suggested_action,
                "estimated_improvement_percent": item["hint"].estimated_improvement_percent,
                "priority_score": item["priority_score"]
            }
            for item in prioritized_suggestions[:limit]
        ]
    
    async def clear_stats(self, older_than_hours: Optional[int] = None) -> int:
        """Clear query statistics"""
        if older_than_hours is None:
            # Clear all stats
            count = len(self._query_stats)
            self._query_stats.clear()
            logger.info(f"Cleared all {count} query statistics")
            return count
        
        # Clear stats older than specified hours
        cutoff_time = datetime.now() - timedelta(hours=older_than_hours)
        keys_to_remove = []
        
        for query_hash, stats in self._query_stats.items():
            if stats.last_execution and stats.last_execution < cutoff_time:
                keys_to_remove.append(query_hash)
        
        for key in keys_to_remove:
            del self._query_stats[key]
        
        logger.info(f"Cleared {len(keys_to_remove)} old query statistics")
        return len(keys_to_remove)
    
    def _generate_query_hash(
        self,
        query_type: QueryType,
        collection: str,
        query_params: Dict[str, Any]
    ) -> str:
        """Generate hash for grouping similar queries"""
        
        # Create normalized representation of query parameters
        normalized_params = {}
        
        # Normalize common parameters
        for key, value in query_params.items():
            if key in ["text_query", "query"]:
                # Normalize text queries (remove specific terms, keep structure)
                if value:
                    normalized_params[key] = "HAS_TEXT_QUERY"
                else:
                    normalized_params[key] = None
            elif key in ["limit", "offset"]:
                # Group by ranges for pagination parameters
                if isinstance(value, int):
                    if value <= 10:
                        normalized_params[key] = "small"
                    elif value <= 50:
                        normalized_params[key] = "medium"
                    else:
                        normalized_params[key] = "large"
            elif key in ["sort_by", "order"]:
                # Keep sorting parameters as-is
                normalized_params[key] = value
            elif isinstance(value, (list, dict)):
                # For complex parameters, just note presence
                normalized_params[key] = "HAS_COMPLEX_VALUE" if value else None
            else:
                # Keep simple parameters as-is
                normalized_params[key] = value
        
        # Create hash from normalized parameters
        param_str = ":".join(f"{k}={v}" for k, v in sorted(normalized_params.items()))
        query_signature = f"{query_type.value}:{collection}:{param_str}"
        
        return str(hash(query_signature))
    
    def _filter_stats(
        self,
        collection: Optional[str] = None,
        query_type: Optional[QueryType] = None
    ) -> List[QueryStats]:
        """Filter query statistics based on criteria"""
        
        filtered = list(self._query_stats.values())
        
        if collection:
            filtered = [stats for stats in filtered if stats.collection == collection]
        
        if query_type:
            filtered = [stats for stats in filtered if stats.query_type == query_type]
        
        return filtered
    
    def _calculate_suggestion_priority(
        self,
        stats: QueryStats,
        hint: OptimizationHint
    ) -> float:
        """Calculate priority score for optimization suggestion"""
        
        # Base score from impact level
        impact_scores = {"low": 1.0, "medium": 2.0, "high": 3.0}
        impact_score = impact_scores.get(hint.impact_level, 1.0)
        
        # Frequency score (more frequent queries get higher priority)
        frequency_score = min(stats.execution_count / 100, 5.0)  # Cap at 5x
        
        # Performance score (slower queries get higher priority)
        performance_score = min(stats.avg_time_ms / self.slow_query_threshold_ms, 10.0)  # Cap at 10x
        
        # Recency score (recent queries get higher priority)
        if stats.last_execution:
            hours_ago = (datetime.now() - stats.last_execution).total_seconds() / 3600
            recency_score = max(1.0, 24.0 - hours_ago) / 24.0  # Decay over 24 hours
        else:
            recency_score = 0.1
        
        # Weighted composite score
        priority_score = (
            impact_score * 0.4 +
            frequency_score * 0.3 +
            performance_score * 0.2 +
            recency_score * 0.1
        )
        
        return priority_score
    
    async def _analyze_and_suggest_optimizations(
        self,
        stats: QueryStats,
        query_params: Dict[str, Any]
    ) -> None:
        """Analyze query and generate optimization hints"""
        
        # Clear existing hints for this execution
        stats.optimization_hints.clear()
        
        # Apply optimization rules
        for rule in self._optimization_rules:
            try:
                hints = rule(stats)
                stats.optimization_hints.extend(hints)
            except Exception as e:
                logger.error(f"Error in optimization rule: {e}")
    
    def _register_optimization_rules(self) -> None:
        """Register optimization analysis rules"""
        
        self._optimization_rules = [
            self._rule_slow_text_search,
            self._rule_large_result_sets,
            self._rule_frequent_similar_queries,
            self._rule_missing_indexes,
            self._rule_inefficient_filters,
            self._rule_pagination_performance,
        ]
    
    def _rule_slow_text_search(self, stats: QueryStats) -> List[OptimizationHint]:
        """Rule for slow text search queries"""
        hints = []
        
        if (stats.query_type == QueryType.SEARCH and 
            stats.avg_time_ms > self.slow_query_threshold_ms * 2):
            
            hints.append(OptimizationHint(
                hint_type="text_search_optimization",
                description=f"Text search queries in {stats.collection} are averaging {stats.avg_time_ms:.1f}ms",
                impact_level="high",
                suggested_action="Consider implementing full-text search indexes or search engine integration",
                estimated_improvement_percent=60.0
            ))
        
        return hints
    
    def _rule_large_result_sets(self, stats: QueryStats) -> List[OptimizationHint]:
        """Rule for queries returning large result sets"""
        hints = []
        
        if stats.avg_time_ms > self.slow_query_threshold_ms * 1.5:
            hints.append(OptimizationHint(
                hint_type="result_set_size",
                description=f"Query may be returning large result sets ({stats.avg_time_ms:.1f}ms avg)",
                impact_level="medium",
                suggested_action="Implement pagination with smaller page sizes or add result limiting",
                estimated_improvement_percent=40.0
            ))
        
        return hints
    
    def _rule_frequent_similar_queries(self, stats: QueryStats) -> List[OptimizationHint]:
        """Rule for frequently executed similar queries"""
        hints = []
        
        if stats.execution_count > 100 and stats.avg_time_ms > 50:
            hints.append(OptimizationHint(
                hint_type="caching_opportunity",
                description=f"Frequently executed query ({stats.execution_count} times) with moderate response time",
                impact_level="medium",
                suggested_action="Implement result caching with appropriate TTL",
                estimated_improvement_percent=70.0
            ))
        
        return hints
    
    def _rule_missing_indexes(self, stats: QueryStats) -> List[OptimizationHint]:
        """Rule for queries that might benefit from indexes"""
        hints = []
        
        if (stats.query_type in [QueryType.SEARCH, QueryType.FILTER] and
            stats.avg_time_ms > self.slow_query_threshold_ms):
            
            hints.append(OptimizationHint(
                hint_type="index_optimization",
                description=f"Filter/search queries on {stats.collection} may benefit from better indexing",
                impact_level="high",
                suggested_action="Analyze query patterns and create compound indexes on frequently filtered fields",
                estimated_improvement_percent=80.0
            ))
        
        return hints
    
    def _rule_inefficient_filters(self, stats: QueryStats) -> List[OptimizationHint]:
        """Rule for inefficient filter operations"""
        hints = []
        
        if (stats.query_type == QueryType.FILTER and 
            stats.slow_query_count > stats.execution_count * 0.3):  # 30% slow queries
            
            hints.append(OptimizationHint(
                hint_type="filter_optimization",
                description=f"Filter operations have high slow query rate ({stats.slow_query_count}/{stats.execution_count})",
                impact_level="medium",
                suggested_action="Review filter combinations and consider pre-computed filter results",
                estimated_improvement_percent=50.0
            ))
        
        return hints
    
    def _rule_pagination_performance(self, stats: QueryStats) -> List[OptimizationHint]:
        """Rule for pagination performance issues"""
        hints = []
        
        # This is a simplified rule - in practice, we'd analyze offset patterns
        if stats.avg_time_ms > self.slow_query_threshold_ms * 3:
            hints.append(OptimizationHint(
                hint_type="pagination_optimization", 
                description=f"High response times may indicate inefficient pagination",
                impact_level="medium",
                suggested_action="Consider cursor-based pagination instead of offset-based pagination",
                estimated_improvement_percent=45.0
            ))
        
        return hints
    
    async def _cleanup_old_stats(self) -> None:
        """Clean up old statistics to prevent memory growth"""
        if len(self._query_stats) <= self.max_stats_entries:
            return
        
        # Remove oldest entries based on last execution time
        sorted_stats = sorted(
            self._query_stats.items(),
            key=lambda x: x[1].last_execution or datetime.min
        )
        
        entries_to_remove = len(self._query_stats) - self.max_stats_entries
        for i in range(entries_to_remove):
            query_hash = sorted_stats[i][0]
            del self._query_stats[query_hash]
        
        logger.debug(f"Cleaned up {entries_to_remove} old query statistics")


# Global query optimizer instance
_query_optimizer: Optional[QueryOptimizer] = None


def get_query_optimizer() -> QueryOptimizer:
    """Get global query optimizer instance"""
    global _query_optimizer
    if _query_optimizer is None:
        _query_optimizer = QueryOptimizer()
    return _query_optimizer