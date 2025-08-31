"""Optimized database queries with batch operations for D&D 5e SRD Editor."""
from __future__ import annotations

import asyncio
import math
import time
from typing import Any, Dict, List, Optional, Tuple, Union
from dataclasses import dataclass

from motor.motor_asyncio import AsyncIOMotorDatabase, AsyncIOMotorCollection
from pymongo import ASCENDING, DESCENDING

from core.cache import get_cache_manager, cached, cache_key_for_collection_counts, cache_key_for_search
from core.logging_config import get_logger, log_database_operation
from core.errors import DatabaseError, ErrorCode, safe_db_operation

logger = get_logger(__name__)


@dataclass
class QueryOptions:
    """Options for database queries."""
    
    skip: int = 0
    limit: int = 20
    sort: Optional[List[Tuple[str, int]]] = None
    projection: Optional[Dict[str, int]] = None
    hint: Optional[str] = None  # Index hint
    max_time_ms: Optional[int] = 30000  # 30 second query timeout
    
    def __post_init__(self):
        if self.sort is None:
            self.sort = [("_sortkey_alpha", ASCENDING)]


@dataclass
class PaginatedResult:
    """Result of paginated query."""
    
    items: List[Dict[str, Any]]
    total_count: int
    page: int
    page_size: int
    total_pages: int
    has_next: bool
    has_previous: bool


@dataclass
class AggregationResult:
    """Result of aggregation query."""
    
    data: List[Dict[str, Any]]
    total_count: int
    execution_time_ms: float


