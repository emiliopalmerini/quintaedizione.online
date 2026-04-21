package web

import (
	"net/http"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/models"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/mappers"
	pkgweb "github.com/emiliopalmerini/quintaedizione.online/pkg/web"
)

// AreaHandler renders a macro-area landing page that lists the area's collections
// as compact category CTAs.
type AreaHandler struct {
	*baseHandler
}

func (h *AreaHandler) handleArea(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	domainGroup, ok := collections.GetGroup(slug)
	if !ok {
		h.ErrorResponse(w, r, pkgweb.NewHTTPError(http.StatusNotFound, "Macro-area non trovata"), "Macro-area non trovata")
		return
	}

	collectionStats, err := h.contentService.GetCollectionStats(r.Context())
	if err != nil {
		h.ErrorResponse(w, r, err, "Errore nel caricamento delle statistiche")
		return
	}

	statsByName := make(map[string]int64, len(collectionStats))
	for _, col := range collectionStats {
		name := mappers.GetString(col, "collection", "")
		statsByName[name] = mappers.GetInt64(col, "count", 0)
	}

	group := models.CollectionGroup{
		Slug:        domainGroup.Slug,
		Label:       domainGroup.Label,
		Description: domainGroup.Description,
	}
	for _, cn := range domainGroup.Collections {
		name := cn.String()
		mc := models.Collection{
			Name:  name,
			Label: h.getCollectionTitle(name),
			Count: statsByName[name],
		}
		if info, ok := collections.GetInfo(name); ok {
			mc.Description = info.Description
		}
		group.Collections = append(group.Collections, mc)
		group.Total += mc.Count
	}

	data := models.AreaPageData{
		PageData: models.PageData{
			Title:       group.Label + " — quintaedizione.online",
			Description: group.Description,
			TotalItems:  group.Total,
		},
		Group: group,
	}

	content, err := h.templateEngine.RenderArea(r.Context(), data)
	if err != nil {
		h.ErrorResponse(w, r, err, "Errore nel rendering della pagina macro-area")
		return
	}

	h.renderHTML(w, content, "home")
}
