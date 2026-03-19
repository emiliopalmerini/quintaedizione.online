"""Rules parser for SRD 5.1 — identical to 5.2.1."""

from ..parsers.rules import parse_rules  # noqa: F401
from . import register

# Re-register under 5.1's registry
register("rules")(parse_rules)
