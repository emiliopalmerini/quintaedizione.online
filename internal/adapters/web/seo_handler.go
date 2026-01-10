package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SEOHandler handles SEO-related requests (robots.txt, sitemap).
type SEOHandler struct {
	*baseHandler
}

// handleRobotsTxt serves the robots.txt file.
func (h *SEOHandler) handleRobotsTxt(c *gin.Context) {
	c.Header("Cache-Control", "max-age=86400, public")
	c.File("./web/static/robots.txt")
}

// handleSitemap generates and serves the XML sitemap.
func (h *SEOHandler) handleSitemap(c *gin.Context) {
	sitemap := h.generateSitemap()
	c.Header("Cache-Control", "max-age=86400, public")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(sitemap))
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
    <loc>` + baseURL + `/` + collection + `</loc>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>`
	}

	sitemap += `
</urlset>`
	return sitemap
}
