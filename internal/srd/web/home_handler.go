package web

import (
	"net/http"
	"sort"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/models"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/mappers"
)

// HomeHandler handles home page requests.
type HomeHandler struct {
	*baseHandler
}

// handleHome renders the home page with collection statistics.
func (h *HomeHandler) handleHome(w http.ResponseWriter, r *http.Request) {
	collectionStats, err := h.contentService.GetCollectionStats(r.Context())
	if err != nil {
		collectionStats = h.getDefaultCollections()
	}

	// Build a lookup of collection models by name
	collectionMap := make(map[string]models.Collection, len(collectionStats))
	total := int64(0)
	allCollections := make([]models.Collection, 0, len(collectionStats))

	for _, col := range collectionStats {
		name := mappers.GetString(col, "collection", "")
		count := mappers.GetInt64(col, "count", 0)

		mc := models.Collection{
			Name:  name,
			Count: count,
		}
		mc.Label = h.getCollectionTitle(name)
		if info, ok := collections.GetInfo(name); ok {
			mc.Description = info.Description
		}

		if count > 0 {
			total += count
		}

		collectionMap[name] = mc
		allCollections = append(allCollections, mc)
	}

	// Build grouped collections in display order
	domainGroups := collections.GetGroups()
	groups := make([]models.CollectionGroup, 0, len(domainGroups))
	for _, dg := range domainGroups {
		mg := models.CollectionGroup{
			Slug:        dg.Slug,
			Label:       dg.Label,
			Description: dg.Description,
		}
		for _, cn := range dg.Collections {
			if mc, ok := collectionMap[cn.String()]; ok {
				mg.Collections = append(mg.Collections, mc)
				mg.Total += mc.Count
			}
		}
		groups = append(groups, mg)
	}

	data := models.HomePageData{
		PageData: models.PageData{
			Title:       "quintaedizione.online",
			Description: "Il Fantastico Visualizzatore di SRD (5e 2024)",
			TotalItems:  total,
		},
		Collections: allCollections,
		Groups:      groups,
		Total:       total,
		Editions:    2,
	}

	content, err := h.templateEngine.RenderHome(r.Context(), data)
	if err != nil {
		h.ErrorResponse(w, r, err, "Errore nel rendering della pagina home")
		return
	}

	h.renderHTML(w, content, "home")
}

// getDefaultCollections returns default collection data when stats are unavailable.
func (h *HomeHandler) getDefaultCollections() []map[string]any {
	allCollections := collections.GetAllWithInfo()
	result := make([]map[string]any, 0, len(allCollections))

	for _, info := range allCollections {
		result = append(result, map[string]any{
			"name":  info.Name.String(),
			"label": info.Title,
			"count": 0,
		})
	}

	// Sort alphabetically by name
	sort.Slice(result, func(i, j int) bool {
		return result[i]["name"].(string) < result[j]["name"].(string)
	})

	return result
}
