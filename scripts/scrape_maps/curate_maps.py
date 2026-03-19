#!/usr/bin/env python3
"""Transform raw Dyson Logos scrape data into curated Italian mappe.json.

Usage:
    python scripts/scrape_maps/curate_maps.py [--input FILE] [--output FILE]

Reads dyson_raw.json and produces a curated mappe.json with:
- Italian translated names
- Assigned categories
- Italian tags
- Generated slugs
"""

import argparse
import json
import re
import unicodedata


# ── Name translation ──────────────────────────────────────────────────────────

# Word-level translations (lowercase key → Italian)
WORD_MAP = {
    # Articles & prepositions
    "the": "", "of": "di", "on": "su", "in": "in", "at": "al",
    "under": "sotto", "beneath": "sotto", "below": "sotto",
    "above": "sopra", "over": "sopra",
    "between": "tra", "beyond": "oltre", "through": "attraverso",
    "near": "presso", "into": "dentro", "from": "da",
    "and": "e", "with": "con", "without": "senza",

    # Common nouns - places
    "temple": "tempio", "shrine": "santuario", "church": "chiesa",
    "abbey": "abbazia", "monastery": "monastero", "chapel": "cappella",
    "cathedral": "cattedrale", "fane": "tempio",
    "castle": "castello", "keep": "fortino", "fort": "forte",
    "fortress": "fortezza", "stronghold": "roccaforte", "citadel": "cittadella",
    "tower": "torre", "spire": "guglia", "lighthouse": "faro",
    "dungeon": "sotterraneo", "crypt": "cripta", "tomb": "tomba",
    "catacombs": "catacombe", "catacomb": "catacomba",
    "sepulchre": "sepolcro", "ossuary": "ossario",
    "barrow": "tumulo", "cairn": "tumulo", "mound": "tumulo",
    "graveyard": "cimitero", "cemetery": "cimitero",
    "cave": "caverna", "cavern": "caverna", "caverns": "caverne",
    "caves": "caverne", "grotto": "grotta",
    "mine": "miniera", "mines": "miniere",
    "sewer": "fogna", "sewers": "fogne",
    "labyrinth": "labirinto", "maze": "labirinto",
    "city": "citta", "town": "villaggio", "village": "villaggio",
    "hamlet": "borgata", "settlement": "insediamento",
    "market": "mercato", "square": "piazza", "plaza": "piazza",
    "street": "strada", "alley": "vicolo", "road": "strada",
    "bridge": "ponte", "gate": "porta", "gates": "porte",
    "wall": "muro", "walls": "mura",
    "inn": "locanda", "tavern": "taverna", "bar": "taverna",
    "shop": "bottega", "store": "negozio",
    "manor": "maniero", "mansion": "palazzo", "palace": "palazzo",
    "house": "casa", "home": "dimora", "hall": "sala",
    "guildhall": "gilda", "library": "biblioteca",
    "prison": "prigione", "jail": "prigione", "dungeon": "prigione",
    "observatory": "osservatorio", "workshop": "laboratorio",
    "farm": "fattoria", "mill": "mulino",
    "island": "isola", "islands": "isole", "isle": "isola",
    "lake": "lago", "river": "fiume", "pool": "pozza",
    "waterfall": "cascata", "spring": "sorgente",
    "mountain": "montagna", "mountains": "montagne",
    "hill": "collina", "hills": "colline", "peak": "vetta",
    "forest": "foresta", "woods": "bosco", "wood": "bosco",
    "swamp": "palude", "marsh": "palude", "bog": "palude",
    "desert": "deserto", "dune": "duna", "dunes": "dune",
    "volcano": "vulcano",
    "ruins": "rovine", "ruin": "rovina",
    "courtyard": "cortile", "garden": "giardino",
    "chamber": "camera", "room": "stanza", "rooms": "stanze",
    "passage": "passaggio", "corridor": "corridoio",
    "pit": "fossa", "well": "pozzo", "shaft": "pozzo",
    "portal": "portale", "door": "porta",
    "stairs": "scale", "stairway": "scalinata",
    "basement": "seminterrato",

    # Common nouns - creatures/people
    "dragon": "drago", "dragons": "draghi",
    "goblin": "goblin", "goblins": "goblin",
    "dwarf": "nano", "dwarves": "nani", "dwarven": "nanico",
    "elf": "elfo", "elves": "elfi", "elven": "elfico",
    "drow": "drow",
    "orc": "orco", "orcs": "orchi",
    "giant": "gigante", "giants": "giganti",
    "demon": "demone", "demons": "demoni",
    "wizard": "mago", "witch": "strega",
    "king": "re", "queen": "regina", "prince": "principe",
    "knight": "cavaliere", "lord": "signore",
    "thief": "ladro", "thieves": "ladri",
    "assassin": "assassino", "bandit": "bandito", "bandits": "banditi",
    "pirate": "pirata", "pirates": "pirati",
    "druid": "druido",
    "monk": "monaco", "monks": "monaci",
    "priest": "sacerdote",
    "necromancer": "negromante",
    "emperor": "imperatore", "empress": "imperatrice",

    # Adjectives
    "ancient": "antico", "old": "vecchio",
    "dark": "oscuro", "black": "nero", "shadow": "ombra",
    "white": "bianco", "red": "rosso", "blue": "blu", "green": "verde",
    "golden": "dorato", "gold": "oro", "silver": "argento",
    "iron": "ferro", "stone": "pietra", "crystal": "cristallo",
    "hidden": "nascosto", "secret": "segreto", "lost": "perduto",
    "fallen": "caduto", "broken": "spezzato", "crumbling": "fatiscente",
    "ruined": "rovinato", "shattered": "infranto",
    "sunken": "sommerso", "flooded": "allagato", "submerged": "sommerso",
    "frozen": "ghiacciato", "burning": "ardente",
    "sacred": "sacro", "holy": "sacro", "cursed": "maledetto",
    "haunted": "infestato", "ghostly": "spettrale",
    "great": "grande", "small": "piccolo", "little": "piccolo",
    "deep": "profondo", "deeper": "piu profondo",
    "upper": "superiore", "lower": "inferiore",
    "northern": "settentrionale", "southern": "meridionale",
    "eastern": "orientale", "western": "occidentale",
    "north": "nord", "south": "sud", "east": "est", "west": "ovest",
    "new": "nuovo", "twisted": "contorto",
    "abandoned": "abbandonato", "forgotten": "dimenticato",

    # Other common words
    "lair": "tana", "den": "tana", "nest": "nido",
    "throne": "trono", "altar": "altare",
    "treasure": "tesoro", "vault": "volta",
    "camp": "accampamento",
    "tomb": "tomba",
    "deeps": "profondita",
    "depths": "profondita",
    "level": "livello",
    "map": "mappa",
    "building": "edificio",
    "part": "parte", "parts": "parti",
    "two": "due", "three": "tre", "four": "quattro", "five": "cinque",
}

