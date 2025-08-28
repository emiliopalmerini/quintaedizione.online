# Data Model (SRD)

Status: Accepted

## Context
- I dati dell'SRD sono numerosi ed eterogenei. Serve un modello normalizzato per l'editor/visualizzatore.

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

## Normalizzazione
- Nei testi estratti, i trattini en/em (–, —) sono normalizzati quando necessario.
- Il titolo dei documenti (`titolo`) è l'`H1` se presente; in assenza si usa lo slug derivato dal filename.
