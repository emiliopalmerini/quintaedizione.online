from __future__ import annotations

import curses
import os
from dataclasses import dataclass
from pathlib import Path
from typing import List, Optional, Tuple

from .parsers.spells import parse_spells
from .parsers.magic_items import parse_magic_items
from .parsers.equipment import parse_equipment
from .parsers.rules import parse_rules_glossary
from .parsers.monsters import parse_monsters
from .parsers.classes import parse_classes

try:
    from pymongo import MongoClient  # optional: only required for upsert
except Exception:  # pragma: no cover
    MongoClient = None  # type: ignore


@dataclass
class UiState:
    input_dir: Path
    mongo_uri: str
    db_name: str
    dry_run: bool = True
    selected: List[int] = None  # indices into DEFAULT_WORK
    cursor: int = 0

    def __post_init__(self) -> None:
        if self.selected is None:
            self.selected = []


@dataclass
class WorkItem:
    filename: str
    collection: str
    parser: callable


DEFAULT_WORK: List[WorkItem] = [
    WorkItem("08_b_spellsaz.md", "spells", parse_spells),
    WorkItem("07_magic_items.md", "magic_items", parse_magic_items),
    WorkItem("07_armor_items.md", "armor", parse_equipment),
    WorkItem("07_weapons_items.md", "weapons", parse_equipment),
    WorkItem("07_tools_items.md", "tools", parse_equipment),
    WorkItem("07_mounts_vehicles_items.md", "mounts_vehicles", parse_equipment),
    WorkItem("07_services_items.md", "services", parse_equipment),
    WorkItem("09_rules_glossary.md", "rules_glossary", parse_rules_glossary),
    WorkItem("13_monsters_az.md", "monsters", parse_monsters),
    WorkItem("14_animals.md", "animals", parse_monsters),
    WorkItem("04_classes.md", "classes", parse_classes),
]


def _read_lines(path: Path) -> List[str]:
    try:
        return path.read_text(encoding="utf-8").splitlines()
    except Exception as e:
        return [f"ERROR: Failed to read {path}: {e}"]


HELP_LINES = [
    "↑/↓ or k/j: move  •  Space: toggle  •  a: all  •  n: none",
    "d: toggle dry-run  •  u: toggle upsert (dry-run off)  •  e: edit input dir",
    "m: edit Mongo URI  •  b: edit DB name  •  r/Enter: run  •  q: quit",
]


