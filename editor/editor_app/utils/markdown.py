from __future__ import annotations

from typing import Optional

try:
    import markdown as _markdown
except Exception:  # pragma: no cover
    _markdown = None  # type: ignore


def render_md(text: Optional[str]) -> str:
    """Render markdown to HTML using Python-Markdown with common extensions.

    Falls back to basic HTML-escaped text with <br> separation if the
    dependency is unavailable at runtime.
    """
    src = (text or "").strip()
    if not src:
        return ""
    if _markdown is None:
        # Minimal fallback: escape and preserve paragraphs
        import html

        esc = html.escape(src)
        return "<p>" + esc.replace("\n\n", "</p><p>") + "</p>"
    try:
        return _markdown.markdown(
            src,
            extensions=[
                "extra",  # includes: abbr, attr_list, def_list, fenced_code, footnotes, tables
                "sane_lists",
                "toc",
                "admonition",
                "smarty",
            ],
            output_format="xhtml",
        )
    except Exception:
        # Be resilient
        return src

