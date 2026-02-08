"""Feats (Talenti) parser.

Structure in PDF:
  H1: Talenti
  H2: Descrizioni dei talenti (intro)
  H4: Talenti Origini / Talenti Generali / Talenti Stile di combattimento / Talenti Dono epico
    H5: Individual feat name
      First content paragraph: starts with BODY_ITALIC "Talento {Category} (prerequisito: ...)"
        followed by BODY spans with the feat description.
"""

from __future__ import annotations

import re

from ..classify import SpanRole
from ..heading_tree import HeadingNode
from ..markdown_gen import paragraph_to_markdown
from ..merge import ClassifiedSpan, Paragraph
from ..schemas import Feat
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

# Matches: "Talento {Category}" or "Talento {Category} (prerequisito: ...)"
_SUBTITLE_RE = re.compile(
    r"^Talento\s+(.+?)(?:\s*\(prerequisito:\s*(.+?)\))?\s*$",
    re.IGNORECASE,
)

# Map H4 category headings to canonical names
_CATEGORY_MAP = {
    "talenti origini": "Origini",
    "talenti generali": "Generale",
    "talenti stile di combattimento": "Stile di combattimento",
    "talenti dono epico": "Dono epico",
}


def _extract_feat(name: str, content: list[Paragraph], category_from_parent: str) -> Feat:
    """Extract a Feat from its heading name and content paragraphs.

    The first content paragraph typically starts with an italic "Talento {Category}"
    span, followed by body text spans in the same paragraph.
    """
    category = category_from_parent
    prerequisite = ""
    repeatable = False
    body_parts: list[str] = []

    for para_idx, para in enumerate(content):
        # On the first paragraph, check for category subtitle in spans
        if para_idx == 0 and para.spans:
            first_span = para.spans[0]
            if first_span.role == SpanRole.BODY_ITALIC and first_span.text.strip().startswith("Talento"):
                # Collect all leading BODY_ITALIC spans to form the full subtitle
                subtitle_parts: list[str] = []
                body_start_idx = 0
                for si, span in enumerate(para.spans):
                    if span.role == SpanRole.BODY_ITALIC:
                        subtitle_parts.append(span.text)
                    elif span.text.strip() == "":
                        continue  # skip whitespace between italic spans
                    else:
                        body_start_idx = si
                        break
                else:
                    body_start_idx = len(para.spans)

                subtitle_text = " ".join(subtitle_parts).strip()
                # Normalize internal whitespace
                subtitle_text = " ".join(subtitle_text.split())
                m = _SUBTITLE_RE.match(subtitle_text)
                if m:
                    category = m.group(1).strip()
                    if m.group(2):
                        prerequisite = m.group(2).strip()

                # Build markdown from remaining spans (skip subtitle spans)
                remaining_spans = para.spans[body_start_idx:]
                while remaining_spans and remaining_spans[0].text.strip() == "":
                    remaining_spans = remaining_spans[1:]
                if remaining_spans:
                    remaining_para = Paragraph(
                        spans=remaining_spans,
                        role=SpanRole.BODY,
                        page_num=para.page_num,
                    )
                    md = paragraph_to_markdown(remaining_para)
                    if md.strip():
                        body_parts.append(md)
                continue

        # Regular paragraphs
        md = paragraph_to_markdown(para)
        if md.strip():
            body_parts.append(md)

    full_body = "\n\n".join(body_parts)
    if "Ripetibile." in full_body:
        repeatable = True

    return Feat(
        id=slugify(name),
        name=name,
        category=category,
        prerequisite=prerequisite,
        repeatable=repeatable,
        benefit=full_body,
    )


def _collect_feats_from_tree(nodes: list[HeadingNode], parent_category: str = "") -> list[Feat]:
    """Recursively collect feats from heading tree."""
    feats: list[Feat] = []

    for node in nodes:
        # H4 = category heading
        if node.level == 4:
            cat = _CATEGORY_MAP.get(node.title.lower().strip(), parent_category)
            feats.extend(_collect_feats_from_tree(node.children, cat))
            continue

        # H5 = individual feat
        if node.level == 5:
            feat = _extract_feat(node.title, node.content, parent_category)
            feats.append(feat)
            continue

        # H1, H2 — descend
        feats.extend(_collect_feats_from_tree(node.children, parent_category))

    return feats


@register("feats")
def parse_feats(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[Feat]:
    """Parse feats from the Talenti section (pages 98-100)."""
    return _collect_feats_from_tree(tree)
