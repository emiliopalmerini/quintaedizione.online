package parsers

import (
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// CreateDefaultWork creates the default Italian work items configuration
func CreateDefaultWork() []domain.WorkItem {
	return []domain.WorkItem{
		// Document pages (Italian)
		{
			Filename:   "ita/01_informazioni_legali.md",
			Collection: "documenti",
			Parser:     ParseDocument("01_informazioni_legali.md"),
		},
		{
			Filename:   "ita/02_giocare_il_gioco.md",
			Collection: "documenti", 
			Parser:     ParseDocument("02_giocare_il_gioco.md"),
		},
		{
			Filename:   "ita/03_creazione_personaggio.md",
			Collection: "documenti",
			Parser:     ParseDocument("03_creazione_personaggio.md"),
		},
		{
			Filename:   "ita/04_classi.md",
			Collection: "documenti",
			Parser:     ParseDocument("04_classi.md"),
		},
		{
			Filename:   "ita/05_origini_personaggio.md",
			Collection: "documenti",
			Parser:     ParseDocument("05_origini_personaggio.md"),
		},
		{
			Filename:   "ita/06_talenti.md",
			Collection: "documenti",
			Parser:     ParseDocument("06_talenti.md"),
		},
		{
			Filename:   "ita/07_equipaggiamento.md",
			Collection: "documenti",
			Parser:     ParseDocument("07_equipaggiamento.md"),
		},
		{
			Filename:   "ita/08_equipaggiamento_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("08_equipaggiamento_items.md"),
		},
		{
			Filename:   "ita/09_armi_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("09_armi_items.md"),
		},
		{
			Filename:   "ita/10_oggetti_magici_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("10_oggetti_magici_items.md"),
		},
		{
			Filename:   "ita/11_armatura_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("11_armatura_items.md"),
		},
		{
			Filename:   "ita/12_strumenti_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("12_strumenti_items.md"),
		},
		{
			Filename:   "ita/13_servizi_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("13_servizi_items.md"),
		},
		{
			Filename:   "ita/14_cavalcature_veicoli_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("14_cavalcature_veicoli_items.md"),
		},
		{
			Filename:   "ita/15_incantesimi.md",
			Collection: "documenti",
			Parser:     ParseDocument("15_incantesimi.md"),
		},
		{
			Filename:   "ita/16_incantesimi_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("16_incantesimi_items.md"),
		},
		{
			Filename:   "ita/17_glossario_regole.md",
			Collection: "documenti",
			Parser:     ParseDocument("17_glossario_regole.md"),
		},
		{
			Filename:   "ita/18_gameplay_toolbox.md",
			Collection: "documenti",
			Parser:     ParseDocument("18_gameplay_toolbox.md"),
		},
		{
			Filename:   "ita/19_mostri.md",
			Collection: "documenti",
			Parser:     ParseDocument("19_mostri.md"),
		},
		{
			Filename:   "ita/20_mostri_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("20_mostri_items.md"),
		},
		{
			Filename:   "ita/21_animali.md",
			Collection: "documenti",
			Parser:     ParseDocument("21_animali.md"),
		},

		// Structured data (Italian only)
		{
			Filename:   "ita/04_classi.md",
			Collection: "classi",
			Parser:     ParseClasses,
		},
		{
			Filename:   "ita/05_origini_personaggio.md",
			Collection: "backgrounds",
			Parser:     ParseBackgrounds,
		},
		{
			Filename:   "ita/09_armi_items.md",
			Collection: "armi",
			Parser:     ParseWeapons,
		},
		{
			Filename:   "ita/11_armatura_items.md",
			Collection: "armature",
			Parser:     ParseArmor,
		},
		{
			Filename:   "ita/12_strumenti_items.md",
			Collection: "strumenti",
			Parser:     ParseTools,
		},
		{
			Filename:   "ita/13_servizi_items.md",
			Collection: "servizi",
			Parser:     ParseServices,
		},
		{
			Filename:   "ita/08_equipaggiamento_items.md",
			Collection: "equipaggiamento",
			Parser:     ParseGear,
		},
		{
			Filename:   "ita/10_oggetti_magici_items.md",
			Collection: "oggetti_magici",
			Parser:     ParseMagicItems,
		},
		{
			Filename:   "ita/16_incantesimi_items.md",
			Collection: "incantesimi",
			Parser:     ParseSpells,
		},
		{
			Filename:   "ita/06_talenti.md",
			Collection: "talenti",
			Parser:     ParseFeats,
		},
		{
			Filename:   "ita/20_mostri_items.md",
			Collection: "mostri",
			Parser:     ParseMonstersMonster,
		},
		{
			Filename:   "ita/21_animali.md",
			Collection: "animali",
			Parser:     ParseMonstersAnimal,
		},
	}
}

// Note: ParseDocument is implemented in documents.go

func ParseClasses(lines []string) ([]map[string]interface{}, error) {
	// TODO: Implement classes parser
	return []map[string]interface{}{}, nil
}

func ParseBackgrounds(lines []string) ([]map[string]interface{}, error) {
	// TODO: Implement backgrounds parser
	return []map[string]interface{}{}, nil
}

func ParseWeapons(lines []string) ([]map[string]interface{}, error) {
	// TODO: Implement weapons parser
	return []map[string]interface{}{}, nil
}

func ParseArmor(lines []string) ([]map[string]interface{}, error) {
	// TODO: Implement armor parser
	return []map[string]interface{}{}, nil
}

func ParseTools(lines []string) ([]map[string]interface{}, error) {
	// TODO: Implement tools parser
	return []map[string]interface{}{}, nil
}

func ParseServices(lines []string) ([]map[string]interface{}, error) {
	// TODO: Implement services parser
	return []map[string]interface{}{}, nil
}

func ParseGear(lines []string) ([]map[string]interface{}, error) {
	// TODO: Implement gear parser
	return []map[string]interface{}{}, nil
}

func ParseMagicItems(lines []string) ([]map[string]interface{}, error) {
	// TODO: Implement magic items parser
	return []map[string]interface{}{}, nil
}

func ParseFeats(lines []string) ([]map[string]interface{}, error) {
	// TODO: Implement feats parser
	return []map[string]interface{}{}, nil
}

func ParseMonstersMonster(lines []string) ([]map[string]interface{}, error) {
	// TODO: Implement monsters parser
	return []map[string]interface{}{}, nil
}

func ParseMonstersAnimal(lines []string) ([]map[string]interface{}, error) {
	// TODO: Implement animals parser
	return []map[string]interface{}{}, nil
}