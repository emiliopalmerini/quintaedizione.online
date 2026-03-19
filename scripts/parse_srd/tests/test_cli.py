# /// script
# requires-python = ">=3.12"
# dependencies = ["pytest"]
# ///
"""Tests for CLI --profile flag and profile/section resolution."""

from __future__ import annotations

import pytest

from ..profiles import PROFILE_521, PROFILE_51
from ..section_split import SECTIONS, SECTIONS_51
from .._cli import resolve_profile


class TestResolveProfile:
    """resolve_profile maps CLI string to (FontProfile, sections) pair."""

    def test_default_is_521(self):
        profile, sections = resolve_profile(None)
        assert profile is PROFILE_521
        assert sections is SECTIONS

    def test_explicit_521(self):
        profile, sections = resolve_profile("5.2.1")
        assert profile is PROFILE_521
        assert sections is SECTIONS

    def test_explicit_51(self):
        profile, sections = resolve_profile("5.1")
        assert profile is PROFILE_51
        assert sections is SECTIONS_51

    def test_unknown_profile_raises(self):
        with pytest.raises(ValueError, match="Unknown profile"):
            resolve_profile("3.5")
