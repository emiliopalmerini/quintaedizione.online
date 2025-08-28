from __future__ import annotations

from typing import Any, Dict, Mapping, Optional, Tuple

from editor_app.core.search import QFilterOptions, q_filter


def rx(val: str) -> Dict[str, Any]:
    return {"$regex": val, "$options": "i"}


def parse_bool(val: Optional[str]) -> Optional[bool]:
    if val is None:
        return None
    v = val.strip().lower()
    if v in ("1", "true", "yes", "y", "si", "s"):
        return True
    if v in ("0", "false", "no", "n"):
        return False
    return None


def build_filter(
    q: str | None, collection: str, params: Mapping[str, str], *, quick: bool = False
) -> Dict[str, Any]:
    filt: Dict[str, Any] = {}

    if q:
        if quick:
            q_opts = QFilterOptions(
                fields=["name", "term", "title", "titolo", "nome"],
                min_len=1,
                prefix=True,
                raw_regex=False,
                whole_words=False,
            )
        else:
            q_opts = QFilterOptions(
                fields=["name", "term", "title", "titolo", "nome", "description", "description_md", "content"],
                min_len=2,
            )
        qf = q_filter(q, options=q_opts)
        if qf:
            filt.update(qf)

    if collection == "spells":
        lvl = params.get("level")
        if lvl and lvl.isdigit():
            filt["level"] = int(lvl)
        school = params.get("school")
        if school:
            filt["school"] = rx(school)
        ritual = parse_bool(params.get("ritual"))
        if ritual is not None:
            filt["ritual"] = ritual
        classes = params.get("classes")
        if classes:
            filt["classes"] = {"$elemMatch": rx(classes)}

    elif collection == "magic_items":
        rarity = params.get("rarity")
        if rarity:
            filt["rarity"] = rx(rarity)
        itype = params.get("type")
        if itype:
            filt["type"] = rx(itype)
        att = parse_bool(params.get("attunement"))
        if att is not None:
            filt["attunement"] = att

    elif collection == "monsters":
        size = params.get("size")
        if size:
            filt["size"] = rx(size)
        mtype = params.get("type")
        if mtype:
            filt["type"] = rx(mtype)
        align = params.get("alignment")
        if align:
            filt["alignment"] = rx(align)
        cr = params.get("cr")
        if cr:
            try:
                cr_num = float(cr) if "." in cr or "/" not in cr else None
            except Exception:
                cr_num = None
            if cr_num is not None:
                filt.setdefault("$and", []).append({"$or": [{"challenge_rating": cr_num}, {"cr": cr_num}]})
            else:
                r = rx(cr)
                filt.setdefault("$and", []).append({"$or": [{"challenge_rating": r}, {"cr": r}]})

    return filt


def alpha_sort_expr() -> Dict[str, Any]:
    return {
        "$toLower": {
            "$ifNull": [
                "$_sortkey_alpha",
                {
                    "$ifNull": [
                        "$slug",
                        {
                            "$ifNull": [
                                "$name",
                                {
                                    "$ifNull": [
                                        "$term",
                                        {
                                            "$ifNull": [
                                                "$title",
                                                {"$ifNull": ["$titolo", {"$ifNull": ["$nome", ""]}]},
                                            ]
                                        },
                                    ]
                                },
                            ]
                        },
                    ]
                },
            ]
        }
    }


async def neighbors_alpha_repo(repo, collection: str, cur_key: str, filt: Dict[str, Any]) -> Tuple[Optional[str], Optional[str]]:
    key = (cur_key or "").lower()
    prev_pipe = [
        {"$match": filt},
        {"$addFields": {"_sortkey": alpha_sort_expr()}},
        {"$match": {"_sortkey": {"$lt": key}}},
        {"$sort": {"_sortkey": -1, "_id": -1}},
        {"$limit": 1},
        {"$project": {"_id": 1}},
    ]
    next_pipe = [
        {"$match": filt},
        {"$addFields": {"_sortkey": alpha_sort_expr()}},
        {"$match": {"_sortkey": {"$gt": key}}},
        {"$sort": {"_sortkey": 1, "_id": 1}},
        {"$limit": 1},
        {"$project": {"_id": 1}},
    ]
    prev = await repo.aggregate_list(collection, prev_pipe)
    nxt = await repo.aggregate_list(collection, next_pipe)
    prev_id = str(prev[0].get("_id")) if prev else None
    next_id = str(nxt[0].get("_id")) if nxt else None
    return prev_id, next_id
