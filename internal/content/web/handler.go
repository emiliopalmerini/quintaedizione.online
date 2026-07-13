package web

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/emiliopalmerini/quintaedizione.online/internal/content/catalog"
	"github.com/emiliopalmerini/quintaedizione.online/internal/content/routing"
)

const entityPage = `<!doctype html>
<html lang="it">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Entity.Name}} | Quinta Edizione Online</title>
  <meta name="description" content="{{.Entity.Name}}, contenuto {{.Entity.Kind}} per l'edizione {{.Entity.Edition}}.">
  <link rel="canonical" href="{{.CanonicalPath}}">
</head>
<body>
  <main>
    <article data-entity-id="{{.Entity.ID}}" data-concept-id="{{.Entity.ConceptID}}">
      <header>
        <p>Edizione {{.Entity.Edition}}</p>
        <h1>{{.Entity.Name}}</h1>
      </header>
      <dl>
        <dt>Fonte</dt>
        <dd>{{.Entity.Source.Document}}</dd>
      </dl>
    </article>
  </main>
</body>
</html>
`

type Handler struct {
	entities *catalog.Catalog
	routes   *routing.Registry
	page     *template.Template
}

type pageData struct {
	Entity        catalog.Entity
	CanonicalPath string
}

func NewHandler(entities *catalog.Catalog, routes *routing.Registry) (*Handler, error) {
	if entities == nil {
		return nil, errors.New("content catalog is required")
	}
	if routes == nil {
		return nil, errors.New("route registry is required")
	}
	page, err := template.New("entity-page").Parse(entityPage)
	if err != nil {
		return nil, err
	}
	return &Handler{entities: entities, routes: routes, page: page}, nil
}

func (handler *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		response.Header().Set("Allow", "GET, HEAD")
		http.Error(response, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if request.URL.RawPath != "" {
		http.NotFound(response, request)
		return
	}

	resolved, exists := handler.routes.Resolve(request.URL.Path)
	if !exists {
		http.NotFound(response, request)
		return
	}
	if resolved.Redirect {
		http.Redirect(response, request, resolved.CanonicalPath, http.StatusMovedPermanently)
		return
	}
	entity, exists := handler.entities.Entity(resolved.EntityID)
	if !exists {
		http.NotFound(response, request)
		return
	}

	var body bytes.Buffer
	if err := handler.page.Execute(&body, pageData{Entity: entity, CanonicalPath: resolved.CanonicalPath}); err != nil {
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	response.Header().Set("Content-Length", strconv.Itoa(body.Len()))
	if request.Method == http.MethodHead {
		return
	}
	_, _ = response.Write(body.Bytes())
}
