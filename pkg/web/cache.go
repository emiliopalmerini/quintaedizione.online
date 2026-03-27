package web

import (
	"fmt"
	"net/http"
)

// SetCacheHeaders sets Cache-Control headers on the response.
// If maxAge > 0, it sets a public cache with the given duration in seconds.
// If maxAge <= 0, it sets no-cache headers.
func SetCacheHeaders(w http.ResponseWriter, maxAge int) {
	if maxAge > 0 {
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", maxAge))
	} else {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	}
}
