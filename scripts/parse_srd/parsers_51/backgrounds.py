"""Backgrounds parser for SRD 5.1.

Structure: H2("Background") > H3(BackgroundName) with traits in content.
"""

from __future__ import annotations

from ..classify import SpanRole
from ..heading_tree import HeadingNode
from ..markdown_gen import paragraphs_to_markdown
from ..merge import Paragraph
from ..schemas import Background
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

_KNOWN_BACKGROUNDS = {
    "accolito", "criminale", "sapiente", "soldato",
}


def _extract_field(content: list[Paragraph], label: str) -> str:
    for para in content:
        text = para.text.strip()
        if text.lower().startswith(label.lower()):
            rest = text[len(label):].strip()
            if rest.startswith(":"):
                rest = rest[1:].strip()
            return rest
    return ""


def _extract_background(name: str, content: list[Paragraph]) -> Background:
    ability_scores = _extract_field(content, "Punteggi di caratteristica")
    feat = _extract_field(content, "Talento")
    skills = _extract_field(content, "Competenze nelle abilità")
    tool = _extract_field(content, "Competenza negli strumenti")
    equipment = _extract_field(content, "Equipaggiamento")

    body: list[Paragraph] = []
    meta_labels = {
        "punteggi di caratteristica", "talento", "competenze nelle abilità",
        "competenza negli strumenti", "equipaggiamento", "linguaggi",
    }
    for para in content:
        text = para.text.strip().lower()
        is_meta = any(text.startswith(label) for label in meta_labels)
        if not is_meta and para.role not in (SpanRole.TABLE_HEADER, SpanRole.TABLE_HEADER_SMALL):
            body.append(para)

    return Background(
        id=slugify(name),
        name=name,
        ability_scores=ability_scores,
        feat=feat,
        skill_proficiencies=skills,
        tool_proficiency=tool,
        equipment=equipment,
        description=paragraphs_to_markdown(body),
    )


@register("backgrounds")
def parse_backgrounds(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[Background]:
    """Parse backgrounds from 5.1 SRD.

    Structure: H2("Background") > H3(BackgroundName).
    """
    results: list[Background] = []

    for node in tree:
        title_lower = node.title.lower().strip()

        if title_lower == "background":
            for child in node.children:
                if child.level == 3 and child.title.lower().strip() in _KNOWN_BACKGROUNDS:
                    all_content = list(child.content)
                    bg = _extract_background(child.title.strip(), all_content)
                    # Append sub-feature content
                    child_md_parts: list[str] = []
                    for gc in child.children:
                        gc_md = paragraphs_to_markdown(gc.content)
                        if gc_md:
                            child_md_parts.append(f"##### {gc.title.strip()}\n\n{gc_md}")
                    if child_md_parts:
                        extra = "\n\n".join(child_md_parts)
                        bg["description"] = (bg["description"] + "\n\n" + extra).strip()
                    results.append(bg)
            continue

        # Descend H1 wrappers
        if node.level == 1:
            sub = parse_backgrounds(section, paragraphs, node.children)
            results.extend(sub)

    return results
