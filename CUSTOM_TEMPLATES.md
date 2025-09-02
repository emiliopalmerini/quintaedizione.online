# Template Personalizzati per @editor/

Questo documento descrive i template personalizzati creati per ogni collezione del D&D 5e SRD Editor, con filtri dinamici, metadati specializzati e funzionalità HTMX avanzate.

## 🎯 Overview

Sono stati creati **7 template specializzati** per le principali collezioni del SRD, ciascuno ottimizzato per la tipologia di contenuto e con funzionalità specifiche per migliorare l'esperienza utente.

## 📋 Template Creati

### 1. **Incantesimi** - `show_spells.html`
**Collezioni:** `spells`, `incantesimi`

**Filtri Rapidi:**
- Livello incantesimo (0-9)
- Scuola di magia
- Classi che possono lanciarlo

**Metadati Specializzati:**
- Dettagli completi (livello, scuola, tempo lancio, gittata, componenti, durata)
- Indicatori booleani (Rituale, Concentrazione)
- Liste classi cliccabili per filtri
- Livelli superiori in sezione dedicata

**Funzionalità Interattive:**
- Copia blocco incantesimo formattato
- Aggiunta al grimorio personale (localStorage)
- Pannello laterale con statistiche complete

### 2. **Mostri** - `show_monsters.html`
**Collezioni:** `monsters`, `mostri`

**Filtri Rapidi:**
- Grado di Sfida (GS/CR)
- Taglia creatura
- Tipo creatura
- Allineamento

**Metadati Specializzati:**
- Statistiche combattimento (CA, PF, velocità)
- Caratteristiche con modificatori calcolati
- Sezioni organizzate (tiri salvezza, abilità, resistenze, sensi, linguaggi)

**Funzionalità Interattive:**
- Copia blocco statistiche completo
- Aggiunta all'incontro corrente
- Tiro iniziativa con modificatori
- Pannello statistiche sticky

### 3. **Oggetti Magici** - `show_magic_items.html`
**Collezioni:** `magic_items`, `oggetti_magici`

**Filtri Rapidi:**
- Rarità (Comune → Artefatto)
- Tipo oggetto
- Requisito sintonizzazione

**Metadati Specializzati:**
- Rarità con colori distintivi
- Proprietà organizzate e peso/valore
- Sezioni specializzate (attivazione, maledizioni, varianti)
- Incantesimi contenuti cliccabili

**Funzionalità Interattive:**
- Copia scheda oggetto completa
- Aggiunta all'inventario personale
- Tiro cariche con pattern dadi
- Gestione sintonizzazione

### 4. **Armi** - `show_weapons.html`
**Collezioni:** `weapons`, `armi`

**Filtri Rapidi:**
- Categoria (Semplici/Da guerra)
- Maestria arma
- Proprietà specifiche

**Metadati Specializzati:**
- Tabella riassuntiva con tutte le statistiche
- Spiegazione proprietà automatica
- Consigli tattici per categoria
- Calcoli danno e modificatori

**Funzionalità Interattive:**
- Tiro danni normale e critico
- Tiro per colpire con risultati speciali
- Copia statistiche formattate
- Consigli build automatici

### 5. **Armature** - `show_armor.html`
**Collezioni:** `armor`, `armature`

**Filtri Rapidi:**
- Categoria (Leggera/Media/Pesante)
- Classe Armatura
- Requisiti forza

**Metadati Specializzati:**
- Calcolatore CA interattivo
- Tabella comparativa categorie
- Informazioni tattiche dettagliate
- Requisiti e limitazioni

**Funzionalità Interattive:**
- Calcolatore CA con bonus Destrezza
- Sistema confronto armature
- Consigli build per classe
- Copia statistiche complete

### 6. **Classi** - `show_classes.html`
**Collezioni:** `classes`, `classi`

**Filtri Rapidi:**
- Dado vita
- Incantatori (Sì/No)
- Abilità primarie

**Metadati Specializzati:**
- Panoramica ruolo e complessità
- Competenze organizzate (armature, armi, tiri salvezza)
- Lista incantesimi per incantatori
- Sottoclassi con descrizioni
- Build suggerite per principianti/esperti

**Funzionalità Interattive:**
- Generatore personaggio completo (4d6 drop lowest)
- Navigazione rapida sezioni
- Aggiunta ai preferiti
- Copia riassunto classe

### 7. **Strumenti** - `show_tools.html`
**Collezioni:** `tools`, `strumenti`

**Filtri Rapidi:**
- Categoria (Artigianato/Gioco/Musicale/etc.)
- Abilità associate
- Competenza richiesta

