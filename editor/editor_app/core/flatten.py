from __future__ import annotations

from dataclasses import dataclass, field
from datetime import date, datetime
from decimal import Decimal
from typing import Any, Callable, Dict, Iterable, List, Set, Tuple, Union

ScalarTypes = (str, int, float, bool, type(None))
ExtraScalarTypes = (datetime, date, Decimal)  # serializzati

Path = str
FlatItem = Tuple[Path, Any]
Predicate = Callable[[Path, Any], bool]


@dataclass(frozen=True)
class FlattenOptions:
    sep: str = "."  # separatore tra segmenti
    index_style: str = "dot"  # "dot" -> a.0; "bracket" -> a[0]
    exclude_keys: Set[str] = field(default_factory=lambda: {"_id"})
    include: Predicate | None = None  # True => includi
    exclude: Predicate | None = None  # True => escludi
    scalar_pred: Callable[[Any], bool] | None = None
    stringify_unknown: bool = True  # gestisce tipi non attesi
    max_depth: int | None = None  # None = illimitato
    sort: bool = True


def _default_is_scalar(x: Any) -> bool:
    return isinstance(x, ScalarTypes + ExtraScalarTypes)


def _coerce(value: Any) -> Any:
    if isinstance(value, (datetime, date)):
        return value.isoformat()
    if isinstance(value, Decimal):
        return float(value)
    return value


def _join(seg: List[str], opts: FlattenOptions) -> str:
    if not seg:
        return ""
    if opts.index_style == "bracket":
        out: List[str] = []
        for s in seg:
            if s.isdigit():
                out[-1] = f"{out[-1]}[{s}]"
            else:
                out.append(s)
        return opts.sep.join(out)
    return opts.sep.join(seg)


def flatten_for_form(
    doc: Dict[str, Any], *, options: FlattenOptions | None = None
) -> List[FlatItem]:
    opts = options or FlattenOptions()
    is_scalar = opts.scalar_pred or _default_is_scalar

    out: List[FlatItem] = []
    stack: List[Tuple[Any, List[str], int]] = [(doc, [], 0)]

    while stack:
        node, path_segs, depth = stack.pop()
        path = _join(path_segs, opts)

        # filtri chiave diretta
        if path in opts.exclude_keys:
            continue

        # profondità
        if opts.max_depth is not None and depth > opts.max_depth:
            continue

        # foglia scalare
        if is_scalar(node):
            val = _coerce(node)
            if opts.include and not opts.include(path, val):
                continue
            if opts.exclude and opts.exclude(path, val):
                continue
            out.append((path, val))
            continue

        # dict
        if isinstance(node, dict):
            # iterazione stabile ma senza costi elevati
            for k, v in node.items():
                segs = path_segs + [k] if path_segs else [k]
                stack.append((v, segs, depth + 1))
            continue

        # list/tuple
        if isinstance(node, (list, tuple)):
            for i, v in enumerate(node):
                segs = path_segs + [str(i)]
                stack.append((v, segs, depth + 1))
            continue

        # tipo non previsto
        if opts.stringify_unknown:
            val = str(node)
            if opts.include and not opts.include(path, val):
                continue
            if opts.exclude and opts.exclude(path, val):
                continue
            out.append((path, val))

    if opts.sort:
        out.sort(key=lambda kv: kv[0])
    return out
