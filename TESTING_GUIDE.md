# Guida al Testing dell'Architettura Esagonale

Questa guida fornisce istruzioni complete per testare tutte le modifiche implementate, inclusa l'architettura esagonale e le nuove entità D&D 5e.

## 🚀 Panoramica delle Modifiche

### Cosa È Stato Implementato

1. **Architettura Esagonale Completa**
   - Parser Service (write-side) con CQRS
   - Editor Service (read-side) con query ottimizzate
   - Shared Domain con tutte le entità D&D 5e

2. **Nuove Entità Domain**
   - ✅ **13 tipologie di entità** complete (classi, incantesimi, mostri, equipaggiamento, etc.)
   - ✅ **Value Objects** immutabili con validazione
   - ✅ **Repository Pattern** con interfacce astratte
   - ✅ **Event-Driven Architecture** con domain events

3. **Infrastruttura**
   - ✅ **Dependency Injection** containers per entrambi i servizi
   - ✅ **CQRS Query Models** ottimizzati per UI
   - ✅ **Validation Services** con business rules

## 🧪 Metodi di Testing

### 1. **Testing Manuale (Raccomandato)**

#### A. Verifica Servizi Attivi

```bash
# Avvia tutto il sistema
make up

# Verifica stato containers
make logs

# Check dei servizi
curl -I http://localhost:8000/    # Editor
curl -I http://localhost:8100/    # Parser
```

**Expected Results:**
- Editor: HTTP 200, contenuto HTML
- Parser: HTTP 200, interfaccia parser

#### B. Test Database Connectivity

```bash
# Test connessione database via Editor
curl -s http://localhost:8000/classi | head -20

# Test connessione database via Parser
curl -s http://localhost:8100/test-conn
```

**Expected Results:**
- Editor: Pagina classi caricata con successo
- Parser: Risultato test connessione MongoDB

#### C. Test Funzionalità Tradizionali

Naviga manualmente a:

1. **Editor Tradizionale**: http://localhost:8000/
   - ✅ Homepage caricata
   - ✅ Menu collezioni visibile
   - ✅ Ricerca funzionante

2. **Parser Tradizionale**: http://localhost:8100/
   - ✅ Interfaccia parser visibile
   - ✅ Form di configurazione presente

#### D. Test Architettura Esagonale

**IMPORTANTE**: Le route esagonali potrebbero non essere attive se l'Editor usa un main.py diverso.

Per attivarle:

```bash
# 1. Verifica quale main.py viene usato
docker exec dnd-editor ls -la main*

# 2. Se necessario, ferma e rebuilda
make down
make build-editor
make up
```

Poi testa:

- **Hexagonal Demo**: http://localhost:8000/hex/ (potrebbe dare 404)
- **Hexagonal Classes**: http://localhost:8000/hex/classes (potrebbe dare 404)

### 2. **Testing Programmatico**

#### A. Test Importazioni Domain Model

```bash
# Testa importazioni base
python3 -c "
from shared_domain.entities import DndClass, ClassId, Ability, HitDie
print('✅ Core entities imported')
"

# Test creazione entità
python3 -c "
from shared_domain.entities import DndClass, ClassId, Ability, HitDie
class_obj = DndClass(
    id=ClassId('guerriero'),
    name='Guerriero',
    primary_ability=Ability.FORZA,
    hit_die=HitDie(10),
    version='1.0',
    source='Test'
)
print('✅ Entity created:', class_obj.name)
"
```

#### B. Test Container Access (Avanzato)

```bash
# Testa accesso ai container (richiede dipendenze)
python3 -c "
import sys
sys.path.append('.')
try:
    from editor.infrastructure.container import get_container
    container = get_container()
    print('✅ Editor container accessible')
except Exception as e:
    print('❌ Editor container error:', e)
"
```

### 3. **Test Database Operations**

#### A. Verifica Collezioni

```bash
# Lista collezioni MongoDB (se hai accesso diretto)
docker exec dnd-mongo mongosh dnd --eval "db.runCommand('listCollections')"

# Oppure attraverso l'Editor
curl -s "http://localhost:8000/classi" | grep -o "Barbaro\|Guerriero\|Mago" | head -3
```

#### B. Test Ricerca

```bash
# Test ricerca classi
curl -s "http://localhost:8000/classi?q=barbaro" | grep -i barbaro

# Test ricerca incantesimi
curl -s "http://localhost:8000/incantesimi?q=fireball" | grep -i fireball
```

