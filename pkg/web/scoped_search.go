package web

import (
	"net/http"
	"net/url"
)

// ScopedSearchHandler sends the shared search form to the selected content area.
func ScopedSearchHandler(w http.ResponseWriter, r *http.Request) {
	destination := "/srd/search"
	switch r.URL.Query().Get("scope") {
	case "mappe":
		destination = "/mappe"
	case "generatori":
		destination = "/generatori"
	}

	query := url.Values{"q": {r.URL.Query().Get("q")}}
	http.Redirect(w, r, destination+"?"+query.Encode(), http.StatusSeeOther)
}
