"""
MongoDB implementation for generic document queries optimized for read operations
Handles equipment, backgrounds, feats and other content types
"""
from typing import Dict, List, Optional, Any
import logging
from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorCollection

from shared_domain.document_entities import Document, DocumentId, DocumentQueryRepository
from shared_domain.query_models import DocumentSearchQuery, DocumentSummary, DocumentDetail, DocumentContentMatch

logger = logging.getLogger(__name__)


class MongoDBDocumentQueryRepository(DocumentQueryRepository):
    """MongoDB implementation optimized for generic document read operations"""
    
    def __init__(self, connection_string: str, database_name: str):
        self.client = AsyncIOMotorClient(connection_string)
        self.db = self.client[database_name]
        
        # Collection mapping for different document types
        self.collection_mapping = {
            "equipment": "equipaggiamento",
            "weapons": "armi", 
            "armor": "armature",
            "magic_items": "oggetti_magici",
            "backgrounds": "backgrounds",
            "feats": "talenti",
            "services": "servizi",
            "tools": "strumenti",
            "species": "specie",
        }
        
        # Ensure read-optimized indexes for all collections
        self._ensure_read_indexes()
    
    async def _ensure_read_indexes(self) -> None:
        """Create indexes optimized for read operations on all document collections"""
        try:
            for collection_name in self.collection_mapping.values():
                collection = self.db[collection_name]
                
                # Text search index for common fields
                await collection.create_index([
                    ("name", "text"),
                    ("nome", "text"),
                    ("title", "text"),
                    ("titolo", "text"),
                    ("description", "text"),
                    ("descrizione", "text")
                ])
                
                # Common filter indexes
                await collection.create_index("categoria")
                await collection.create_index("tipo")
                await collection.create_index("rarita")
                
            logger.info("Document read-optimized indexes ensured")
            
        except Exception as e:
            logger.warning(f"Could not create document indexes: {e}")
    
    def _get_collection(self, document_type: str) -> AsyncIOMotorCollection:
        """Get collection for document type"""
        collection_name = self.collection_mapping.get(document_type, document_type)
        return self.db[collection_name]
    
    async def find_by_id(self, document_id: DocumentId, document_type: str) -> Optional[Document]:
        """Find document by ID with full details"""
        try:
            collection = self._get_collection(document_type)
            doc = await collection.find_one({"_id": document_id.value})
            if not doc:
                return None
            return self._document_to_entity(doc, document_type)
        except Exception as e:
            logger.error(f"Error finding document by ID {document_id.value}: {e}")
            return None
    
    async def search_documents(self, query: DocumentSearchQuery) -> List[DocumentSummary]:
        """Search documents with filtering and return summaries"""
        try:
            collection = self._get_collection(query.document_type)
            mongo_query = self._build_search_query(query)
            
            # Use projection for performance - only summary fields
            projection = {
                "_id": 1,
                "name": 1,
                "nome": 1,
                "title": 1,
                "titolo": 1,
                "categoria": 1,
                "tipo": 1,
                "rarita": 1,
                "costo": 1,
                "peso": 1,
                "description": 1,
                "descrizione": 1
            }
            
            # Apply sorting and limits
            cursor = collection.find(mongo_query, projection)
            
            if query.sort_by == "name":
                cursor = cursor.sort([("nome", 1), ("name", 1), ("titolo", 1), ("title", 1)])
            elif query.sort_by == "category":
                cursor = cursor.sort("categoria", 1)
            elif query.sort_by == "type":
                cursor = cursor.sort("tipo", 1)
            elif query.sort_by == "cost":
                cursor = cursor.sort("costo", 1)
            
            if query.limit:
                cursor = cursor.limit(query.limit)
            if query.offset:
                cursor = cursor.skip(query.offset)
            
            docs = await cursor.to_list(length=None)
            # Return raw docs for now to preserve slug field for URL generation
            # TODO: Extend DocumentSummary to include slug field
            return docs
            
        except Exception as e:
            logger.error(f"Error in document search: {e}")
            return []
    
    async def get_documents_by_category(self, document_type: str, category: str) -> List[DocumentSummary]:
        """Get all documents of specific category"""
        try:
            collection = self._get_collection(document_type)
            docs = await collection.find(
                {"categoria": category},
                {
                    "_id": 1, "name": 1, "nome": 1, "title": 1, "titolo": 1,
                    "categoria": 1, "tipo": 1, "rarita": 1, "costo": 1, "peso": 1
                }
            ).sort([("nome", 1), ("name", 1), ("titolo", 1), ("title", 1)]).to_list(length=None)
            
            return docs
        except Exception as e:
            logger.error(f"Error getting documents by category {category}: {e}")
            return []
    
    async def get_documents_by_type(self, document_type: str, item_type: str) -> List[DocumentSummary]:
        """Get all documents of specific type"""
        try:
            collection = self._get_collection(document_type)
            docs = await collection.find(
                {"tipo": item_type},
                {
                    "_id": 1, "name": 1, "nome": 1, "title": 1, "titolo": 1,
                    "categoria": 1, "tipo": 1, "rarita": 1, "costo": 1, "peso": 1
                }
            ).sort([("nome", 1), ("name", 1), ("titolo", 1), ("title", 1)]).to_list(length=None)
            
            return docs
        except Exception as e:
            logger.error(f"Error getting documents by type {item_type}: {e}")
            return []
    
    async def get_distinct_values(self, document_type: str, field: str) -> List[str]:
        """Get distinct values for a field in document type"""
        try:
            collection = self._get_collection(document_type)
            values = await collection.distinct(field)
            # Filter out None, empty strings, and sort
            filtered_values = [v for v in values if v is not None and v != ""]
            return sorted(filtered_values)
        except Exception as e:
            logger.error(f"Error getting distinct values for {field}: {e}")
            return []
    
    # Implement missing abstract methods
    async def get_document_detail(self, document_id: DocumentId) -> Optional[DocumentDetail]:
        """Get detailed document information by ID"""
        try:
            # Try to find in any collection - we'll search through all mapped collections
            for doc_type, collection_name in self.collection_mapping.items():
                collection = self.db[collection_name]
                doc = await collection.find_one({"_id": document_id.value})
                if doc:
                    return self._document_to_detail(doc, doc_type)
            return None
        except Exception as e:
            logger.error(f"Error getting document detail for {document_id.value}: {e}")
            return None
    
    async def search_document_content(self, text_query: str) -> List[DocumentContentMatch]:
        """Search for content within documents across all collections"""
        try:
            matches = []
            query_lower = text_query.lower()
            
            # Search across all document collections
            for doc_type, collection_name in self.collection_mapping.items():
                collection = self.db[collection_name]
                
                # Find documents that match the text query
                mongo_query = {
                    "$or": [
                        {"name": {"$regex": text_query, "$options": "i"}},
                        {"nome": {"$regex": text_query, "$options": "i"}},
                        {"title": {"$regex": text_query, "$options": "i"}},
                        {"titolo": {"$regex": text_query, "$options": "i"}},
                        {"description": {"$regex": text_query, "$options": "i"}},
                        {"descrizione": {"$regex": text_query, "$options": "i"}},
                        {"contenuto": {"$regex": text_query, "$options": "i"}},
                        {"markdown": {"$regex": text_query, "$options": "i"}}
                    ]
                }
                
                cursor = collection.find(mongo_query).limit(20)  # Limit results per collection
                docs = await cursor.to_list(length=None)
                
                for doc in docs:
                    # Extract title
                    title = (doc.get("name") or doc.get("nome") or 
                           doc.get("title") or doc.get("titolo") or "")
                    
                    # Extract content for excerpt
                    content = (doc.get("description") or doc.get("descrizione") or 
                             doc.get("contenuto") or doc.get("markdown") or "")
                    
                    # Find the matching part for excerpt
                    content_lower = content.lower()
                    match_pos = content_lower.find(query_lower)
                    
                    if match_pos >= 0:
                        # Extract excerpt around the match
                        start = max(0, match_pos - 50)
                        end = min(len(content), match_pos + len(text_query) + 50)
                        excerpt = content[start:end].strip()
                        if start > 0:
                            excerpt = "..." + excerpt
                        if end < len(content):
                            excerpt = excerpt + "..."
                    else:
                        # Fallback - take first 100 chars
                        excerpt = content[:100] + "..." if len(content) > 100 else content
                    
                    # Calculate simple match score based on position and frequency
                    score = 1.0
                    if query_lower in title.lower():
                        score += 0.5  # Title matches get higher score
                    
                    matches.append(DocumentContentMatch(
                        document_id=str(doc["_id"]),
                        document_title=title,
                        paragraph_title=title,  # For now, use document title
                        subparagraph_title=None,
                        content_excerpt=excerpt,
                        match_score=score
                    ))
            
            # Sort by score descending and limit total results
            matches.sort(key=lambda x: x.match_score, reverse=True)
            return matches[:50]  # Return top 50 matches
            
        except Exception as e:
            logger.error(f"Error searching document content: {e}")
            return []
    
    async def get_document_table_of_contents(self, document_id: DocumentId) -> Optional[List[Dict[str, Any]]]:
        """Get table of contents for a document"""
        try:
            # Try to find in any collection
            for doc_type, collection_name in self.collection_mapping.items():
                collection = self.db[collection_name]
                doc = await collection.find_one({"_id": document_id.value})
                if doc:
                    return self._generate_table_of_contents(doc, doc_type)
            return None
        except Exception as e:
            logger.error(f"Error getting table of contents for {document_id.value}: {e}")
            return None
    
    def _build_search_query(self, query: DocumentSearchQuery) -> Dict[str, Any]:
        """Build MongoDB query from search parameters"""
        mongo_query = {}
        
        # Text search
        if query.text_query:
            mongo_query["$or"] = [
                {"name": {"$regex": query.text_query, "$options": "i"}},
                {"nome": {"$regex": query.text_query, "$options": "i"}},
                {"title": {"$regex": query.text_query, "$options": "i"}},
                {"titolo": {"$regex": query.text_query, "$options": "i"}},
                {"description": {"$regex": query.text_query, "$options": "i"}},
                {"descrizione": {"$regex": query.text_query, "$options": "i"}}
            ]
        
        # Common filters
        if query.category:
            mongo_query["categoria"] = query.category
        
        if query.item_type:
            mongo_query["tipo"] = query.item_type
        
        if query.rarity:
            mongo_query["rarita"] = query.rarity
        
        # Custom filters based on document type
        if query.filters:
            for key, value in query.filters.items():
                if value is not None and value != "":
                    mongo_query[key] = value
        
        return mongo_query
    
    def _document_to_summary(self, doc: Dict[str, Any], document_type: str) -> DocumentSummary:
        """Convert MongoDB document to DocumentSummary"""
        name = (doc.get("name") or doc.get("nome") or 
                doc.get("title") or doc.get("titolo") or "")
        
        # Get description preview (first 100 characters)
        description = (doc.get("description") or doc.get("descrizione") or "")
        description_preview = description[:100] + "..." if len(description) > 100 else description
        
        return DocumentSummary(
            id=str(doc["_id"]),
            titolo=name,
            categoria=doc.get("categoria"),
            pagina=doc.get("pagina"),
            paragraph_count=len(description.split('\n\n')) if description else 0,
            content_length=len(description),
            keywords=doc.get("keywords", [])
        )
    
    def _document_to_entity(self, doc: Dict[str, Any], document_type: str) -> Document:
        """Convert MongoDB document to full domain entity"""
        name = (doc.get("name") or doc.get("nome") or 
                doc.get("title") or doc.get("titolo") or "")
        description = (doc.get("description") or doc.get("descrizione") or "")
        
        return Document(
            id=DocumentId(str(doc["_id"])),
            name=name,
            document_type=document_type,
            content=description,
            metadata={
                "categoria": doc.get("categoria", ""),
                "tipo": doc.get("tipo", ""),
                "rarita": doc.get("rarita", ""),
                "costo": doc.get("costo", ""),
                "peso": doc.get("peso", ""),
                # Include all other fields as metadata
                **{k: v for k, v in doc.items() 
                   if k not in ["_id", "name", "nome", "title", "titolo", 
                               "description", "descrizione"]}
            },
            source="SRD"
        )
    
    def _document_to_detail(self, doc: Dict[str, Any], document_type: str) -> DocumentDetail:
        """Convert MongoDB document to DocumentDetail"""
        title = (doc.get("name") or doc.get("nome") or 
                doc.get("title") or doc.get("titolo") or "")
        
        # Generate table of contents structure
        toc = self._generate_table_of_contents(doc, document_type)
        
        # Extract basic content metrics
        content = (doc.get("description") or doc.get("descrizione") or 
                  doc.get("contenuto") or doc.get("markdown") or "")
        
        # Count paragraphs (simple heuristic)
        paragraph_count = len([p for p in content.split('\n\n') if p.strip()]) if content else 1
        subparagraph_count = content.count('###') if content else 0
        
        # Extract keywords from various fields
        keywords = []
        if doc.get("parole_chiave"):
            keywords.extend(doc.get("parole_chiave", []))
        if doc.get("categoria"):
            keywords.append(doc.get("categoria"))
        if doc.get("tipo"):
            keywords.append(doc.get("tipo"))
            
        return DocumentDetail(
            id=str(doc["_id"]),
            titolo=title,
            categoria=doc.get("categoria"),
            pagina=doc.get("pagina"),
            sommario=doc.get("sommario") or doc.get("summary"),
            paragraph_count=paragraph_count,
            subparagraph_count=subparagraph_count,
            total_content_length=len(content),
            keywords=keywords,
            table_of_contents=toc
        )
    
    def _generate_table_of_contents(self, doc: Dict[str, Any], document_type: str) -> List[Dict[str, Any]]:
        """Generate table of contents from document structure"""
        toc = []
        
        # Get the main content
        content = (doc.get("description") or doc.get("descrizione") or 
                  doc.get("contenuto") or doc.get("markdown") or "")
        
        if not content:
            return toc
        
        # Parse markdown headers for structure
        lines = content.split('\n')
        current_section = None
        section_counter = 1
        subsection_counter = 1
        
        for line in lines:
            line = line.strip()
            
            # Main sections (## headers)
            if line.startswith('## '):
                if current_section:
                    toc.append(current_section)
                
                title = line[3:].strip()
                current_section = {
                    "numero": section_counter,
                    "titolo": title,
                    "type": "paragraph",
                    "content_length": 0,
                    "sottoparagrafi": []
                }
                section_counter += 1
                subsection_counter = 1
            
            # Subsections (### headers)
            elif line.startswith('### ') and current_section:
                title = line[4:].strip()
                subsection = {
                    "numero": subsection_counter,
                    "titolo": title,
                    "type": "subparagraph",
                    "content_length": 0
                }
                current_section["sottoparagrafi"].append(subsection)
                subsection_counter += 1
            
            # Count content length (simple heuristic)
            elif line and current_section:
                current_section["content_length"] += len(line)
                if current_section["sottoparagrafi"]:
                    current_section["sottoparagrafi"][-1]["content_length"] += len(line)
        
        # Add the last section
        if current_section:
            toc.append(current_section)
        
        # If no headers found, create a simple single-section TOC
        if not toc:
            title = (doc.get("name") or doc.get("nome") or 
                    doc.get("title") or doc.get("titolo") or "Content")
            toc.append({
                "numero": 1,
                "titolo": title,
                "type": "paragraph",
                "content_length": len(content),
                "sottoparagrafi": []
            })
        
        return toc
    
    async def close(self) -> None:
        """Close database connection"""
        self.client.close()