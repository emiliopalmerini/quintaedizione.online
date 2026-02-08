"""Validation and summary report for parsed output."""

from __future__ import annotations

import json
from pathlib import Path


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

        # Check for empty required fields
        for entry in data:
            if not isinstance(entry, dict):
                continue
            entry_id = entry.get("id", "?")
            name = entry.get("name", entry.get("title", entry.get("term", "")))

            if not name:
                warnings += 1
                print(f"    WARN: {entry_id} has empty name/title")

            # Check for empty description (skip if entry has children)
            desc = entry.get("description", entry.get("content", entry.get("definition", entry.get("benefit", ""))))
            has_children = bool(entry.get("children"))
            if isinstance(desc, str) and not desc.strip() and not has_children:
                warnings += 1
                if count < 50:  # Only show for small collections
                    print(f"    WARN: {name or entry_id} has empty description")

        total_warnings += warnings
        total_errors += errors

    print(f"\n{'─' * 60}")
    print(f"TOTAL: {total_entries} entries, {total_warnings} warnings, {total_errors} errors")
    print("=" * 60)
