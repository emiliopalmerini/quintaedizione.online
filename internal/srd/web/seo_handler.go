package web

import (
	"net/http"
)

// SEOHandler handles SEO-related requests (robots.txt, sitemap).
type SEOHandler struct {
	*baseHandler
}

// handleRobotsTxt serves the robots.txt file.
func (h *SEOHandler) handleRobotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=86400, public")
	http.ServeFile(w, r, "./web/static/robots.txt")
}

// handleSitemap generates and serves the XML sitemap.
func (h *SEOHandler) handleSitemap(w http.ResponseWriter, r *http.Request) {
	sitemap := h.generateSitemap()
	w.Header().Set("Cache-Control", "max-age=86400, public")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(sitemap))
}

// generateSitemap creates the XML sitemap with all collections.
func (h *SEOHandler) generateSitemap() string {
	baseURL := "https://quintaedizione.online"
	sitemap := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>` + baseURL + `/</loc>
    <changefreq>weekly</changefreq>
    <priority>1.0</priority>
  </url>`

	collectionNames := getValidCollections()
	for _, collection := range collectionNames {
		sitemap += `
  <url>
    <loc>` + baseURL + `/srd/` + collection + `</loc>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>`
	}

	sitemap += `
</urlset>`
	return sitemap
}
