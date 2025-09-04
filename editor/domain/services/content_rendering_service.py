"""
Content Rendering Domain Service
Pure business logic for document content rendering without infrastructure dependencies
"""
from typing import Dict, Any, Optional, Tuple
from dataclasses import dataclass
from enum import Enum


class ContentFormat(Enum):
    """Content format types"""
    MARKDOWN = "markdown"
    HTML = "html"
    PLAIN_TEXT = "plain_text"
    MIXED = "mixed"


@dataclass
class ContentMetadata:
    """Content metadata for rendering decisions"""
    has_markdown_syntax: bool
    has_headers: bool
    content_length: int
    field_name: str
    is_markdown_field: bool


@dataclass
class RenderedContent:
    """Result of content rendering operation"""
    html_content: str
    raw_content: str
    format_used: ContentFormat
    metadata: ContentMetadata


class ContentRenderingService:
    """Domain service for document content rendering business logic"""
    
    @staticmethod
    def analyze_content_format(content: str, field_name: str) -> ContentMetadata:
        """Analyze content to determine optimal rendering strategy"""
        if not content:
            return ContentMetadata(
                has_markdown_syntax=False,
                has_headers=False,
                content_length=0,
                field_name=field_name,
                is_markdown_field=False
            )
        
        # Check for markdown field naming convention
        is_markdown_field = field_name.endswith("_md")
        
        # Analyze content for markdown patterns
        lines = content.split('\n')
        has_headers = any(
            line.strip().startswith(('#', '##', '###', '####', '#####', '######')) 
            for line in lines
        )
        
        markdown_indicators = ['**', '*', '`', '---', '###', '##', '#', '> ', '- ', '* ', '1. ']
        has_markdown_syntax = any(
            indicator in content 
            for indicator in markdown_indicators
        )
        
        return ContentMetadata(
            has_markdown_syntax=has_markdown_syntax,
            has_headers=has_headers,
            content_length=len(content),
            field_name=field_name,
            is_markdown_field=is_markdown_field
        )
    
    @staticmethod
    def should_render_as_markdown(metadata: ContentMetadata) -> bool:
        """Business rule: decide whether to render content as markdown"""
        # Rule 1: Always render markdown fields as markdown
        if metadata.is_markdown_field:
            return True
        
        # Rule 2: Render as markdown if content field has clear markdown syntax
        if (metadata.field_name == "content" and 
            (metadata.has_headers or metadata.has_markdown_syntax)):
            return True
        
        # Rule 3: Don't render other fields as markdown
        return False
    
    @staticmethod
    def extract_content_from_document(document: Dict[str, Any]) -> Tuple[Optional[str], Optional[str]]:
        """Extract content and field name from document following priority rules"""
        # Priority order for content extraction
        content_field_priority = [
            "description_md", "descrizione_md",  # Highest priority: markdown fields
            "content",                           # Medium priority: content field
            "description", "descrizione"         # Lowest priority: plain description
        ]
        
        for field_name in content_field_priority:
            if field_name in document and document[field_name]:
                return document[field_name], field_name
        
        return None, None
    
    @staticmethod
    def prepare_content_for_rendering(raw_content: str, metadata: ContentMetadata) -> str:
        """Apply content preprocessing before rendering"""
        if not raw_content:
            return ""
        
        # Business rule: Clean up common formatting issues
        content = raw_content.strip()
        
        # Remove excessive line breaks (more than 2 consecutive)
        import re
        content = re.sub(r'\n{3,}', '\n\n', content)
        
        # Ensure proper spacing around headers
        if metadata.has_headers:
            content = re.sub(r'\n(#{1,6}\s)', r'\n\n\1', content)
            content = re.sub(r'(#{1,6}\s[^\n]+)\n(?!\n)', r'\1\n\n', content)
        
        return content
    
    def render_document_content(
        self, 
        document: Dict[str, Any],
        markdown_renderer=None
    ) -> Optional[RenderedContent]:
        """
        Main business logic for rendering document content
        Accepts optional markdown_renderer to avoid infrastructure dependency
        """
        # Step 1: Extract content
        raw_content, field_name = self.extract_content_from_document(document)
        if not raw_content or not field_name:
            return None
        
        # Step 2: Analyze content
        metadata = self.analyze_content_format(raw_content, field_name)
        
        # Step 3: Prepare content
        prepared_content = self.prepare_content_for_rendering(raw_content, metadata)
        
        # Step 4: Determine rendering strategy
        should_markdown = self.should_render_as_markdown(metadata)
        
        if should_markdown and markdown_renderer:
            # Render as markdown using provided renderer
            html_content = markdown_renderer(prepared_content)
            format_used = ContentFormat.MARKDOWN
        elif should_markdown and not markdown_renderer:
            # Fallback: return as plain text with warning
            html_content = prepared_content
            format_used = ContentFormat.PLAIN_TEXT
        else:
            # Render as plain text
            html_content = prepared_content
            format_used = ContentFormat.PLAIN_TEXT
        
        return RenderedContent(
            html_content=html_content,
            raw_content=raw_content,
            format_used=format_used,
            metadata=metadata
        )


class ContentValidationService:
    """Domain service for content validation rules"""
    
    @staticmethod
    def validate_document_completeness(document: Dict[str, Any]) -> Dict[str, Any]:
        """Validate document has required fields and content quality"""
        validation_result = {
            "is_valid": True,
            "warnings": [],
            "errors": [],
            "completeness_score": 100
        }
        
        # Required fields check
        required_fields = ["name", "nome"]  # At least one name field required
        if not any(field in document and document[field] for field in required_fields):
            validation_result["errors"].append("Missing document name")
            validation_result["is_valid"] = False
            validation_result["completeness_score"] -= 50
        
        # Content presence check
        content_fields = ["description", "descrizione", "content", "description_md", "descrizione_md"]
        has_content = any(field in document and document[field] for field in content_fields)
        if not has_content:
            validation_result["warnings"].append("No content field found")
            validation_result["completeness_score"] -= 30
        
        # Content quality check
        for field in content_fields:
            if field in document and document[field]:
                content = document[field]
                if len(content) < 20:
                    validation_result["warnings"].append(f"Short content in {field}")
                    validation_result["completeness_score"] -= 10
        
        return validation_result
    
    @staticmethod
    def sanitize_document_for_display(document: Dict[str, Any]) -> Dict[str, Any]:
        """Apply business rules for safe document display"""
        sanitized = document.copy()
        
        # Remove internal fields that shouldn't be displayed
        internal_fields = ["_id", "created_at", "updated_at", "version"]
        for field in internal_fields:
            sanitized.pop(field, None)
        
        # Ensure display name is available
        if "nome" not in sanitized and "name" in sanitized:
            sanitized["nome"] = sanitized["name"]
        elif "name" not in sanitized and "nome" in sanitized:
            sanitized["name"] = sanitized["nome"]
        
        return sanitized