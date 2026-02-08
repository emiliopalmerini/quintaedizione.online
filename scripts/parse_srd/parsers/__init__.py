"""Parser registry — maps parser names to handler functions."""

from __future__ import annotations

from typing import Callable

from ..heading_tree import HeadingNode
from ..merge import Paragraph
from ..section_split import SectionDef

ParserFunc = Callable[
    [SectionDef, list[Paragraph], list[HeadingNode]],
    list | dict,
]

_REGISTRY: dict[str, ParserFunc] = {}


def register(name: str):
    """Decorator to register a parser function."""
    def decorator(func: ParserFunc) -> ParserFunc:
        _REGISTRY[name] = func
        return func
    return decorator


def get_parser(name: str) -> ParserFunc | None:
    """Get a parser by name."""
    return _REGISTRY.get(name)


# Import all parser modules to trigger registration
from . import (  # noqa: E402, F401
    backgrounds,
    classes,
    equipment,
    feats,
    magic_items,
    monsters,
    rules,
    species,
    spells,
)
