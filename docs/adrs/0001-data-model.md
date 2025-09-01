# Data Model (SRD)

Status: Accepted

## Context
- I dati dell'SRD sono numerosi ed eterogenei. Serve un modello normalizzato per il visualizzatore.

## Decisione

Dividiamo i dati nelle seguenti collection.

- `documenti`: pagine SRD italiane (Markdown completo, con titolo e numero pagina quando disponibile).

Esempio:

```json
{
  "_id": "...",
  "slug": "note-legali",
  "titolo": "Note Legali",
  "content": "# Note Legali\n...",
  "numero_di_pagina": 1
}
```

- `classi`: una per ogni classe. Alcuni campi sono opzionali e presenti solo dove hanno senso.

```json
{
  "slug": "stregone",
  "nome": "Stregone",
  "dado_vita": "d6",
  "caratteristica_primaria": "Carisma",
  "salvezze_competenze": ["Costituzione", "Carisma"],
  "abilità_competenze_opzioni": {"scegli": 2, "opzioni": ["Arcana", "Inganno", "Intuizione", "Intimidire", "Persuasione", "Religione"]},
  "armi_competenze": ["Armi semplici"],
  "armature_competenze": [],
  "equipaggiamento_iniziale_opzioni": [
    {"etichetta": "Opzione A", "oggetti": ["Lancia", "2 Pugnali", "Focus arcano (cristallo)", "Zaino dell'Esploratore di Dungeon", "28 mo"]},
    {"etichetta": "Opzione B", "oggetti": ["50 mo"]}
  ],
  "tabella_livelli": [
    {
      "livello": 1,
      "bonus_competenza": 2,
      "privilegi_di_classe": ["Lancio di Incantesimi", "Magia Innata"],
      "punti_stregoneria": null,
      "trucchetti_conosciuti": 4,
      "incantesimi_preparati": 2,
      "slot_incantesimo": {"1": 2}
    },
    {
      "livello": 2,
      "bonus_competenza": 2,
      "capacità": ["Fonte di Magia", "Metamagia"],
      "punti_stregoneria": 2,
      "trucchetti_conosciuti": 4,
      "incantesimi_preparati": 4,
      "slot_incantesimo": {"1": 3}
    }
  ],
  "privilegi_di_classe": [
    {"nome": "Lancio di Incantesimi", "livello": 1, "descrizione": "..."},
    {"nome": "Magia Innata", "livello": 1, "descrizione": "..."},
    {"nome": "Fonte di Magia", "livello": 2, "descrizione": "..."}
  ],
  "lanciare_incantesimi": {
    "trucchetti": ["Luce", "Prestidigitazione", "Presa Folgorante", "Scoppio Stregonesco"],
    "lista_incantesimi": {"1": ["Mani Brucianti", "Individuazione del Magico", "Scudo"], "2": ["Invisibilità", "Raggio Rovente"]}
  },
  "sottoclassi": [
    {
      "slug": "stregoneria-draconica",
      "nome": "Stregoneria Draconica",
      "privilegi_sottoclasse": [
        {"nome": "Resilienza Draconica", "livello": 3, "descrizione": "..."},
        {"nome": "Affinità Elementale", "livello": 6, "descrizione": "..."}
      ],
      "incantesimi_aggiuntivi": {"3": ["Alterare Sé Stesso", "Sfera Cromatica"], "5": ["Paura", "Volare"]}
    }
  ]
}
```

Visto che le classi sono diverse, i parser per le classi possono divergere e produrre un superset di campi. I campi non applicabili rimangono assenti.

- `backgrounds`: estratti da `05_origini_personaggio.md`.

```md
#### Accolito

**Punteggi di Caratteristica:** Intelligenza, Saggezza, Carisma
**Talento:** Iniziato alla Magia (Chierico) (vedi “Talenti”)
**Competenze in Abilità:** Intuizione e Religione
**Competenza negli Strumenti:** Strumenti da Calligrafo
**Equipaggiamento:** *Scegli A o B:* (A) ...; oppure (B) ...
```

```json
{
  "punteggi_caratteristica": ["Intelligenza", "Saggezza", "Carisma"],
  "talento": "Iniziato alla Magia (Chierico)",
  "abilità_competenze": ["Intuizione", "Religione"],
  "strumenti_competenze": ["Strumenti da Calligrafo"],
  "equipaggiamento_iniziale_opzioni": [
    {"etichetta": "Opzione A", "oggetti": ["..."]},
    {"etichetta": "Opzione B", "oggetti": ["..."]}
  ]
}
```

