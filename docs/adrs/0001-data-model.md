# Data Model (SRD)

Status: Accepted

## Context
- I dati dell'SRD sono numerosi ed eterogenei. Serve un modello normalizzato per il visualizzatore.

## Decisione

Dividiamo i dati nelle seguenti collection.

- `documenti`: pagine SRD italiane (Markdown completo, con titolo e numero pagina quando disponibile). Parsiamo in un json che ne rappresenti la struttura.

Esempio:

```json
{
  "pagina": 1,
  "titolo": "Giocare",
  "paragrafi": [
    {
      "numero": 1,
      "titolo": "Ritmo di gioco",
      "corpo": {
        "markdown": "I tre pilastri principali del gioco di D&D sono l’interazione sociale, l’esplorazione e il combattimento. Qualunque di questi tu stia vivendo, il gioco si sviluppa secondo questo schema di base:\n\n1. **Il Game Master descrive una scena.** Il GM racconta ai giocatori dove si trovano gli avventurieri e cosa c’è attorno a loro (quante porte ci sono in una stanza, cosa c’è su un tavolo, e così via).\n\n2. **I giocatori descrivono cosa fanno i loro personaggi.** Tipicamente i personaggi restano insieme mentre viaggiano in un dungeon o in un altro ambiente. A volte avventurieri diversi fanno cose diverse: uno può cercare in uno scrigno, un altro esaminare un simbolo inciso su un muro e un terzo fare la guardia ai mostri. Fuori dal combattimento, il GM garantisce che ogni personaggio abbia la possibilità di agire e decide come risolvere le loro attività. In combattimento i personaggi agiscono a turni.\n\n3. **Il GM narra i risultati delle azioni degli avventurieri.** A volte risolvere un compito è facile. Se un avventuriero attraversa una stanza e prova ad aprire una porta, il GM può dire che la porta si apre e descrivere cosa c’è oltre. Ma la porta potrebbe essere chiusa a chiave, il pavimento potrebbe nascondere una trappola, o altre circostanze potrebbero rendere difficile portare a termine l’azione. In quei casi il GM può chiedere di tirare un dado per determinare cosa accade. La descrizione del risultato porta spesso a un nuovo punto decisionale, riportando il gioco al passo 1.\n\nQuesto schema si applica in ogni sessione (ogni volta che ci si siede a giocare), che gli avventurieri stiano parlando con un nobile, esplorando una rovina o combattendo un drago. In certe situazioni—specialmente nel combattimento—l’azione è più strutturata e tutti agiscono a turno.",
        "sottoparagrafi": [
          {
            "numero": 1,
            "titolo": "Le eccezioni superano le regole generali",
            "markdown": "> **Le eccezioni superano le regole generali**\n>\n> Le regole generali governano ogni parte del gioco. Per esempio, le regole sul combattimento dicono che gli attacchi in mischia usano Forza e quelli a distanza usano Destrezza. Questa è una regola generale, ed è valida finché qualcosa nel gioco non dice esplicitamente il contrario.  \n> Il gioco include anche elementi—privilegi di classe, talenti, proprietà delle armi, incantesimi, oggetti magici, capacità dei mostri e simili—che a volte contraddicono una regola generale. Quando un’eccezione e una regola generale si scontrano, vince l’eccezione. Per esempio, se un privilegio dice che puoi usare Carisma per gli attacchi in mischia, puoi farlo, anche se ciò contraddice la regola generale."
          }
        ]
      }
    },
    {
      "numero": 2,
      "titolo": "Le sei caratteristiche",
      "corpo": {
        "markdown": "Tutte le creature—personaggi e mostri—hanno sei caratteristiche che misurano tratti fisici e mentali, come mostrato nella tabella delle descrizioni.\n\nTabella: Descrizioni delle caratteristiche\n\n| Caratteristica | Misura …                         |\n|----------------|----------------------------------|\n| Forza          | Potenza fisica                   |\n| Destrezza      | Agilità, riflessi ed equilibrio  |\n| Costituzione   | Salute e resistenza              |\n| Intelligenza   | Ragionamento e memoria           |\n| Saggezza       | Percettività e forza di volontà  |\n| Carisma        | Fiducia, presenza e fascino      |",
        "sottoparagrafi": [
          {
            "numero": 1,
            "titolo": "Punteggi di caratteristica",
            "markdown": "Ogni caratteristica ha un punteggio da 1 a 20, anche se alcuni mostri arrivano fino a 30. Il punteggio rappresenta la grandezza di una caratteristica. La tabella dei punteggi riassume il significato.\n\nTabella: Punteggi di caratteristica\n\n| Punteggio | Significato                                                                 |\n|-----------|-----------------------------------------------------------------------------|\n| 1         | Minimo possibile. Se un effetto riduce un punteggio a 0, quell’effetto spiega cosa accade. |\n| 2–9       | Capacità debole.                                                            |\n| 10–11     | Media umana.                                                                |\n| 12–19     | Capacità elevata.                                                           |\n| 20        | Massimo raggiungibile da un avventuriero salvo diversa indicazione.         |\n| 21–29     | Capacità straordinaria.                                                     |\n| 30        | Massimo assoluto.                                                           |"
          }
        ]
      }
    }
  ]
}

```

