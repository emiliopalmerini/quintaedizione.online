"""Validation and summary report for parsed output."""

from __future__ import annotations

import json
from pathlib import Path

# ── Enum sets for validation ────────────────────────────────────────────────

_SPELL_SCHOOLS = {
    "abiurazione", "ammaliamento", "divinazione", "evocazione",
    "illusione", "invocazione", "necromanzia", "trasmutazione",
}

_MAGIC_ITEM_RARITIES = {
    "comune", "non comune", "raro", "rara", "molto raro", "molto rara",
    "leggendario", "leggendaria", "manufatto", "varia", "rarità variabile",
}

_EQUIPMENT_CATEGORIES = {
    "weapons", "armor", "tools", "gear", "mounts", "services",
}


# ── Per-collection validators ───────────────────────────────────────────────

def _validate_spells(data: list[dict], verbose: bool = True) -> tuple[int, int]:
    """Validate spell entries. Returns (warnings, errors)."""
    warnings = 0
    errors = 0
    required = ("school", "casting_time", "range", "components", "duration", "description")
    for entry in data:
        name = entry.get("name", "?")
        for field in required:
            val = entry.get(field)
            if not val or (isinstance(val, str) and not val.strip()) or (isinstance(val, list) and len(val) == 0):
                errors += 1
                if verbose:
                    print(f"    ERROR: {name} — empty {field}")
        # Enum: school
        school = entry.get("school", "").lower()
        if school and school not in _SPELL_SCHOOLS:
            warnings += 1
            if verbose:
                print(f"    WARN: {name} — unknown school '{entry.get('school')}'")
    return warnings, errors


def _validate_monsters(data: list[dict], verbose: bool = True) -> tuple[int, int]:
    """Validate monster entries. Returns (warnings, errors)."""
    warnings = 0
    errors = 0
    required = ("type", "size", "ac", "hp", "cr")
    for entry in data:
        name = entry.get("name", "?")
        for field in required:
            val = entry.get(field)
            if not val or (isinstance(val, str) and not val.strip()):
                errors += 1
                if verbose:
                    print(f"    ERROR: {name} — empty {field}")
    return warnings, errors


def _validate_magic_items(data: list[dict], verbose: bool = True) -> tuple[int, int]:
    """Validate magic item entries. Returns (warnings, errors)."""
    warnings = 0
    errors = 0
    required = ("type", "rarity", "description")
    for entry in data:
        name = entry.get("name", "?")
        for field in required:
            val = entry.get(field)
            if not val or (isinstance(val, str) and not val.strip()):
                errors += 1
                if verbose:
                    print(f"    ERROR: {name} — empty {field}")
        # Enum: rarity
        rarity = entry.get("rarity", "").lower()
        if rarity and rarity not in _MAGIC_ITEM_RARITIES:
            warnings += 1
            if verbose:
                print(f"    WARN: {name} — unknown rarity '{entry.get('rarity')}'")
    return warnings, errors


def _validate_equipment(data: list[dict], verbose: bool = True) -> tuple[int, int]:
    """Validate equipment entries. Returns (warnings, errors)."""
    warnings = 0
    errors = 0
    required = ("category", "description")
    for entry in data:
        name = entry.get("name", "?")
        for field in required:
            val = entry.get(field)
            if not val or (isinstance(val, str) and not val.strip()):
                errors += 1
                if verbose:
                    print(f"    ERROR: {name} — empty {field}")
        # Empty subcategory
        if not entry.get("subcategory", "").strip():
            warnings += 1
            if verbose:
                print(f"    WARN: {name} — empty subcategory")
        # Enum: category
        cat = entry.get("category", "")
        if cat and cat not in _EQUIPMENT_CATEGORIES:
            warnings += 1
            if verbose:
                print(f"    WARN: {name} — unknown category '{cat}'")
    return warnings, errors


