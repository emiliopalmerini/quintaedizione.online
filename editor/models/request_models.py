"""Request validation models for D&D 5e SRD Editor."""
from __future__ import annotations

from typing import Optional, Literal, Any, Dict
from pydantic import BaseModel, Field, validator, root_validator
import re


class PaginationParams(BaseModel):
    """Pagination parameters validation."""
    
    page: int = Field(default=1, ge=1, le=10000, description="Page number")
    page_size: int = Field(default=20, ge=1, le=100, description="Items per page")
    
    class Config:
        extra = "forbid"


class LanguageParams(BaseModel):
    """Language selection parameters."""
    
    lang: Literal["it", "en"] = Field(default="it", description="Content language")
    
    class Config:
        extra = "forbid"


class SearchQuery(BaseModel):
    """Search query validation."""
    
    q: Optional[str] = Field(
        default=None, 
        min_length=1, 
        max_length=200,
        description="Search query string"
    )
    collection: Optional[str] = Field(
        default=None,
        description="Collection to search in"
    )
    
    @validator('q')
    def validate_search_query(cls, v):
        if not v:
            return v
            
        # Remove potentially dangerous patterns
        if re.search(r'[{}$]', v):
            raise ValueError('Query contains invalid characters')
        
        # Strip whitespace and normalize
        v = v.strip()
        if not v:
            raise ValueError('Query cannot be empty after cleaning')
            
        return v
    
    @validator('collection')
    def validate_collection(cls, v):
        if not v:
            return v
            
        # Collection name validation (alphanumeric + underscore)
        if not re.match(r'^[a-zA-Z][a-zA-Z0-9_]*$', v):
            raise ValueError('Invalid collection name format')
            
        return v
    
    class Config:
        extra = "forbid"


class FilterParams(BaseModel):
    """Advanced filtering parameters."""
    
    level: Optional[int] = Field(default=None, ge=0, le=9, description="Spell level")
    school: Optional[str] = Field(default=None, max_length=50, description="Magic school")
    class_name: Optional[str] = Field(default=None, max_length=50, description="Character class")
    ritual: Optional[bool] = Field(default=None, description="Ritual spell filter")
    concentration: Optional[bool] = Field(default=None, description="Concentration spell filter")
    translated: Optional[bool] = Field(default=None, description="Show only translated items")
    modified: Optional[bool] = Field(default=None, description="Show only modified items")
    
    @validator('school', 'class_name')
    def validate_string_fields(cls, v):
        if not v:
            return v
            
        # Allow only letters, spaces, hyphens, and apostrophes
        if not re.match(r'^[a-zA-Z\s\-\']+$', v):
            raise ValueError('Field contains invalid characters')
            
        return v.strip()
    
    class Config:
        extra = "forbid"


class ListPageParams(PaginationParams, LanguageParams, FilterParams, SearchQuery):
    """Combined parameters for list page requests."""
    
    sort: Optional[Literal["alpha", "level", "school", "modified"]] = Field(
        default="alpha",
        description="Sort order"
    )
    
    class Config:
        extra = "forbid"


class ShowPageParams(LanguageParams):
    """Parameters for show page requests."""
    
    slug: str = Field(..., min_length=1, max_length=200, description="Document slug")
    
    @validator('slug')
    def validate_slug(cls, v):
        # Slug validation: alphanumeric, hyphens, underscores
        if not re.match(r'^[a-zA-Z0-9\-_]+$', v):
            raise ValueError('Invalid slug format')
        return v
    
    class Config:
        extra = "forbid"


class EditPageParams(LanguageParams):
    """Parameters for edit page requests."""
    
    slug: str = Field(..., min_length=1, max_length=200, description="Document slug")
    
    @validator('slug')
    def validate_slug(cls, v):
        if not re.match(r'^[a-zA-Z0-9\-_]+$', v):
            raise ValueError('Invalid slug format')
        return v
    
    class Config:
        extra = "forbid"


class DocumentUpdateData(BaseModel):
    """Document update validation."""
    
    title: Optional[str] = Field(default=None, min_length=1, max_length=500)
    content: Optional[str] = Field(default=None, min_length=1, max_length=50000)
    slug: Optional[str] = Field(default=None, min_length=1, max_length=200)
    
    # Spell-specific fields
    level: Optional[int] = Field(default=None, ge=0, le=9)
    school: Optional[str] = Field(default=None, max_length=50)
    ritual: Optional[bool] = Field(default=None)
    concentration: Optional[bool] = Field(default=None)
    casting_time: Optional[str] = Field(default=None, max_length=100)
    range: Optional[str] = Field(default=None, max_length=100)
    components: Optional[str] = Field(default=None, max_length=200)
    duration: Optional[str] = Field(default=None, max_length=100)
    classes: Optional[list[str]] = Field(default=None)
    
    # Item-specific fields
    type: Optional[str] = Field(default=None, max_length=50)
    rarity: Optional[str] = Field(default=None, max_length=50)
    attunement: Optional[bool] = Field(default=None)
    
    # Metadata
    tags: Optional[list[str]] = Field(default=None)
    notes: Optional[str] = Field(default=None, max_length=1000)
    
    @validator('title', 'content', 'casting_time', 'range', 'components', 'duration', 'type', 'rarity')
    def validate_text_fields(cls, v):
        if v is None:
            return v
        return v.strip()
    
    @validator('classes', 'tags')
    def validate_string_lists(cls, v):
        if not v:
            return v
        
        # Validate each item in the list
        cleaned = []
        for item in v:
            if not isinstance(item, str):
                raise ValueError('List items must be strings')
            
            cleaned_item = item.strip()
            if cleaned_item and len(cleaned_item) <= 50:
                cleaned.append(cleaned_item)
        
        return cleaned if cleaned else None
    
    @validator('slug')
    def validate_slug(cls, v):
        if not v:
            return v
        if not re.match(r'^[a-zA-Z0-9\-_]+$', v):
            raise ValueError('Invalid slug format')
        return v
    
    @root_validator(skip_on_failure=True)
    def validate_at_least_one_field(cls, values):
        """Ensure at least one field is provided for update."""
        if not any(v is not None for v in values.values()):
            raise ValueError('At least one field must be provided for update')
        return values
    
    class Config:
        extra = "forbid"


class CollectionParams(LanguageParams):
    """Collection validation parameters."""
    
    collection: str = Field(..., min_length=1, max_length=50, description="Collection name")
    
    @validator('collection')
    def validate_collection_name(cls, v):
        # Validate against known collections
        from core.config import COLLECTIONS
        
        if v not in COLLECTIONS:
            raise ValueError(f'Unknown collection: {v}. Available: {", ".join(COLLECTIONS)}')
        
        return v
    
    class Config:
        extra = "forbid"


class HealthCheckResponse(BaseModel):
    """Health check response model."""
    
    status: Literal["healthy", "unhealthy"]
    timestamp: float
    database: bool = Field(description="Database connectivity status")
    details: Optional[Dict[str, Any]] = Field(default=None)
    
    class Config:
        extra = "forbid"


class ErrorResponse(BaseModel):
    """Standard error response model."""
    
    error: bool = Field(default=True)
    error_code: str
    message: str
    details: Optional[Dict[str, Any]] = Field(default=None)
    context: Optional[Dict[str, Any]] = Field(default=None)
    error_id: Optional[str] = Field(default=None)
    
    class Config:
        extra = "forbid"
