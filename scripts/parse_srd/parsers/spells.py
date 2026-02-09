"""Spells (Incantesimi) parser.

Structure in PDF:
  H2: Descrizioni degli incantesimi (marker — spells start after this)
  H5: Spell name
    BODY_ITALIC: "{Scuola} di {N}º livello ({classi})" or "Trucchetto di {Scuola} ({classi})"
    TABLE_HEADER_SMALL: "Tempo di lancio: ..."
    TABLE_HEADER_SMALL: "Gittata: ..."
    TABLE_HEADER_SMALL: "Componenti: ..."
    TABLE_HEADER_SMALL: "Durata: ..."
    BODY paragraphs: description
      May contain BODY_BOLD_ITALIC: "Utilizzo di uno slot incantesimo di livello superiore."
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

# Cantrip: "Trucchetto di {Scuola} ({classi})"
_CANTRIP_RE = re.compile(
    r"Trucchetto\s+di\s+(.+?)\s*\((.+?)\)",
    re.IGNORECASE,
)

# Leveled spell: "{Scuola} di {N}º livello ({classi})"
_LEVELED_RE = re.compile(
    r"(.+?)\s+di\s+(\d+)º\s+livello\s*\((.+?)\)",
    re.IGNORECASE,
)

# Higher levels marker
_HIGHER_LEVELS_MARKER = "Utilizzo di uno slot incantesimo di livello superiore."


def _extract_field(paragraphs: list[Paragraph], label: str) -> str:
    """Extract a metadata field from TABLE_HEADER_SMALL paragraphs.

    Handles continuation on the next SIDEBAR paragraph (e.g. long component text
    that wraps to a second line with a different font role).
    Also handles the edge case where all metadata is merged into the subtitle.
    """
    for i, para in enumerate(paragraphs):
        text = para.text.strip()
        if text.lower().startswith(label.lower()):
            rest = text[len(label):].strip()
            if rest.startswith(":"):
                rest = rest[1:].strip()
            # Check for continuation in next SIDEBAR paragraph
            if i + 1 < len(paragraphs) and paragraphs[i + 1].role == SpanRole.SIDEBAR:
                rest += " " + paragraphs[i + 1].text.strip()
            return rest

    # Fallback: search inside merged subtitle paragraphs (e.g. entire spell in one line)
    label_lower = label.lower() + ":"
    for para in paragraphs:
        text = para.text
        idx = text.lower().find(label_lower)
        if idx >= 0:
            rest = text[idx + len(label_lower):].strip()
            # Value ends at the next known label
            next_labels = ["Tempo di lancio:", "Gittata:", "Componenti:", "Componente:", "Durata:"]
            end = len(rest)
            for nl in next_labels:
                ni = rest.find(nl)
                if ni > 0:
                    end = min(end, ni)
            value = rest[:end].strip()
            # For last field (no next label found), truncate at sentence boundary
            # Duration values are short (e.g. "istantanea", "1 ora")
            if end == len(rest) and len(value) > 50:
                # Find first uppercase letter after a space following a lowercase char
                for ci in range(2, len(value)):
                    if value[ci].isupper() and value[ci - 1] == " " and value[ci - 2].islower():
                        value = value[:ci].strip()
                        break
            return value

    # Third pass: scan individual spans for the label text regardless of role.
    # Some spells have metadata in BODY_ITALIC or other non-TABLE_HEADER spans.
    label_lower = label.lower()
    next_labels_lower = ["tempo di lancio", "gittata", "componenti", "componente", "durata"]
    for para in paragraphs:
        for i, span in enumerate(para.spans):
            span_text = span.text.strip()
            if span_text.lower().startswith(label_lower):
                rest = span_text[len(label):].strip()
                if rest.startswith(":"):
                    rest = rest[1:].strip()
                # Collect text from subsequent spans until a next label
                for j in range(i + 1, len(para.spans)):
                    next_span_text = para.spans[j].text.strip()
                    if any(next_span_text.lower().startswith(nl) for nl in next_labels_lower):
                        break
                    rest += " " + next_span_text if next_span_text else ""
                rest = rest.strip()
                # Truncate at next label if found within rest
                for nl in next_labels_lower:
                    ni = rest.lower().find(nl)
                    if ni > 0:
                        rest = rest[:ni].strip()
                if rest:
                    return rest
    return ""


def _parse_subtitle(text: str) -> tuple[int, str, list[str]]:
    """Parse spell subtitle into (level, school, classes).

    Returns (0, school, classes) for cantrips.
    """
    text = " ".join(text.split())  # normalize whitespace

    m = _CANTRIP_RE.search(text)
    if m:
        school = m.group(1).strip()
        classes = [c.strip() for c in m.group(2).split(",")]
        return 0, school, classes

    m = _LEVELED_RE.search(text)
    if m:
        school = m.group(1).strip()
        level = int(m.group(2))
        classes = [c.strip() for c in m.group(3).split(",")]
        return level, school, classes

    return -1, "", []


def _extract_spell_body(content: list[Paragraph]) -> tuple[str, str]:
    """Extract description and at-higher-levels from body paragraphs.

    Returns (description_md, at_higher_levels_md).
    """
    # The body paragraphs include the subtitle and metadata fields at the start.
    # We need to skip those and find the actual description body.
    body_paras: list[Paragraph] = []
    meta_labels = {"tempo di lancio", "gittata", "componenti", "durata"}

    subtitle_skipped = False
    in_meta = True
    for para in content:
        text = para.text.strip()
        # Skip the first subtitle (BODY_ITALIC with spell school/level pattern)
        if not subtitle_skipped and para.role == SpanRole.BODY_ITALIC and (
            "Trucchetto" in text or _LEVELED_RE.search(text)
        ):
            subtitle_skipped = True
            # Edge case: entire spell merged into one paragraph (subtitle + meta + body)
            if "Durata:" in text:
                dur_idx = text.index("Durata:")
                after_dur = text[dur_idx + len("Durata:"):]
                # Find end of duration value (next sentence start)
                parts = after_dur.split(None, 1)
                if len(parts) > 1:
                    # Duration value + body text after it
                    # Split at first uppercase letter after duration
                    dur_and_body = after_dur.strip()
                    for ci, ch in enumerate(dur_and_body):
                        if ci > 0 and ch.isupper() and dur_and_body[ci - 1] == " ":
                            body_text = dur_and_body[ci:].strip()
                            if body_text:
                                from ..merge import ClassifiedSpan
                                body_paras.append(Paragraph(
                                    spans=[ClassifiedSpan(text=body_text, role=SpanRole.BODY)],
                                    role=SpanRole.BODY,
                                    page_num=para.page_num,
                                ))
                            break
            continue
        # Skip metadata fields
        if para.role in (SpanRole.TABLE_HEADER_SMALL, SpanRole.TABLE_HEADER):
            in_meta = True
            continue
        # Skip SIDEBAR continuations of metadata (e.g. long component text)
        if in_meta and para.role == SpanRole.SIDEBAR:
            continue
        in_meta = False
        body_paras.append(para)

    # Now split body into description and at-higher-levels
    # The higher-levels marker is a BODY_BOLD_ITALIC span within a body paragraph
    desc_parts: list[str] = []
    higher_parts: list[str] = []
    in_higher = False

    for para in body_paras:
        # Check if this paragraph contains the higher-levels marker.
        # The marker text "Utilizzo di uno slot incantesimo di livello superiore."
        # may be split across multiple consecutive BODY_BOLD_ITALIC spans.
        marker_start = -1
        marker_end = -1
        bi_text = ""
        bi_start = -1
        for i, span in enumerate(para.spans):
            if span.role == SpanRole.BODY_BOLD_ITALIC:
                if bi_start < 0:
                    bi_start = i
                bi_text += span.text
            elif span.text.strip() == "" and bi_start >= 0:
                bi_text += span.text  # whitespace between bold-italic spans
            else:
                if bi_start >= 0 and "livello supe" in bi_text:
                    marker_start = bi_start
                    marker_end = i
                    break
                bi_text = ""
                bi_start = -1
        # Check at end of spans too
        if marker_start < 0 and bi_start >= 0 and "livello supe" in bi_text:
            marker_start = bi_start
            marker_end = len(para.spans)

        if marker_start >= 0 and not in_higher:
            # Split: before marker = description, after marker = higher levels
            before_spans = para.spans[:marker_start]
            after_spans = para.spans[marker_end:]

            while before_spans and before_spans[-1].text.strip() == "":
                before_spans = before_spans[:-1]
            if before_spans:
                desc_para = Paragraph(
                    spans=before_spans, role=SpanRole.BODY, page_num=para.page_num,
                )
                md = paragraph_to_markdown(desc_para)
                if md.strip():
                    desc_parts.append(md)

            # Skip leading whitespace in after spans
            while after_spans and after_spans[0].text.strip() == "":
                after_spans = after_spans[1:]
            if after_spans:
                higher_para = Paragraph(
                    spans=after_spans, role=SpanRole.BODY, page_num=para.page_num,
                )
                md = paragraph_to_markdown(higher_para)
                if md.strip():
                    higher_parts.append(md)
            in_higher = True
            continue

        md = paragraph_to_markdown(para)
        if md.strip():
            if in_higher:
                higher_parts.append(md)
            else:
                desc_parts.append(md)

    # Post-processing fallback: if the marker text ended up in description
    # (e.g. because bold-italic spans were split by a non-bold-italic span),
    # search the rendered description for the marker and split there.
    if not higher_parts and desc_parts:
        full_desc = "\n\n".join(desc_parts)
        # Search for the marker in various rendered forms
        marker_variants = [
            "_**Utilizzo di uno slot incantesimo di livello superiore.**_",
            "***Utilizzo di uno slot incantesimo di livello superiore.***",
            "Utilizzo di uno slot incantesimo di livello superiore.",
        ]
        for marker in marker_variants:
            idx = full_desc.find(marker)
            if idx >= 0:
                before = full_desc[:idx].strip()
                after = full_desc[idx + len(marker):].strip()
                if after:
                    desc_parts = [before] if before else []
                    higher_parts = [after]
                break

    return "\n\n".join(desc_parts), "\n\n".join(higher_parts)


def _is_ritual(casting_time: str) -> bool:
    """Check if the spell is a ritual from casting time field."""
    return "rituale" in casting_time.lower()


@register("spells")
def parse_spells(
    section: SectionDef,
    paragraphs: list[Paragraph],
    tree: list[HeadingNode],
) -> list[Spell]:
    """Parse spells from the Incantesimi section (pages 118-201).

    Skips all content before the "Descrizioni degli incantesimi" H2 heading.
    """
    spells: list[Spell] = []

    # Find the "Descrizioni degli incantesimi" marker in the tree
    spell_nodes: list[HeadingNode] = []
    for node in tree:
        if node.level == 2 and "Descrizioni degli incantesimi" in node.title:
            spell_nodes = node.children
            break
        # Also check children of H1
        for child in node.children:
            if child.level == 2 and "Descrizioni degli incantesimi" in child.title:
                spell_nodes = child.children
                break
        if spell_nodes:
            break

    # If not found in tree, fall back to collecting all H5 nodes after the marker
    if not spell_nodes:
        past_marker = False
        for node in tree:
            if node.level == 2 and "Descrizioni" in node.title:
                past_marker = True
                continue
            if past_marker and node.level == 5:
                spell_nodes.append(node)
            for child in node.children:
                if child.level == 2 and "Descrizioni" in child.title:
                    past_marker = True
                if past_marker and child.level == 5:
                    spell_nodes.append(child)

    for node in spell_nodes:
        if node.level != 5:
            # Collect H5 children from deeper levels
            for child in node.children:
                if child.level == 5:
                    spell_nodes.append(child)
            continue

        name = node.title.strip()
        content = node.content

        # Parse subtitle (first BODY_ITALIC paragraph)
        level = -1
        school = ""
        classes: list[str] = []
        for para in content:
            if para.role == SpanRole.BODY_ITALIC:
                subtitle_text = " ".join(para.text.split())
                level, school, classes = _parse_subtitle(subtitle_text)
                break

        if level < 0:
            warnings.warn(f"Spell '{name}': failed to parse subtitle (level=-1)")
            continue  # Not a valid spell entry

        # Extract metadata
        casting_time = _extract_field(content, "Tempo di lancio")
        range_ = _extract_field(content, "Gittata")
        components = _extract_field(content, "Componenti") or _extract_field(content, "Componente")
        duration = _extract_field(content, "Durata")

        # Warn on empty required fields
        for _label, _val in [
            ("casting_time", casting_time), ("range", range_),
            ("components", components), ("duration", duration),
        ]:
            if not _val:
                warnings.warn(f"Spell '{name}': empty {_label}")

        # Extract body
        description, at_higher_levels = _extract_spell_body(content)

        spells.append(Spell(
            id=slugify(name),
            name=name,
            level=level,
            school=school,
            classes=classes,
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
