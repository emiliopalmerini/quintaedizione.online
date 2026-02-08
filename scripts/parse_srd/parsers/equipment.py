"""Equipment parser — weapons, armor, tools, gear, mounts, services.

These sections are table-heavy. We generate structured entries with
the table data extracted from the heading tree.
"""

from __future__ import annotations

from ..classify import SpanRole
from ..heading_tree import HeadingNode
from ..markdown_gen import paragraphs_to_markdown
from ..merge import Paragraph
from ..schemas import EquipmentItem
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

# Map section names to equipment categories
_SECTION_CATEGORY = {
    "Armi": "weapons",
    "Armature": "armor",
    "Strumenti": "tools",
    "Equipaggiamenti": "gear",
    "Cavalcature e Veicoli": "mounts",
    "Servizi": "services",
}


@register("equipment")
def parse_equipment(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[EquipmentItem]:
    """Parse equipment sections into structured items.

    Each section is converted to a single entry with its full content
    as markdown, since tables need to be preserved as-is for the
    Go pipeline to process.
    """
    category = _SECTION_CATEGORY.get(section.name, "gear")
    items: list[EquipmentItem] = []

    def _collect(nodes: list[HeadingNode], parent_sub: str = "") -> None:
        for node in nodes:
            # H1/H2 are section headers — descend
            if node.level <= 2:
                _collect(node.children, node.title.strip())
                continue

            # H3-H5 = individual equipment entries or sub-sections
            name = node.title.strip()
            content_md = paragraphs_to_markdown(node.content)

            # Also include children content
            for child in node.children:
                child_md = paragraphs_to_markdown(child.content)
                if child_md:
                    content_md += f"\n\n### {child.title}\n\n{child_md}"

            items.append(EquipmentItem(
                id=slugify(name),
                name=name,
                category=category,
                subcategory=parent_sub,
                properties={},
                description=content_md,
            ))

    _collect(tree)

    # If no individual items found, create one entry for the whole section
    if not items:
        description = paragraphs_to_markdown(paragraphs)
        items.append(EquipmentItem(
            id=slugify(section.name),
            name=section.name,
            category=category,
            subcategory="",
            properties={},
            description=description,
        ))

    return items