class OptimizedQueryService:
    """Service for optimized database queries with caching and batch operations."""
    
    def __init__(self, db: AsyncIOMotorDatabase):
        self.db = db
        self.cache = None  # Will be set lazily
    
    async def _get_cache(self):
        """Get cache manager instance."""
        if self.cache is None:
            self.cache = await get_cache_manager()
        return self.cache
    
    async def get_collection_counts_batch(self, collections: List[str], lang: str = "it") -> Dict[str, int]:
        """Get document counts for multiple collections efficiently."""
        cache = await self._get_cache()
        cache_key = cache_key_for_collection_counts(lang)
        
        # Try cache first
        cached_counts = await cache.get(cache_key)
        if cached_counts is not None:
            logger.debug(f"Collection counts cache hit for language: {lang}")
            return {col: cached_counts.get(col, 0) for col in collections}
        
        logger.info(f"Fetching collection counts for {len(collections)} collections", extra={"lang": lang})
        start_time = time.time()
        
        # Build collection name mapping
        collection_mapping = {}
        for collection in collections:
            db_collection = self._get_db_collection_name(collection, lang)
            collection_mapping[collection] = db_collection
        
        # Execute count operations concurrently
        async def count_collection(collection_name: str, db_collection_name: str) -> Tuple[str, int]:
            try:
                count = await self.db[db_collection_name].estimated_document_count()
                logger.debug(f"Collection {collection_name} ({db_collection_name}): {count} documents")
                return collection_name, count
            except Exception as e:
                logger.warning(f"Failed to count collection {collection_name}: {e}")
                return collection_name, 0
        
        # Execute all counts concurrently
        tasks = [
            count_collection(col_name, db_col_name)
            for col_name, db_col_name in collection_mapping.items()
        ]
        
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        # Process results
        counts = {}
        for result in results:
            if isinstance(result, Exception):
                logger.error(f"Error in batch count operation: {result}")
                continue
            
            collection_name, count = result
            counts[collection_name] = count
        
        execution_time = (time.time() - start_time) * 1000
        log_database_operation(
            logger,
            "batch_count",
            f"{len(collections)} collections",
            duration_ms=execution_time
        )
        
        # Cache results for 5 minutes
        await cache.set(cache_key, counts, ttl=300)
        
        return counts
    
    async def paginated_find(
        self,
        collection: str,
        filter_query: Dict[str, Any],
        options: QueryOptions,
        lang: str = "it"
    ) -> PaginatedResult:
        """Execute optimized paginated find query."""
        db_collection = self._get_db_collection_name(collection, lang)
        collection_obj = self.db[db_collection]
        
        logger.debug(f"Executing paginated find on {db_collection}", extra={
            "filter": filter_query,
            "skip": options.skip,
            "limit": options.limit,
            "sort": options.sort
        })
        
        start_time = time.time()
        
        # Use aggregation pipeline for better performance with counting
        pipeline = [
            {"$match": filter_query},
            {
                "$facet": {
                    "data": [
                        {"$sort": dict(options.sort)},
                        {"$skip": options.skip},
                        {"$limit": options.limit}
                    ],
                    "count": [
                        {"$count": "total"}
                    ]
                }
            }
        ]
        
        # Add projection if specified
        if options.projection:
            pipeline[1]["$facet"]["data"].insert(-2, {"$project": options.projection})
        
        # Set hint if specified
        aggregate_options = {}
        if options.hint:
            aggregate_options["hint"] = options.hint
        if options.max_time_ms:
            aggregate_options["maxTimeMS"] = options.max_time_ms
        
        try:
            cursor = collection_obj.aggregate(pipeline, **aggregate_options)
            result = await cursor.to_list(length=1)
            
            if not result:
                return PaginatedResult(
                    items=[],
                    total_count=0,
                    page=1,
                    page_size=options.limit,
                    total_pages=0,
                    has_next=False,
                    has_previous=False
                )
            
            data = result[0]
            items = data.get("data", [])
            total_count = data.get("count", [{}])[0].get("total", 0)
            
            # Calculate pagination metadata
            page = (options.skip // options.limit) + 1
            total_pages = math.ceil(total_count / options.limit) if total_count > 0 else 1
            has_next = page < total_pages
            has_previous = page > 1
            
            execution_time = (time.time() - start_time) * 1000
            log_database_operation(
                logger,
                "paginated_find",
                db_collection,
                filter_doc=filter_query,
                duration_ms=execution_time
            )
            
            return PaginatedResult(
                items=items,
                total_count=total_count,
                page=page,
                page_size=options.limit,
                total_pages=total_pages,
                has_next=has_next,
                has_previous=has_previous
            )
            
        except Exception as e:
            logger.error(f"Paginated find failed on {db_collection}: {e}", exc_info=e)
            raise DatabaseError(
                f"Paginated query failed: {str(e)}",
                ErrorCode.DATABASE_OPERATION_FAILED,
                operation="paginated_find",
                collection=db_collection
            )
    
    async def search_documents(
        self,
        collection: str,
        search_query: str,
        filter_query: Dict[str, Any],
        options: QueryOptions,
        lang: str = "it"
    ) -> PaginatedResult:
        """Execute optimized text search with caching."""
        cache = await self._get_cache()
        cache_key = cache_key_for_search(collection, search_query, filter_query, lang)
        
        # Try cache first for expensive search operations
        if search_query:  # Only cache actual search queries
            cached_result = await cache.get(cache_key)
            if cached_result is not None:
                logger.debug(f"Search cache hit for: {search_query}")
                return PaginatedResult(**cached_result)
        
        db_collection = self._get_db_collection_name(collection, lang)
        collection_obj = self.db[db_collection]
        
        logger.info(f"Executing search on {db_collection}", extra={
            "query": search_query,
            "filter": filter_query
        })
        
        start_time = time.time()
        
        # Build search pipeline
        pipeline = []
        
        # Add text search stage if query provided
        if search_query:
            # Use regex search for flexibility (text index might not be available)
            search_conditions = {
                "$or": [
                    {"title": {"$regex": search_query, "$options": "i"}},
                    {"content": {"$regex": search_query, "$options": "i"}},
                ]
            }
            
            # Combine with filter query
            if filter_query:
                combined_filter = {"$and": [search_conditions, filter_query]}
            else:
                combined_filter = search_conditions
        else:
            combined_filter = filter_query or {}
        
        pipeline.append({"$match": combined_filter})
        
        # Add score-based sorting for search queries
        if search_query:
            # Add relevance scoring based on title matches
            pipeline.extend([
                {
                    "$addFields": {
                        "relevance_score": {
                            "$cond": [
                                {"$regexMatch": {"input": "$title", "regex": search_query, "options": "i"}},
                                10,  # Higher score for title matches
                                1    # Lower score for content matches
                            ]
                        }
                    }
                },
                {"$sort": {"relevance_score": DESCENDING, "_sortkey_alpha": ASCENDING}}
            ])
        else:
            pipeline.append({"$sort": dict(options.sort)})
        
        # Add faceted aggregation for pagination
        pipeline.append({
            "$facet": {
                "data": [
                    {"$skip": options.skip},
                    {"$limit": options.limit}
                ],
                "count": [
                    {"$count": "total"}
                ]
            }
        })
        
        # Add projection if specified
        if options.projection:
            pipeline[-1]["$facet"]["data"].insert(0, {"$project": options.projection})
        
        try:
            cursor = collection_obj.aggregate(pipeline, maxTimeMS=options.max_time_ms)
            result = await cursor.to_list(length=1)
            
            if not result:
                return PaginatedResult(
                    items=[],
                    total_count=0,
                    page=1,
                    page_size=options.limit,
                    total_pages=0,
                    has_next=False,
                    has_previous=False
                )
            
            data = result[0]
            items = data.get("data", [])
            total_count = data.get("count", [{}])[0].get("total", 0)
            
            # Calculate pagination metadata
            page = (options.skip // options.limit) + 1
            total_pages = math.ceil(total_count / options.limit) if total_count > 0 else 1
            
            result_obj = PaginatedResult(
                items=items,
                total_count=total_count,
                page=page,
                page_size=options.limit,
                total_pages=total_pages,
                has_next=page < total_pages,
                has_previous=page > 1
            )
            
            execution_time = (time.time() - start_time) * 1000
            log_database_operation(
                logger,
                "search_documents",
                db_collection,
                filter_doc=combined_filter,
                duration_ms=execution_time
            )
            
            # Cache search results for 2 minutes
            if search_query:
                await cache.set(cache_key, {
                    "items": items,
                    "total_count": total_count,
                    "page": page,
                    "page_size": options.limit,
                    "total_pages": total_pages,
                    "has_next": result_obj.has_next,
                    "has_previous": result_obj.has_previous
                }, ttl=120)
            
            return result_obj
            
        except Exception as e:
            logger.error(f"Search failed on {db_collection}: {e}", exc_info=e)
            raise DatabaseError(
                f"Search query failed: {str(e)}",
                ErrorCode.SEARCH_ENGINE_ERROR,
                operation="search_documents",
                collection=db_collection,
                context={"query": search_query, "filter": filter_query}
            )
    
    async def get_document_with_navigation(
        self,
        collection: str,
        slug: str,
        lang: str = "it"
    ) -> Dict[str, Any]:
        """Get document with previous/next navigation efficiently."""
        db_collection = self._get_db_collection_name(collection, lang)
        collection_obj = self.db[db_collection]
        
        logger.debug(f"Getting document with navigation: {slug} from {db_collection}")
        start_time = time.time()
        
        # First, get the current document
        current_doc = await collection_obj.find_one({"slug": slug})
        if not current_doc:
            return {"doc": None, "prev_doc": None, "next_doc": None}
        
        # Use aggregation to get prev/next efficiently
        current_sort_key = current_doc.get("_sortkey_alpha", "")
        
        pipeline = [
            {
                "$facet": {
                    "prev": [
                        {"$match": {"_sortkey_alpha": {"$lt": current_sort_key}}},
                        {"$sort": {"_sortkey_alpha": DESCENDING}},
                        {"$limit": 1},
                        {"$project": {"title": 1, "slug": 1, "_sortkey_alpha": 1}}
                    ],
                    "next": [
                        {"$match": {"_sortkey_alpha": {"$gt": current_sort_key}}},
                        {"$sort": {"_sortkey_alpha": ASCENDING}},
                        {"$limit": 1},
                        {"$project": {"title": 1, "slug": 1, "_sortkey_alpha": 1}}
                    ]
                }
            }
        ]
        
        try:
            cursor = collection_obj.aggregate(pipeline)
            nav_result = await cursor.to_list(length=1)
            
            prev_doc = nav_result[0]["prev"][0] if nav_result and nav_result[0]["prev"] else None
            next_doc = nav_result[0]["next"][0] if nav_result and nav_result[0]["next"] else None
            
            execution_time = (time.time() - start_time) * 1000
            log_database_operation(
                logger,
                "get_document_with_navigation",
                db_collection,
                filter_doc={"slug": slug},
                duration_ms=execution_time
            )
            
            return {
                "doc": current_doc,
                "prev_doc": prev_doc,
                "next_doc": next_doc
            }
            
        except Exception as e:
            logger.error(f"Navigation query failed for {slug}: {e}", exc_info=e)
            # Fall back to just the document without navigation
            return {"doc": current_doc, "prev_doc": None, "next_doc": None}
    
    async def aggregate_with_caching(
        self,
        collection: str,
        pipeline: List[Dict[str, Any]],
        lang: str = "it",
        cache_ttl: int = 300
    ) -> AggregationResult:
        """Execute aggregation pipeline with caching."""
        cache = await self._get_cache()
        cache_key = cache._make_key("agg", collection, lang, cache._hash_key(pipeline))
        
        # Try cache first
        cached_result = await cache.get(cache_key)
        if cached_result is not None:
            logger.debug(f"Aggregation cache hit for collection: {collection}")
            return AggregationResult(**cached_result)
        
        db_collection = self._get_db_collection_name(collection, lang)
        collection_obj = self.db[db_collection]
        
        logger.info(f"Executing aggregation on {db_collection}", extra={"stages": len(pipeline)})
        start_time = time.time()
        
        try:
            cursor = collection_obj.aggregate(pipeline)
            data = await cursor.to_list(length=None)
            
            execution_time = (time.time() - start_time) * 1000
            
            result = AggregationResult(
                data=data,
                total_count=len(data),
                execution_time_ms=execution_time
            )
            
            log_database_operation(
                logger,
                "aggregation",
                db_collection,
                duration_ms=execution_time
            )
            
            # Cache result
            await cache.set(cache_key, {
                "data": data,
                "total_count": len(data),
                "execution_time_ms": execution_time
            }, ttl=cache_ttl)
            
            return result
            
        except Exception as e:
            logger.error(f"Aggregation failed on {db_collection}: {e}", exc_info=e)
            raise DatabaseError(
                f"Aggregation query failed: {str(e)}",
                ErrorCode.DATABASE_OPERATION_FAILED,
                operation="aggregation",
                collection=db_collection
            )
    
    async def batch_find_by_ids(
        self,
        collection: str,
        ids: List[str],
        lang: str = "it",
        projection: Optional[Dict[str, int]] = None
    ) -> Dict[str, Dict[str, Any]]:
        """Efficiently find multiple documents by ID."""
        if not ids:
            return {}
        
        db_collection = self._get_db_collection_name(collection, lang)
        collection_obj = self.db[db_collection]
        
        logger.debug(f"Batch find {len(ids)} documents from {db_collection}")
        start_time = time.time()
        
        try:
            query = {"_id": {"$in": ids}}
            cursor = collection_obj.find(query, projection)
            documents = await cursor.to_list(length=None)
            
            # Index by ID for easy lookup
            result = {doc["_id"]: doc for doc in documents}
            
            execution_time = (time.time() - start_time) * 1000
            log_database_operation(
                logger,
                "batch_find_by_ids",
                db_collection,
                filter_doc={"count": len(ids)},
                duration_ms=execution_time
            )
            
            return result
            
        except Exception as e:
            logger.error(f"Batch find failed on {db_collection}: {e}", exc_info=e)
            raise DatabaseError(
                f"Batch find query failed: {str(e)}",
                ErrorCode.DATABASE_OPERATION_FAILED,
                operation="batch_find_by_ids",
                collection=db_collection
            )
    
    def _get_db_collection_name(self, collection: str, lang: str) -> str:
        """Get database collection name based on logical collection and language."""
        # This should match the logic from core.config.db_collection_for
        if lang and lang.lower().startswith("en"):
            if collection in ["incantesimi", "oggetti_magici", "mostri", "classi", "razze"]:
                return f"{collection}_en"
            elif collection == "documenti":
                return "documenti_en"
        
        return collection


# Factory function
def create_optimized_query_service(db: AsyncIOMotorDatabase) -> OptimizedQueryService:
    """Create optimized query service instance."""
    return OptimizedQueryService(db)