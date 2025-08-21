from __future__ import annotations
import json, re
from typing import Any, Dict
from bson import ObjectId

def to_jsonable(obj: Any) -> Any:
    if isinstance(obj, ObjectId): return str(obj)
    if isinstance(obj, list): return [to_jsonable(x) for x in obj]
    if isinstance(obj, dict): return {k: to_jsonable(v) for k, v in obj.items()}
    return obj

def coerce_scalar(s: str) -> Any:
    v = s.strip()
    if v == "": return ""
    low = v.lower()
    if low in ("true","false"): return low == "true"
    if low in ("null","none"): return None
    if re.fullmatch(r"-?\d+", v): 
        try: return int(v)
        except: pass
    if re.fullmatch(r"-?\d+\.\d+", v):
        try: return float(v)
        except: pass
    try:
        return json.loads(v)
    except: 
        return v

def rebuild_document_from_form(flat: Dict[str, str]) -> Dict[str, Any]:
    """
    Ricostruisce il documento a partire da input tipo:
    f.name=... , f.abilities.STR.score=..., f.actions.0.name=...
    """
    doc: Dict[str, Any] = {}
    keys = sorted([k for k in flat if k.startswith("f.")])
    for k in keys:
        path = k[2:]  # remove 'f.'
        parts = path.split(".")
        cur: Any = doc
        for i, part in enumerate(parts):
            is_last = i == len(parts) - 1
            is_index = part.isdigit()
            if is_last:
                if is_index:
                    # assegnazione ad indice lista
                    if not isinstance(cur, list):
                        raise ValueError(f"List expected before '{part}' in '{path}'")
                    idx = int(part)
                    while len(cur) <= idx: cur.append(None)
                    cur[idx] = coerce_scalar(flat[k])
                else:
                    if isinstance(cur, list):
                        raise ValueError(f"Cannot set key '{part}' on list in '{path}'")
                    cur[part] = coerce_scalar(flat[k])
            else:
                nxt = parts[i+1]
                next_is_index = nxt.isdigit()
                if is_index:
                    idx = int(part)
                    if not isinstance(cur, list):
                        raise ValueError(f"List expected before '{part}' in '{path}'")
                    while len(cur) <= idx:
                        cur.append([] if next_is_index else {})
                    if cur[idx] is None:
                        cur[idx] = [] if next_is_index else {}
                    cur = cur[idx]
                else:
                    if part not in cur or not isinstance(cur[part], (dict, list)):
                        cur[part] = [] if next_is_index else {}
                    cur = cur[part]
    return doc