- `classi`: una per ogni classe. Alcuni campi sono opzionali e presenti solo dove hanno senso.
L'intera classe markdown è da inserire in tutte le classi. Viene inserito solo nel primo esempio.
```json
{
  "slug": "<identificatore_minuscolo>",
  "nome": "<Nome Classe>",
  "sottotitolo": "<tagline breve>",
  "dado_vita": "d6|d8|d10|d12",
  "caratteristica_primaria": ["Forza","Destrezza","Costituzione","Intelligenza","Saggezza","Carisma"],
  "salvezze_competenze": ["<Car1>","<Car2>"],
  "abilità_competenze_opzioni": {
    "scegli": <n>,
    "opzioni": ["<Abilità1>","<Abilità2>", "..."]
  },
  "armi_competenze": ["<arma o gruppo>","..."],
  "armature_competenze": ["Leggere","Medie","Pesanti","Scudi"],
  "strumenti_competenze": ["<strumento>","..."],
  "equipaggiamento_iniziale_opzioni": [
    { "etichetta": "Opzione A", "oggetti": ["<oggetto>", "..."] },
    { "etichetta": "Opzione B", "oggetti": ["<oggetto>", "..."] },
    { "etichetta": "Ricchezza", "oggetti": ["<mo>"] }
  ],

  "multiclasse": {
    "prerequisiti": ["<Car> <valore>","..."],
    "tratti_acquisiti": ["<dado_vita>","<competenze armi>","<armature>","<strumenti/abilità extra>"],
    "note": "<regole slot o eccezioni>"
  },

  "progressioni": {
    "maestria_armi": { "livelli": { "1": 0, "4": 1, "8": 2, "16": 3 } },
    "stili_combattimento": { "livelli": { "1": 1, "7": 2 }, "scelte": ["<Stile>","..."] },
    "attacchi_extra": { "livelli": { "5": 1, "11": 2, "20": 3 } },
    "risorse": [
      { "chiave": "ripresa_rapida|punti_concentrazione|canalizzare_divinita|forma_selvatica|…", "livelli": { "1": 2, "2": 2, "10": 3, "18": 4 } }
    ],
    "aumenti_caratteristica": [4,8,12,16],
    "dono_epico": 19
  },

  "magia": {
    "ha_incantesimi": true,
    "lista_riferimento": "<Chierico|Druido|Paladino|Ranger|Stregone|Warlock|Mago|Classe>",
    "caratteristica_incantatore": "<Carisma|Saggezza|Intelligenza>",
    "preparazione": "prepared|known|none",
    "focus": "<simbolo sacro|focus druidico|focus arcano|—>",
    "rituali": "nessuno|solo_lista|da_libro",
    "trucchetti": { "1": 2, "4": 3, "10": 4 },
    "incantesimi_preparati_o_notI": { "1": 4, "5": 9, "10": 15, "20": 22 },
    "slot": {
      "1": [2,0,0,0,0,0,0,0,0],
      "2": [3,0,0,0,0,0,0,0,0],
      "3": [4,2,0,0,0,0,0,0,0],
      "4": [4,3,0,0,0,0,0,0,0],
      "5": [4,3,2,0,0,0,0,0,0],
      "6": [4,3,3,0,0,0,0,0,0],
      "7": [4,3,3,1,0,0,0,0,0],
      "8": [4,3,3,2,0,0,0,0,0],
      "9": [4,3,3,3,1,0,0,0,0],
      "10": [4,3,3,3,2,0,0,0,0],
      "11": [4,3,3,3,2,1,0,0,0],
      "12": [4,3,3,3,2,1,0,0,0],
      "13": [4,3,3,3,2,1,1,0,0],
      "14": [4,3,3,3,2,1,1,0,0],
      "15": [4,3,3,3,2,1,1,1,0],
      "16": [4,3,3,3,2,1,1,1,0],
      "17": [4,3,3,3,2,1,1,1,1],
      "18": [4,3,3,3,3,1,1,1,1],
      "19": [4,3,3,3,3,2,1,1,1],
      "20": [4,3,3,3,3,2,2,1,1]
    },
    "patto_warlock": { "slot": { "1":1, "2":2, "3":2, "5":2, "11":3, "17":4 }, "livello_slot": { "1":1, "3":2, "5":3, "7":4, "9":5 } },
    "arcani_mistici": { "11":6, "13":7, "15":8, "17":9 }
  },

  "tabella_livelli": [
    {
      "livello": 1,
      "bonus_competenza": 2,
      "privilegi_di_classe": ["<Priv1>","<Priv2>"],
      "risorse": { "<chiave_risorsa>": 2 },
      "trucchetti": 0,
      "incantesimi_preparati": 0,
      "slot": [2,0,0,0,0,0,0,0,0],
      "note": "<campi opzionali specifici (es. Canalizzare Divinità, Forma selvatica, Maestria armi, ecc.)>"
    }
    /* …ripetere fino a 20, variando colonne rilevanti per la classe … */
  ],

  "privilegi_di_classe": [
    { "nome": "<Nome Privilegio>", "livello": <n>, "descrizione": "<testo breve>"},
    { "nome": "<Nome Privilegio>", "livello": <n>, "descrizione": "<testo breve>"}
  ],

  "sottoclassi": [
    {
      "slug": "<identificatore>",
      "nome": "<Nome Sottoclasse>",
      "descrizione": "<frase tema>",
      "privilegi_sottoclasse": [
        { "nome": "<Privilegio>", "livello": 3, "descrizione": "<testo>"},
        { "nome": "<Privilegio>", "livello": 6, "descrizione": "<testo>"},
        { "nome": "<Privilegio>", "livello": 10, "descrizione": "<testo>"},
        { "nome": "<Privilegio>", "livello": 14, "descrizione": "<testo>"}
      ],
      "incantesimi_sempre_preparati": {
        "3": ["<inc1>","<inc2>"],
        "5": ["<inc>"],
        "7": ["<inc>"],
        "9": ["<inc>"]
      }
    }
  ],

  "liste_incantesimi": {
    "0": ["<Trucchetto>","..."],
    "1": ["<Incantesimo>","..."],
    "2": ["..."]
    /* fino a 9; opzionale per classi non-incantatrici */
  },

  "raccomandazioni": {
    "trucchetti_cons": ["<nome>", "..."],
    "incantesimi_iniziali_cons": ["<nome>", "..."],
    "equip_iniziale_cons": "Opzione A|B|C",
    "talenti_cons": ["<Stile di combattimento/Talento>", "..."],
    "dono_epico_cons": "<nome>"
  }
}

```
Il json sotto è da considerarsi solamente un esempio. I valori devono essere parsati esattamente da ../../data/ita/04_classi_items.md


