package factories

import (
	"fmt"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
)

// ParserFactory creates language-aware parsers using the Factory pattern
type ParserFactory struct {
	registry        *parsers.Registry
	languageConfigs map[parsers.LanguageCode]*parsers.LanguageConfig
	logger          parsers.Logger
}

// NewParserFactory creates a new parser factory
func NewParserFactory(registry *parsers.Registry, logger parsers.Logger) *ParserFactory {
	return &ParserFactory{
		registry:        registry,
		languageConfigs: make(map[parsers.LanguageCode]*parsers.LanguageConfig),
		logger:          logger,
	}
}

// LoadLanguageConfig loads and stores language-specific configuration
func (f *ParserFactory) LoadLanguageConfig(language parsers.LanguageCode, configPath string) error {
	config, err := LoadLanguageConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load language config for %s: %w", language, err)
	}
	
	f.languageConfigs[language] = config
	return nil
}

// LoadAllLanguageConfigs loads all supported language configurations
func (f *ParserFactory) LoadAllLanguageConfigs() error {
	configs, err := LoadAllLanguageConfigs()
	if err != nil {
		return err
	}
	
	f.languageConfigs = configs
	return nil
}

// CreateParser creates a parser for specific content type and language
func (f *ParserFactory) CreateParser(contentType parsers.ContentType, language parsers.LanguageCode) (parsers.ParsingStrategy, error) {
	config, exists := f.languageConfigs[language]
	if !exists {
		return nil, fmt.Errorf("language config not found: %s", language)
	}

	switch contentType {
	case parsers.ContentTypeSpells:
		return f.createSpellsParser(language, config)
	case parsers.ContentTypeMonsters:
		return f.createMonstersParser(language, config)
	case parsers.ContentTypeClasses:
		return f.createClassesParser(language, config)
	case parsers.ContentTypeWeapons:
		return f.createWeaponsParser(language, config)
	case parsers.ContentTypeArmor:
		return f.createArmorParser(language, config)
	case parsers.ContentTypeBackgrounds:
		return f.createBackgroundsParser(language, config)
	case parsers.ContentTypeMagicItems:
		return f.createMagicItemsParser(language, config)
	case parsers.ContentTypeFeats:
		return f.createFeatsParser(language, config)
	case parsers.ContentTypeAnimals:
		return f.createAnimalsParser(language, config)
	case parsers.ContentTypeDocuments:
		return f.createDocumentsParser(language, config)
	case parsers.ContentTypeGear:
		return f.createEquipmentParser(language, config)
	case parsers.ContentTypeTools:
		return f.createEquipmentParser(language, config) // same as gear
	case parsers.ContentTypeServices:
		return f.createEquipmentParser(language, config) // same as gear
	default:
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
}

// RegisterParsersForLanguage creates and registers all parsers for a language
func (f *ParserFactory) RegisterParsersForLanguage(language parsers.LanguageCode) error {
	if _, exists := f.languageConfigs[language]; !exists {
		return fmt.Errorf("language config not found: %s", language)
	}

	// Get all supported content types
	contentTypes := parsers.GetAllContentTypes()

	for _, contentType := range contentTypes {
		parser, err := f.CreateParser(contentType, language)
		if err != nil {
			f.logger.Info("Failed to create parser for %s/%s: %v", contentType, language, err)
			continue
		}

		// Register with unique key (contentType + language)
		key := fmt.Sprintf("%s_%s", contentType, language)
		f.registry.Register(key, parser)
		
		f.logger.Info("Registered parser: %s", key)
	}

	return nil
}

// RegisterAllParsers registers parsers for all loaded languages
func (f *ParserFactory) RegisterAllParsers() error {
	for language := range f.languageConfigs {
		if err := f.RegisterParsersForLanguage(language); err != nil {
			return fmt.Errorf("failed to register parsers for %s: %w", language, err)
		}
	}
	return nil
}

// ===== FACTORY METHODS FOR SPECIFIC PARSERS =====

func (f *ParserFactory) createSpellsParser(language parsers.LanguageCode, config *parsers.LanguageConfig) (parsers.ParsingStrategy, error) {
	return &parsers.SpellsStrategy{
		BaseParser: parsers.NewBaseParserWithLanguage(
			parsers.ContentTypeSpells,
			"Spells Parser",
			fmt.Sprintf("Parses D&D 5e spells from %s SRD markdown content", language),
			language,
			config,
		).WithLogger(f.logger),
	}, nil
}

func (f *ParserFactory) createMonstersParser(language parsers.LanguageCode, config *parsers.LanguageConfig) (parsers.ParsingStrategy, error) {
	return &parsers.MonstersStrategy{
		BaseParser: parsers.NewBaseParserWithLanguage(
			parsers.ContentTypeMonsters,
			"Monsters Parser",
			fmt.Sprintf("Parses D&D 5e monsters from %s SRD markdown content", language),
			language,
			config,
		).WithLogger(f.logger),
	}, nil
}

func (f *ParserFactory) createClassesParser(language parsers.LanguageCode, config *parsers.LanguageConfig) (parsers.ParsingStrategy, error) {
	return &parsers.ClassesStrategy{
		BaseParser: parsers.NewBaseParserWithLanguage(
			parsers.ContentTypeClasses,
			"Classes Parser",
			fmt.Sprintf("Parses D&D 5e classes from %s SRD markdown content", language),
			language,
			config,
		).WithLogger(f.logger),
	}, nil
}

func (f *ParserFactory) createWeaponsParser(language parsers.LanguageCode, config *parsers.LanguageConfig) (parsers.ParsingStrategy, error) {
	return &parsers.WeaponsStrategy{
		BaseParser: parsers.NewBaseParserWithLanguage(
			parsers.ContentTypeWeapons,
			"Weapons Parser",
			fmt.Sprintf("Parses D&D 5e weapons from %s SRD markdown content", language),
			language,
			config,
		).WithLogger(f.logger),
	}, nil
}

func (f *ParserFactory) createArmorParser(language parsers.LanguageCode, config *parsers.LanguageConfig) (parsers.ParsingStrategy, error) {
	return &parsers.ArmorStrategy{
		BaseParser: parsers.NewBaseParserWithLanguage(
			parsers.ContentTypeArmor,
			"Armor Parser",
			fmt.Sprintf("Parses D&D 5e armor from %s SRD markdown content", language),
			language,
			config,
		).WithLogger(f.logger),
	}, nil
}

func (f *ParserFactory) createBackgroundsParser(language parsers.LanguageCode, config *parsers.LanguageConfig) (parsers.ParsingStrategy, error) {
	return &parsers.BackgroundsStrategy{
		BaseParser: parsers.NewBaseParserWithLanguage(
			parsers.ContentTypeBackgrounds,
			"Backgrounds Parser",
			fmt.Sprintf("Parses D&D 5e backgrounds from %s SRD markdown content", language),
			language,
			config,
		).WithLogger(f.logger),
	}, nil
}

func (f *ParserFactory) createMagicItemsParser(language parsers.LanguageCode, config *parsers.LanguageConfig) (parsers.ParsingStrategy, error) {
	return &parsers.MagicItemsStrategy{
		BaseParser: parsers.NewBaseParserWithLanguage(
			parsers.ContentTypeMagicItems,
			"Magic Items Parser",
			fmt.Sprintf("Parses D&D 5e magic items from %s SRD markdown content", language),
			language,
			config,
		).WithLogger(f.logger),
	}, nil
}

func (f *ParserFactory) createFeatsParser(language parsers.LanguageCode, config *parsers.LanguageConfig) (parsers.ParsingStrategy, error) {
	return &parsers.FeatsStrategy{
		BaseParser: parsers.NewBaseParserWithLanguage(
			parsers.ContentTypeFeats,
			"Feats Parser",
			fmt.Sprintf("Parses D&D 5e feats from %s SRD markdown content", language),
			language,
			config,
		).WithLogger(f.logger),
	}, nil
}

func (f *ParserFactory) createAnimalsParser(language parsers.LanguageCode, config *parsers.LanguageConfig) (parsers.ParsingStrategy, error) {
	return &parsers.AnimalsStrategy{
		BaseParser: parsers.NewBaseParserWithLanguage(
			parsers.ContentTypeAnimals,
			"Animals Parser",
			fmt.Sprintf("Parses D&D 5e animals from %s SRD markdown content", language),
			language,
			config,
		).WithLogger(f.logger),
	}, nil
}

func (f *ParserFactory) createDocumentsParser(language parsers.LanguageCode, config *parsers.LanguageConfig) (parsers.ParsingStrategy, error) {
	return &parsers.DocumentsStrategy{
		BaseParser: parsers.NewBaseParserWithLanguage(
			parsers.ContentTypeDocuments,
			"Documents Parser",
			fmt.Sprintf("Parses D&D 5e documents from %s SRD markdown content", language),
			language,
			config,
		).WithLogger(f.logger),
	}, nil
}

func (f *ParserFactory) createEquipmentParser(language parsers.LanguageCode, config *parsers.LanguageConfig) (parsers.ParsingStrategy, error) {
	return &parsers.EquipmentStrategy{
		BaseParser: parsers.NewBaseParserWithLanguage(
			parsers.ContentTypeGear,
			"Equipment Parser",
			fmt.Sprintf("Parses D&D 5e equipment from %s SRD markdown content", language),
			language,
			config,
		).WithLogger(f.logger),
	}, nil
}