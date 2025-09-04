"""
Administrative Endpoints for Hexagonal Architecture Monitoring
"""
from typing import Dict, Any, Optional
from fastapi import APIRouter, HTTPException, Query
from fastapi.responses import JSONResponse

import logging

# Import hexagonal architecture components if available
try:
    from infrastructure.performance import get_cache_manager, get_metrics_collector, get_query_optimizer
    from infrastructure.command_setup import get_command_bus, execute_cache_preload, execute_search_optimization
    from infrastructure.event_setup import get_metrics_summary, reset_metrics, get_cache_invalidation_keys
    from domain.commands import (
        CacheContentCommand, InvalidateCacheCommand, 
        OptimizeSearchCommand, RecordAnalyticsCommand
    )
    ADMIN_AVAILABLE = True
except ImportError as e:
    logging.warning(f"Admin endpoints not fully available: {e}")
    ADMIN_AVAILABLE = False

router = APIRouter(prefix="/admin", tags=["admin"])
logger = logging.getLogger(__name__)


@router.get("/status")
async def admin_status():
    """Administrative status overview"""
    if not ADMIN_AVAILABLE:
        return {
            "status": "limited",
            "message": "Advanced monitoring not available",
            "available_features": ["basic_health"]
        }
    
    try:
        # Get basic system status
        cache_manager = get_cache_manager()
        metrics_collector = get_metrics_collector()
        query_optimizer = get_query_optimizer()
        
        cache_stats = await cache_manager.get_stats()
        metrics_summary = await get_metrics_summary()
        
        return {
            "status": "fully_operational",
            "architecture": "hexagonal",
            "components": {
                "cache_manager": {
                    "status": "active",
                    "hit_ratio": cache_stats.hit_ratio,
                    "entries": cache_stats.total_entries,
                    "size_mb": cache_stats.total_size_bytes / 1024 / 1024
                },
                "metrics_collector": {
                    "status": "active" if metrics_collector._is_collecting else "inactive",
                    "total_events": metrics_summary.get("total_events", 0),
                    "collection_interval": metrics_collector.collection_interval_seconds
                },
                "query_optimizer": {
                    "status": "active",
                    "tracked_queries": len(query_optimizer._query_stats),
                    "slow_query_threshold_ms": query_optimizer.slow_query_threshold_ms
                }
            },
            "available_endpoints": [
                "/admin/cache-stats",
                "/admin/metrics", 
                "/admin/performance",
                "/admin/query-optimization",
                "/admin/cache-preload",
                "/admin/cache-invalidate"
            ]
        }
    except Exception as e:
        logger.error(f"Error getting admin status: {e}")
        raise HTTPException(status_code=500, detail=f"Error getting admin status: {e}")


@router.get("/cache-stats")
async def get_cache_stats():
    """Get detailed cache statistics"""
    if not ADMIN_AVAILABLE:
        raise HTTPException(status_code=503, detail="Cache statistics not available")
    
    try:
        cache_manager = get_cache_manager()
        cache_info = await cache_manager.get_cache_info()
        
        return {
            "timestamp": cache_info.get("generated_at"),
            "cache_statistics": cache_info["stats"],
            "cache_usage": cache_info["usage"], 
            "cache_configuration": cache_info["configuration"],
            "top_entries": cache_info["top_entries"][:10],  # Limit to top 10
            "recommendations": await _generate_cache_recommendations(cache_info)
        }
    except Exception as e:
        logger.error(f"Error getting cache stats: {e}")
        raise HTTPException(status_code=500, detail=f"Error getting cache stats: {e}")


@router.get("/metrics")
async def get_metrics():
    """Get comprehensive system metrics"""
    if not ADMIN_AVAILABLE:
        raise HTTPException(status_code=503, detail="Metrics not available")
    
    try:
        metrics_collector = get_metrics_collector()
        metrics_summary = await metrics_collector.get_metrics_summary()
        
        return {
            "timestamp": metrics_summary.get("timestamp"),
            "system_metrics": metrics_summary.get("system", {}),
            "application_metrics": metrics_summary.get("application", {}),
            "cache_metrics": metrics_summary.get("cache", {}),
            "database_metrics": metrics_summary.get("database", {}),
            "performance_trends": metrics_summary.get("trends", {}),
            "active_alerts": metrics_summary.get("alerts", []),
            "collection_info": metrics_summary.get("collection_info", {})
        }
    except Exception as e:
        logger.error(f"Error getting metrics: {e}")
        raise HTTPException(status_code=500, detail=f"Error getting metrics: {e}")


