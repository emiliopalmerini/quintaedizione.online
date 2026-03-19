"""Feats parser for SRD 5.1.

Structure: H1("Talenti") > H3("Lottatore") — single feat, no categories.
"""

from __future__ import annotations

from ..heading_tree import HeadingNode
from ..markdown_gen import paragraph_to_markdown
from ..merge import Paragraph
from ..schemas import Feat
from ..section_split import SectionDef
from ..slugify import slugify
from . import register


@register("feats")
def parse_feats(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[Feat]:
    """Parse feats from 5.1 SRD. Feats are H3 nodes."""
    feats: list[Feat] = []

    def _collect(nodes: list[HeadingNode]) -> None:
        for node in nodes:
            if node.level in (3, 5):
                body_parts = [
                    paragraph_to_markdown(para)
                    for para in node.content
                    if paragraph_to_markdown(para).strip()
                ]
                feats.append(Feat(
                    id=slugify(node.title),
                    name=node.title.strip(),
                    category="",
                    prerequisite="",
                    repeatable=False,
                    benefit="\n\n".join(body_parts),
                ))
            else:
                _collect(node.children)

    _collect(tree)
    return feats
