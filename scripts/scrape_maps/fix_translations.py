#!/usr/bin/env python3
"""Fix Italian translations in curated map data.

Applies manual name overrides with proper Italian grammar,
then regenerates slugs.

Usage:
    python scripts/scrape_maps/fix_translations.py
"""

import json
import re
import unicodedata

# Complete manual translations: english name → Italian name
# Proper nouns (character/place names) are kept as-is
# Series identifiers kept in original form for consistency
NAME_OVERRIDES = {
    # 0-14: Recent maps + Index Card Dungeon II series
    "The Dreadwarren": "Il Dreadwarren",
    "Hexatomb": "Hexatomb",
    "Shrine on the Mosswater": "Santuario sul Mosswater",
    "Index Card Dungeon II - Map 12": "Dungeon su Scheda II - Mappa 12",
    "Index Card Dungeon II - Map 11": "Dungeon su Scheda II - Mappa 11",
    "Index Card Dungeon II - Map 10": "Dungeon su Scheda II - Mappa 10",
    "Index Card Dungeon II - Map 9 - Ancient Sewers": "Dungeon su Scheda II - Mappa 9 - Fogne Antiche",
    "Index Card Dungeon II - Map 8 - Linking Caverns": "Dungeon su Scheda II - Mappa 8 - Caverne Comunicanti",
    "Index Card Dungeon II – Map 7 – North Sub-Dungeons": "Dungeon su Scheda II - Mappa 7 - Sotterranei Nord",
    "Index Card Dungeon II - Map 6 - Dungeon Ruins": "Dungeon su Scheda II - Mappa 6 - Rovine del Dungeon",
    "Index Card Dungeon II - Map 5 - Tower Dungeons": "Dungeon su Scheda II - Mappa 5 - Sotterranei della Torre",
    "Index Card Dungeon II - Map 4 - Tower Ruins": "Dungeon su Scheda II - Mappa 4 - Rovine della Torre",
    "Index Card Dungeon II - Map 3 - The Ghost Tower": "Dungeon su Scheda II - Mappa 3 - La Torre Fantasma",
    "Index Card Dungeon II – Map 2 – Northern Ruins": "Dungeon su Scheda II - Mappa 2 - Rovine Settentrionali",
    "Index Card Dungeon II - Map 1 - Tower Base": "Dungeon su Scheda II - Mappa 1 - Base della Torre",

    # 15-29
    "The Village of Millbrook": "Il Villaggio di Millbrook",
    "Building 13 - Netmaker": "Edificio 13 - Il Retaio",
    "Building 12 - General Goods Store": "Edificio 12 - Emporio",
    "The Finger": "Il Dito",
    "Beneath the Finger": "Sotto il Dito",
    "Deeper Beneath the Finger": "Ancora Piu in Profondita sotto il Dito",
    "The deeps far below \"The Finger\"": "Le Profondita sotto \"Il Dito\"",
    "A Dungeon In Two Parts": "Un Sotterraneo in Due Parti",
    "Eastmeadow Manor": "Maniero di Eastmeadow",
    "Under Eastmeadow Manor": "Sotto il Maniero di Eastmeadow",
    "Allerton Hold": "La Fortezza di Allerton",
    "Crumbling Gate": "La Porta Fatiscente",
    "The Fallen Temple": "Il Tempio Caduto",
    "Beneath the Black Cairn": "Sotto il Tumulo Nero",
    "The Pipeworks": "Le Tubature",

    # 30-49
    "The Sinkhole Crypt": "La Cripta della Voragine",
    "Building 11 - Carpenter": "Edificio 11 - Il Carpentiere",
    "The Greywater Temple": "Il Tempio di Greywater",
    "Griffon Tower Dungeons": "Sotterranei della Torre del Grifone",
    "Griffon Tower": "La Torre del Grifone",
    "Ruins of the Waste Treatment Facility": "Rovine dell'Impianto di Trattamento",
    "Building 10 - Wit & Wisdom Apothecary": "Edificio 10 - Speziale Wit & Wisdom",
    "Crypts and Worms": "Cripte e Vermi",
    "The Cinereous Basilica": "La Basilica Cinerea",
    "Building 9 - The Riddle of Steel": "Edificio 9 - L'Enigma dell'Acciaio",
    "The Autumn Lands - Map I": "Le Terre d'Autunno - Mappa I",
    "The Autumn Lands - Map H": "Le Terre d'Autunno - Mappa H",
    "The Autumn Lands - Map G": "Le Terre d'Autunno - Mappa G",
    "The Autumn Lands - Hex Map F": "Le Terre d'Autunno - Mappa Esagonale F",
    "The Autumn Lands - Hex Map E": "Le Terre d'Autunno - Mappa Esagonale E",
    "The Autumn Lands Map 4": "Le Terre d'Autunno - Mappa 4",
    "The Autumn Lands - Map C": "Le Terre d'Autunno - Mappa C",
    "The Autumn Lands - Map B": "Le Terre d'Autunno - Mappa B",
    "The Autumn Lands - Map A": "Le Terre d'Autunno - Mappa A",
    "Temple Of The Worm - Upper Levels": "Tempio del Verme - Livelli Superiori",

    # 50-69
    "Temple of the Worm": "Tempio del Verme",
    "Building 8 - Leather Worker": "Edificio 8 - Il Conciatore",
    "Iseldec's Drop - Levels 20-23": "La Caduta di Iseldec - Livelli 20-23",
    "Brezchyn's Halls": "Le Sale di Brezchyn",
    "Building 7 - Random Market Tavern": "Edificio 7 - Taverna del Mercato",
    "Building 6 - Wine Garden": "Edificio 6 - Il Giardino dei Vini",
    "Iseldec's Drop - Levels 17-19": "La Caduta di Iseldec - Livelli 17-19",
    "Iseldec's Drop - Levels 13-16": "La Caduta di Iseldec - Livelli 13-16",
    "Building 5 - Loans / Bank": "Edificio 5 - Prestiti / Banca",
    "Shop 4 - The Fruit Shop": "Bottega 4 - Il Fruttivendolo",
    "Shop #3 - Carpets": "Bottega 3 - Tappeti",
    "Iseldec's Drop (Levels 9-12)": "La Caduta di Iseldec (Livelli 9-12)",
    "Iseldec\u2019s Drop (Levels 9-12)": "La Caduta di Iseldec (Livelli 9-12)",
    "Stonewill Estate": "Tenuta di Stonewill",
    "The Golden Fishmarket": "Il Mercato del Pesce Dorato",
    "Rok's Pottery Workshop": "Il Laboratorio di Ceramiche di Rok",
    "Grudge Match at the Underground Market": "Resa dei Conti al Mercato Sotterraneo",
    "Menrina's Library & The Salty Tavern": "La Biblioteca di Menrina e la Taverna Salata",
    "Dungeon of the Bad Egg": "Il Sotterraneo dell'Uovo Marcio",
    "Iseldec's Drop (Levels 5-8)": "La Caduta di Iseldec (Livelli 5-8)",
    "Ierades's Isle": "L'Isola di Ierades",

    # 70-89
    "Goretooth's Grotto": "La Grotta di Goretooth",
    "Sister's Ford": "Il Guado delle Sorelle",
    "Iseldec's Drop (Levels 1-4)": "La Caduta di Iseldec (Livelli 1-4)",
    "The Tower of Mourning": "La Torre del Lutto",
    "The Temple of Love (2024 Remix)": "Il Tempio dell'Amore (2024 Remix)",
    "The Sleeping Goat Inn": "Locanda della Capra Dormiente",
    "Temple of Lost Ormus": "Tempio del Perduto Ormus",
    "The Lame Barghest Tavern": "La Taverna del Barghest Zoppo",
    "The Lost Watch": "La Vedetta Perduta",
    "Gath-Am's Beacon": "Il Faro di Gath-Am",
    "The Strangled Imp": "Il Diavoletto Strangolato",
    "The Sordid Rhinoceros": "Il Rinoceronte Sordido",
    "Gnarsh's Domain": "Il Dominio di Gnarsh",
    "The House of Kalaxar (Linear Dungeon Experiment 2)": "La Casa di Kalaxar (Esperimento Lineare 2)",
    "Pit of the Sand Wraiths (Linear Dungeon Experiment 1)": "La Fossa degli Spettri di Sabbia (Esperimento Lineare 1)",
    "Edge Strip Dungeon": "Sotterraneo a Striscia",
    "Miserth Keep - Main Level": "Fortino di Miserth - Livello Principale",
    "Miserth Keep – Upper Levels": "Fortino di Miserth - Livelli Superiori",
    "Miserth Keep – Dungeons": "Fortino di Miserth - Sotterranei",
    "The Black Skulls Tomb": "La Tomba dei Teschi Neri",

    # 90-109
    "Archon's Tower": "La Torre dell'Arconte",
    "Beneath the Archon's Tower": "Sotto la Torre dell'Arconte",
    "Desert ClanHold": "Fortezza del Clan del Deserto",
    "The Deep Sepulchre": "Il Sepolcro Profondo",
    "Temple of the Divinity in Copper": "Tempio della Divinita in Rame",
    "Pitmann Manse": "La Dimora di Pitmann",
    "The Halls of Lost Heroes": "Le Sale degli Eroi Perduti",
    "Bolukbasti Grotto": "La Grotta di Bolukbasti",
    "Drurdelm Tombs": "Le Tombe di Drurdelm",
    "Greyrock Tower": "La Torre di Greyrock",
    "Statues of the Thrall Gods": "Le Statue degli Dei Schiavi",
    "Flooded Catacombs (1200 dpi)": "Catacombe Allagate",
    "Sanctuary of the Magi (1200 dpi)": "Santuario dei Magi",
    "Windsong Hall": "La Sala di Windsong",
    "Willowstone Hall": "La Sala di Willowstone",
    "The Garden Tower": "La Torre del Giardino",
    "Gladhold Estate": "Tenuta di Gladhold",
    "The Rumbledown Ruins": "Le Rovine Rumbledown",
    "Ency Glowlands - Map 2": "Le Terre Luminose di Ency - Mappa 2",
    "Gascon's Pit": "La Fossa di Gascon",

    # 110-129
    "The Stony Shore - Map 1": "La Riva Rocciosa - Mappa 1",
    "The Stony Shore - Map 2": "La Riva Rocciosa - Mappa 2",
    "The Stony Shore – Map 3": "La Riva Rocciosa - Mappa 3",
    "The Delren Street Sewers": "Le Fogne di Via Delren",
    "The Whispering Outpost (B&W)": "L'Avamposto Sussurrante",
    "Bore Facility P44": "Impianto di Trivellazione P44",
    "Zinik's Stones": "Le Pietre di Zinik",
    "Beneath the Claw of Sunsets - Map 3": "Sotto l'Artiglio dei Tramonti - Mappa 3",
    "Ruins of the Claw of Sunsets - Map 3": "Rovine dell'Artiglio dei Tramonti - Mappa 3",
    "Proving Grounds of the Mad Ogre Lord": "Campo di Prova del Signore Ogre Folle",
    "The Yellow Fane in Ruins (1200dpi)": "Il Tempio Giallo in Rovine",
    "Under the Lighthouse": "Sotto il Faro",
    "The Dry Lighthouse": "Il Faro Secco",
    "Beneath the Claw of Sunsets – Map 2": "Sotto l'Artiglio dei Tramonti - Mappa 2",
    "Ruins of the Claw of Sunsets – Map 2": "Rovine dell'Artiglio dei Tramonti - Mappa 2",
    "Sanctum Cay (B&W)": "L'Isolotto Sacro",
    "The Long Halls of Taqash Thesk": "Le Lunghe Sale di Taqash Thesk",
    "Science Tower": "La Torre della Scienza",
    "Opal's Steps - Stepped Pyramid": "I Gradini di Opal - Piramide a Gradoni",
    "Beneath the Claw of Sunsets – Map 1": "Sotto l'Artiglio dei Tramonti - Mappa 1",

    # 130-149
    "Ruins of the Claw of Sunsets - Map 1": "Rovine dell'Artiglio dei Tramonti - Mappa 1",
    "Temple of the Communion of Zeviax": "Tempio della Comunione di Zeviax",
    "Dungeoneer Survival Guide Projection 1": "Guida alla Sopravvivenza - Proiezione 1",
    "Dungeons of the Twin Demons": "Sotterranei dei Demoni Gemelli",
    "Gateway of the Twin Demons": "Il Portale dei Demoni Gemelli",
    "The Sirens' Grotto": "La Grotta delle Sirene",
    "The Maw": "Le Fauci",
    "Pits of the Black Moon (B&W)": "Le Fosse della Luna Nera",
    "Lost Tomb Complex": "Complesso Tombale Perduto",
    "The Ioun Tower": "La Torre di Ioun",
    "Hnálla's Tower of Pillars": "La Torre dei Pilastri di Hnálla",
    "The Grotto Beneath Lazuni Hill": "La Grotta sotto la Collina Lazuni",
    "The Shrine atop Lazuni Hill": "Il Santuario sulla Collina Lazuni",
    "Ephara's Alithinero (1200 dpi)": "L'Alithinero di Ephara",
    "Violet Chambers atop Kákri Midállu (1200 dpi)": "Le Camere Viola sopra Kákri Midállu",
    "The Eleint Passages – Ruined Fane": "I Passaggi di Eleint - Tempio in Rovina",
    "The Eleint Passages - The Tower": "I Passaggi di Eleint - La Torre",
    "The Eleint Passages - South": "I Passaggi di Eleint - Sud",
    "The Eleint Passages - North": "I Passaggi di Eleint - Nord",
    "The Basilisk's Caves": "Le Caverne del Basilisco",

    # 150-169
    "Cult Basement": "Il Seminterrato del Culto",
    "Chambers of the Copper Skulls": "Le Camere dei Teschi di Rame",
    "Ruins of the Hill Fort": "Rovine del Forte sulla Collina",
    "Temple Cave of the Ruinous Ministers": "La Grotta-Tempio dei Ministri Rovinosi",
    "Vund Home": "La Dimora di Vund",
    "The Ency Glowlands Map 1": "Le Terre Luminose di Ency - Mappa 1",
    "Dungeon of Holding": "Il Sotterraneo della Custodia",
    "Sharn Skybridges III": "Ponti Celesti di Sharn III",
    "Dead Goblin Hole": "La Tana del Goblin Morto",
    "Bradrig's Hall": "La Sala di Bradrig",
    "Ruined Cities of Yoth": "Le Citta in Rovina di Yoth",
    "Smallharbour": "Porto Piccolo",
    "Hunter's Descent": "La Discesa del Cacciatore",
    "The Idol Pit": "La Fossa dell'Idolo",
    "Sanctuary of Xeeus": "Santuario di Xeeus",
    "Beneath the Temperance Stone": "Sotto la Pietra della Temperanza",
    "Tombs of the Silver Army": "Le Tombe dell'Esercito d'Argento",
    "Garnet's Spring": "La Sorgente di Garnet",
    "A Dungeon of Impossible Stairs (1200 dpi)": "Un Sotterraneo dalle Scale Impossibili",
    "Gavreth's Stones": "Le Pietre di Gavreth",

    # 170-189
    "Veghul's Drop": "Il Precipizio di Veghul",
    "Hardbrand Tower": "La Torre di Hardbrand",
    "The Last Stones of Sagemane Castle": "Le Ultime Pietre del Castello di Sagemane",
    "The Jade Catacombs": "Le Catacombe di Giada",
    "The Bloodied Stones - a druidic circle": "Le Pietre Insanguinate - Circolo Druidico",
    "Tombs of the Throl Tribe": "Le Tombe della Tribu Throl",
    "The Andesite Pyramid": "La Piramide di Andesite",
    "The Red Temple under Béy Sǘ": "Il Tempio Rosso sotto Béy Sǘ",
    "Barrow of the Flayer": "Il Tumulo dello Scorticatore",
    "Release The Kraken On The Warkings Vault": "La Volta dei Warkings",
    "Release The Kraken On The Mastervale Estate": "Tenuta di Mastervale",
    "The Village of Winten Ford": "Il Villaggio di Winten Ford",
    "The Maze of Yivh'Kthaloth": "Il Labirinto di Yivh'Kthaloth",
    "Dohrlegond's Tombs": "Le Tombe di Dohrlegond",
    "Passages Beneath": "I Passaggi Sottostanti",
    "Screaming Hall of the Ur-Goblin": "La Sala Urlante dell'Ur-Goblin",
    "Gaerborin Keep": "Il Fortino di Gaerborin",
    "Raining Cave": "La Caverna della Pioggia",
    "Blackstone Bastion": "Il Bastione di Pietranera",
    "Oreney's Watch": "La Vedetta di Oreney",

    # 190-209
    "Crypt of the Queen of Bones": "La Cripta della Regina delle Ossa",
    "Sumerian Three Story Home": "Dimora Sumera a Tre Piani",
    "Frogsmead Inn & Tavern": "Locanda e Taverna Frogsmead",
    "Page o' Little Ruins": "Pagina di Piccole Rovine",
    "Under the Observatory": "Sotto l'Osservatorio",
    "Unrol Kaz'ad Watch": "La Vedetta di Unrol Kaz'ad",
    "Temple Ruins / Froghemoth Nest": "Rovine del Tempio / Nido del Froghemoth",
    "The Sunken Tower": "La Torre Sommersa",
    "The Goblin Warrens at Fort Redshield": "Le Tane dei Goblin al Forte Redshield",
    "Ruins of Brollmoreth": "Rovine di Brollmoreth",
    "Ruins of the East Gate of Steldin Dorg": "Rovine della Porta Orientale di Steldin Dorg",
    "The Ruins of Shagunat Keep": "Le Rovine del Fortino di Shagunat",
    "Geomorphic Halls Level 1": "Sale Geomorfiche Livello 1",
    "Adventures around Jalovhec BW": "Avventure intorno a Jalovhec",
    "Redford Citadel": "La Cittadella di Redford",
    "Nuzur Hollow": "La Conca di Nuzur",
    "Darklingtown - The Upper Tunnels": "Darklingtown - I Tunnel Superiori",
    "Darklingtown - Cavern & Spillways District": "Darklingtown - Distretto delle Caverne e degli Sfioratori",
    "The Dwarven Shrine at Mount Thorrien": "Il Santuario Nanico al Monte Thorrien",
    "Ruldroc Castle": "Il Castello di Ruldroc",

    # 210-229
    "Oceanwatch": "La Vedetta sull'Oceano",
    "Darklingtown - The Tunnels District": "Darklingtown - Distretto dei Tunnel",
    "Darklingtown - Frog Tower": "Darklingtown - La Torre della Rana",
    "Darklingtown - The Mushroom Caves": "Darklingtown - Le Caverne dei Funghi",
    "Sharn Heights - Skybridge Nexus": "Sharn Heights - Nodo dei Ponti Celesti",
    "Krelava Manor": "Il Maniero di Krelava",
    "Daglan's Cave": "La Caverna di Daglan",
    "Sewer Complex": "Complesso Fognario",
    "Darklingtown - East Frogsport": "Darklingtown - Frogsport Est",
    "Darklingtown - Frogsport": "Darklingtown - Frogsport",
    "Bastion of the Prince of Clubs": "Il Bastione del Principe di Bastoni",
    "Kamrorth's Cairn": "Il Tumulo di Kamrorth",
    "Pirate Booty Island": "L'Isola del Tesoro dei Pirati",
    "Darklingtown South": "Darklingtown Sud",
    "Darklingtown (North District)": "Darklingtown (Distretto Nord)",
    "Imp Tower": "La Torre del Diavoletto",
    "Creeping Sands": "Le Sabbie Striscianti",
    "Business Cards 2021 - Set 1": "Geomorfi su Biglietto - Set 1",
    "2021 Business Card Geomorphs - Set 2 Promotional": "Geomorfi su Biglietto - Set 2",
    "The Goblin Vault": "La Volta dei Goblin",

    # 230-249
    "Heart of Darkling - The Stairs": "Cuore di Darkling - Le Scale",
    "The Champion's Retreat": "Il Rifugio del Campione",
    "Mayer's Fort": "Il Forte di Mayer",
    "Nephilim's Hall": "La Sala del Nephilim",
    "Skaldon's Dome": "La Cupola di Skaldon",
    "Pentagonal Monuments in the Ghost Dunes": "Monumenti Pentagonali nelle Dune Fantasma",
    "Heart of Darkling - The Pillars": "Cuore di Darkling - I Pilastri",
    "Greyfalls Cave": "La Caverna di Greyfalls",
    "Catacombs of the Flayed Minotaur": "Catacombe del Minotauro Scorticato",
    "The Dragon Shrine (1200 dpi)": "Il Santuario del Drago",
    "The Tyrant's Ruins": "Le Rovine del Tiranno",
    "The Ritual Pool": "La Pozza del Rituale",
    "Heart of Darkling - Deception's Bridge": "Cuore di Darkling - Il Ponte dell'Inganno",
    "Collapsed Tomb of Mosogret (1200 dpi)": "La Tomba Crollata di Mosogret",
    "Lords of the Aldeiron Peaks (1200 dpi)": "I Signori delle Vette di Aldeiron",
    "Undercrofts (1200 dpi)": "Le Cripte Sotterranee",
    "The Fallen House of Githaleon (1200 dpi)": "La Casa Caduta di Githaleon",
    "Heart of Darkling - The Cold Caves": "Cuore di Darkling - Le Caverne Fredde",
    "Heart of Darkling - The Drop": "Cuore di Darkling - Il Precipizio",
    "Red Talon's Lair": "La Tana dell'Artiglio Rosso",

    # 250-269
    "The Granite Shore (1200 dpi)": "La Riva di Granito",
    "The Dragon Temple (1200 dpi)": "Il Tempio del Drago",
    "The Four-Toothed Drunk (1200 dpi)": "L'Ubriaco dai Quattro Denti",
    "Heart of Darkling - Gibberling Lake (1200 dpi)": "Cuore di Darkling - Il Lago dei Gibberling",
    "Heart of Darkling - The Darkling Galleries": "Cuore di Darkling - Le Gallerie Darkling",
    "Lost City of the Naga Queens": "La Citta Perduta delle Regine Naga",
    "Island Tomb": "La Tomba sull'Isola",
    "The Bottomless Tombs - Part 4 (1200 dpi)": "Le Tombe Senza Fondo - Parte 4",
    "Lost Reliquary (1200 dpi)": "Il Reliquiario Perduto",
    "Landing Facility (with grids)": "Struttura di Atterraggio",
    "Temple of the 4 Gods (ground level)": "Tempio dei 4 Dei (Piano Terra)",
    "Tarsakh Manor - Upper Floors (1200 dpi)": "Maniero di Tarsakh - Piani Superiori",
    "Tarsakh Manor Grounds (1200 dpi)": "I Terreni del Maniero di Tarsakh",
    "Tarsakh Village (1200 dpi)": "Il Villaggio di Tarsakh",
    "Forlorn Halls of the Mongrelfolk (1200 dpi)": "Le Sale Desolate dei Mongrelfolk",
    "Twilight Descent (1200 dpi)": "La Discesa del Crepuscolo",
    "Return to Durahn's Tomb": "Ritorno alla Tomba di Durahn",
    "Workshop and Store": "Laboratorio e Negozio",
    "Tunnels of the Shrouded Emperor (1200 dpi)": "I Tunnel dell'Imperatore Velato",
    "Well Gurath (1200 dpi)": "Il Pozzo di Gurath",

    # 270-289
    "The Old Throne (1200 dpi)": "Il Vecchio Trono",
    "Skull Maze (1200 dpi)": "Il Labirinto del Teschio",
    "Forbidden Halls": "Le Sale Proibite",
    "Alturiak Manor (1200 dpi)": "Maniero di Alturiak",
    "Dungeons of the Iron Star - Level 2 (1200 dpi)": "Sotterranei della Stella di Ferro - Livello 2",
    "Borderlands Caves - Level 2 East - (1200 dpi)": "Caverne delle Terre di Confine - Livello 2 Est",
    "Quellport & the Isle of Seven Bees (1200 dpi)": "Quellport e l'Isola delle Sette Api",
    "Sietch of Morning (1200 dpi)": "Il Sietch del Mattino",
    "Abbey of the Iron Star (1200 dpi)": "Abbazia della Stella di Ferro",
    "Dungeons of the Iron Star (1200 dpi)": "Sotterranei della Stella di Ferro",
    "Dungeons of the Grand Illusionist (1200 dpi)": "Sotterranei del Grande Illusionista",
    "The Ruined Keep Of Madrual": "Il Fortino in Rovina di Madrual",
    "Walled Temple": "Il Tempio Murato",
    "The Citadel at Sabre Lake": "La Cittadella sul Lago Sabre",
    "Pharyka's Walk": "Il Sentiero di Pharyka",
    "The ruins of Boar Isle Tower (1200 dpi)": "Le Rovine della Torre sull'Isola del Cinghiale",
    "There's a Hole in the Dungeon (1200 dpi)": "C'e un Buco nel Sotterraneo",
    "Return to Appletree Pond": "Ritorno allo Stagno del Melo",
    "the Demon Pillars of Iv": "I Pilastri Demoniaci di Iv",
    "Cooper's Hole (1200 dpi)": "La Buca di Cooper",

    # 290-309
    "Lautuntown": "Lautuntown",
    "Shrine of the Demon Eskarna (1200 dpi)": "Santuario del Demone Eskarna",
    "Borderlands Caves (1200 dpi)": "Caverne delle Terre di Confine",
    "Somerrich Cays": "Gli Isolotti di Somerrich",
    "The False Tombs (1200 dpi)": "Le False Tombe",
    "Release The Kraken On Axebridge Over Blackbay": "Il Ponte dell'Ascia su Blackbay",
    "Sanctum of the Blind Protean": "Il Sanctum del Proteo Cieco",
    "Bloodmarket Cave (1200 dpi)": "La Caverna del Mercato di Sangue",
    "Last Home of the Three Heretics of Xaeen": "L'Ultima Dimora dei Tre Eretici di Xaeen",
    "Bitterchains Tombs": "Le Tombe di Bitterchains",
    "Barrow Mounds of the Lich and Famous": "I Tumuli del Lich e dei Famosi",
    "Lair of the Golden Wolf": "La Tana del Lupo Dorato",
    "Foxtail Grotto": "La Grotta di Codadivolpe",
    "Dread Shrine of the Magi in Sapphire": "Il Santuario Terribile dei Magi in Zaffiro",
    "Pentagon Cove": "La Baia del Pentagono",
    "The Goat Shrine": "Il Santuario della Capra",
    "Tombs of the Steel Makers": "Le Tombe dei Forgiatori d'Acciaio",
    "Crowspine Labyrinth": "Il Labirinto di Crowspine",
    "Sanvild's Delve": "Lo Scavo di Sanvild",
    "The Raven's Rum & Roosts": "Il Rum e i Posatoi del Corvo",

    # 310-329
    "Isometric Tomb of Illhan the Binder": "Tomba Isometrica di Illhan il Legatore",
    "Crowned Hill": "La Collina Incoronata",
    "Four Pagodas of Kwantoom": "Le Quattro Pagode di Kwantoom",
    "Springhollow": "Vallesorgente",
    "The Bottomless Tombs - Level 3": "Le Tombe Senza Fondo - Livello 3",
    "Moonset Street Shops": "I Negozi di Via Moonset",
    "The Warlock's Arms": "Le Armi del Warlock",
    "Gladiator's Temple D&D Map": "Il Tempio del Gladiatore",
    "Vigilance Trail": "Il Sentiero della Vigilanza",
    "The Infested Hall": "La Sala Infestata",
    "The Ruins near Elverston Hold": "Le Rovine presso la Fortezza di Elverston",
    "Principalities of Black Sphinx Bay": "Principati della Baia della Sfinge Nera",
    "Guddur's Wicked Teahouse": "La Casa del Te Malvagia di Guddur",
    "Spell-Eater's Spring": "La Sorgente del Divoraincantesimi",
    "Drowning Point": "Il Punto dell'Annegamento",
    "The Juicer": "Lo Spremitore",
    "The Old Fort": "Il Vecchio Forte",
    "The Old Fort Ruins & Dungeon": "Rovine e Sotterranei del Vecchio Forte",
    "The Twin Norkers (tagged)": "I Norker Gemelli",
    "Seaside Passage": "Il Passaggio Costiero",

    # 330-349
    "Seever's Mill": "Il Mulino di Seever",
    "The Holy City of Guerras-El-Estat": "La Citta Santa di Guerras-El-Estat",
    "Bloodied Warrens": "Le Tane Insanguinate",
    "Beneath the Marching Tankard": "Sotto il Boccale in Marcia",
    "The Marching Tankard - Upstairs - Tagged": "Il Boccale in Marcia - Piano Superiore",
    "The Vault of Dahlver-Nar": "La Volta di Dahlver-Nar",
    "Barrow Mounds of the Lich & Famous III": "I Tumuli del Lich e dei Famosi III",
    "The Cinder Throne": "Il Trono di Cenere",
    "The Hydra's Alehouse": "La Birreria dell'Idra",
    "Church of the Oracles in Onyx": "La Chiesa degli Oracoli in Onice",
    "The Lost Temple of Aphosh the Haunted": "Il Tempio Perduto di Aphosh l'Infestato",
    "Bloodied Axe Shrine": "Il Santuario dell'Ascia Insanguinata",
    "Jacob's Spur - colour 300dpi": "Lo Sperone di Jacob",
    "The Dungeon in 12 Parts": "Il Sotterraneo in 12 Parti",
    "The Bast Inn": "La Locanda di Bast",
    "Royal Catacombs of Adrih": "Le Catacombe Reali di Adrih",
    "Pillar of the Igesej Loremaster": "Il Pilastro del Maestro del Sapere Igesej",
    "Kabrel's Tower": "La Torre di Kabrel",
    "Uogralas, City of the Frogs": "Uogralas, Citta delle Rane",
    "Hurren, City of the Elders": "Hurren, Citta degli Anziani",

    # 350-369
    "The Blind Lamia's Cave": "La Caverna della Lamia Cieca",
    "Iyesgarten Regional Map": "Mappa Regionale di Iyesgarten",
    "The Iyesgarten Inn": "La Locanda di Iyesgarten",
    "The Village of Iyesgarten": "Il Villaggio di Iyesgarten",
    "Ssa-Tun's Lake of Milk": "Il Lago di Latte di Ssa-Tun",
    "Spectre's Tower": "La Torre dello Spettro",
    "The Temple Walk": "La Passeggiata del Tempio",
    "Chambers of the Absent City": "Le Camere della Citta Assente",
    "Shrine of the Emperor of Bones": "Santuario dell'Imperatore delle Ossa",
    "The Behkon Inn": "La Locanda di Behkon",
    "Sewer Connectors": "Raccordi Fognari",
    "Lady White's Ruins": "Le Rovine di Lady White",
    "Isle of Kheyus (Colour)": "L'Isola di Kheyus",
    "Summerthorpe": "Summerthorpe",
    "Labhruinn's Tavern": "La Taverna di Labhruinn",
    "An-Nayyir's Pyramid": "La Piramide di An-Nayyir",
    "Defiled Waters": "Le Acque Profanate",
    "Cage Street Sewers": "Le Fogne di Via Cage",
    "The Dark Caverns of Turr": "Le Caverne Oscure di Turr",
    "Sewer Elements": "Elementi Fognari",

    # 370-389
    "Temple Crypts of the Wraith Priests": "Le Cripte del Tempio dei Sacerdoti Spettrali",
    "the Stone Troll's Lantern": "La Lanterna del Troll di Pietra",
    "West Sewers Complex": "Complesso Fognario Ovest",
    "Psychedelic Cellar of the Stone Giants": "La Cantina Psichedelica dei Giganti di Pietra",
    "South Sewers Hideout": "Il Nascondiglio delle Fogne Sud",
    "Smugglers-Lodge": "Il Rifugio dei Contrabbandieri",
    "Gauntlet of the Flintcrowned Ghouls": "Il Guanto di Sfida dei Ghoul dalla Corona di Selce",
    "Shieldrick's Tower Inn": "La Locanda della Torre di Shieldrick",
    "Rosewood Street Sewers": "Le Fogne di Via Rosewood",
    "The Red Descent": "La Discesa Rossa",
    "The Old Turnip Inn": "La Locanda della Vecchia Rapa",
    "The Phoenix Diadem": "Il Diadema della Fenice",
    "Guild Temple": "Il Tempio della Gilda",
    "Crypt of the Smith": "La Cripta del Fabbro",
    "The Ruins of Greymail Clanhold": "Le Rovine della Fortezza del Clan Greymail",
    "The Vanshiro Reliquary": "Il Reliquiario di Vanshiro",
    "Onyx Hill Ruins": "Rovine della Collina di Onice",
    "The Half-Cask Tavern": "La Taverna del Mezzo Barile",
    "The Court of Summer Wines": "La Corte dei Vini Estivi",
    "Letath": "Letath",

    # 390-409
    "Oubliette of the Forgotten Magus": "L'Oubliette del Mago Dimenticato",
    "Ashryn Spire": "La Guglia di Ashryn",
    "The Fire Beetle Ale House": "La Birreria dello Scarabeo di Fuoco",
    "Robrus Beach Cave": "La Caverna della Spiaggia di Robrus",
    "Control": "Controllo",
    "The Blessed Monastery": "Il Monastero Benedetto",
    "The Bubble City of Oublos": "La Citta delle Bolle di Oublos",
    "The Bottomless Tombs - Level 2": "Le Tombe Senza Fondo - Livello 2",
    "Twelve Goats Tavern": "La Taverna dei Dodici Capri",
    "Black Armoury of the Mad King": "L'Armeria Nera del Re Folle",
    "Will-o-the-Wisp": "Fuoco Fatuo",
    "Sanhelter Keep": "Il Fortino di Sanhelter",
    "The Wooden Duck Inn": "La Locanda dell'Anatra di Legno",
    "The Vault of Tranquility": "La Volta della Tranquillita",
    "the Ruins of Charnesse": "Le Rovine di Charnesse",
    "Yruvex Swamps": "Le Paludi di Yruvex",
    "The Banshee's Tower": "La Torre della Banshee",
    "The Tower-Faced Demon": "Il Demone dalla Faccia di Torre",
    "Lesser Temple of the Heretics": "Il Tempio Minore degli Eretici",
    "Temple of the Seven Heretics": "Il Tempio dei Sette Eretici",

    # 410-439
    "Brenovale Castle": "Il Castello di Brenovale",
    "Lorean's Manor": "Il Maniero di Lorean",
    "Paradise Control": "Controllo del Paradiso",
    "The Portal Nexus (with grid)": "Il Nodo dei Portali",
    "Rose Point Manor": "Il Maniero di Rose Point",
    "Beneath Rose Point Manor": "Sotto il Maniero di Rose Point",
    "Vault of the Blue Golem": "La Volta del Golem Blu",
    "Strange Ruins": "Rovine Misteriose",
    "The Many Chambers of Izzet's Folly": "Le Molte Camere della Follia di Izzet",
    "Cliffstable on Kerstal": "Cliffstable su Kerstal",
    "HollowStone Bandit Camp": "L'Accampamento dei Banditi di Hollowstone",
    "The Arcane Waters": "Le Acque Arcane",
    "Brentil Tower": "La Torre di Brentil",
    "Palace of the Sands": "Il Palazzo delle Sabbie",
    "Dungeon of the Third Eye": "Il Sotterraneo del Terzo Occhio",
    "Prince's Harbour - Map 3": "Il Porto del Principe - Mappa 3",
    "Stony Hill": "La Collina Rocciosa",
    "Old Cruik Hollow": "La Vecchia Conca di Cruik",
    "White Crag Fortress": "La Fortezza della Rupe Bianca",
    "Prince's Harbour - Map 2": "Il Porto del Principe - Mappa 2",
    "Greth's Island Keep": "Il Fortino sull'Isola di Greth",
    "Tomb of the Kirin-Born Prince": "La Tomba del Principe Nato dal Kirin",
    "Wyvernseeker Rock": "La Roccia di Wyvernseeker",
    "Prior's Hill": "La Collina del Priore",
    "Prince's Harbour - Map 1": "Il Porto del Principe - Mappa 1",
    "Redrock Cays": "Gli Isolotti di Redrock",
    "Circle Crypts of the Ophidian Emperor": "Le Cripte Circolari dell'Imperatore Serpente",
    "Hevlod Manor": "Il Maniero di Hevlod",
    "Briar Keep 2018 Edition": "Il Fortino di Briar (Edizione 2018)",
    "Drow Spire Fortress": "La Fortezza della Guglia Drow",
}