@router.get("/performance")
async def get_performance_report(
    collection: Optional[str] = Query(None, description="Filter by collection"),
    hours: int = Query(1, ge=1, le=24, description="Hours of history to analyze")
):
    """Get detailed performance analysis"""
    if not ADMIN_AVAILABLE:
        raise HTTPException(status_code=503, detail="Performance analysis not available")
    
    try:
        query_optimizer = get_query_optimizer()
        
        # Get performance report
        report = await query_optimizer.get_performance_report(collection=collection)
        
        # Get optimization suggestions
        suggestions = []
        if collection:
            suggestions = await query_optimizer.get_optimization_suggestions(collection, limit=10)
        
        return {
            "analysis_period_hours": hours,
            "collection_filter": collection,
            "summary": report.get("summary", {}),
            "slowest_queries": report.get("slow_queries", [])[:10],
            "optimization_suggestions": suggestions,
            "collection_breakdown": report.get("collection_analysis", {}),
            "generated_at": report.get("generated_at")
        }
    except Exception as e:
        logger.error(f"Error getting performance report: {e}")
        raise HTTPException(status_code=500, detail=f"Error getting performance report: {e}")


@router.get("/query-optimization")
async def get_query_optimization(
    collection: str = Query(..., description="Collection to optimize"),
    limit: int = Query(10, ge=1, le=50, description="Number of suggestions")
):
    """Get query optimization suggestions for a collection"""
    if not ADMIN_AVAILABLE:
        raise HTTPException(status_code=503, detail="Query optimization not available")
    
    try:
        query_optimizer = get_query_optimizer()
        suggestions = await query_optimizer.get_optimization_suggestions(collection, limit)
        
        return {
            "collection": collection,
            "suggestions_count": len(suggestions),
            "optimization_suggestions": suggestions,
            "recommendation_summary": await _generate_optimization_summary(suggestions)
        }
    except Exception as e:
        logger.error(f"Error getting query optimization: {e}")
        raise HTTPException(status_code=500, detail=f"Error getting query optimization: {e}")


@router.post("/cache-preload")
async def execute_cache_preload_endpoint(
    collections: Optional[list] = None
):
    """Execute cache preload for specified collections"""
    if not ADMIN_AVAILABLE:
        raise HTTPException(status_code=503, detail="Cache preload not available")
    
    try:
        if collections is None:
            collections = ["incantesimi", "mostri", "classi", "armi"]
        
        result = await execute_cache_preload(collections)
        
        return {
            "operation": "cache_preload",
            "collections": collections,
            "success": result["success"],
            "message": result["message"],
            "execution_time_ms": result["execution_time_ms"],
            "preloaded_items": result.get("data", {}).get("preloaded_count", 0)
        }
    except Exception as e:
        logger.error(f"Error executing cache preload: {e}")
        raise HTTPException(status_code=500, detail=f"Error executing cache preload: {e}")


@router.post("/cache-invalidate")
async def invalidate_cache_endpoint(
    cache_keys: Optional[list] = None,
    cache_patterns: Optional[list] = None,
    invalidate_all: bool = False
):
    """Invalidate cache entries"""
    if not ADMIN_AVAILABLE:
        raise HTTPException(status_code=503, detail="Cache invalidation not available")
    
    try:
        command_bus = get_command_bus()
        
        command = InvalidateCacheCommand(
            cache_keys=cache_keys or [],
            cache_patterns=cache_patterns or [],
            invalidate_all=invalidate_all
        )
        
        result = await command_bus.execute(command)
        
        return {
            "operation": "cache_invalidate",
            "success": result.success,
            "message": result.message,
            "invalidated_keys": result.data.get("invalidated_keys", []) if result.data else [],
            "execution_time_ms": result.execution_time_ms
        }
    except Exception as e:
        logger.error(f"Error invalidating cache: {e}")
        raise HTTPException(status_code=500, detail=f"Error invalidating cache: {e}")


