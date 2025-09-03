# Shared domain for both Editor and Parser services
# Complete D&D 5e SRD domain model following ADR data model specification

from .complete_entities import *
from .entities import *
from .query_models import *

# Make SRDDomainModel easily accessible
from .complete_entities import SRDDomainModel

__all__ = [
    "SRDDomainModel",  # Main facade for domain model
    # All entities and components are exported via complete_entities
]