```json
{
  "slug": "barbaro",
  "nome": "Barbaro",
  "markdown": "...", <- intera classe in markdown
  "dado_vita": "d12",
  "caratteristica_primaria": "Forza",
  "salvezze_competenze": ["Forza", "Costituzione"],
  "abilità_competenze_opzioni": {
    "scegli": 2,
    "opzioni": ["Addestrare Animali", "Atletica", "Intimidire", "Natura", "Percezione", "Sopravvivenza"]
  },
  "armi_competenze": ["Armi semplici", "Armi da guerra"],
  "armature_competenze": ["Armature leggere", "Armature medie", "Scudi"],
  "strumenti_competenze": [],
  "equipaggiamento_iniziale_opzioni": [
    { "etichetta": "Opzione A", "oggetti": ["Ascia bipenne", "4 Asce da lancio", "Zaino da esploratore", "15 mo"] },
    { "etichetta": "Opzione B", "oggetti": ["75 mo"] }
  ],
  "multiclasse": {
    "tratti_acquisiti": ["Dado Punti Ferita", "Competenza armi da guerra", "Addestramento con Scudi"],
    "note": "Ottieni anche i privilegi di classe di 1° livello elencati nella tabella."
  },
  "progressioni_speciali": {
    "risorse": {
      "per_livello": [
        { "livello": 1, "voci": { "usi_ira": 2, "danni_da_ira": 2, "dado_ispirazione": null } },
        { "livello": 2, "voci": { "usi_ira": 2, "danni_da_ira": 2, "dado_ispirazione": null } },
        { "livello": 3, "voci": { "usi_ira": 3, "danni_da_ira": 2, "dado_ispirazione": null } },
        { "livello": 4, "voci": { "usi_ira": 3, "danni_da_ira": 2, "dado_ispirazione": null } },
        { "livello": 5, "voci": { "usi_ira": 3, "danni_da_ira": 2, "dado_ispirazione": null } },
        { "livello": 6, "voci": { "usi_ira": 4, "danni_da_ira": 2, "dado_ispirazione": null } },
        { "livello": 7, "voci": { "usi_ira": 4, "danni_da_ira": 2, "dado_ispirazione": null } },
        { "livello": 8, "voci": { "usi_ira": 4, "danni_da_ira": 2, "dado_ispirazione": null } },
        { "livello": 9, "voci": { "usi_ira": 4, "danni_da_ira": 3, "dado_ispirazione": null } },
        { "livello": 10, "voci": { "usi_ira": 4, "danni_da_ira": 3, "dado_ispirazione": null } },
        { "livello": 11, "voci": { "usi_ira": 4, "danni_da_ira": 3, "dado_ispirazione": null } },
        { "livello": 12, "voci": { "usi_ira": 5, "danni_da_ira": 3, "dado_ispirazione": null } },
        { "livello": 13, "voci": { "usi_ira": 5, "danni_da_ira": 3, "dado_ispirazione": null } },
        { "livello": 14, "voci": { "usi_ira": 5, "danni_da_ira": 3, "dado_ispirazione": null } },
        { "livello": 15, "voci": { "usi_ira": 5, "danni_da_ira": 3, "dado_ispirazione": null } },
        { "livello": 16, "voci": { "usi_ira": 5, "danni_da_ira": 4, "dado_ispirazione": null } },
        { "livello": 17, "voci": { "usi_ira": 6, "danni_da_ira": 4, "dado_ispirazione": null } },
        { "livello": 18, "voci": { "usi_ira": 6, "danni_da_ira": 4, "dado_ispirazione": null } },
        { "livello": 19, "voci": { "usi_ira": 6, "danni_da_ira": 4, "dado_ispirazione": null } },
        { "livello": 20, "voci": { "usi_ira": 6, "danni_da_ira": 4, "dado_ispirazione": null } }
      ]
    },
    "maestrie_armi_per_livello": [
      { "livello": 1, "conteggio": 2 },
      { "livello": 4, "conteggio": 3 },
      { "livello": 10, "conteggio": 4 }
    ]
  },
  "tabella_livelli": [
    { "livello": 1, "bonus_competenza": 2, "privilegi_di_classe": ["Ira", "Difesa senza armatura", "Maestria nelle armi"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 2, "bonus_competenza": 2, "privilegi_di_classe": ["Senso del pericolo", "Attacco sconsiderato"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 3, "bonus_competenza": 2, "privilegi_di_classe": ["Sottoclasse del Barbaro", "Conoscenza primordiale"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 4, "bonus_competenza": 2, "privilegi_di_classe": ["Aumento dei punteggi di caratteristica"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 5, "bonus_competenza": 3, "privilegi_di_classe": ["Attacco extra", "Movimento veloce"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 6, "bonus_competenza": 3, "privilegi_di_classe": ["Privilegio di sottoclasse"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 7, "bonus_competenza": 3, "privilegi_di_classe": ["Istinto ferino", "Balzo istintivo"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 8, "bonus_competenza": 3, "privilegi_di_classe": ["Aumento dei punteggi di caratteristica"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 9, "bonus_competenza": 4, "privilegi_di_classe": ["Colpo brutale"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 10, "bonus_competenza": 4, "privilegi_di_classe": ["Privilegio di sottoclasse"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 11, "bonus_competenza": 4, "privilegi_di_classe": ["Ira implacabile"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 12, "bonus_competenza": 4, "privilegi_di_classe": ["Aumento dei punteggi di caratteristica"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 13, "bonus_competenza": 5, "privilegi_di_classe": ["Colpo brutale migliorato"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 14, "bonus_competenza": 5, "privilegi_di_classe": ["Privilegio di sottoclasse"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 15, "bonus_competenza": 5, "privilegi_di_classe": ["Ira persistente"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 16, "bonus_competenza": 5, "privilegi_di_classe": ["Aumento dei punteggi di caratteristica"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 17, "bonus_competenza": 6, "privilegi_di_classe": ["Colpo brutale migliorato"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 18, "bonus_competenza": 6, "privilegi_di_classe": ["Potenza indomita"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 19, "bonus_competenza": 6, "privilegi_di_classe": ["Dono epico"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } },
    { "livello": 20, "bonus_competenza": 6, "privilegi_di_classe": ["Campione primordiale"], "trucchetti_conosciuti": null, "incantesimi_preparati": null, "slot_incantesimo": { "1":0,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0 } }
  ],
  "privilegi_di_classe": [
    { "nome": "Ira", "livello": 1, "descrizione": "Resistenze B/P/T; bonus danno su attacchi con Forza; vantaggio a prove e TS Forza; niente concentrazione/incantesimi; estendibile ogni turno; max 10 minuti." },
    { "nome": "Difesa senza armatura", "livello": 1, "descrizione": "CA = 10 + Des + Cos; lo Scudo si applica." },
    { "nome": "Maestria nelle armi", "livello": 1, "descrizione": "Scegli 2 armi da mischia semplici/da guerra; cambi una scelta a fine Riposo lungo. Aumenta ai livelli indicati." },
    { "nome": "Senso del pericolo", "livello": 2, "descrizione": "Vantaggio ai TS Des se non Incapacitato." },
    { "nome": "Attacco sconsiderato", "livello": 2, "descrizione": "Vantaggio ai tiri per colpire con Forza fino al prossimo turno; gli attacchi contro di te hanno vantaggio." },
    { "nome": "Conoscenza primordiale", "livello": 3, "descrizione": "1 abilità extra. In Ira puoi trattare Acrobazia, Intimidire, Percezione, Furtività, Sopravvivenza come prove di Forza." },
    { "nome": "Attacco extra", "livello": 5, "descrizione": "Due attacchi con l'Azione Attacco." },
    { "nome": "Movimento veloce", "livello": 5, "descrizione": "Velocità +3 m se non armatura pesante." },
    { "nome": "Istinto ferino", "livello": 7, "descrizione": "Vantaggio ai tiri di Iniziativa." },
    { "nome": "Balzo istintivo", "livello": 7, "descrizione": "Entrando in Ira, ti muovi fino a metà velocità come Azione bonus." },
    { "nome": "Colpo brutale", "livello": 9, "descrizione": "Rinunci al vantaggio di Attacco sconsiderato per 1d10 danni extra e un effetto (spinta o rallentamento)." },
    { "nome": "Ira implacabile", "livello": 11, "descrizione": "A 0 PF: TS Cos CD 10; se riesci, PF = 2× livello da Barbaro; CD +5 cumulativa fino al riposo." },
    { "nome": "Colpo brutale migliorato", "livello": 13, "descrizione": "Nuovi effetti: stordente (svantaggio al prossimo TS, niente AO) o squarciante (+5 al prossimo tiro per colpire di un alleato)." },
    { "nome": "Ira persistente", "livello": 15, "descrizione": "Tiri Iniziativa: ripristini tutti gli usi di Ira (1/LL). L'Ira dura 10 minuti senza estensioni." },
    { "nome": "Colpo brutale migliorato", "livello": 17, "descrizione": "Danni extra 2d10 e puoi applicare due effetti." },
    { "nome": "Potenza indomita", "livello": 18, "descrizione": "Prove/TS di Forza minimi pari al punteggio di Forza." },
    { "nome": "Dono epico", "livello": 19, "descrizione": "Ottieni un Dono epico." },
    { "nome": "Campione primordiale", "livello": 20, "descrizione": "Forza e Costituzione +4 fino a max 25." }
  ],
  "regole_classe": {
    "durate": { "ira": "fino a fine del tuo turno successivo; estendibile; max 10 minuti" },
    "limitazioni": { "ira_armatura_pesante": true, "ira_no_concentrazione_incantesimi": true },
    "formule": { "CA_senza_armatura": "10 + Des + Cos" }
  },
  "lanciare_incantesimi": { "presente": false, "caratteristica_incantatore": null, "focus": null, "trucchetti_consigliati": [], "incantesimi_iniziali_consigliati": [], "lista_incantesimi": {} },
  "sottoclassi": [
    {
      "slug": "sentiero-del-berserker",
      "nome": "Sentiero del Berserker",
      "privilegi_sottoclasse": [
        { "nome": "Frenesia", "livello": 3, "descrizione": "In Ira, con Attacco sconsiderato, danni extra al primo bersaglio pari a d6 × bonus Danni da Ira." },
        { "nome": "Ira insensata", "livello": 6, "descrizione": "Immunità Affascinato e Spaventato in Ira; terminano entrando in Ira." },
        { "nome": "Ritorsione", "livello": 10, "descrizione": "Reazione: attacco in mischia contro chi ti danneggia entro 1,5 m." },
        { "nome": "Presenza intimidatoria", "livello": 14, "descrizione": "AB: emanazione 9 m; TS Sag (CD 8 + mod For + competenza) o Spaventato 1 min; RI a fine turno; 1/LL o spendi 1 uso Ira." }
      ],
      "incantesimi_aggiuntivi": {}
    }
  ]
},
{
  "slug": "bardo",
  "nome": "Bardo",
  "dado_vita": "d8",
  "caratteristica_primaria": "Carisma",
  "salvezze_competenze": ["Destrezza", "Carisma"],
  "abilità_competenze_opzioni": { "scegli": 3, "opzioni": ["Qualsiasi abilità"] },
  "armi_competenze": ["Armi semplici"],
  "armature_competenze": ["Armature leggere"],
  "strumenti_competenze": ["Scegli 3 strumenti musicali"],
  "equipaggiamento_iniziale_opzioni": [
    { "etichetta": "Opzione A", "oggetti": ["Armatura di cuoio", "2 Pugnali", "1 Strumento musicale a scelta", "Zaino da intrattenitore", "19 mo"] },
    { "etichetta": "Opzione B", "oggetti": ["90 mo"] }
  ],
  "multiclasse": {
    "tratti_acquisiti": ["Dado Punti Ferita", "1 abilità a scelta", "1 strumento musicale a scelta", "Addestramento armature leggere"],
    "note": "Gli slot incantesimo seguono le regole del multiclasse incantatori."
  },
  "progressioni_speciali": {
    "risorse": {
      "per_livello": [
        { "livello": 1, "voci": { "dado_ispirazione": 6 } },
        { "livello": 5, "voci": { "dado_ispirazione": 8 } },
        { "livello": 10, "voci": { "dado_ispirazione": 10 } },
        { "livello": 15, "voci": { "dado_ispirazione": 12 } }
      ]
    },
    "maestrie_armi_per_livello": []
  },
  "tabella_livelli": [
    { "livello":1,"bonus_competenza":2,"privilegi_di_classe":["Ispirazione bardica","Incantesimi"],"trucchetti_conosciuti":2,"incantesimi_preparati":4,"slot_incantesimo":{"1":2,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0}},
    { "livello":2,"bonus_competenza":2,"privilegi_di_classe":["Maestria","Tuttofare"],"trucchetti_conosciuti":2,"incantesimi_preparati":5,"slot_incantesimo":{"1":3,"2":0,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0}},
    { "livello":3,"bonus_competenza":2,"privilegi_di_classe":["Sottoclasse del Bardo"],"trucchetti_conosciuti":2,"incantesimi_preparati":6,"slot_incantesimo":{"1":4,"2":2,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0}},
    { "livello":4,"bonus_competenza":2,"privilegi_di_classe":["Aumento dei punteggi di caratteristica"],"trucchetti_conosciuti":3,"incantesimi_preparati":7,"slot_incantesimo":{"1":4,"2":3,"3":0,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0}},
    { "livello":5,"bonus_competenza":3,"privilegi_di_classe":["Fonte di Ispirazione"],"trucchetti_conosciuti":3,"incantesimi_preparati":9,"slot_incantesimo":{"1":4,"2":3,"3":2,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0}},
    { "livello":6,"bonus_competenza":3,"privilegi_di_classe":["Privilegio di sottoclasse"],"trucchetti_conosciuti":3,"incantesimi_preparati":10,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":0,"5":0,"6":0,"7":0,"8":0,"9":0}},
    { "livello":7,"bonus_competenza":3,"privilegi_di_classe":["Controincanto"],"trucchetti_conosciuti":3,"incantesimi_preparati":11,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":1,"5":0,"6":0,"7":0,"8":0,"9":0}},
    { "livello":8,"bonus_competenza":3,"privilegi_di_classe":["Aumento dei punteggi di caratteristica"],"trucchetti_conosciuti":3,"incantesimi_preparati":12,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":2,"5":0,"6":0,"7":0,"8":0,"9":0}},
    { "livello":9,"bonus_competenza":4,"privilegi_di_classe":["Maestria"],"trucchetti_conosciuti":3,"incantesimi_preparati":14,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":3,"5":1,"6":0,"7":0,"8":0,"9":0}},
    { "livello":10,"bonus_competenza":4,"privilegi_di_classe":["Segreti magici"],"trucchetti_conosciuti":4,"incantesimi_preparati":15,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":3,"5":2,"6":0,"7":0,"8":0,"9":0}},
    { "livello":11,"bonus_competenza":4,"privilegi_di_classe":[],"trucchetti_conosciuti":4,"incantesimi_preparati":16,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":3,"5":2,"6":1,"7":0,"8":0,"9":0}},
    { "livello":12,"bonus_competenza":4,"privilegi_di_classe":["Aumento dei punteggi di caratteristica"],"trucchetti_conosciuti":4,"incantesimi_preparati":16,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":3,"5":2,"6":1,"7":0,"8":0,"9":0}},
    { "livello":13,"bonus_competenza":5,"privilegi_di_classe":[],"trucchetti_conosciuti":4,"incantesimi_preparati":17,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":3,"5":2,"6":1,"7":1,"8":0,"9":0}},
    { "livello":14,"bonus_competenza":5,"privilegi_di_classe":["Privilegio di sottoclasse"],"trucchetti_conosciuti":4,"incantesimi_preparati":17,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":3,"5":2,"6":1,"7":1,"8":0,"9":0}},
    { "livello":15,"bonus_competenza":5,"privilegi_di_classe":[],"trucchetti_conosciuti":4,"incantesimi_preparati":18,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":3,"5":2,"6":1,"7":1,"8":1,"9":0}},
    { "livello":16,"bonus_competenza":5,"privilegi_di_classe":["Aumento dei punteggi di caratteristica"],"trucchetti_conosciuti":4,"incantesimi_preparati":18,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":3,"5":2,"6":1,"7":1,"8":1,"9":0}},
    { "livello":17,"bonus_competenza":6,"privilegi_di_classe":[],"trucchetti_conosciuti":4,"incantesimi_preparati":19,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":3,"5":2,"6":1,"7":1,"8":1,"9":1}},
    { "livello":18,"bonus_competenza":6,"privilegi_di_classe":["Ispirazione superiore"],"trucchetti_conosciuti":4,"incantesimi_preparati":20,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":3,"5":3,"6":1,"7":1,"8":1,"9":1}},
    { "livello":19,"bonus_competenza":6,"privilegi_di_classe":["Dono epico"],"trucchetti_conosciuti":4,"incantesimi_preparati":21,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":3,"5":3,"6":2,"7":1,"8":1,"9":1}},
    { "livello":20,"bonus_competenza":6,"privilegi_di_classe":["Parole di Creazione"],"trucchetti_conosciuti":4,"incantesimi_preparati":22,"slot_incantesimo":{"1":4,"2":3,"3":3,"4":3,"5":3,"6":2,"7":2,"8":1,"9":1}}
  ],
  "privilegi_di_classe": [
    { "nome": "Ispirazione bardica", "livello": 1, "descrizione": "AB: conferisci un dado a una creatura entro 18 m; 1 uso per punto di mod Car; d6→d8 al 5°, d10 al 10°, d12 al 15°; si usa su fallimenti con d20 entro 1 ora." },
    { "nome": "Incantesimi", "livello": 1, "descrizione": "Carisma come caratteristica; focus: strumento musicale; prepari secondo tabella; recupero slot con Riposo lungo." },
    { "nome": "Maestria", "livello": 2, "descrizione": "Scegli 2 abilità competenti; altre 2 al 9°." },
    { "nome": "Tuttofare", "livello": 2, "descrizione": "Metà bonus competenza a prove senza competenza." },
    { "nome": "Fonte di Ispirazione", "livello": 5, "descrizione": "Recuperi usi con Riposo breve/lungo; puoi spendere slot per recuperarne uno." },
    { "nome": "Controincanto", "livello": 7, "descrizione": "Reazione: te o alleato entro 9 m ripetete TS vs Affascinato/Spaventato con vantaggio." },
    { "nome": "Segreti magici", "livello": 10, "descrizione": "Puoi preparare incantesimi anche da Chierico, Druido, Mago." },
    { "nome": "Ispirazione superiore", "livello": 18, "descrizione": "All'inizio di un combattimento: se hai <2 usi, sali a 2." },
    { "nome": "Parole di Creazione", "livello": 20, "descrizione": "Hai sempre preparati Parola di guarigione suprema e Parola di morte; bersagli aggiuntivi a 3 m." }
  ],
  "regole_classe": {
    "durate": {},
    "limitazioni": {},
    "formule": {}
  },
  "lanciare_incantesimi": {
    "presente": true,
    "caratteristica_incantatore": "Carisma",
    "focus": "Strumento musicale",
    "trucchetti_consigliati": ["Luci danzanti", "Irrisione crudele"],
    "incantesimi_iniziali_consigliati": ["Ammaliare persone", "Spruzzo colorato", "Sussurri dissonanti", "Parola curativa"],
    "lista_incantesimi": {}
  },
  "sottoclassi": [
    {
      "slug": "collegio-della-sapienza",
      "nome": "Collegio della Sapienza",
      "privilegi_sottoclasse": [
        { "nome": "Competenze aggiuntive", "livello": 3, "descrizione": "Ottieni competenza in tre abilità a scelta." },
        { "nome": "Parole taglienti", "livello": 3, "descrizione": "Reazione: spendi Ispirazione; sottrai il dado a un tiro riuscito/danni entro 18 m." },
        { "nome": "Scoperte magiche", "livello": 6, "descrizione": "Impari 2 incantesimi da Chierico/Druido/Mago; sempre preparati; sostituibili a ogni livello." },
        { "nome": "Abilità impareggiabile", "livello": 14, "descrizione": "Quando fallisci una prova/attacco, spendi Ispirazione; se resti in fallimento, il dado non è speso." }
      ],
      "incantesimi_aggiuntivi": {}
    }
  ]
},
```

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

