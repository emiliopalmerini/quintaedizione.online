"""Font profiles for different SRD PDF editions.

Each profile captures the font families, sizes, and colors used in a specific
PDF so the classifier can map raw spans to semantic roles.
"""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True, slots=True)
class FontProfile:
    """Font/color calibration for a specific SRD PDF."""

    name: str

    # Heading font family and color
    heading_font: str
    heading_color: int

    # Body text
    body_font: str
    body_color: int

    # Sidebar / table font (same family in 5.1, different in 5.2.1)
    sidebar_font: str

    # Stat block — distinct font/color in 5.2.1, None in 5.1
    stat_font: str | None
    stat_color: int | None

    # Link color
    link_color: int

    # Footer color (None means footer uses body_color)
    footer_color: int | None


PROFILE_521 = FontProfile(
    name="SRD 5.2.1",
    heading_font="GillSans",
    heading_color=0x8C2220,
    body_font="Cambria",
    body_color=0x231F20,
    sidebar_font="GillSans",
    stat_font="Optima",
    stat_color=0x540000,
    link_color=0x1E5E9E,
    footer_color=0x808285,
)

PROFILE_51 = FontProfile(
    name="SRD 5.1",
    heading_font="Calibri",
    heading_color=0x943634,
    body_font="Cambria",
    body_color=0x000000,
    sidebar_font="Calibri",
    stat_font=None,
    stat_color=None,
    link_color=0x0000FF,
    footer_color=None,
)
