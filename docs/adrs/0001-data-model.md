# Data Model (SRD)

**Status:** Accepted

## Context

* I dati SRD sono numerosi ed eterogenei. Serve un modello normalizzato per il visualizzatore.

## Decisione

Dividiamo i dati nelle seguenti collection. Ogni entry ha `slug` e `contenuto` (markdown integrale della sezione) quando applicabile.

---

## documenti

### Schema

```json
{
  "pagina": 1,
  "slug": "giocare",
  "titolo": "Giocare",
  "contenuto": "# Giocare\n…"
}
```

---

## classi

### Schema

```json
{
  "slug": "<identificatore_minuscolo>",
  "nome": "<Nome Classe>",
  "sottotitolo": "<tagline breve>",
  "markdown": "<intera classe in markdown>",
  "dado_vita": "d6|d8|d10|d12",
  "caratteristica_primaria": ["Forza","Destrezza","Costituzione","Intelligenza","Saggezza","Carisma"],
  "salvezze_competenze": ["<Car1>","<Car2>"],
  "abilità_competenze_opzioni": { "scegli": 2, "opzioni": ["<Abilità1>", "..."] },
  "armi_competenze": ["<arma o gruppo>", "..."],
  "armature_competenze": ["Leggere","Medie","Pesanti","Scudi"],
  "strumenti_competenze": ["<strumento>", "..."],
  "equipaggiamento_iniziale_opzioni": [
    { "etichetta": "Opzione A", "oggetti": ["<oggetto>", "..."] },
    { "etichetta": "Opzione B", "oggetti": ["<oggetto>", "..."] },
    { "etichetta": "Ricchezza", "oggetti": ["<mo>"] }
  ],
  "multiclasse": {
    "prerequisiti": ["<Car> <valore>", "..."],
    "tratti_acquisiti": ["<dado_vita>", "<competenze armi>", "<armature>", "<strumenti/abilità extra>"],
    "note": "<regole slot o eccezioni>"
  },
  "progressioni": {
    "maestria_armi": { "livelli": { "1": 0 } },
    "stili_combattimento": { "livelli": { }, "scelte": [] },
    "attacchi_extra": { "livelli": { } },
    "risorse": [ { "chiave": "<risorsa>", "livelli": { "1": 2 } } ],
    "aumenti_caratteristica": [4,8,12,16],
    "dono_epico": 19
  },
  "magia": {
    "ha_incantesimi": false,
    "lista_riferimento": "<Classe o Lista>",
    "caratteristica_incantatore": "<Carisma|Saggezza|Intelligenza|null>",
    "preparazione": "prepared|known|none",
    "focus": "<tipo|—>",
    "rituali": "nessuno|solo_lista|da_libro",
    "trucchetti": { },
    "incantesimi_preparati_o_noti": { },
    "slot": { "1": [2,0,0,0,0,0,0,0,0] },
    "patto_warlock": { "slot": {}, "livello_slot": {} },
    "arcani_mistici": { }
  },
  "tabella_livelli": [
    {
      "livello": 1,
      "bonus_competenza": 2,
      "privilegi_di_classe": ["<Priv1>"],
      "risorse": { },
      "trucchetti": 0,
      "incantesimi_preparati": 0,
      "slot": [0,0,0,0,0,0,0,0,0],
      "note": ""
    }
  ],
  "privilegi_di_classe": [
    { "nome": "<Nome Privilegio>", "livello": 1, "descrizione": "<testo breve>" }
  ],
  "sottoclassi": [
    {
      "slug": "<id>",
      "nome": "<Nome Sottoclasse>",
      "descrizione": "<frase tema>",
      "privilegi_sottoclasse": [
        { "nome": "<Privilegio>", "livello": 3, "descrizione": "<testo>" }
      ],
      "incantesimi_sempre_preparati": { "3": ["<inc1>"] }
    }
  ],
  "liste_incantesimi": { "0": ["<Trucchetto>"] },
  "raccomandazioni": {
    "trucchetti_cons": [],
    "incantesimi_iniziali_cons": [],
    "equip_iniziale_cons": "",
    "talenti_cons": [],
    "dono_epico_cons": ""
  },
  "contenuto": "## <Nome Classe>\n…"
}
```

