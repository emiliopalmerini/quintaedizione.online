# 0004 – Hexagonal Architecture (Dominio)

**Status:** Accepted

## Context

- Adottiamo un’architettura esagonale per separare dominio, applicazione e adattatori.
- Serve esplicitare le entità di dominio (aggregate e value object) derivate dallo SRD per mantenere il core indipendente da persistenza e UI.

## Decisione

- Definiamo gli aggregate root e i principali value object del dominio estraendoli dal data model SRD.
- Ogni aggregate ha `slug` come identificatore naturale e conserva `contenuto` (markdown integrale) per il rendering fedele.
- I repository sono porte (ports) specifiche per aggregate; gli adattatori mapperanno 1:1 con le collection dati.

## Entità di dominio

Aggregate root (uno per collection SRD):

- Documento
  - Chiave: `slug` (più `pagina` per ordinamento).
  - Campi principali: `titolo`, `paragrafi[]` (con `numero`, `titolo`, `corpo.markdown`, `sottoparagrafi[]`), `contenuto`.
- Classe
  - Chiave: `slug`.
  - Campi principali: `nome`, `sottotitolo`, `markdown` (testo completo), `dado_vita`, `caratteristica_primaria` (enum abilità),
    competenze (`salvezze_competenze`, `abilità_competenze_opzioni`, `armi_competenze`, `armature_competenze`, `strumenti_competenze`),
    `equipaggiamento_iniziale_opzioni[]`, `multiclasse{}`, `progressioni{}`, `magia{}`, `tabella_livelli[]`, `privilegi_di_classe[]`,
    `sottoclassi[]`, `liste_incantesimi`, `raccomandazioni{}`, `contenuto`.
  - Sotto‑entità: PrivilegioDiClasse, Sottoclasse (con PrivilegioSottoclasse), TabellaLivelliRiga, Progressioni, Magia.
- Background (Origine)
  - Chiave: `slug`.
  - Campi principali: `nome`, `punteggi_caratteristica[]`, `talento`, competenze (`abilità_competenze`, `strumenti_competenze`),
    `equipaggiamento_iniziale_opzioni[]`, `contenuto`.
- Incantesimo
  - Chiave: `slug`.
  - Campi principali: `nome`, `livello` (0–9), `scuola`, `classi[]`, `lancio{tempo,gittata,componenti,materiali?,durata}`, `contenuto`.
- Arma
  - Chiave: `slug`.
  - Campi principali: `nome`, `costo{}`, `peso{}`, `danno` (es. "1d4 Perforante"), `categoria`, `proprieta[]`, `maestria`, `gittata{}`, `contenuto`.
- Armatura
  - Chiave: `slug`.
  - Campi principali: `nome`, `costo{}`, `peso{}`, `categoria`, `classe_armatura{base,modificatore_des,limite_des?}`,
    `forza_richiesta?`, `svantaggio_furtivita`, `contenuto`.
- Strumento
  - Chiave: `slug`.
  - Campi principali: `nome`, `costo{}`, `peso{}`, `categoria`, `abilità_associata`, `utilizzi[]`, `creazioni[]`, `contenuto`.
- Servizio
  - Chiave: `slug`.
  - Campi principali: `nome`, `costo{}`, `categoria`, `descrizione`, `contenuto`.
- Equipaggiamento (generico)
  - Chiave: `slug`.
  - Campi principali: `nome`, `costo{}`, `peso{}`, attributi opzionali specifici (es. `capacita{}`), `note?`, `contenuto`.
- OggettoMagico
  - Chiave: `slug`.
  - Campi principali: `nome`, `tipo`, `rarita`, `sintonizzazione` (bool/testo), `contenuto`.
- Mostro
  - Chiave: `slug`.
  - Campi principali: `nome`, `taglia`, `tipo`, `allineamento`, `ac`, `hp`, `velocita`, `caratteristiche{}`,
    `sensibilita{}`, `tiri_salvezza{}`, `abilità{}`, `immunita{danni[],condizioni[]}`, `tratti[]`, `azioni[]`, `reazioni[]`,
    `azioni_leggendarie[]`, `incantesimi{cd,attacco,lista[]}`, `contenuto`.
- Animale
  - Chiave: `slug`.
  - Campi principali: `nome`, `taglia`, `tipo`, `ac`, `hp`, `velocita`, `caratteristiche{}`, `tratti[]`, `azioni[]`, `contenuto`.
- Talento
  - Chiave: `slug`.
  - Campi principali: `nome`, `categoria`, `prerequisiti?`, `benefici[]`, `contenuto`.
- CavalcaturaOVeicolo
  - Chiave: `slug`.
  - Campi principali: `nome`, `tipo`, `costo{}`, `velocita{}`, `capacita_carico{}`, `equipaggio?`, `passeggeri?`, `ca?`, `pf?`,
    `soglia_danni?`, `descrizione`, `contenuto`.

### Value Object comuni

- Slug: identificatore testuale normalizzato, unico per aggregate.
- Costo: `{ valore:number, valuta:string }`.
- Peso: `{ valore:number, unita:string }`.
- Gittata: `{ normale:string, lunga?:string }` o testo per incantesimi.
- ClasseArmatura: `{ base:number, modificatore_des:boolean, limite_des?:number }`.
- Caratteristiche: `{ str,dex,con,int,wis,cha }` (interi).
- Livello: intero con vincoli di dominio (1–20 per progressioni; 0–9 per incantesimi).
- OpzioniDiScelta: strutture tipo `{ scegli:n, opzioni:string[] }` per competenze/equipaggiamento.
- Markdown: `contenuto`/`markdown` con il testo integrale per rendering fedele.

## Porte (Ports) del dominio

- Repository per ciascun aggregate: `DocumentiRepository`, `ClassiRepository`, `BackgroundsRepository`, `IncantesimiRepository`,
  `ArmiRepository`, `ArmatureRepository`, `StrumentiRepository`, `ServiziRepository`, `EquipaggiamentoRepository`,
  `OggettiMagiciRepository`, `MostriRepository`, `AnimaliRepository`, `TalentiRepository`, `CavalcatureVeicoliRepository`.
- Le porte espongono operazioni tipiche: lookup per `slug`, liste ordinate alfabeticamente, filtri specifici (es. per livello incantesimi).

## Conseguenze

- Modello coerente con lo SRD e indipendente da persistenza/transport.
- Adattatori possono mappare 1:1 le collection ai repository senza logica di business.
- Le UI possono usare `contenuto` per render fidele e i campi strutturati per feature interattive e filtri.