### 4. **Test Parser Operations**

#### A. Test Connessione Database

```bash
# Test connettività parser
curl -X POST \
  http://localhost:8100/test-conn \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "mongo_instance=mongo:27017&db_name=dnd"
```

#### B. Test Dry-Run Parsing

```bash
# Test parsing in modalità dry-run
curl -X POST \
  http://localhost:8100/run \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "input_dir=data&mongo_instance=mongo:27017&db_name=dnd&dry_run=on&selected=0"
```

## 📋 Checklist di Verifica

### ✅ Sistema Base
- [ ] Editor accessibile su porta 8000
- [ ] Parser accessibile su porta 8100
- [ ] MongoDB attivo e connesso
- [ ] Collezioni principali caricate (classi, incantesimi, etc.)

### ✅ Architettura Tradizionale
- [ ] Homepage Editor funzionante
- [ ] Navigazione collezioni funzionante
- [ ] Ricerca testuale funzionante
- [ ] Parser interface caricata
- [ ] Test connessione database parser OK

### ✅ Architettura Esagonale (se attivata)
- [ ] Route `/hex/` accessibile
- [ ] Route `/hex/classes` funzionante
- [ ] Dependency Injection containers funzionanti
- [ ] Domain model importabile
- [ ] Validation services funzionanti

### ✅ Nuove Entità
- [ ] Domain entities importabili
- [ ] Value objects validano correttamente
- [ ] Repository interfaces definite
- [ ] Query models CQRS disponibili

## 🐛 Troubleshooting

### Problema: Route Esagonali 404

**Causa**: L'Editor potrebbe usare un main.py diverso che non include le route esagonali.

**Soluzione**:
```bash
# Verifica quale main viene usato
docker exec dnd-editor cat main.py | head -10

# Se necessario, rebuilda con le modifiche
make down && make build-editor && make up
```

### Problema: Import Error su Domain Model

**Causa**: Dipendenze mancanti o conflitti di import.

**Soluzione**:
```bash
# Test import semplificato
python3 -c "from shared_domain.entities import Ability; print('✅ Basic import OK')"

# Se fallisce, verifica path
python3 -c "import sys; print(sys.path)"
```

### Problema: Container Access Errors

**Causa**: Dipendenze Python mancanti (pymongo, motor, etc.).

**Soluzione**:
```bash
# Installa dipendenze se necessario
pip install pymongo motor fastapi

# Oppure testa all'interno del container
docker exec dnd-editor python3 -c "from shared_domain.entities import DndClass; print('✅')"
```

## 🎯 Expected Results Summary

### ✅ **Successo Minimo** (Core funzionante)
- Editor e Parser accessibili
- Database connesso
- Navigazione base funzionante
- Domain model importabile

### ✅ **Successo Completo** (Tutto funzionante)
- Tutto del successo minimo +
- Route esagonali attive
- Container DI funzionanti
- Tutte le 13 entità accessibili
- Validation services operativi

### ⚠️ **Stato Parziale** (Cosa aspettarsi)
Dato che si tratta di un refactoring importante, è normale che:
- Route esagonali potrebbero non essere attive inizialmente
- Alcuni import potrebbero fallire per dipendenze mancanti
- Container DI potrebbero richiedere configurazione aggiuntiva

L'importante è che **il sistema base continui a funzionare** e le **nuove entità siano accessibili**.

## 📚 Test Steps Raccomandati

### Per Sviluppatori

1. **Quick Test**: `make up && curl http://localhost:8000/`
2. **Domain Test**: `python3 -c "from shared_domain import SRDDomainModel; print(SRDDomainModel.get_domain_info())"`
3. **Full Test**: Eseguire tutti i test manuali sopra

### Per Deployment

1. Eseguire testing automatizzato se disponibile
2. Verificare tutte le collezioni caricate
3. Test ricerca e navigazione
4. Verificare performance database

### Per Utenti Finali

1. Navigare homepage: http://localhost:8000/
2. Testare ricerca su diverse collezioni
3. Verificare caricamento contenuti
4. Testare copia rapida contenuti

Il sistema dovrebbe essere **retrocompatibile**, quindi tutte le funzionalità esistenti devono continuare a funzionare normalmente mentre le nuove funzionalità dell'architettura esagonale vengono gradualmente integrate.