def _validate_feats(data: list[dict], verbose: bool = True) -> tuple[int, int]:
    """Validate feat entries. Returns (warnings, errors)."""
    warnings = 0
    errors = 0
    required = ("benefit",)
    for entry in data:
        name = entry.get("name", "?")
        for field in required:
            val = entry.get(field)
            if not val or (isinstance(val, str) and not val.strip()):
                errors += 1
                if verbose:
                    print(f"    ERROR: {name} — empty {field}")
    return warnings, errors


def _validate_backgrounds(data: list[dict], verbose: bool = True) -> tuple[int, int]:
    """Validate background entries. Returns (warnings, errors)."""
    warnings = 0
    errors = 0
    required = ("description",)
    for entry in data:
        name = entry.get("name", "?")
        for field in required:
            val = entry.get(field)
            if not val or (isinstance(val, str) and not val.strip()):
                errors += 1
                if verbose:
                    print(f"    ERROR: {name} — empty {field}")
    return warnings, errors


def _validate_classes(data: list[dict], verbose: bool = True) -> tuple[int, int]:
    """Validate class entries. Returns (warnings, errors)."""
    warnings = 0
    errors = 0
    required = ("hit_die", "description")
    for entry in data:
        name = entry.get("name", "?")
        for field in required:
            val = entry.get(field)
            if not val or (isinstance(val, str) and not val.strip()):
                errors += 1
                if verbose:
                    print(f"    ERROR: {name} — empty {field}")
    return warnings, errors


# Map output filename to validator
_VALIDATORS: dict[str, object] = {
    "spells.json": _validate_spells,
    "monsters.json": _validate_monsters,
    "magic_items.json": _validate_magic_items,
    "equipment.json": _validate_equipment,
    "feats.json": _validate_feats,
    "backgrounds.json": _validate_backgrounds,
    "classes.json": _validate_classes,
}


def validate_output(output_dir: Path) -> None:
    """Validate all JSON output files and print a summary report."""
    print("\n" + "=" * 60)
    print("QUALITY REPORT")
    print("=" * 60)

    total_entries = 0
    total_warnings = 0
    total_errors = 0

    for json_file in sorted(output_dir.glob("*.json")):
        data = json.loads(json_file.read_text(encoding="utf-8"))

        if not isinstance(data, list):
            print(f"\n  {json_file.name}: not a list, skipping")
            continue

        count = len(data)
        total_entries += count
        warnings = 0
        errors = 0

        # Check for unique IDs
        ids = [entry.get("id", "") for entry in data if isinstance(entry, dict)]
        unique_ids = set(ids)
        if len(ids) != len(unique_ids):
            dupes = len(ids) - len(unique_ids)
            errors += dupes
            print(f"\n  {json_file.name}: {count} entries, {dupes} DUPLICATE IDs")
        else:
            print(f"\n  {json_file.name}: {count} entries")

        # Generic checks: empty name/description
        for entry in data:
            if not isinstance(entry, dict):
                continue
            entry_id = entry.get("id", "?")
            name = entry.get("name", entry.get("title", entry.get("term", "")))

            if not name:
                warnings += 1
                print(f"    WARN: {entry_id} has empty name/title")

            # Check for empty description (skip if entry has children or structured content)
            desc = entry.get("description", entry.get("content", entry.get("definition", entry.get("benefit", ""))))
            has_children = bool(entry.get("children"))
            has_features = bool(entry.get("traits") or entry.get("actions"))
            if isinstance(desc, str) and not desc.strip() and not has_children and not has_features:
                warnings += 1
                if count < 50:  # Only show for small collections
                    print(f"    WARN: {name or entry_id} has empty description")

        # Per-collection validation
        validator = _VALIDATORS.get(json_file.name)
        if validator:
            cw, ce = validator(data)  # type: ignore[operator]
            warnings += cw
            errors += ce

        total_warnings += warnings
        total_errors += errors

    print(f"\n{'─' * 60}")
    print(f"TOTAL: {total_entries} entries, {total_warnings} warnings, {total_errors} errors")
    print("=" * 60)