# Full name overrides for tricky cases (lowercase → Italian)
NAME_OVERRIDES = {
    "a dungeon in two parts": "Un Sotterraneo in Due Parti",
    "the pipeworks": "Le Tubature",
}


def slugify(text: str) -> str:
    """Generate a URL-safe slug from Italian text."""
    text = text.lower().strip()
    text = unicodedata.normalize("NFKD", text)
    text = text.encode("ascii", "ignore").decode("ascii")
    text = re.sub(r"[^\w\s-]", "", text)
    text = re.sub(r"[-\s]+", "-", text)
    return text.strip("-")


def translate_name(english_name: str) -> str:
    """Best-effort translation of map name to Italian."""
    lower = english_name.lower().strip()

    # Check full name overrides first
    if lower in NAME_OVERRIDES:
        return NAME_OVERRIDES[lower]

    # Handle series patterns like "Index Card Dungeon II - Map 12"
    # Keep series identifiers, translate the rest
    series_match = re.match(r"^(Index Card Dungeon (?:II )?- Map \d+)(?:\s*[-–]\s*(.+))?$", english_name)
    if series_match:
        series_part = series_match.group(1)
        subtitle = series_match.group(2)
        if subtitle:
            translated_subtitle = _translate_phrase(subtitle)
            return f"{series_part} - {translated_subtitle}"
        return series_part

    # Handle "Building N - Name" pattern
    building_match = re.match(r"^(Building \d+)\s*[-–]\s*(.+)$", english_name)
    if building_match:
        num = building_match.group(1).replace("Building", "Edificio")
        subtitle = _translate_phrase(building_match.group(2))
        return f"{num} - {subtitle}"

    return _translate_phrase(english_name)