- `incantesimi`: una entry per incantesimo (lista da `16_incantesimi_items.md`). Ogni documento contiene un campo strutturato e il markdown originale della sezione per un rendering fedele.

```json
{
  "slug": "acid-arrow",
  "nome": "Acid Arrow",
  "livello": 2,
  "scuola": "Evocation",
  "classi": ["Wizard"],
  "lancio": {"tempo": "Action", "gittata": "90 feet", "componenti": ["V", "S", "M", "powdered rhubarb leaf"], "durata": "Instantaneous"},
  "contenuto": "#### Acid Arrow\n*Level 2 Evocation (Wizard)*\n..."
}
```

- `armi`: una entry per arma (`09_armi_items.md`) .

```json
{
  "slug": "pugnale",
  "nome": "Pugnale",
  "costo": "2 mo",
  "peso": "0,5 kg",
  "danno": "1d4 Perforante",
  "categoria": "Semplice da Mischia",
  "proprieta": ["Accurata", "Leggera", "Da Lancio"],
  "maestria": "Fendente Rapido",
  "gittata": {"normale": "6 m", "lunga": "18 m"},
  "contenuto": "## Pugnale\n**Costo:** ..."
}
```

- `armature`: una entry per ogni armatura/scudo (`11_armatura_items.md`).

