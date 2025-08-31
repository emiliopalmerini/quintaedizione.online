"""Domain models for D&D 5e SRD content."""
from __future__ import annotations

from typing import Optional, Any, Dict, List, Union
from pydantic import BaseModel, Field, validator
from datetime import datetime


class DocumentBase(BaseModel):
    """Base document model with common fields."""
    
    title: str = Field(..., min_length=1, max_length=500)
    content: str = Field(..., min_length=1)
    slug: str = Field(..., min_length=1, max_length=200)
    numero_di_pagina: Optional[int] = Field(default=None, ge=1)
    _sortkey_alpha: Optional[str] = Field(default=None)
    
    # Metadata
    created_at: Optional[datetime] = Field(default=None)
    updated_at: Optional[datetime] = Field(default=None)
    modified: Optional[bool] = Field(default=False)
    translated: Optional[bool] = Field(default=False)
    
    @validator('slug')
    def validate_slug(cls, v):
        import re
        if not re.match(r'^[a-zA-Z0-9\-_]+$', v):
            raise ValueError('Invalid slug format')
        return v
    
    class Config:
        extra = "allow"  # Allow additional fields for flexibility
        json_encoders = {
            datetime: lambda v: v.isoformat() if v else None
        }


class SpellDocument(DocumentBase):
    """Spell document model."""
    
    level: int = Field(..., ge=0, le=9, description="Spell level")
    school: str = Field(..., max_length=50, description="School of magic")
    ritual: bool = Field(default=False, description="Is ritual spell")
    concentration: bool = Field(default=False, description="Requires concentration")
    
    casting_time: str = Field(..., max_length=100, description="Casting time")
    range: str = Field(..., max_length=100, description="Spell range")
    components: str = Field(..., max_length=200, description="Spell components")
    duration: str = Field(..., max_length=100, description="Spell duration")
    
    classes: List[str] = Field(default_factory=list, description="Available classes")
    
    @validator('classes')
    def validate_classes(cls, v):
        # Clean and validate class names
        if not v:
            return []
        
        valid_classes = [
            'Bard', 'Chierico', 'Druido', 'Mago', 'Paladino', 'Ranger', 
            'Stregone', 'Warlock', 'Wizard', 'Artificer'
        ]
        
        cleaned = []
        for class_name in v:
            if isinstance(class_name, str) and class_name.strip():
                cleaned.append(class_name.strip())
        
        return cleaned
    
    class Config:
        extra = "allow"


class ItemDocument(DocumentBase):
    """Magic item document model."""
    
    type: str = Field(..., max_length=50, description="Item type")
    rarity: str = Field(..., max_length=50, description="Item rarity")
    attunement: bool = Field(default=False, description="Requires attunement")
    
    @validator('rarity')
    def validate_rarity(cls, v):
        valid_rarities = ['Common', 'Uncommon', 'Rare', 'Very Rare', 'Legendary', 'Artifact']
        if v and v not in valid_rarities:
            # Allow Italian translations
            italian_rarities = ['Comune', 'Non comune', 'Raro', 'Molto raro', 'Leggendario', 'Artefatto']
            if v not in italian_rarities:
                raise ValueError(f'Invalid rarity: {v}')
        return v
    
    class Config:
        extra = "allow"


class MonsterDocument(DocumentBase):
    """Monster document model."""
    
    size: Optional[str] = Field(default=None, max_length=20)
    type: Optional[str] = Field(default=None, max_length=50)
    alignment: Optional[str] = Field(default=None, max_length=50)
    armor_class: Optional[int] = Field(default=None, ge=1, le=30)
    hit_points: Optional[int] = Field(default=None, ge=1)
    speed: Optional[str] = Field(default=None, max_length=100)
    challenge_rating: Optional[str] = Field(default=None, max_length=10)
    
    # Ability scores
    strength: Optional[int] = Field(default=None, ge=1, le=30)
    dexterity: Optional[int] = Field(default=None, ge=1, le=30)
    constitution: Optional[int] = Field(default=None, ge=1, le=30)
    intelligence: Optional[int] = Field(default=None, ge=1, le=30)
    wisdom: Optional[int] = Field(default=None, ge=1, le=30)
    charisma: Optional[int] = Field(default=None, ge=1, le=30)
    
    class Config:
        extra = "allow"


class GenericDocument(DocumentBase):
    """Generic document for other content types."""
    
    category: Optional[str] = Field(default=None, max_length=50)
    subcategory: Optional[str] = Field(default=None, max_length=50)
    tags: List[str] = Field(default_factory=list)
    
    class Config:
        extra = "allow"


# Union type for all document types
DocumentModel = Union[SpellDocument, ItemDocument, MonsterDocument, GenericDocument]


class CollectionInfo(BaseModel):
    """Collection metadata model."""
    
    name: str = Field(..., description="Collection name")
    label: str = Field(..., description="Display label")
    count: int = Field(..., ge=0, description="Document count")
    type: str = Field(..., description="Collection type")
    
    class Config:
        extra = "forbid"


class PaginatedResponse(BaseModel):
    """Paginated response model."""
    
    items: List[Dict[str, Any]] = Field(..., description="Response items")
    total: int = Field(..., ge=0, description="Total item count")
    page: int = Field(..., ge=1, description="Current page")
    pages: int = Field(..., ge=1, description="Total pages")
    page_size: int = Field(..., ge=1, description="Items per page")
    
    # Navigation
    has_prev: bool = Field(..., description="Has previous page")
    has_next: bool = Field(..., description="Has next page")
    prev_page: Optional[int] = Field(default=None, description="Previous page number")
    next_page: Optional[int] = Field(default=None, description="Next page number")
    
    class Config:
        extra = "forbid"


class SearchResult(BaseModel):
    """Search result model."""
    
    document: Dict[str, Any] = Field(..., description="Document data")
    score: Optional[float] = Field(default=None, description="Search relevance score")
    highlights: Optional[Dict[str, List[str]]] = Field(default=None, description="Search highlights")
    
    class Config:
        extra = "forbid"


class NavigationContext(BaseModel):
    """Navigation context for document browsing."""
    
    current_doc: Dict[str, Any] = Field(..., description="Current document")
    prev_doc: Optional[Dict[str, Any]] = Field(default=None, description="Previous document")
    next_doc: Optional[Dict[str, Any]] = Field(default=None, description="Next document")
    
    collection: str = Field(..., description="Collection name")
    total_count: int = Field(..., ge=0, description="Total documents in collection")
    current_position: int = Field(..., ge=1, description="Current position in collection")
    
    class Config:
        extra = "forbid"