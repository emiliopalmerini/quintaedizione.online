package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler_Healthy(t *testing.T) {
	checker := &HealthChecker{
		Checks: []HealthCheck{
			{Name: "store", Check: func() bool { return true }},
		},
	}

	handler := checker.Handler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %v", body["status"])
	}

	checks, ok := body["checks"].(map[string]any)
	if !ok {
		t.Fatal("expected 'checks' object in response")
	}
	if checks["store"] != "ok" {
		t.Errorf("expected store check 'ok', got %v", checks["store"])
	}
}

func TestHealthHandler_Unhealthy(t *testing.T) {
	checker := &HealthChecker{
		Checks: []HealthCheck{
			{Name: "store", Check: func() bool { return false }},
		},
	}

	handler := checker.Handler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "unhealthy" {
		t.Errorf("expected status 'unhealthy', got %v", body["status"])
	}

	checks, ok := body["checks"].(map[string]any)
	if !ok {
		t.Fatal("expected 'checks' object in response")
	}
	if checks["store"] != "fail" {
		t.Errorf("expected store check 'fail', got %v", checks["store"])
	}
}

func TestHealthHandler_MultipleChecks(t *testing.T) {
	checker := &HealthChecker{
		Checks: []HealthCheck{
			{Name: "store", Check: func() bool { return true }},
			{Name: "search", Check: func() bool { return false }},
		},
	}

	handler := checker.Handler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Any failing check should make the whole response 503
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503 when any check fails, got %d", rec.Code)
	}
}
