package domain

// GroupInfo defines metadata for a generator group.
type GroupInfo struct {
	ID          string
	Label       string
	Description string
	Order       int
}

// GroupRegistry maps group IDs to their metadata, ordered by display priority.
var GroupRegistry = []GroupInfo{
	{
		ID:    "core-adventure",
		Label: "Generatore di Avventure",
		Description: "Le tabelle di questa sezione possono aiutarti a generare un'avventura fantasy basata " +
			"sul concetto tradizionale di essere ingaggiati da un patrono o un altro PNG per intraprendere " +
			"una missione in un luogo specifico. Spesso queste avventure si svolgono in piccoli insediamenti " +
			"circondati da rovine antiche e tane di mostri ai confini della civiltà.\n\n" +
			"Usa queste tabelle insieme per generare e ispirare avventure complete, oppure usa le singole " +
			"tabelle per riempire i dettagli di altre avventure che crei o giochi. Questo generatore " +
			"(e in particolare la tabella dei Mostri del Dungeon e quella dei Tesori) è pensato per " +
			"personaggi dal 1° al 4° livello, ma può essere facilmente modificato per avventure di livello superiore.",
		Order: 1,
	},
	{
		ID:    "npc-generator",
		Label: "Generatore di PNG",
		Description: "I PNG danno vita ai mondi di gioco. Puoi usare i generatori di questa sezione per " +
			"creare rapidamente PNG da inserire nella tua partita, tirando sulle tabelle seguenti per " +
			"generare le caratteristiche di base. Per dare davvero vita al PNG, puoi poi modellare la " +
			"sua personalità e interpretazione ispirandoti a personaggi dei tuoi libri, serie TV o film " +
			"preferiti, cambiando genere e altri tratti per renderli originali.",
		Order: 2,
	},
	{
		ID:    "treasure-generator",
		Label: "Generatore di Tesori",
		Description: "Cumuli di monete, gemme scintillanti e potenti reliquie nascoste nelle profondità del " +
			"mondo attendono avventurieri abbastanza coraggiosi da cercarli. Questa sezione offre un semplice " +
			"set di tabelle e linee guida che ti permettono di assegnare rapidamente tesori per il tuo GDR " +
			"fantasy, e che funzionano bene insieme alle regole più dettagliate sui tesori del gioco.",
		Order: 3,
	},
	{
		ID:    "random-traps",
		Label: "Generatore di Trappole",
		Description: "Usa queste liste per generare trappole semplici o complesse, incorporando molteplici " +
			"caratteristiche, più danni energetici o condizioni.\n\n" +
			"Per generare una trappola semplice, tira solo sulla lista Tipo e sulla tabella Innesco. Per una " +
			"trappola più pericolosa, aggiungi un effetto dalla tabella Variante per dare un tocco unico ai " +
			"danni o imporre una condizione debilitante. Per una trappola davvero diabolica, puoi tirare sulla " +
			"tabella Variante e Tipo due volte, combinando le caratteristiche in combinazioni letali come " +
			"\"bolas soporifere e pilastri schiaccianti tonanti, innescati da un teschio demoniaco di onice\".",
		Order: 4,
	},
	{
		ID:    "random-chambers",
		Label: "Stanze Casuali",
		Description: "Queste pagine contengono liste di stanze per quindici \"dungeon\" comuni. Usa queste stanze " +
			"per riempire le sale di luoghi più grandi o per ispirare le tue idee. Arricchisci le stanze con " +
			"ulteriori dettagli da altre tabelle casuali secondo necessità.\n\n" +
			"Queste liste sono ordinate con le stanze tipiche in fondo e le stanze fantastiche o pericolose " +
			"in cima. Tira un dado più piccolo per stanze più tipiche e un dado più grande per stanze più " +
			"fantastiche o pericolose.",
		Order: 5,
	},
	{
		ID:    "random-items",
		Label: "Oggetti Casuali",
		Description: "Le seguenti liste ti permettono di generare reliquie e oggetti utili, da scoperte " +
			"mondane a potenti artefatti magici. Se vuoi creare un'arma magica interessante, ad esempio, " +
			"puoi tirare sulle tabelle Condizione, Origine, Arma ed Effetto Magico. Se vuoi solo un oggetto " +
			"mondano bizzarro, tira sulle tabelle Condizione, Origine e Oggetto Mondano senza aggiungere " +
			"alcun effetto.\n\n" +
			"Alcune strane reliquie potrebbero consentire un singolo uso di un potente incantesimo. Tira " +
			"sulle tabelle Condizione, Origine, Oggetto Mondano ed Effetto Magico per generare una reliquia " +
			"magica unica a uso singolo.",
		Order: 6,
	},
	{
		ID:    "town-events",
		Label: "Eventi in Città",
		Description: "Ogni volta che i personaggi entrano in una nuova città o iniziano una nuova sessione lì, " +
			"aggiungere dettagli e contesto all'ambientazione può aiutare a dare vita alle cose. Queste liste " +
			"di \"Eventi in Città\" aiutano a determinare cosa potrebbe succedere in città, come si sentono " +
			"gli abitanti, com'è il tempo e quale evento mondano o fantastico potrebbe aver luogo.",
		Order: 7,
	},
	{
		ID:    "dungeon-monsters",
		Label: "Mostri Casuali del Dungeon",
		Description: "Le seguenti tabelle ti permettono di selezionare casualmente mostri in base al \"livello " +
			"del dungeon\". Sebbene queste tabelle siano pensate per l'esplorazione classica dei dungeon, puoi " +
			"usarle per generare mostri incontrati casualmente in qualsiasi ambientazione — una rovina, una " +
			"vecchia chiesa, caverne, catacombe, la torre di un vecchio mago o qualche altra tana dimenticata.\n\n" +
			"Per usare queste tabelle, decidi prima su quale livello del dungeon si trovano i personaggi. " +
			"Poi tira un d20 e consulta la tabella del livello per determinare quale lista di mostri usare. " +
			"Vai quindi alla lista indicata e tira un altro d20 per determinare quale mostro potrebbe comparire. " +
			"Invece di usare i livelli del dungeon, puoi semplicemente scegliere la lista di mostri più " +
			"adatta alle circostanze.",
		Order: 8,
	},
}

// GetGroupInfo returns the metadata for a given group ID.
func GetGroupInfo(id string) (GroupInfo, bool) {
	for _, g := range GroupRegistry {
		if g.ID == id {
			return g, true
		}
	}
	return GroupInfo{}, false
}