- `strumenti`: una entry per ogni set di strumenti (`12_strumenti_items.md`).

- `servizi`: una entry per ogni servizio (`13_servizi_items.md`).

- `equipaggiamento`: una entry per ogni oggetto di equipaggiamento (`08_equipaggiamento_items.md`).

- `oggetti_magici`: una entry per ogni oggetto magico (sezione A–Z in `10_oggetti_magici_items.md`).

```json
{
  "slug": "adamantine-armor",
  "nome": "Adamantine Armor",
  "tipo": "Armor (Any Medium or Heavy, Except Hide Armor)",
  "rarita": "Uncommon",
  "sintonizzazione": false,
  "contenuto": "### Adamantine Armor\n*Armor (Any Medium or Heavy, Except Hide Armor), Uncommon*\n..."
}
```

- `mostri`: una entry per creatura (sezione A–Z in `20_mostri_items.md`).

```json
{
  "slug": "aboleth",
  "nome": "Aboleth",
  "tag": {"taglia": "Large", "tipo": "Aberration", "allineamento": "Lawful Evil"},
  "ac": 17,
  "hp": "150 (20d10 + 40)",
  "velocita": "10 ft., Swim 40 ft.",
  "caratteristiche": {"str": 21, "dex": 9, "con": 15, "int": 18, "wis": 15, "cha": 18},
  "contenuto": "## Aboleth\n*Large Aberration, Lawful Evil*\n- **Armor Class:** 17\n..."
}
```

- `animali`: una entry per ogni animale (`21_animali.md`). Struttura analoga a `mostri`.

- `talenti`: una entry per talento (sezioni in `06_talenti.md`).

```json
{
  "slug": "allerta",
  "nome": "Allerta",
  "categoria": "Talento di Origine",
  "prerequisiti": "",
  "benefici": ["Competenza all’Iniziativa", "Scambio di Iniziativa"],
  "contenuto": "#### Allerta\n*Talento di Origine*\n..."
}
```

## Normalizzazione
- Nei testi estratti, i trattini en/em (–, —) sono normalizzati quando necessario.
- Il titolo dei documenti (`titolo`) è l'`H1` se presente; in assenza si usa lo slug derivato dal filename.
 - Ogni entità conserva un campo `contenuto` con il markdown integrale della propria sezione per una resa fedele.