- `incantesimi` / `spells_en`: una entry per incantesimo (lista da `16_incantesimi_items.md` / `08_spells_items.md`). Ogni documento contiene un campo strutturato e il markdown originale della sezione per un rendering fedele.

```json
{
  "shared_id": "spell:0001",
  "slug": "acid-arrow",
  "nome": "Acid Arrow",
  "livello": 2,
  "scuola": "Evocation",
  "classi": ["Wizard"],
  "lancio": {"tempo": "Action", "gittata": "90 feet", "componenti": "V, S, M (powdered rhubarb leaf)", "durata": "Instantaneous"},
  "content": "#### Acid Arrow\n*Level 2 Evocation (Wizard)*\n..."
}
```

- `armi` / `weapons_en`: una entry per arma (`09_armi_items.md` / `07_weapons_items.md`).

```json
{
  "shared_id": "weapon:0002",
  "slug": "pugnale",
  "nome": "Pugnale",
  "costo": "2 mo",
  "peso": "0,5 kg",
  "danno": "1d4 Perforante",
  "categoria": "Semplice da Mischia",
  "proprieta": ["Accurata", "Leggera", "Da Lancio"],
  "maestria": "Fendente Rapido",
  "gittata": {"normale": "6 m", "lunga": "18 m"},
  "content": "## Pugnale\n**Costo:** ..."
}
```

- `armature` / `armor_en`: una entry per ogni armatura/scudo (`11_armatura_items.md` / `07_armor_items.md`).

- `strumenti` / `tools_en`: una entry per ogni set di strumenti (`12_strumenti_items.md` / `07_tools_items.md`).

- `servizi` / `services_en`: una entry per ogni servizio (`13_servizi_items.md` / `07_services_items.md`).

- `equipaggiamento` / `adventuring_gear_en`: una entry per ogni oggetto di equipaggiamento (`08_equipaggiamento_items.md` / `07_adventuring_gear.md`).

- `oggetti_magici` / `magic_items_en`: una entry per ogni oggetto magico (sezione A–Z in `10_oggetti_magici_items.md` / `07_magic_items.md`).

```json
{
  "shared_id": "magicitem:0042",
  "slug": "adamantine-armor",
  "nome": "Adamantine Armor",
  "tipo": "Armor (Any Medium or Heavy, Except Hide Armor)",
  "rarita": "Uncommon",
  "sintonizzazione": false,
  "content": "### Adamantine Armor\n*Armor (Any Medium or Heavy, Except Hide Armor), Uncommon*\n..."
}
```

- `mostri` / `monsters_en`: una entry per creatura (sezione A–Z in `20_mostri_items.md` / `13_monsters_items.md`).

```json
{
  "shared_id": "monster:0100",
  "slug": "aboleth",
  "nome": "Aboleth",
  "tag": {"taglia": "Large", "tipo": "Aberration", "allineamento": "Lawful Evil"},
  "ac": 17,
  "hp": "150 (20d10 + 40)",
  "velocita": "10 ft., Swim 40 ft.",
  "caratteristiche": {"str": 21, "dex": 9, "con": 15, "int": 18, "wis": 15, "cha": 18},
  "content": "## Aboleth\n*Large Aberration, Lawful Evil*\n- **Armor Class:** 17\n..."
}
```

- `animali` / `animals_en`: una entry per ogni animale (`21_animali.md` / `14_animals_items.md`). Struttura analoga a `mostri`.

- `talenti` / `feats_en`: una entry per talento (sezioni in `06_talenti.md` / `06_feats.md`).

```json
{
  "shared_id": "feat:0007",
  "slug": "allerta",
  "nome": "Allerta",
  "categoria": "Talento di Origine",
  "prerequisiti": "",
  "benefici": ["Competenza all’Iniziativa", "Scambio di Iniziativa"],
  "content": "#### Allerta\n*Talento di Origine*\n..."
}
```

Identificatore condiviso
- Ogni documento estratto ha un campo `shared_id` con forma `<namespace>:NNNN` (es. `spell:0001`), così da collegare le versioni IT/EN della stessa entità anche se sono in collezioni distinte.
- Il `shared_id` è deterministico all’interno del file sorgente: è basato sull’ordine degli elementi (stabile tra IT/EN nelle fonti SRD) e sul tipo di entità (namespace). Questo evita dipendenze da nomi tradotti.

## Normalizzazione
- Nei testi estratti, i trattini en/em (–, —) sono normalizzati quando necessario.
- Il titolo dei documenti (`titolo`) è l'`H1` se presente; in assenza si usa lo slug derivato dal filename.
 - Ogni entità conserva un campo `content` con il markdown integrale della propria sezione per una resa fedele.