def slugify(text: str) -> str:
    text = text.lower().strip()
    text = unicodedata.normalize("NFKD", text)
    text = text.encode("ascii", "ignore").decode("ascii")
    text = re.sub(r"[^\w\s-]", "", text)
    text = re.sub(r"[-\s]+", "-", text)
    return text.strip("-")


def main():
    with open("mappe_curated.json", encoding="utf-8") as f:
        maps = json.load(f)

    fixed = 0
    unfixed = []
    for m in maps:
        original = m["nome_originale"]
        if original in NAME_OVERRIDES:
            m["nome"] = NAME_OVERRIDES[original]
            m["slug"] = slugify(m["nome"])
            m["immagine"] = f"{m['slug']}.png"
            fixed += 1
        else:
            unfixed.append(original)

    # Deduplicate slugs
    seen_slugs = {}
    for m in maps:
        base = m["slug"]
        if base in seen_slugs:
            counter = 2
            while f"{base}-{counter}" in seen_slugs:
                counter += 1
            m["slug"] = f"{base}-{counter}"
            m["immagine"] = f"{m['slug']}.png"
        seen_slugs[m["slug"]] = True

    with open("mappe_curated.json", "w", encoding="utf-8") as f:
        json.dump(maps, f, indent=2, ensure_ascii=False)
        f.write("\n")

    print(f"Fixed {fixed}/{len(maps)} translations")
    if unfixed:
        print(f"\nMissing overrides ({len(unfixed)}):")
        for name in unfixed:
            print(f"  - {name}")


if __name__ == "__main__":
    main()
