import re
from typing import Dict, Any

def q_filter(q: str) -> Dict[str, Any]:
    if not q:
        return {}
    r = {"$regex": re.escape(q), "$options": "i"}
    return {"$or": [
        {"name": r}, {"term": r}, {"description": r},
        {"description_md": r}, {"title": r}
    ]}