---

## backgrounds

### Schema

```json
{
  "slug": "<id>",
  "nome": "<Nome Origine>",
  "punteggi_caratteristica": ["<Car>", "..."],
  "talento": "<Nome Talento>",
  "abilità_competenze": ["<Abilità>", "..."],
  "strumenti_competenze": ["<Strumento>", "..."],
  "equipaggiamento_iniziale_opzioni": [
    {"etichetta": "Opzione A", "oggetti": ["..."]},
    {"etichetta": "Opzione B", "oggetti": ["..."]}
  ],
  "contenuto": "#### <Nome Origine>\n…"
}
```

---

## specie 

### Schema

```json
{
  "slug": "<id>",
  "nome": "<Nome Origine>",
  "punteggi_caratteristica": ["<Car>", "..."],
  "talento": "<Nome Talento>",
  "abilità_competenze": ["<Abilità>", "..."],
  "strumenti_competenze": ["<Strumento>", "..."],
  "equipaggiamento_iniziale_opzioni": [
    {"etichetta": "Opzione A", "oggetti": ["..."]},
    {"etichetta": "Opzione B", "oggetti": ["..."]}
  ],
  "contenuto": "#### <Nome Origine>\n…"
}
```

---

## incantesimi

### Schema

```json
{
  "slug": "<acid-arrow>",
  "nome": "<Acid Arrow>",
  "livello": 2,
  "scuola": "<Evocation>",
  "classi": ["<Wizard>", "..."],
  "lancio": {
    "tempo": "<Azione|Reazione|…>",
    "gittata": "<90 ft|18 m>",
    "componenti": ["V","S","M"],
    "materiali": "testo opzionale",
    "durata": "<Istantanea|Concentrazione X min>"
  },
  "contenuto": "#### <Nome>\n*Level …*\n…"
}
```

---

## armi

### Schema

```json
{
  "slug": "<pugnale>",
  "nome": "<Pugnale>",
  "costo": { "valore": 2, "valuta": "mo" },
  "peso": { "valore": 0.5, "unita": "kg" },
  "danno": "1d4 Perforante",
  "categoria": "Semplice da Mischia",
  "proprieta": ["Accurata","Leggera","Da Lancio"],
  "maestria": "Fendente Rapido",
  "gittata": { "normale": "6 m", "lunga": "18 m" },
  "contenuto": "## <Pugnale>\n…"
}
```

---

## armature

### Schema

```json
{
  "slug": "<armatura-imbottita>",
  "nome": "<Armatura Imbottita>",
  "costo": { "valore": 5, "valuta": "mo" },
  "peso": { "valore": 3.5, "unita": "kg" },
  "categoria": "Leggera",
  "classe_armatura": { "base": 11, "modificatore_des": true, "limite_des": null },
  "forza_richiesta": null,
  "svantaggio_furtivita": true,
  "contenuto": "## <Armatura Imbottita>\n…"
}
```

---

## strumenti

### Schema

```json
{
  "slug": "<strumenti-da-alchimista>",
  "nome": "<Strumenti da alchimista>",
  "costo": { "valore": 50, "valuta": "mo" },
  "peso": { "valore": 3.5, "unita": "kg" },
  "categoria": "Strumenti",
  "abilità_associata": "Intelligenza",
  "utilizzi": [
    { "descrizione": "Identificare una sostanza", "cd": 15 }
  ],
  "creazioni": ["Acido","Fuoco dell’alchimista"],
  "contenuto": "## <Strumenti da alchimista>\n…"
}
```

---

