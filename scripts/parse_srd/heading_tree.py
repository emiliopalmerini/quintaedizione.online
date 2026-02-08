"""Build heading hierarchy from classified paragraphs.

Uses a stack-based approach: push on deeper heading, pop on same/higher level.
"""

from __future__ import annotations

from dataclasses import dataclass, field

from .classify import SpanRole
from .merge import Paragraph


# Map SpanRole headings to numeric levels
_HEADING_LEVELS: dict[SpanRole, int] = {
    SpanRole.H1: 1,
    SpanRole.H2: 2,
    SpanRole.H3: 3,
    SpanRole.H4: 4,
    SpanRole.H5: 5,
    SpanRole.H6: 6,
}


@dataclass
class HeadingNode:
    level: int
    title: str
    role: SpanRole
    content: list[Paragraph] = field(default_factory=list)
    children: list[HeadingNode] = field(default_factory=list)
    page_num: int = 0


def build_heading_tree(paragraphs: list[Paragraph], min_level: int = 1) -> list[HeadingNode]:
    """Build a tree of headings from a flat list of paragraphs.

    Args:
        paragraphs: Classified paragraphs in reading order.
        min_level: Minimum heading level to consider as a tree root.

    Returns:
        List of root-level HeadingNode trees.
    """
    roots: list[HeadingNode] = []
    stack: list[HeadingNode] = []

    for para in paragraphs:
        level = _HEADING_LEVELS.get(para.role, 0)

        if level > 0 and level >= min_level:
            title = para.text.strip()

            # Merge continuation: if previous node is at the same level and
            # the new title starts with lowercase, it's a wrapped heading line
            if stack and stack[-1].level == level and title and title[0].islower():
                stack[-1].title += " " + title
                continue

            node = HeadingNode(
                level=level,
                title=title,
                role=para.role,
                page_num=para.page_num,
            )

            # Pop stack until we find a parent with a lower level
            while stack and stack[-1].level >= level:
                stack.pop()

            if stack:
                stack[-1].children.append(node)
            else:
                roots.append(node)

            stack.append(node)

        elif stack:
            # Non-heading paragraph — attach to current deepest heading
            stack[-1].content.append(para)
        # else: content before any heading — discard or attach to a virtual root

    return roots


def walk_tree(nodes: list[HeadingNode], depth: int = 0):
    """Debug: print heading tree structure."""
    for node in nodes:
        indent = "  " * depth
        content_count = len(node.content)
        children_count = len(node.children)
        print(f"{indent}[H{node.level}] {node.title} ({content_count} paragraphs, {children_count} children)")
        walk_tree(node.children, depth + 1)
