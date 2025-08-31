from __future__ import annotations

from typing import Any, Dict, Mapping, Optional, Tuple

from core.search import QFilterOptions, q_filter


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
        lvl = params.get("level") or params.get("livello")
        try:
            if lvl and str(lvl).strip().isdigit():
                v = int(str(lvl).strip())
                filt.setdefault("$or", []).extend([
                    {"level": v},
                    {"livello": v},
                ])
        except Exception:
            pass
        school = params.get("school") or params.get("scuola")
        if school:
            r = rx(school)
            filt.setdefault("$or", []).extend([
                {"school": r},
                {"scuola": r},
            ])
        ritual = parse_bool(params.get("ritual") or params.get("rituale"))
        if ritual is not None:
            filt.setdefault("$or", []).extend([
                {"ritual": ritual},
                {"rituale": ritual},
            ])
        classes = params.get("classes") or params.get("classi")
        if classes:
            r = rx(classes)
            filt.setdefault("$or", []).extend([
                {"classes": {"$elemMatch": r}},
                {"classi": {"$elemMatch": r}},
            ])

    elif collection == "magic_items":
        rarity = params.get("rarity") or params.get("rarita")
        if rarity:
            r = rx(rarity)
            filt.setdefault("$or", []).extend([
                {"rarity": r},
                {"rarita": r},
            ])
        itype = params.get("type") or params.get("tipo")
        if itype:
            r = rx(itype)
            filt.setdefault("$or", []).extend([
                {"type": r},
                {"tipo": r},
            ])
        att = parse_bool(params.get("attunement") or params.get("sintonizzazione"))
        if att is not None:
            filt.setdefault("$or", []).extend([
                {"attunement": att},
                {"sintonizzazione": att},
            ])

    elif collection == "monsters":
        size = params.get("size") or params.get("taglia")
        if size:
            r = rx(size)
            filt.setdefault("$or", []).extend([
                {"size": r},
                {"tag.taglia": r},
            ])
        mtype = params.get("type") or params.get("tipo")
        if mtype:
            r = rx(mtype)
            filt.setdefault("$or", []).extend([
                {"type": r},
                {"tag.tipo": r},
            ])
        align = params.get("alignment") or params.get("allineamento")
        if align:
            r = rx(align)
            filt.setdefault("$or", []).extend([
                {"alignment": r},
                {"tag.allineamento": r},
            ])
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

    elif collection == "weapons":
        category = params.get("category") or params.get("categoria")
        if category:
            r = rx(category)
            filt.setdefault("$or", []).extend([
                {"category": r},
                {"categoria": r},
            ])
        mastery = params.get("mastery") or params.get("maestria")
        if mastery:
            r = rx(mastery)
            filt.setdefault("$or", []).extend([
                {"mastery": r},
                {"maestria": r},
            ])
        prop = params.get("property") or params.get("proprieta")
        if prop:
            r = rx(prop)
            filt.setdefault("$or", []).extend([
                {"properties": {"$elemMatch": r}},
                {"proprieta": {"$elemMatch": r}},
            ])

    elif collection == "armor":
        category = params.get("category") or params.get("categoria")
        if category:
            r = rx(category)
            filt.setdefault("$or", []).extend([
                {"category": r},
                {"categoria": r},
            ])
        stealth = parse_bool(params.get("stealth") or params.get("svantaggio"))
        if stealth is not None:
            filt.setdefault("$or", []).extend([
                {"stealth_disadvantage": stealth},
                {"svantaggio_furtivita": stealth},
            ])
        strength = params.get("strength") or params.get("forza")
        if strength:
            r = rx(strength)
            filt.setdefault("$or", []).extend([
                {"strength": r},
                {"forza": r},
            ])

    elif collection == "tools":
        ability = params.get("ability") or params.get("abilita")
        if ability:
            r = rx(ability)
            filt.setdefault("$or", []).extend([
                {"ability": r},
                {"abilita": r},
            ])
        category = params.get("category") or params.get("categoria")
        if category:
            r = rx(category)
            filt.setdefault("$or", []).extend([
                {"category": r},
                {"categoria": r},
            ])
        craft = params.get("craft") or params.get("crea")
        if craft:
            r = rx(craft)
            filt.setdefault("$or", []).extend([
                {"craft": {"$elemMatch": r}},
                {"crea": {"$elemMatch": r}},
            ])

    elif collection == "services":
        category = params.get("category") or params.get("categoria")
        if category:
            r = rx(category)
            filt.setdefault("$or", []).extend([
                {"category": r},
                {"categoria": r},
            ])
        avail = params.get("availability") or params.get("disponibilita") or params.get("disponibilità")
        if avail:
            r = rx(avail)
            filt.setdefault("$or", []).extend([
                {"availability": r},
                {"disponibilita": r},
                {"disponibilità": r},
            ])

    elif collection == "gear":
        weight = params.get("weight") or params.get("peso")
        if weight:
            r = rx(weight)
            filt.setdefault("$or", []).extend([
                {"weight": r},
                {"peso": r},
            ])

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