## servizi

### Schema

```json
{
  "slug": "<tenore-di-vita-miserabile>",
  "nome": "<Tenore di vita - Miserabile>",
  "costo": { "valore": 0, "valuta": "gratuito" },
  "categoria": "Servizio",
  "descrizione": "…",
  "contenuto": "## <Tenore di vita - Miserabile>\n…"
}
```

---

## equipaggiamento

### Schema

```json
{
  "slug": "<otre>",
  "nome": "<Otre>",
  "costo": { "valore": 2, "valuta": "ma" },
  "peso": { "valore": 2.3, "unita": "kg" },
  "capacita": { "valore": 2, "unita": "l" },
  "note": "…",
  "contenuto": "## <Otre>\n…"
}
```

---

## oggetti\_magici

### Schema

```json
{
  "slug": "<adamantine-armor>",
  "nome": "<Adamantine Armor>",
  "tipo": "Armor (Any Medium or Heavy, Except Hide Armor)",
  "rarita": "Uncommon|…",
  "sintonizzazione": false,
  "contenuto": "### <Adamantine Armor>\n…"
}
```

---

## mostri

### Schema

```json
{
  "slug": "<aboleth>",
  "nome": "<Aboleth>",
  "taglia": "Grande",
  "tipo": "Aberrazione",
  "allineamento": "Legale Malvagio",
  "gs": 10,
  "pe": { "base": 5900, "tana": 7200},
  "ac": 17,
  "hp": "150 (20d10 + 40)",
  "velocita": "3 m, Nuotare 12 m",
  "caratteristiche": { "str": 21, "dex": 9, "con": 15, "int": 18, "wis": 15, "cha": 18 },
  "sensibilita": { },
  "tiri_salvezza": { },
  "abilità": { },
  "immunita": { "danni": [], "condizioni": [] },
  "azioni": [],
  "tratti": [],
  "reazioni": [],
  "azioni_leggendarie": [],
  "incantesimi": { "cd": null, "attacco": null, "lista": [] },
  "contenuto": "## <Aboleth>\n…"
}
```

---

## animali

### Schema

```json
{
  "slug": "<mulo>",
  "nome": "<Mulo>",
  "taglia": "Media",
  "tipo": "Animale",
  "ac": 10,
  "hp": "11 (2d8 + 2)",
  "velocita": "12 m",
  "caratteristiche": { "str": 14, "dex": 10, "con": 12, "int": 2, "wis": 10, "cha": 5 },
  "tratti": [],
  "azioni": [],
  "contenuto": "## <Mulo>\n…"
}
```

---

## talenti

### Schema

```json
{
  "slug": "<allerta>",
  "nome": "<Allerta>",
  "categoria": "Talento di Origine|Generale|…",
  "prerequisiti": "",
  "benefici": ["…"],
  "contenuto": "#### <Allerta>\n…"
}
```

---

## cavalcature\_e\_veicoli

### Schema

```json
{
  "slug": "<id>",
  "nome": "<Nome>",
  "tipo": "cavalcatura | nave | veicolo | altro",
  "costo": { "valore": 0, "valuta": "mo|ma|…" },
  "velocita": { "valore": null, "unita": "km/h|nodi|m/round" },
  "capacita_carico": { "valore": null, "unita": "kg|tonnellate" },
  "equipaggio": null,
  "passeggeri": null,
  "ca": null,
  "pf": null,
  "soglia_danni": null,
  "descrizione": "…",
  "contenuto": "## <Nome>\n…"
}
```

---

## Normalizzazione

* Normalizziamo i trattini en/em (–, —) quando necessario.
* Il titolo dei documenti è l’`H1` se presente, altrimenti lo slug derivato dal filename.
* Ogni entità conserva `contenuto` con il markdown integrale per il rendering fedele.
* Le unità sono in sistema metrico; riportiamo le originali in nota solo se presenti nella fonte.