**Metadati Specializzati:**
- Tabelle CD per difficoltà
- Usi comuni per categoria
- Sinergie con altre competenze
- Tempi e costi creazione (artigianato)

**Funzionalità Interattive:**
- Tiro prova abilità con modificatori
- Suggerimenti usi casuali
- Simulatore creazione oggetti
- Calcolatore tempi/costi

### 8. **Background** - `show_background.html`
**Collezioni:** `backgrounds`
*Template esistente mantenuto*

## 🚀 Caratteristiche Tecniche

### HTMX Integration
- **Navigazione dinamica:** Tutti i filtri utilizzano HTMX per aggiornamenti senza reload
- **Target specifici:** `hx-target="#main-content"` per sostituire contenuto
- **History management:** `hx-push-url="true"` per URL navigabili
- **Indicatori loading:** Feedback visivo durante le richieste

### Responsive Design
- **Layout griglia:** Sidebar 1/4 + contenuto 3/4 su desktop
- **Mobile-first:** Riordino automatico su dispositivi piccoli
- **Sticky sidebar:** Informazioni sempre visibili durante scroll
- **Tabelle responsive:** Scroll orizzontale quando necessario

### Accessibility
- **ARIA labels:** Tutte le azioni e navigazioni etichettate
- **Keyboard navigation:** Supporto completo navigazione tastiera
- **Screen reader friendly:** Struttura semantica e descrizioni
- **Focus management:** Stati focus chiari e visibili

### Performance
- **Template caching:** Riutilizzo template engine Jinja2
- **LocalStorage:** Dati utente (inventario, preferiti) salvati localmente
- **Lazy loading:** Contenuti pesanti caricati on-demand
- **Minimal JavaScript:** Funzionalità essenziali, nessuna dipendenza pesante

## 📂 Struttura File

```
editor/templates/
├── show.html                    # Template base (fallback)
├── show_spells.html            # Incantesimi
├── show_monsters.html          # Mostri  
├── show_magic_items.html       # Oggetti magici
├── show_weapons.html           # Armi
├── show_armor.html             # Armature
├── show_classes.html           # Classi
├── show_tools.html             # Strumenti
├── show_background.html        # Background (esistente)
├── show_class.html             # Classi (deprecato)
└── show_template_overview.html # Panoramica template
```

## 🔧 Configurazione Router

Il router in `routers/pages.py` è stato aggiornato per mappare automaticamente le collezioni ai template corretti:

```python
template_mapping = {
    "classi": "show_classes.html",
    "classes": "show_classes.html", 
    "spells": "show_spells.html",
    "incantesimi": "show_spells.html",
    "magic_items": "show_magic_items.html",
    "oggetti_magici": "show_magic_items.html",
    "monsters": "show_monsters.html", 
    "mostri": "show_monsters.html",
    "weapons": "show_weapons.html",
    "armi": "show_weapons.html",
    "armor": "show_armor.html",
    "armature": "show_armor.html", 
    "tools": "show_tools.html",
    "strumenti": "show_tools.html",
}
```

## 🎨 Sistema Colori

Ogni collezione ha colori distintivi:
- **Incantesimi:** Blu (`text-blue-800`, `bg-blue-50`)
- **Mostri:** Rosso (`text-red-800`, `bg-red-50`)
- **Oggetti Magici:** Viola (`text-purple-800`, `bg-purple-50`)
- **Armi:** Arancione (`text-orange-800`, `bg-orange-50`)
- **Armature:** Teal (`text-teal-800`, `bg-teal-50`)
- **Classi:** Indaco (`text-indigo-800`, `bg-indigo-50`)
- **Strumenti:** Smeraldo (`text-emerald-800`, `bg-emerald-50`)

## 🧪 Testing

Per testare i filtri e funzionalità:

1. **Filtri HTMX:** Verificare che clicking sui filtri rapidi aggiorni la lista senza reload
2. **Funzioni JavaScript:** Testare dadi, calcolatori e copia negli appunti
3. **LocalStorage:** Verificare salvataggio inventario, preferiti e grimorio
4. **Responsive:** Testare layout su mobile, tablet e desktop
5. **Accessibilità:** Navigazione con Tab e screen reader

## 🔮 Funzionalità Future

Possibili estensioni:
- **Export PDF:** Esportazione schede personaggio/creature
- **Confronti avanzati:** Tabelle comparative multi-oggetto  
- **Builder integrati:** Creatori personaggio/incontro completi
- **Sync cloud:** Sincronizzazione dati utente
- **Temi personalizzabili:** Dark mode e varianti colore
- **Plugin system:** API per estensioni community

---

**Autore:** Claude AI Assistant  
**Data:** Gennaio 2025  
**Versione Editor:** Due Draghi 5e SRD v2024