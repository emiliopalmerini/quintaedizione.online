from srd_parser.parsers.documents import _slug_from_filename, _page_from_filename, parse_document


def test_slug_and_page_from_filename():
    assert _slug_from_filename("01_legal_information.md") == "legal-information"
    assert _slug_from_filename("foo_bar.md") == "foo-bar"
    assert _page_from_filename("21_animals.md") == 21
    assert _page_from_filename("animals.md") is None


def test_parse_document_title_fallback():
    lines = ["Some content without H1", "Another paragraph"]
    out = parse_document(lines, "03_character_creation.md")
    assert len(out) == 1
    doc = out[0]
    assert doc["slug"] == "character-creation"
    # falls back to humanized title from filename
    assert doc["titolo"] == "Character Creation"

