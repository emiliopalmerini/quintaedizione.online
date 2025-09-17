package domain

import (
	"fmt"
)

// ClasseBuilder provides a fluent interface for building Classe entities
type ClasseBuilder struct {
	classe *Classe
	errors []error
}

// NewClasseBuilder creates a new ClasseBuilder with default values
func NewClasseBuilder() *ClasseBuilder {
	return &ClasseBuilder{
		classe: &Classe{
			// Initialize with sensible defaults
			CaratteristicaPrimaria:         []Caratteristica{},
			SalvezzeCompetenze:             []NomeCaratteristica{},
			ArmiCompetenze:                 []string{},
			ArmatureCompetenze:             []CompetenzaArmatura{},
			StrumentiCompetenze:            []StrumentoSlug{},
			EquipaggiamentoInizialeOpzioni: []EquipaggiamentoOpzione{},
			PrivilegiDiClasse:              []Privilegio{},
			Sottoclassi:                    []Sottoclasse{},
		},
		errors: []error{},
	}
}

// SetNome sets the nome field and generates the slug
func (b *ClasseBuilder) SetNome(nome string) *ClasseBuilder {
	if nome == "" {
		b.errors = append(b.errors, fmt.Errorf("nome cannot be empty"))
		return b
	}

	slug, err := NewSlug(nome)
	if err != nil {
		b.errors = append(b.errors, fmt.Errorf("failed to generate slug from nome '%s': %w", nome, err))
		return b
	}

	b.classe.Nome = nome
	b.classe.Slug = slug
	return b
}

// SetSottotitolo sets the sottotitolo field
func (b *ClasseBuilder) SetSottotitolo(sottotitolo string) *ClasseBuilder {
	b.classe.Sottotitolo = sottotitolo
	return b
}

// SetMarkdown sets the markdown field
func (b *ClasseBuilder) SetMarkdown(markdown string) *ClasseBuilder {
	b.classe.Markdown = markdown
	return b
}

// SetDadoVita sets the dado vita field
func (b *ClasseBuilder) SetDadoVita(dadoVita Dadi) *ClasseBuilder {
	b.classe.DadoVita = dadoVita
	return b
}

// SetCaratteristicaPrimaria sets the primary characteristics
func (b *ClasseBuilder) SetCaratteristicaPrimaria(caratteristiche []Caratteristica) *ClasseBuilder {
	if caratteristiche == nil {
		caratteristiche = []Caratteristica{}
	}
	b.classe.CaratteristicaPrimaria = caratteristiche
	return b
}

// AddCaratteristicaPrimaria adds a primary characteristic
func (b *ClasseBuilder) AddCaratteristicaPrimaria(caratteristica Caratteristica) *ClasseBuilder {
	b.classe.CaratteristicaPrimaria = append(b.classe.CaratteristicaPrimaria, caratteristica)
	return b
}

// SetSalvezzeCompetenze sets the saving throw proficiencies
func (b *ClasseBuilder) SetSalvezzeCompetenze(salvezze []NomeCaratteristica) *ClasseBuilder {
	if salvezze == nil {
		salvezze = []NomeCaratteristica{}
	}
	b.classe.SalvezzeCompetenze = salvezze
	return b
}

// AddSalvezzaCompetenza adds a saving throw proficiency
func (b *ClasseBuilder) AddSalvezzaCompetenza(salvezza NomeCaratteristica) *ClasseBuilder {
	b.classe.SalvezzeCompetenze = append(b.classe.SalvezzeCompetenze, salvezza)
	return b
}

// SetAbilitaCompetenzeOpzioni sets the skill proficiency options
func (b *ClasseBuilder) SetAbilitaCompetenzeOpzioni(abilita Scelta) *ClasseBuilder {
	b.classe.AbilitaCompetenzeOpzioni = abilita
	return b
}

// SetArmiCompetenze sets the weapon proficiencies
func (b *ClasseBuilder) SetArmiCompetenze(armi []string) *ClasseBuilder {
	if armi == nil {
		armi = []string{}
	}
	b.classe.ArmiCompetenze = armi
	return b
}

// AddArmaCompetenza adds a weapon proficiency
func (b *ClasseBuilder) AddArmaCompetenza(arma string) *ClasseBuilder {
	if arma != "" {
		b.classe.ArmiCompetenze = append(b.classe.ArmiCompetenze, arma)
	}
	return b
}

// SetArmatureCompetenze sets the armor proficiencies
func (b *ClasseBuilder) SetArmatureCompetenze(armature []CompetenzaArmatura) *ClasseBuilder {
	if armature == nil {
		armature = []CompetenzaArmatura{}
	}
	b.classe.ArmatureCompetenze = armature
	return b
}

// AddArmaturaCompetenza adds an armor proficiency
func (b *ClasseBuilder) AddArmaturaCompetenza(armatura CompetenzaArmatura) *ClasseBuilder {
	b.classe.ArmatureCompetenze = append(b.classe.ArmatureCompetenze, armatura)
	return b
}

