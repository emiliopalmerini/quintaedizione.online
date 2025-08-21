from __future__ import annotations
from typing import Any, Dict, List, Tuple

Scalar = (str, int, float, bool, type(None))


def _is_scalar(x: Any) -> bool:
    return isinstance(x, Scalar)


def flatten_for_form(doc: Dict[str, Any]) -> List[Tuple[str, Any]]:
    """
    Ritorna solo campi *scalari* come coppie (path, value).
    - Niente nodo root.
    - Niente contenitori dict/list.
    - Per le liste: genera path "field.0", "field.1", ...
    - Esclude _id (verrà mostrato come read-only separatamente).
    """
    out: List[Tuple[str, Any]] = []

    def walk(node: Any, path: str) -> None:
        if path == "_id":
            return
        if _is_scalar(node):
            out.append((path, node))
            return
        if isinstance(node, dict):
            for k, v in node.items():
                new_path = f"{path}.{k}" if path else k
                walk(v, new_path)
            return
        if isinstance(node, list):
            for i, v in enumerate(node):
                new_path = f"{path}.{i}" if path else str(i)
                walk(v, new_path)
            return
        # altri tipi non previsti: serializza a stringa
        out.append((path, str(node)))

    walk(doc, "")
    # ordina per path per stabilità
    out.sort(key=lambda kv: kv[0])
    return out