def _translate_phrase(phrase: str) -> str:
    """Translate a phrase word by word, preserving proper nouns."""
    words = phrase.split()
    translated = []

    i = 0
    while i < len(words):
        word = words[i]
        clean = re.sub(r"[^\w']", "", word.lower())

        if clean in WORD_MAP:
            replacement = WORD_MAP[clean]
            if replacement:  # Skip empty replacements (like "the")
                # Preserve punctuation
                prefix = ""
                suffix = ""
                if word and not word[0].isalnum():
                    prefix = word[0]
                if word and not word[-1].isalnum():
                    suffix = word[-1]
                translated.append(prefix + replacement + suffix)
        else:
            # Keep as-is (proper noun or unknown word)
            translated.append(word)
        i += 1

    result = " ".join(translated)

    # Capitalize first letter
    if result:
        result = result[0].upper() + result[1:]

    # Apply Italian article rules for common patterns
    result = _fix_articles(result)

    return result


def _fix_articles(name: str) -> str:
    """Add/fix Italian articles where appropriate."""
    # "di il" → "del", "di la" → "della", etc.
    name = re.sub(r"\bdi il\b", "del", name, flags=re.IGNORECASE)
    name = re.sub(r"\bdi lo\b", "dello", name, flags=re.IGNORECASE)
    name = re.sub(r"\bdi la\b", "della", name, flags=re.IGNORECASE)
    name = re.sub(r"\bdi i\b", "dei", name, flags=re.IGNORECASE)
    name = re.sub(r"\bdi le\b", "delle", name, flags=re.IGNORECASE)
    name = re.sub(r"\bsu il\b", "sul", name, flags=re.IGNORECASE)
    name = re.sub(r"\bsu la\b", "sulla", name, flags=re.IGNORECASE)
    return name


# ── Category assignment ───────────────────────────────────────────────────────

# Priority-ordered: first matching rule wins
CATEGORY_RULES = [
    # Specific types first
    ({"Temple", "Shrine", "Church", "Abbey", "Monastery", "Fane", "Chapel"}, "tempio"),
    ({"Cave", "Cavern", "Underdark", "Stalagmite"}, "caverna"),
    ({"Tower", "Wizard Tower", "Wizard's Tower"}, "torre"),
    ({"Castle", "Keep", "Fort", "Fortre", "Fortification", "Citadel", "Stronghold"}, "fortezza"),
    ({"Tomb", "Crypt", "Barrow Mound", "Graveyard", "Ossuary", "Sepulchre"}, "tomba"),
    ({"Sewer"}, "fogna"),
    ({"Manor", "Mansion", "House", "Home", "Haunted House"}, "residenza"),
    ({"Inn", "Tavern", "Bar"}, "locanda"),
    ({"Store", "Shop", "Small Shops Serie", "Blacksmith", "Apothecary", "Herbalist"}, "negozio"),
    ({"City", "Town", "Village", "Urban", "Thorpe"}, "citta"),
    ({"Island", "Coastal", "Ocean", "Lake", "River", "Waterfall", "Pool"}, "acquatico"),
    ({"Mountain", "Hill", "Volcano", "Desert", "Dune", "Swamp", "Forest"}, "natura"),
    ({"Regional", "Regional Map", "Hex Map", "Hexmap", "Hex Crawl"}, "regionale"),
    ({"Isometric", "Side View", "Perspective"}, "vista"),
    ({"Geomorph", "Dungeon Geomorph"}, "geomorph"),
    ({"Ruin"}, "rovine"),
    ({"Dungeon", "Megadungeon", "Mega Dungeon"}, "dungeon"),
]


def assign_category(tags: list[str]) -> str:
    """Assign an Italian category based on English tags."""
    tag_set = set(tags)
    for trigger_tags, category in CATEGORY_RULES:
        if tag_set & trigger_tags:
            return category
    return "altro"


# ── Tag translation ───────────────────────────────────────────────────────────

