package web

import (
	"encoding/json"
	"net/http"
)

// HealthCheck represents a single named readiness check.
type HealthCheck struct {
	Name  string
	Check func() bool
}

// HealthChecker aggregates multiple health checks and exposes an HTTP handler.
type HealthChecker struct {
	Checks []HealthCheck
}

// Handler returns an http.HandlerFunc that runs all checks and responds with
// HTTP 200 if all pass, or HTTP 503 if any fail.
func (hc *HealthChecker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		healthy := true
		checks := make(map[string]string, len(hc.Checks))

		for _, c := range hc.Checks {
			if c.Check() {
				checks[c.Name] = "ok"
			} else {
				checks[c.Name] = "fail"
				healthy = false
			}
		}

		status := "healthy"
		httpStatus := http.StatusOK
		if !healthy {
			status = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		json.NewEncoder(w).Encode(map[string]any{
			"status": status,
			"checks": checks,
		})
	}
}
