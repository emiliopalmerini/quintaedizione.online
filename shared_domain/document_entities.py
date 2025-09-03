"""
Document entities for D&D 5e SRD following ADR data model
Handles structured SRD documentation with nested paragraphs
"""
from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Any
import re


@dataclass(frozen=True)
class DocumentId:
    """Document entity identifier"""
    value: str
    
    def __post_init__(self):
        if not self.value or not self.value.strip():
            raise ValueError("DocumentId cannot be empty")
        if not re.match(r'^[a-z][a-z0-9-]*$', self.value):
            raise ValueError(f"Invalid document ID format: {self.value}")


@dataclass
class DocumentSubparagraph:
    """Sub-paragraph within a document paragraph"""
    numero: int
    titolo: str
    markdown: str
    
    def __post_init__(self):
        if self.numero < 1:
            raise ValueError("Subparagraph number must be >= 1")
        if not self.titolo.strip():
            raise ValueError("Subparagraph title cannot be empty")
        if not self.markdown.strip():
            raise ValueError("Subparagraph markdown cannot be empty")


@dataclass
class DocumentParagraphBody:
    """Body content of a document paragraph"""
    markdown: str
    sottoparagrafi: List[DocumentSubparagraph] = field(default_factory=list)
    
    def __post_init__(self):
        if not self.markdown.strip():
            raise ValueError("Paragraph markdown cannot be empty")
    
    def has_subparagraphs(self) -> bool:
        """Check if paragraph has sub-paragraphs"""
        return bool(self.sottoparagrafi)
    
    def get_subparagraph_count(self) -> int:
        """Get number of sub-paragraphs"""
        return len(self.sottoparagrafi)
    
    def get_full_content_length(self) -> int:
        """Get total character count including subparagraphs"""
        total = len(self.markdown)
        for sub in self.sottoparagrafi:
            total += len(sub.markdown)
        return total


@dataclass
class DocumentParagraph:
    """Main paragraph in SRD document"""
    numero: int
    titolo: str
    corpo: DocumentParagraphBody
    
    def __post_init__(self):
        if self.numero < 1:
            raise ValueError("Paragraph number must be >= 1")
        if not self.titolo.strip():
            raise ValueError("Paragraph title cannot be empty")
    
    def is_complex(self) -> bool:
        """Check if paragraph has complex structure"""
        return self.corpo.has_subparagraphs()
    
    def get_total_content_length(self) -> int:
        """Get total character count for paragraph"""
        return self.corpo.get_full_content_length()


@dataclass
class Document:
    """D&D 5e SRD Document entity"""
    id: DocumentId
    titolo: str
    paragrafi: List[DocumentParagraph]
    
    # Optional fields
    pagina: Optional[int] = None
    categoria: Optional[str] = None  # e.g., "Regole Base", "Creazione Personaggio"
    sommario: Optional[str] = None
    parole_chiave: List[str] = field(default_factory=list)
    fonte: str = "SRD"
    versione: str = "1.0"
    
    def __post_init__(self):
        if not self.titolo.strip():
            raise ValueError("Document title cannot be empty")
        if not self.paragrafi:
            raise ValueError("Document must have at least one paragraph")
    
    def get_paragraph_count(self) -> int:
        """Get number of main paragraphs"""
        return len(self.paragrafi)
    
    def get_subparagraph_count(self) -> int:
        """Get total number of sub-paragraphs across all paragraphs"""
        return sum(para.corpo.get_subparagraph_count() for para in self.paragrafi)
    
    def get_total_content_length(self) -> int:
        """Get total character count for entire document"""
        return sum(para.get_total_content_length() for para in self.paragrafi)
    
    def has_page_number(self) -> bool:
        """Check if document has page reference"""
        return self.pagina is not None
    
    def is_complex_document(self) -> bool:
        """Check if document has complex nested structure"""
        return any(para.is_complex() for para in self.paragrafi)
    
    def get_paragraph_by_number(self, numero: int) -> Optional[DocumentParagraph]:
        """Get paragraph by its number"""
        for para in self.paragrafi:
            if para.numero == numero:
                return para
        return None
    
    def get_subparagraph_by_path(self, para_num: int, sub_num: int) -> Optional[DocumentSubparagraph]:
        """Get subparagraph by paragraph and subparagraph numbers"""
        para = self.get_paragraph_by_number(para_num)
        if not para:
            return None
        
        for sub in para.corpo.sottoparagrafi:
            if sub.numero == sub_num:
                return sub
        return None
    
    def search_content(self, query: str) -> List[str]:
        """Search for content within document, return matching paragraph titles"""
        query_lower = query.lower()
        matches = []
        
        # Search in paragraph titles and content
        for para in self.paragrafi:
            if query_lower in para.titolo.lower() or query_lower in para.corpo.markdown.lower():
                matches.append(para.titolo)
            
            # Search in subparagraphs
            for sub in para.corpo.sottoparagrafi:
                if query_lower in sub.titolo.lower() or query_lower in sub.markdown.lower():
                    matches.append(f"{para.titolo} > {sub.titolo}")
        
        return matches
    
    def get_table_of_contents(self) -> List[Dict[str, Any]]:
        """Generate table of contents structure"""
        toc = []
        
        for para in self.paragrafi:
            para_entry = {
                "numero": para.numero,
                "titolo": para.titolo,
                "type": "paragraph",
                "content_length": para.get_total_content_length()
            }
            
            if para.corpo.has_subparagraphs():
                para_entry["sottoparagrafi"] = [
                    {
                        "numero": sub.numero,
                        "titolo": sub.titolo,
                        "type": "subparagraph",
                        "content_length": len(sub.markdown)
                    }
                    for sub in para.corpo.sottoparagrafi
                ]
            
            toc.append(para_entry)
        
        return toc
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "id": self.id.value,
            "titolo": self.titolo,
            "paragraph_count": self.get_paragraph_count(),
            "subparagraph_count": self.get_subparagraph_count(),
            "total_content_length": self.get_total_content_length(),
            "has_page_number": self.has_page_number(),
            "is_complex_document": self.is_complex_document(),
            "categoria": self.categoria,
            "parole_chiave": self.parole_chiave,
            "fonte": self.fonte,
            "versione": self.versione
        }