TAG_MAP = {
    # Places
    "Dungeon": "sotterraneo",
    "Megadungeon": "megadungeon",
    "Mega Dungeon": "megadungeon",
    "Cave": "caverna",
    "Cavern": "caverna",
    "Crypt": "cripta",
    "Tomb": "tomba",
    "Temple": "tempio",
    "Shrine": "santuario",
    "Church": "chiesa",
    "Tower": "torre",
    "Castle": "castello",
    "Keep": "fortino",
    "Fort": "forte",
    "Fortre": "fortezza",
    "Fortification": "fortificazione",
    "Stronghold": "roccaforte",
    "Citadel": "cittadella",
    "Ruin": "rovine",
    "Sewer": "fogna",
    "Mine": "miniera",
    "Inn": "locanda",
    "Tavern": "taverna",
    "Manor": "maniero",
    "Mansion": "palazzo",
    "House": "casa",
    "Building": "edificio",
    "City": "citta",
    "Town": "villaggio",
    "Village": "villaggio",
    "Urban": "urbano",
    "Island": "isola",
    "River": "fiume",
    "Lake": "lago",
    "Coastal": "costiero",
    "Mountain": "montagna",
    "Hill": "collina",
    "Swamp": "palude",
    "Desert": "deserto",
    "Volcano": "vulcano",
    "Forest": "foresta",
    "Underdark": "sottosuolo",
    "Prison": "prigione",
    "Bridge": "ponte",
    "Graveyard": "cimitero",
    "Barrow Mound": "tumulo",
    "Farm": "fattoria",
    "Store": "negozio",
    "Shop": "bottega",
    "Market": "mercato",
    "Library": "biblioteca",
    "Courtyard": "cortile",
    "Palace": "palazzo",
    "Lighthouse": "faro",
    "Labyrinth": "labirinto",
    "Maze": "labirinto",
    "Basement": "seminterrato",
    "Pool": "pozza",
    "Waterfall": "cascata",
    "Outdoor": "esterno",
    "Portal": "portale",
    "Pagoda": "pagoda",
    "Pyramid": "piramide",
    "Catacomb": "catacomba",
    "Ossuary": "ossario",
    "Sepulchre": "sepolcro",
    "Monastery": "monastero",
    "Abbey": "abbazia",

    # Creatures/people
    "Dragon": "drago",
    "Goblin": "goblin",
    "Dwarf": "nano",
    "Dwarve": "nano",
    "Giant": "gigante",
    "Demon": "demone",
    "Drow": "drow",
    "Dark Elve": "elfo oscuro",
    "Wizard": "mago",
    "Druid": "druido",
    "Pirate": "pirata",
    "Bandit": "bandito",
    "Assassin": "assassino",
    "Cult": "culto",
    "Banshee": "banshee",
    "Githyanki": "githyanki",
    "Serpent Folk": "uomini serpente",

    # Features
    "Water": "acqua",
    "Mushroom": "funghi",
    "Fungu": "funghi",
    "Standing Stone": "menhir",
    "Pipe": "condutture",
    "Treasure Map": "mappa del tesoro",
    "Sinkhole": "voragine",

    # Descriptors
    "Regional": "regionale",
    "Regional Map": "mappa regionale",
    "Hex Map": "mappa esagonale",
    "Hexmap": "mappa esagonale",
    "Hex Crawl": "esplorazione esagonale",
    "Hex": "esagonale",
    "Isometric": "isometrico",
    "Side View": "vista laterale",
    "Perspective": "prospettiva",
    "Vertical": "verticale",
    "Geomorph": "geomorfo",
    "Dungeon Geomorph": "geomorfo",
    "Mini-Map": "mini-mappa",
    "Player Map": "mappa giocatore",
    "Handout": "handout",
    "Index Card": "scheda",
    "Five Room Dungeon": "dungeon cinque stanze",
    "Horror": "horror",
    "Weird": "bizzarro",
    "Weird Fantasy": "fantasy bizzarro",
    "Post Apocalyptic": "post-apocalittico",
    "Post-Apocalyptic": "post-apocalittico",
    "Science Fiction": "fantascienza",
    "Science Fantasy": "fantasy scientifico",
    "Halloween": "halloween",
}

