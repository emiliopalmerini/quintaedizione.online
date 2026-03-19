"""Spells parser for SRD 5.1.

Key differences from 5.2.1:
- Subtitle uses ° (U+00B0) not º (U+00BA), no class lists, optional (rituale)
- All metadata merged into one BODY_ITALIC paragraph
- Duration extraction needs sentence-boundary heuristic
"""

from __future__ import annotations

import re
import warnings

from ..classify import SpanRole
from ..heading_tree import HeadingNode
from ..markdown_gen import paragraph_to_markdown
from ..merge import Paragraph
from ..schemas import Spell
from ..section_split import SectionDef
from ..slugify import slugify
from . import register

_CANTRIP_RE = re.compile(
    r"Trucchetto\s+di\s+(\S+)",
    re.IGNORECASE,
)

_LEVELED_RE = re.compile(
    r"(.+?)\s+di\s+(\d+)[º°]\s+livello",
    re.IGNORECASE,
)

# Manual overrides for column-break artifacts
_OVERRIDES: dict[str, dict[str, str]] = {
    "creazione": {"duration": "Speciale"},
    "dominare persone": {"school": "Ammaliamento"},
}


def _parse_subtitle(text: str) -> tuple[int, str]:
    """Parse spell subtitle. Returns (level, school)."""
    text = " ".join(text.split())

    m = _CANTRIP_RE.search(text)
    if m:
        return 0, m.group(1).strip()

    m = _LEVELED_RE.search(text)
    if m:
        return int(m.group(2)), m.group(1).strip()

    return -1, ""


def _extract_field(paragraphs: list[Paragraph], label: str) -> str:
    """Extract a metadata field, handling merged paragraphs."""
    # First pass: standalone paragraphs starting with the label
    for i, para in enumerate(paragraphs):
        text = para.text.strip()
        if text.lower().startswith(label.lower()):
            rest = text[len(label):].strip()
            if rest.startswith(":"):
                rest = rest[1:].strip()
            if i + 1 < len(paragraphs) and paragraphs[i + 1].role == SpanRole.SIDEBAR:
                rest += " " + paragraphs[i + 1].text.strip()
            return rest

    # Second pass: search inside merged paragraphs
    label_lower = label.lower() + ":"
    for para in paragraphs:
        text = para.text
        idx = text.lower().find(label_lower)
        if idx >= 0:
            rest = text[idx + len(label_lower):].strip()
            next_labels = ["Tempo di lancio:", "Gittata:", "Componenti:", "Componente:", "Durata:"]
            end = len(rest)
            for nl in next_labels:
                ni = rest.find(nl)
                if ni > 0:
                    end = min(end, ni)
            value = rest[:end].strip()
            # Truncate at description start
            if end == len(rest):
                normalized = " ".join(value.split())
                for ci in range(2, len(normalized)):
                    if normalized[ci].isupper() and normalized[ci - 1] == " ":
                        prev = normalized[ci - 2]
                        if prev.islower() or prev in ".!?":
                            value = normalized[:ci].strip().rstrip(".")
                            break
            return value

    # Third pass: scan spans
    label_lower_no_colon = label.lower()
    next_labels_lower = ["tempo di lancio", "gittata", "componenti", "componente", "durata"]
    for para in paragraphs:
        for i, span in enumerate(para.spans):
            span_text = span.text.strip()
            if span_text.lower().startswith(label_lower_no_colon):
                rest = span_text[len(label):].strip()
                if rest.startswith(":"):
                    rest = rest[1:].strip()
                for j in range(i + 1, len(para.spans)):
                    next_span_text = para.spans[j].text.strip()
                    if any(next_span_text.lower().startswith(nl) for nl in next_labels_lower):
                        break
                    rest += " " + next_span_text if next_span_text else ""
                rest = rest.strip()
                for nl in next_labels_lower:
                    ni = rest.lower().find(nl)
                    if ni > 0:
                        rest = rest[:ni].strip()
                if rest:
                    return rest
    return ""


def _extract_body(content: list[Paragraph], duration: str) -> tuple[str, str]:
    """Extract description and at-higher-levels.

    In 5.1, the subtitle paragraph often merges everything together.
    We find the body by locating the duration value in the full text
    and taking everything after it.
    """
    parts: list[str] = []
    for para in content:
        md = paragraph_to_markdown(para)
        if md.strip():
            parts.append(md)
    full = "\n\n".join(parts)

    # Find body start: after the duration value
    body = ""
    if duration:
        idx = full.find(duration)
        if idx >= 0:
            body = full[idx + len(duration):].strip()

    # Fallback: everything after "Durata:" label
    if not body:
        idx = full.find("Durata:")
        if idx >= 0:
            rest = full[idx + len("Durata:"):].strip()
            # Skip the duration value (first word/phrase before uppercase sentence start)
            normalized = " ".join(rest.split())
            for ci in range(2, len(normalized)):
                if normalized[ci].isupper() and normalized[ci - 1] == " ":
                    prev = normalized[ci - 2]
                    if prev.islower() or prev in ".!?":
                        body = normalized[ci:].strip()
                        break

    # Fallback: paragraphs after the subtitle
    if not body:
        past_subtitle = False
        body_parts: list[str] = []
        for para in content:
            if not past_subtitle and para.role == SpanRole.BODY_ITALIC:
                past_subtitle = True
                continue
            if past_subtitle:
                md = paragraph_to_markdown(para)
                if md.strip():
                    body_parts.append(md)
        body = "\n\n".join(body_parts)

    # Split at-higher-levels
    marker = "Ai livelli superiori."
    idx = body.find(marker)
    if idx >= 0:
        description = body[:idx].strip()
        at_higher_levels = body[idx + len(marker):].strip()
    else:
        description = body
        at_higher_levels = ""

    return description, at_higher_levels


def _is_ritual(text: str) -> bool:
    return "rituale" in text.lower()


@register("spells")
def parse_spells(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[Spell]:
    """Parse spells from 5.1 SRD."""
    spells: list[Spell] = []

    spell_nodes: list[HeadingNode] = []

    def _find_spells(nodes: list[HeadingNode], past_marker: bool = False) -> None:
        for node in nodes:
            if node.level == 2 and "Descrizioni" in node.title:
                past_marker = True
            if past_marker and node.level == 5:
                spell_nodes.append(node)
            _find_spells(node.children, past_marker)

    _find_spells(tree)

    for node in spell_nodes:
        name = node.title.strip()
        content = node.content

        # Parse subtitle
        level = -1
        school = ""
        for para in content:
            if para.role == SpanRole.BODY_ITALIC:
                subtitle_text = " ".join(para.text.split())
                level, school = _parse_subtitle(subtitle_text)
                break

        if level < 0:
            warnings.warn(f"Spell '{name}': failed to parse subtitle")
            continue

        casting_time = _extract_field(content, "Tempo di lancio")
        range_ = _extract_field(content, "Gittata")
        components = _extract_field(content, "Componenti") or _extract_field(content, "Componente")
        duration = _extract_field(content, "Durata")

        # Apply overrides
        overrides = _OVERRIDES.get(name.lower(), {})
        school = overrides.get("school", school)
        duration = overrides.get("duration", duration)

        description, at_higher_levels = _extract_body(content, duration)

        spells.append(Spell(
            id=slugify(name),
            name=name,
            level=level,
            school=school,
            classes=[],  # 5.1 subtitles don't include class lists
            casting_time=casting_time,
            range=range_,
            components=components,
            duration=duration,
            description=description,
            at_higher_levels=at_higher_levels,
            ritual=_is_ritual(casting_time),
            source="incantesimi",
        ))

    return spells
