package filters

import (
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/filters"
)

// RegisterDefaultFilters declares which fields are exposed as facets per
// collection. EnumValues are populated from the loaded data at startup by
// DeriveEnumValues, so adding a new value to the JSON automatically surfaces
// it in the UI.
func RegisterDefaultFilters(registry *InMemoryFilterRegistry) {
	defs := []filters.FilterDefinition{
		// Incantesimi
		{Name: "scuola", FieldPath: "scuola", Collections: []collections.CollectionName{collections.Incantesimi}, Description: "Scuola di magia"},
		{Name: "livello", FieldPath: "livello", Collections: []collections.CollectionName{collections.Incantesimi}, Description: "Livello dell'incantesimo"},
		{Name: "classe", FieldPath: "classe", Collections: []collections.CollectionName{collections.Incantesimi}, Description: "Classe che può lanciare l'incantesimo"},

		// Mostri
		{Name: "tipo", FieldPath: "tipo", Collections: []collections.CollectionName{collections.Mostri}, Description: "Tipo di creatura"},
		{Name: "taglia", FieldPath: "taglia", Collections: []collections.CollectionName{collections.Mostri}, Description: "Taglia della creatura"},
		{Name: "grado_sfida", FieldPath: "grado_sfida", Collections: []collections.CollectionName{collections.Mostri}, Description: "Grado di sfida"},

		// Oggetti Magici
		{Name: "rarita", FieldPath: "rarita", Collections: []collections.CollectionName{collections.OggettiMagici}, Description: "Rarità dell'oggetto magico"},
		{Name: "tipo_oggetto", FieldPath: "tipo_base", Collections: []collections.CollectionName{collections.OggettiMagici}, Description: "Tipo di oggetto magico"},

		// Equipaggiamenti
		{Name: "categoria", FieldPath: "categoria", Collections: []collections.CollectionName{collections.Equipaggiamenti}, Description: "Categoria dell'equipaggiamento"},

		// Glossario
		{Name: "categoria", FieldPath: "categoria", Collections: []collections.CollectionName{collections.Glossario}, Description: "Categoria del termine"},

		// Specie
		{Name: "tipo_creatura", FieldPath: "tipo_creatura", Collections: []collections.CollectionName{collections.Specie}, Description: "Tipo di creatura"},
	}
	for _, d := range defs {
		registry.AddFilter(d)
	}
}

// RegisterEditionFilter adds the edition filter to the registry when multiple
// sources are loaded. With a single source, no filter is added.
func RegisterEditionFilter(registry *InMemoryFilterRegistry, sources []domain.Source) {
	if len(sources) <= 1 {
		return
	}

	enumValues := make([]string, 0, len(sources))
	for _, src := range sources {
		enumValues = append(enumValues, src.ShortName)
	}

	registry.AddFilter(filters.FilterDefinition{
		Name:        "_source_short",
		FieldPath:   "_source_short",
		Collections: nil, // applies to all collections
		EnumValues:  enumValues,
		Description: "Edizione",
	})
}
