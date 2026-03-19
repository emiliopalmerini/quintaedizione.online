"""CLI helpers for profile resolution."""

from __future__ import annotations

from .profiles import FontProfile, PROFILE_521, PROFILE_51
from .section_split import SectionDef, SECTIONS, SECTIONS_51


_PROFILES: dict[str, tuple[FontProfile, list[SectionDef]]] = {
    "5.2.1": (PROFILE_521, SECTIONS),
    "5.1": (PROFILE_51, SECTIONS_51),
}


def resolve_profile(
    name: str | None,
) -> tuple[FontProfile, list[SectionDef]]:
    """Resolve a CLI profile name to a (FontProfile, sections) pair.

    Args:
        name: Profile name ("5.1", "5.2.1") or None for default (5.2.1).

    Raises:
        ValueError: If the profile name is unknown.
    """
    if name is None:
        name = "5.2.1"
    if name not in _PROFILES:
        available = ", ".join(sorted(_PROFILES))
        raise ValueError(f"Unknown profile: {name!r}. Available: {available}")
    return _PROFILES[name]