# Tags to skip entirely (game systems, meta tags, series names)
SKIP_TAGS = {
    "dungeons-and-dragon", "Release the Kraken", "Commercial Map",
    "DnD5e", "D&D5e", "Dungeons & Dragons 5e", "Dungeons And Dragon",
    "AD&D", "Advanced Dungeons & Dragon",
    "Warhammer", "Warhammer Fantasy", "Warhammer 40k", "WFRP1e", "WHFRP",
    "Gamma World", "Mutant Future", "Traveller", "Top Secret", "White Star",
    "Empire of the Petal Throne", "Tekumel", "EPT",
    "Eberron", "Sharn", "Waterdeep", "Waterdeep Dragon Heist", "Dragon Heist",
    "Ghosts of Saltmarsh", "Saltmarsh",
    "Mythic Odysseys of Thero", "Thero",
    "Descent into the Depths of the Earth", "Dwellers of the Forbidden City",
    "Expedition to the Barrier Peak",
    "Kabuki Kaiser", "Mad Monks of Kwantoom",
    "Neoclassical Geek Revival", "NGR", "GOZR",
    "The Fantasy Trip", "SJGame", "Steve Jackson Game",
    "Four Against Darkne", "4AD",
    "Ruins of the Undercity", "DungeonMorph Dice",
    "temple of elemental evil", "Dungeoneer's Survival Guide", "DSG",
    "Solo Gaming", "Review", "Video", "writing", "fiction",
    "Athena", "Greek", "Roman",
    "A New Campaign", "Fallout", "Star War", "Heavy Metal",
    "Business Card", "Body Part", "Pineal Gland", "Silly",
    "Valentine", "Hallucination", "Casino",
    "For He Is The Kumquat Häagen-Daz",
    "5RD", "Fantasy Map", "Map", "Lost Map",
    "Mapvember", "Geomorph Mapping Challenge",
    "Nuked", "Nuclear Storage Facility", "Future",
    "Espionage", "Fortune Teller", "Smith", "Herbalism",
    # Series/location names - keep as proper nouns don't translate well
    "Heart of Darkling", "Darklingtown", "Caves of Chao",
    "Realms of Crawling Chao", "Copper Sea", "Sabre Lake",
    "Eleint Passage", "Prince's Harbour", "Principalities of Black Sphinx Bay",
    "Claw of Sunset", "Iseldec's Drop", "the Absent City",
    "The Finger", "The Wicker Man", "Borderlands 3",
    "The Autumn Land", "Temple of the Worm", "Borderland",
    "Thieves' World", "Lost City",
    # Index card series
    "Index Card Dungeon", "Index Card Dungeon II",
    "Small Shops Serie",
    # Lovecraft
    "Lovecraft", "Tsathoggua", "Yig",
}


def translate_tags(english_tags: list[str]) -> list[str]:
    """Translate English tags to Italian, skipping irrelevant ones."""
    italian_tags = set()
    for tag in english_tags:
        if tag in SKIP_TAGS:
            continue
        if tag in TAG_MAP:
            italian_tags.add(TAG_MAP[tag])
        # Skip unknown tags silently - they're usually game-system-specific
    return sorted(italian_tags)


# ── Main pipeline ─────────────────────────────────────────────────────────────

def curate_map(raw: dict) -> dict:
    """Transform a raw scraped map entry into a curated Italian entry."""
    english_name = raw["name"]
    italian_name = translate_name(english_name)
    slug = slugify(italian_name)

    # Derive image filename from slug
    image_filename = f"{slug}.png"

    return {
        "slug": slug,
        "nome": italian_name,
        "nome_originale": english_name,
        "immagine": image_filename,
        "categoria": assign_category(raw["tags"]),
        "tag": translate_tags(raw["tags"]),
        "autore": "Dyson Logos",
        "licenza": "Commercial Use Allowed",
        "url_originale": raw["source_url"],
        "immagine_url": raw["image_url"],  # Keep for reference during image download
    }


def main():
    parser = argparse.ArgumentParser(description="Curate raw map data into Italian mappe.json")
    parser.add_argument("--input", "-i", default="dyson_raw.json", help="Input raw JSON file")
    parser.add_argument("--output", "-o", default="mappe_curated.json", help="Output curated JSON file")
    args = parser.parse_args()

    with open(args.input, encoding="utf-8") as f:
        raw_maps = json.load(f)

    curated = []
    seen_slugs = set()
    for raw in raw_maps:
        entry = curate_map(raw)

        # Deduplicate slugs
        base_slug = entry["slug"]
        if base_slug in seen_slugs:
            counter = 2
            while f"{base_slug}-{counter}" in seen_slugs:
                counter += 1
            entry["slug"] = f"{base_slug}-{counter}"
        seen_slugs.add(entry["slug"])

        curated.append(entry)

    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(curated, f, indent=2, ensure_ascii=False)
        f.write("\n")

    # Print stats
    from collections import Counter
    cat_counts = Counter(m["categoria"] for m in curated)
    print(f"Curated {len(curated)} maps")
    print(f"\nCategories:")
    for cat, count in cat_counts.most_common():
        print(f"  {cat:20s} {count}")

    tag_counts = Counter()
    for m in curated:
        tag_counts.update(m["tag"])
    print(f"\nTop 30 tags:")
    for tag, count in tag_counts.most_common(30):
        print(f"  {tag:30s} {count}")

    no_tags = sum(1 for m in curated if not m["tag"])
    print(f"\nMaps with no tags: {no_tags}")
    no_img = sum(1 for m in curated if not m["immagine_url"])
    print(f"Maps with no image URL: {no_img}")


if __name__ == "__main__":
    main()