def _edit_prompt(stdscr, title: str, initial: str) -> Optional[str]:
    curses.echo()
    h, w = stdscr.getmaxyx()
    # Ensure minimal window dimensions and positions within bounds
    win_h = 3
    win_w = max(10, w - 4)
    start_y = max(0, min(h - win_h, h // 2 - 1))
    start_x = max(0, min(w - win_w, 2))
    try:
        win = curses.newwin(win_h, win_w, start_y, start_x)
    except curses.error:
        curses.noecho()
        return None
    try:
        win.border()
        title_txt = f" {title} "
        try:
            win.addstr(0, 2, title_txt[: max(0, win_w - 4)])
        except curses.error:
            pass
        try:
            win.addstr(1, 2, initial[: max(0, win_w - 4)])
        except curses.error:
            pass
        win.refresh()
        input_y = start_y + 1
        input_x = start_x + 2 + min(len(initial), max(0, win_w - 4))
        try:
            stdscr.move(input_y, min(input_x, max(0, w - 1)))
        except curses.error:
            pass
        val = stdscr.getstr(input_y, start_x + 2, 1024)
    finally:
        curses.noecho()
    try:
        s = val.decode("utf-8").strip()
        return s or initial
    except Exception:
        return None


def _draw(stdscr, state: UiState, messages: List[str]) -> None:
    stdscr.clear()
    h, w = stdscr.getmaxyx()

    def safe_addstr(y: int, x: int, text: str) -> None:
        # Guard against out-of-bounds and clip text to available width with a small right margin
        if y < 0 or y >= h or x < 0 or x >= w:
            return
        avail = max(0, w - x - 1)  # keep a 1-col margin to avoid wrapping
        if avail:
            try:
                stdscr.addstr(y, x, text[:avail])
            except curses.error:
                pass

    # If terminal is too small, show a minimal hint and return
    if h < 10 or w < 30:
        safe_addstr(0, 1, "SRD Parser TUI")
        safe_addstr(2, 1, "Terminal too small. Resize to continue.")
        stdscr.refresh()
        return

    # Header
    safe_addstr(0, 2, "SRD Parser TUI")
    safe_addstr(1, 2, f"Input dir: {state.input_dir}")
    safe_addstr(2, 2, f"Dry-run: {'ON' if state.dry_run else 'OFF'}  |  Mongo: {state.mongo_uri}  DB: {state.db_name}")

    # Help
    for i, line in enumerate(HELP_LINES):
        y = 4 + i
        if y >= h:
            break
        safe_addstr(y, 2, line)

    # List box
    top = 7
    safe_addstr(top - 1, 2, "Select collections to parse:")
    for idx, witem in enumerate(DEFAULT_WORK):
        y = top + idx
        if y >= h:
            break
        mark = "[x]" if idx in state.selected else "[ ]"
        label = f"{mark} {witem.collection:16}  {witem.filename}"
        try:
            if idx == state.cursor:
                stdscr.attron(curses.A_REVERSE)
                safe_addstr(y, 4, label)
                stdscr.attroff(curses.A_REVERSE)
            else:
                safe_addstr(y, 4, label)
        except curses.error:
            pass

    # Messages area
    msg_top = top + len(DEFAULT_WORK) + 2
    if 0 <= msg_top - 1 < h and w > 2:
        try:
            stdscr.hline(msg_top - 1, 1, ord("-"), max(0, w - 2))
        except curses.error:
            pass
    max_lines = max(0, h - msg_top - 1)
    tail = messages[-max_lines:] if max_lines > 0 else []
    for i, line in enumerate(tail):
        y = msg_top + i
        if y >= h:
            break
        safe_addstr(y, 2, line)
    stdscr.refresh()


def _run_parse(state: UiState) -> List[str]:
    msgs: List[str] = []
    if not state.selected:
        return ["No collections selected."]
    base = state.input_dir
    if not base.exists():
        return [f"Input dir not found: {base}"]

    col_client = None
    db = None
    upsert_many = None
    unique_keys_for = None
    if not state.dry_run:
        if MongoClient is None:
            return ["pymongo not available; cannot upsert. Enable dry-run or install pymongo."]
        try:
            col_client = MongoClient(state.mongo_uri)
            db = col_client[state.db_name]
        except Exception as e:
            return [f"Mongo connect failed: {e}"]
        try:
            from .ingest import upsert_many as _upsert_many, unique_keys_for as _unique_keys_for
            upsert_many = _upsert_many
            unique_keys_for = _unique_keys_for
        except Exception as e:
            return [f"Failed to load ingest helpers: {e}"]

    total = 0
    for idx in state.selected:
        witem = DEFAULT_WORK[idx]
        path = base / witem.filename
        if not path.exists():
            msgs.append(f"Missing file: {path}")
            continue
        msgs.append(f"Parsing {path.name} → {witem.collection}")
        lines = _read_lines(path)
        try:
            docs = witem.parser(lines)
        except Exception as e:
            msgs.append(f"Parser error in {path.name}: {e}")
            continue
        msgs.append(f"Parsed {len(docs)} docs from {path.name}")
        if state.dry_run:
            total += len(docs)
            continue
        try:
            col = db[witem.collection]  # type: ignore[index]
            written = upsert_many(col, unique_keys_for(witem.collection), docs)  # type: ignore[misc]
            total += written
            msgs.append(f"Upserted {written} docs into {state.db_name}.{witem.collection}")
        except Exception as e:
            msgs.append(f"Upsert failed for {witem.collection}: {e}")
    if state.dry_run:
        msgs.append(f"Dry-run complete. Total parsed: {total}")
    else:
        msgs.append(f"Done. Total upserts: {total}")
    return msgs


def _toggle(state: UiState, idx: int) -> None:
    if idx in state.selected:
        state.selected.remove(idx)
    else:
        state.selected.append(idx)
        state.selected.sort()


def main() -> None:
    default_input = Path(os.environ.get("INPUT_DIR", "data"))
    state = UiState(
        input_dir=default_input,
        mongo_uri=os.environ.get("MONGO_URI", "mongodb://localhost:27017"),
        db_name=os.environ.get("DB_NAME", "dnd"),
        dry_run=True,
    )

    messages: List[str] = []

    def loop(stdscr) -> None:
        nonlocal messages
        curses.curs_set(0)
        stdscr.nodelay(False)
        while True:
            _draw(stdscr, state, messages)
            ch = stdscr.getch()
            if ch in (ord("q"), 27):  # q or ESC
                break
            if ch in (curses.KEY_DOWN, ord("j")):
                state.cursor = min(state.cursor + 1, len(DEFAULT_WORK) - 1)
            elif ch in (curses.KEY_UP, ord("k")):
                state.cursor = max(state.cursor - 1, 0)
            elif ch in (ord(" "),):
                _toggle(state, state.cursor)
            elif ch in (ord("a"),):
                state.selected = list(range(len(DEFAULT_WORK)))
            elif ch in (ord("n"),):
                state.selected = []
            elif ch in (ord("d"),):
                state.dry_run = not state.dry_run
            elif ch in (ord("u"),):
                state.dry_run = False
            elif ch in (ord("e"),):
                val = _edit_prompt(stdscr, "Edit input dir", str(state.input_dir))
                if val:
                    state.input_dir = Path(val)
            elif ch in (ord("m"),):
                val = _edit_prompt(stdscr, "Edit Mongo URI", state.mongo_uri)
                if val:
                    state.mongo_uri = val
            elif ch in (ord("b"),):
                val = _edit_prompt(stdscr, "Edit DB name", state.db_name)
                if val:
                    state.db_name = val
            elif ch in (ord("r"), ord("\n")):
                messages.extend(_run_parse(state))
            # loop continues

    curses.wrapper(loop)


if __name__ == "__main__":
    main()
