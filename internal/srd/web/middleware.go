package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
)

// ErrorRecoveryMiddleware recovers from panics and renders an error page
// using the SRD template engine via baseHandler.
func ErrorRecoveryMiddleware(base *baseHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					isProduction := os.Getenv("ENVIRONMENT") == "production"
					if isProduction {
						log.Printf("PANIC recovered: %v", err)
					} else {
						stack := debug.Stack()
						log.Printf("PANIC recovered: %v\n%s", err, stack)
					}

					errMsg := "Si è verificato un errore interno del server"
					base.ErrorResponse(w, r, fmt.Errorf("internal server error"), errMsg)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func isValidCollection(collection string) bool {
	return collections.IsValid(collection)
}

func getValidCollections() []string {
	return collections.GetAllCollectionNames()
}

// CollectionValidationMiddleware validates the {collection} URL parameter.
func CollectionValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		collection := r.PathValue("collection")
		if collection != "" && !isValidCollection(collection) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"error":             "Collezione non valida",
				"valid_collections": getValidCollections(),
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
