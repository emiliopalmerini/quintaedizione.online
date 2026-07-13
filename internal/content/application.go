package content

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/emiliopalmerini/quintaedizione.online/internal/content/catalog"
	"github.com/emiliopalmerini/quintaedizione.online/internal/content/release"
	"github.com/emiliopalmerini/quintaedizione.online/internal/content/routing"
	contentweb "github.com/emiliopalmerini/quintaedizione.online/internal/content/web"
)

type Application struct {
	handler        http.Handler
	datasetVersion string
}

func NewApplication(releaseFiles, routeFiles fs.FS, routeConfigPath string) (*Application, error) {
	canonicalRelease, err := release.Load(releaseFiles)
	if err != nil {
		return nil, fmt.Errorf("load canonical release: %w", err)
	}
	entities, err := catalog.Compile(canonicalRelease)
	if err != nil {
		return nil, fmt.Errorf("compile canonical catalog: %w", err)
	}
	routes, err := routing.LoadRegistry(routeFiles, routeConfigPath, entities)
	if err != nil {
		return nil, fmt.Errorf("load content routes: %w", err)
	}
	handler, err := contentweb.NewHandler(entities, routes)
	if err != nil {
		return nil, fmt.Errorf("create content handler: %w", err)
	}
	return &Application{
		handler:        handler,
		datasetVersion: canonicalRelease.Manifest().DatasetVersion,
	}, nil
}

func (application *Application) Handler() http.Handler {
	return application.handler
}

func (application *Application) DatasetVersion() string {
	return application.datasetVersion
}
