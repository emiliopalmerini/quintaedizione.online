from srd_parser.utils import split_sections, SECTION_H2_RE


def test_split_sections_basic():
    md = [
        "# Title",
        "",
        "## One",
        "para1",
        "## Two",
        "para2",
        "para3",
    ]
    sections = split_sections(md, SECTION_H2_RE)
    assert len(sections) == 2
    assert sections[0][0] == "One"
    assert sections[0][1] == ["para1"]
    assert sections[1][0] == "Two"
    assert sections[1][1] == ["para2", "para3"]

