# /// script
# requires-python = ">=3.12"
# dependencies = ["pytest"]
# ///
"""Tests for section definitions (5.2.1 and 5.1)."""

from __future__ import annotations

import pytest

from ..section_split import SECTIONS, SECTIONS_51, SectionDef, get_sections_for_parser


class TestSectionDef:
    """SectionDef basic properties."""

    def test_sections_521_exist(self):
        assert len(SECTIONS) > 0

    def test_sections_51_exist(self):
        assert len(SECTIONS_51) > 0


class TestSections51:
    """SRD 5.1 section definitions."""

    def test_has_razze_section(self):
        """5.1 has Razze (races), not Specie."""
        names = [s.name for s in SECTIONS_51]
        assert "Razze" in names
        assert "Specie" not in names

    def test_has_spells(self):
        spells = [s for s in SECTIONS_51 if s.output_file == "spells.json"]
        assert len(spells) == 1
        assert spells[0].parser_name == "spells"

    def test_has_monsters(self):
        monsters = [s for s in SECTIONS_51 if s.output_file == "monsters.json"]
        assert len(monsters) >= 1

    def test_has_equipment(self):
        equip = [s for s in SECTIONS_51 if s.output_file == "equipment.json"]
        assert len(equip) >= 1

    def test_has_magic_items(self):
        items = [s for s in SECTIONS_51 if s.output_file == "magic_items.json"]
        assert len(items) >= 1

    def test_has_classes(self):
        classes = [s for s in SECTIONS_51 if s.output_file == "classes.json"]
        assert len(classes) >= 1
        # 5.1 classes span pages 8-59
        assert classes[0].pages[0] == 8

    def test_has_backgrounds(self):
        bgs = [s for s in SECTIONS_51 if s.output_file == "backgrounds.json"]
        assert len(bgs) == 1

    def test_no_overlapping_pages(self):
        """Section page ranges must not overlap."""
        sorted_sections = sorted(SECTIONS_51, key=lambda s: s.pages[0])
        for i in range(len(sorted_sections) - 1):
            curr = sorted_sections[i]
            nxt = sorted_sections[i + 1]
            assert curr.pages[1] < nxt.pages[0], (
                f"{curr.name} ({curr.pages}) overlaps with {nxt.name} ({nxt.pages})"
            )

    def test_all_sections_have_parser_name(self):
        for s in SECTIONS_51:
            assert s.parser_name, f"Section {s.name} has no parser_name"

    def test_pages_are_valid_ranges(self):
        for s in SECTIONS_51:
            assert s.pages[0] <= s.pages[1], f"Section {s.name} has invalid page range"
            assert s.pages[0] >= 1, f"Section {s.name} starts before page 1"


class TestGetSectionsForParser51:
    """get_sections_for_parser works with both SECTIONS lists."""

    def test_get_monsters_51(self):
        monsters = get_sections_for_parser("monsters", SECTIONS_51)
        assert len(monsters) >= 1

    def test_get_unknown_parser(self):
        result = get_sections_for_parser("nonexistent", SECTIONS_51)
        assert result == []