# Repository interfaces
class DocumentRepository(ABC):
    """Repository interface for document write operations"""
    
    @abstractmethod
    async def find_by_id(self, document_id: DocumentId) -> Optional[Document]:
        pass
    
    @abstractmethod
    async def find_by_title(self, title: str) -> Optional[Document]:
        pass
    
    @abstractmethod
    async def find_by_category(self, category: str) -> List[Document]:
        pass
    
    @abstractmethod
    async def save(self, document: Document) -> None:
        pass
    
    @abstractmethod
    async def find_all(self) -> List[Document]:
        pass


class DocumentQueryRepository(ABC):
    """Repository interface for document read operations (CQRS Query)"""
    
    @abstractmethod
    async def search_documents(self, query: 'DocumentSearchQuery') -> List['DocumentSummary']:
        pass
    
    @abstractmethod
    async def get_document_detail(self, document_id: DocumentId) -> Optional['DocumentDetail']:
        pass
    
    @abstractmethod
    async def search_document_content(self, text_query: str) -> List['DocumentContentMatch']:
        pass
    
    @abstractmethod
    async def get_documents_by_category(self, category: str) -> List['DocumentSummary']:
        pass
    
    @abstractmethod
    async def get_document_table_of_contents(self, document_id: DocumentId) -> Optional[List[Dict[str, Any]]]:
        pass


@dataclass
class DocumentValidationService:
    """Domain service for document validation"""
    
    @staticmethod
    def validate_document(document: Document) -> List[str]:
        """Validate document business rules"""
        errors = []
        
        # Check paragraph numbering
        paragraph_numbers = [para.numero for para in document.paragrafi]
        if len(set(paragraph_numbers)) != len(paragraph_numbers):
            errors.append("Document has duplicate paragraph numbers")
        
        # Check for reasonable paragraph count
        if len(document.paragrafi) > 20:
            errors.append("Document has unusually many paragraphs")
        
        # Check content length
        total_length = document.get_total_content_length()
        if total_length < 100:
            errors.append("Document content seems too short")
        elif total_length > 50000:
            errors.append("Document content seems unusually long")
        
        # Validate subparagraph numbering within each paragraph
        for para in document.paragrafi:
            sub_numbers = [sub.numero for sub in para.corpo.sottoparagrafi]
            if len(set(sub_numbers)) != len(sub_numbers):
                errors.append(f"Paragraph '{para.titolo}' has duplicate subparagraph numbers")
        
        return errors
    
    @staticmethod
    def suggest_missing_data(document: Document) -> List[str]:
        """Suggest potentially missing data"""
        suggestions = []
        
        if not document.pagina:
            suggestions.append("Consider adding page reference")
        
        if not document.categoria:
            suggestions.append("Consider adding document category")
        
        if not document.parole_chiave:
            suggestions.append("Consider adding keywords for search")
        
        if not document.sommario:
            suggestions.append("Consider adding document summary")
        
        # Check for very short paragraphs
        short_paragraphs = [
            para.titolo for para in document.paragrafi 
            if para.get_total_content_length() < 50
        ]
        if short_paragraphs:
            suggestions.append(f"Very short paragraphs: {', '.join(short_paragraphs)}")
        
        return suggestions