// SetStrumentiCompetenze sets the tool proficiencies
func (b *ClasseBuilder) SetStrumentiCompetenze(strumenti []StrumentoSlug) *ClasseBuilder {
	if strumenti == nil {
		strumenti = []StrumentoSlug{}
	}
	b.classe.StrumentiCompetenze = strumenti
	return b
}

// AddStrumentoCompetenza adds a tool proficiency
func (b *ClasseBuilder) AddStrumentoCompetenza(strumento StrumentoSlug) *ClasseBuilder {
	b.classe.StrumentiCompetenze = append(b.classe.StrumentiCompetenze, strumento)
	return b
}

// SetEquipaggiamentoInizialeOpzioni sets the starting equipment options
func (b *ClasseBuilder) SetEquipaggiamentoInizialeOpzioni(opzioni []EquipaggiamentoOpzione) *ClasseBuilder {
	if opzioni == nil {
		opzioni = []EquipaggiamentoOpzione{}
	}
	b.classe.EquipaggiamentoInizialeOpzioni = opzioni
	return b
}

// AddEquipaggiamentoOpzione adds a starting equipment option
func (b *ClasseBuilder) AddEquipaggiamentoOpzione(opzione EquipaggiamentoOpzione) *ClasseBuilder {
	b.classe.EquipaggiamentoInizialeOpzioni = append(b.classe.EquipaggiamentoInizialeOpzioni, opzione)
	return b
}

// SetMulticlasse sets the multiclass requirements
func (b *ClasseBuilder) SetMulticlasse(multiclasse Multiclasse) *ClasseBuilder {
	b.classe.Multiclasse = multiclasse
	return b
}

// SetProgressioni sets the class progressions
func (b *ClasseBuilder) SetProgressioni(progressioni Progressioni) *ClasseBuilder {
	b.classe.Progressioni = progressioni
	return b
}

// SetMagia sets the magic information
func (b *ClasseBuilder) SetMagia(magia Magia) *ClasseBuilder {
	b.classe.Magia = magia
	return b
}

// SetPrivilegiDiClasse sets the class features
func (b *ClasseBuilder) SetPrivilegiDiClasse(privilegi []Privilegio) *ClasseBuilder {
	if privilegi == nil {
		privilegi = []Privilegio{}
	}
	b.classe.PrivilegiDiClasse = privilegi
	return b
}

// AddPrivilegioDiClasse adds a class feature
func (b *ClasseBuilder) AddPrivilegioDiClasse(privilegio Privilegio) *ClasseBuilder {
	b.classe.PrivilegiDiClasse = append(b.classe.PrivilegiDiClasse, privilegio)
	return b
}

// SetSottoclassi sets the subclasses
func (b *ClasseBuilder) SetSottoclassi(sottoclassi []Sottoclasse) *ClasseBuilder {
	if sottoclassi == nil {
		sottoclassi = []Sottoclasse{}
	}
	b.classe.Sottoclassi = sottoclassi
	return b
}

// AddSottoclasse adds a subclass
func (b *ClasseBuilder) AddSottoclasse(sottoclasse Sottoclasse) *ClasseBuilder {
	b.classe.Sottoclassi = append(b.classe.Sottoclassi, sottoclasse)
	return b
}

// SetRaccomandazioni sets the recommendations
func (b *ClasseBuilder) SetRaccomandazioni(raccomandazioni Raccomandazioni) *ClasseBuilder {
	b.classe.Raccomandazioni = raccomandazioni
	return b
}

// SetContenuto sets the content
func (b *ClasseBuilder) SetContenuto(contenuto string) *ClasseBuilder {
	b.classe.Contenuto = contenuto
	return b
}

// Build creates the final Classe instance after validation
func (b *ClasseBuilder) Build() (*Classe, error) {
	// Perform final validation
	if err := b.validate(); err != nil {
		return nil, err
	}

	// Return a copy to prevent external modification
	return b.classe, nil
}

// validate performs validation on the built classe
func (b *ClasseBuilder) validate() error {
	// Check for any errors accumulated during building
	if len(b.errors) > 0 {
		return fmt.Errorf("validation failed with %d errors: %v", len(b.errors), b.errors)
	}

	// Validate required fields
	if b.classe.Nome == "" {
		return fmt.Errorf("nome is required")
	}

	if b.classe.Slug == "" {
		return fmt.Errorf("slug is required (should be generated from nome)")
	}

	// Additional business logic validations can be added here
	return nil
}

// GetErrors returns any accumulated errors
func (b *ClasseBuilder) GetErrors() []error {
	return b.errors
}

// HasErrors returns true if there are any accumulated errors
func (b *ClasseBuilder) HasErrors() bool {
	return len(b.errors) > 0
}