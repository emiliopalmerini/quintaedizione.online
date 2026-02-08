"""Rules and Glossary parsers.

Rules sections: "Come si gioca" (5-20), "Creazione del personaggio" (21-31),
"Strumenti di gioco" (220-231).

Glossary: "Glossario delle regole" (202-219) — entries as H5 with optional
[category] descriptor.
"""

from __future__ import annotations

import re

from ..classify import SpanRole
from ..heading_tree import HeadingNode
from ..markdown_gen import paragraphs_to_markdown
from ..merge import Paragraph
from ..schemas import GlossaryEntry, RuleEntry
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

# Pattern: "Term [category]"
_GLOSSARY_BRACKET_RE = re.compile(r"^(.+?)\s*\[(.+?)\]\s*$")

# "Vedi anche" cross-reference pattern
_SEE_ALSO_RE = re.compile(r"[Vv]edi anche[:\s]+(.+?)\.?\s*$")


def _heading_to_rule(node: HeadingNode) -> RuleEntry:
    """Recursively convert a heading node to a RuleEntry."""
    content_md = paragraphs_to_markdown(node.content)
    children = [_heading_to_rule(child) for child in node.children]

    return RuleEntry(
        id=slugify(node.title),
        title=node.title.strip(),
        content=content_md,
        children=children,
    )


@register("rules")
def parse_rules(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[RuleEntry]:
    """Parse a rules section into nested JSON.

    Each heading becomes a RuleEntry with its content and children.
    """
    rules: list[RuleEntry] = []

    for node in tree:
        rules.append(_heading_to_rule(node))

    return rules


@register("glossary")
def parse_glossary(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[GlossaryEntry]:
    """Parse the glossary section.

    Entries are H5 headings under "Definizione delle regole" (H2).
    """
    entries: list[GlossaryEntry] = []

    def _collect(nodes: list[HeadingNode]) -> None:
        for node in nodes:
            if node.level == 5:
                title = node.title.strip()

                # Extract optional [category] from title
                category = ""
                term = title
                m = _GLOSSARY_BRACKET_RE.match(title)
                if m:
                    term = m.group(1).strip()
                    category = m.group(2).strip()

                # Build definition from content
                definition = paragraphs_to_markdown(node.content)

                # Extract "Vedi anche" cross-references
                see_also: list[str] = []
                for para in node.content:
                    text = para.text.strip()
                    m_see = _SEE_ALSO_RE.search(text)
                    if m_see:
                        refs = m_see.group(1)
                        see_also.extend(
                            r.strip().strip(".")
                            for r in re.split(r"[,;]|(?:\se\s)", refs)
                            if r.strip()
                        )

                entries.append(GlossaryEntry(
                    id=slugify(term),
                    term=term,
                    category=category,
                    definition=definition,
                    see_also=see_also,
                ))
            else:
                _collect(node.children)

    _collect(tree)
    return entries
