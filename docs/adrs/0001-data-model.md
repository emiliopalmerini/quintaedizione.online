# ADR Template

Status: Accepted

Context
- i dati dell'srd sono molti e diversi tra loro

Decisione

dividiamo i dati nelle seguenti collection
`documenti`: rappresenta le pagine per intero. per esempio ../../data/ita/01_informazioni_legali.md. il documento json sarà
    ```json
    {
        _id: guid,
        slug: note-legali
        titolo: "Note Legali" //Il titolo sarà parsato dal documento prendendo H1
        content: "" //L'intero markdown del documento compleso di titolo
        "numero_di_pagina: 1 // parsare il numero dal nome del file
    }
`classi`: rappresenta le classi del gioco. All'inizio saranno 12

```
{
  "slug": "stregone",
  "nome": "Stregone",
  "dado_vita": "d6",
  "caratteristica_primaria": "Carisma",
  "salvezze_competenze": [
    "Costituzione",
    "Carisma"
  ],
  "abilità_competenze_opzioni": {
    "scegli": 2,
    "opzioni": [
      "Arcana",
      "Inganno",
      "Intuizione",
      "Intimidire",
      "Persuasione",
      "Religione"
    ]
  },
  "armi_competenze": [
    "Armi semplici"
  ],
  "armature_competenze": [],
  "equipaggiamento_iniziale_opzioni": [
    {
      "etichetta": "Opzione A",
      "oggetti": [
        "Lancia",
        "2 Pugnali",
        "Focus arcano (cristallo)",
        "Zaino dell'Esploratore di Dungeon",
        "28 mo"
      ]
    },
    {
      "etichetta": "Opzione B",
      "oggetti": [
        "50 mo"
      ]
    }
  ],
  "tabella_livelli": [
    {
      "livello": 1,
      "bonus_competenza": 2,
      "privilegi_di_classe": [
        "Lancio di Incantesimi",
        "Magia Innata"
      ],
      "punti_stregoneria": null,
      "trucchetti_conosciuti": 4,
      "incantesimi_preparati": 2,
      "slot_incantesimo": {
        "1": 2
      }
    },
    {
      "livello": 2,
      "bonus_competenza": 2,
      "capacità": [
        "Fonte di Magia",
        "Metamagia"
      ],
      "punti_stregoneria": 2,
      "trucchetti_conosciuti": 4,
      "incantesimi_preparati": 4,
      "slot_incantesimo": {
        "1": 3
      }
    }
  ],
  "privilegi_di_classe": [
    {
      "nome": "Lancio di Incantesimi",
      "livello": 1,
      "descrizione": "Attraendo dalla tua magia innata, puoi lanciare incantesimi..."
    },
    {
      "nome": "Magia Innata",
      "livello": 1,
      "descrizione": "Un evento nel tuo passato ti ha lasciato un marchio indelebile..."
    },
    {
      "nome": "Fonte di Magia",
      "livello": 2,
      "descrizione": "Puoi attingere al pozzo di magia dentro di te, rappresentato dai Punti Stregoneria..."
    }
  ],
  "lanciare_incantesimi": {
    "trucchetti": [
      "Luce",
      "Prestidigitazione",
      "Presa Folgorante",
      "Scoppio Stregonesco"
    ],
    "lista_incantesimi": {
      "1": [
        "Mani Brucianti",
        "Individuazione del Magico",
        "Scudo"
      ],
      "2": [
        "Invisibilità",
        "Raggio Rovente"
      ],
      "3": [
        "Palla di Fuoco",
        "Volare"
      ],
      "4": [
        "Metamorfosi",
        "Invisibilità Superiore"
      ],
      "5": [
        "Telecinesi",
        "Cerchio di Teletrasporto"
      ]
    }
  },
  "sottoclassi": [
    {
      "slug": "stregoneria-draconica",
      "nome": "Stregoneria Draconica",
      "privilegi_sottoclasse": [
        {
          "nome": "Resilienza Draconica",
          "livello": 3,
          "descrizione": "La magia nel tuo corpo manifesta tratti fisici del tuo dono draconico..."
        },
        {
          "nome": "Affinità Elementale",
          "livello": 6,
          "descrizione": "La tua magia draconica ha affinità con un tipo di danno..."
        }
      ],
      "incantesimi_aggiuntivi": {
        "3": [
          "Alterare Sé Stesso",
          "Sfera Cromatica"
        ],
        "5": [
          "Paura",
          "Volare"
        ],
        "7": [
          "Occhio Arcano",
          "Incantare Mostri"
        ],
        "9": [
          "Leggenda",
          "Evoca Drago"
        ]
      }
    }
  ],
  "extra": {
    "nome": "Opzioni di Metamagia",
    "descrizione": "Le seguenti opzioni sono disponibili per i tuoi privilegi **Metamagia**. Sono presentate in ordine alfabetico.",
    "opzioni": [
      {
        "nome": "Nome opzione",
        "costo": 1,
        "descrizione": "descrizione effetto"
      }
    ]
  }
}
```

Visto che le classi sono diverse, bisognerà procedere con un parser ad hoc per ciascuna. I parser implementeranno la stessa interfaccia ma avranno implementazioni leggermente diverse. Il documento json dovrà essere l'insieme più esteso di feature.
Per le classi a cui non serve un campo non verrà utilizzato.

Dobbiamo parsare anche i background. Sono in ../../data/ita/05_origini_personaggio.md sotto l'header `### Descrizioni Background` Per esempio l'Accolito

```md
#### Accolito

**Punteggi di Caratteristica:** Intelligenza, Saggezza, Carisma  

**Talento:** Iniziato alla Magia (Chierico) (vedi “Talenti”)  

**Competenze in Abilità:** Intuizione e Religione  

**Competenza negli Strumenti:** Strumenti da Calligrafo  

**Equipaggiamento:** *Scegli A o B:* (A) Strumenti da Calligrafo, Libro (preghiere), Simbolo Sacro, Pergamena (10 fogli), Veste, 8 mo; oppure (B) 50 mo
```
```json
{
    "punteggi_caratteristica": ["Intelligenza", "Saggezza", "Carisma"],
    "talento": "Iniziato alla Magia (Chierico)",
    "abilità_competenze": ["Intuizione", "Religione"],
    "strumenti_competenze": ["Strumenti da Calligrafo"],
    "strumenti_competenze": ["Strumenti da Calligrafo"],
    "equipaggiamento_iniziale_opzioni": [
    {
      "etichetta": "Opzione A",
      "oggetti": [
      ]
    },
    {
      "etichetta": "Opzione B",
      "oggetti": [
      ]
    }
}