@router.post("/optimize-search")
async def optimize_search_endpoint(
    collection: str = Query(..., description="Collection to optimize")
):
    """Execute search optimization for a collection"""
    if not ADMIN_AVAILABLE:
        raise HTTPException(status_code=503, detail="Search optimization not available")
    
    try:
        result = await execute_search_optimization(collection)
        
        return {
            "operation": "search_optimization",
            "collection": collection,
            "success": result["success"],
            "message": result["message"],
            "optimizations_applied": result.get("data", {}).get("optimizations", []),
            "execution_time_ms": result["execution_time_ms"]
        }
    except Exception as e:
        logger.error(f"Error optimizing search: {e}")
        raise HTTPException(status_code=500, detail=f"Error optimizing search: {e}")


@router.delete("/metrics")
async def reset_metrics_endpoint():
    """Reset all collected metrics"""
    if not ADMIN_AVAILABLE:
        raise HTTPException(status_code=503, detail="Metrics reset not available")
    
    try:
        reset_metrics()
        
        # Also reset query optimizer stats
        query_optimizer = get_query_optimizer()
        cleared_count = await query_optimizer.clear_stats()
        
        return {
            "operation": "metrics_reset",
            "success": True,
            "message": f"Reset all metrics and {cleared_count} query statistics",
            "timestamp": logger.info("Metrics manually reset via admin endpoint")
        }
    except Exception as e:
        logger.error(f"Error resetting metrics: {e}")
        raise HTTPException(status_code=500, detail=f"Error resetting metrics: {e}")


# Helper functions  
async def _generate_cache_recommendations(cache_info: Dict[str, Any]) -> list[Dict[str, str]]:
    """Generate cache optimization recommendations"""
    recommendations = []
    
    stats = cache_info.get("stats", {})
    usage = cache_info.get("usage", {})
    
    # Hit ratio recommendations
    if stats.get("hit_ratio", 0) < 0.5:
        recommendations.append({
            "type": "performance",
            "priority": "high",
            "recommendation": "Cache hit ratio is low (<50%). Consider increasing cache TTL or preloading frequently accessed content."
        })
    
    # Memory usage recommendations
    if usage.get("memory_usage_percent", 0) > 80:
        recommendations.append({
            "type": "memory",
            "priority": "medium", 
            "recommendation": "Cache memory usage is high (>80%). Consider increasing max cache size or implementing more aggressive eviction."
        })
    
    # Entry count recommendations
    if usage.get("entries_usage_percent", 0) > 90:
        recommendations.append({
            "type": "capacity",
            "priority": "medium",
            "recommendation": "Cache is near entry limit (>90%). Consider increasing max entries or reviewing cache strategy."
        })
    
    if not recommendations:
        recommendations.append({
            "type": "status",
            "priority": "info",
            "recommendation": "Cache performance looks good. No immediate optimizations needed."
        })
    
    return recommendations


async def _generate_optimization_summary(suggestions: list[Dict[str, Any]]) -> Dict[str, Any]:
    """Generate optimization summary from suggestions"""
    if not suggestions:
        return {"status": "optimal", "message": "No optimizations needed"}
    
    high_impact = sum(1 for s in suggestions if s.get("impact_level") == "high")
    medium_impact = sum(1 for s in suggestions if s.get("impact_level") == "medium") 
    low_impact = sum(1 for s in suggestions if s.get("impact_level") == "low")
    
    total_improvement = sum(s.get("estimated_improvement_percent", 0) for s in suggestions)
    
    return {
        "total_suggestions": len(suggestions),
        "impact_breakdown": {
            "high": high_impact,
            "medium": medium_impact, 
            "low": low_impact
        },
        "estimated_total_improvement_percent": round(total_improvement, 1),
        "primary_recommendation": suggestions[0].get("suggested_action") if suggestions else None,
        "status": "needs_optimization" if high_impact > 0 else "good"
    }