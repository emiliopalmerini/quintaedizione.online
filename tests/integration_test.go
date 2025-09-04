package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestServiceAvailability tests that both Go services are running
func TestServiceAvailability(t *testing.T) {
	services := []struct {
		name            string
		url             string
		expectedContent string
	}{
		{"Editor", "http://localhost:8000/", "D&D 5e SRD"},
		{"Parser", "http://localhost:8100/", "SRD Parser"},
	}

	for _, service := range services {
		t.Run(service.name, func(t *testing.T) {
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(service.url)
			if err != nil {
				t.Fatalf("%s service is not accessible: %v", service.name, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("%s service returned status %d, expected 200", service.name, resp.StatusCode)
			}

			// Read response body
			buf := make([]byte, 1024)
			n, _ := resp.Body.Read(buf)
			content := string(buf[:n])

			if !strings.Contains(content, service.expectedContent) {
				t.Logf("Response content (first 500 chars): %s", content[:min(len(content), 500)])
				t.Errorf("%s service does not contain expected content '%s'", service.name, service.expectedContent)
			}

			t.Logf("✅ %s service is running at %s", service.name, service.url)
		})
	}
}

// TestDatabaseConnection tests MongoDB connection through editor service
func TestDatabaseConnection(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	// Test collection endpoint
	resp, err := client.Get("http://localhost:8000/c/classi")
	if err != nil {
		t.Fatalf("Failed to access classes endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Classes endpoint returned status %d, expected 200", resp.StatusCode)
	}

	buf := make([]byte, 2048)
	n, _ := resp.Body.Read(buf)
	content := strings.ToLower(string(buf[:n]))

	if !strings.Contains(content, "classi") && !strings.Contains(content, "classes") {
		t.Errorf("Database connection issue - no classes content found in response")
		t.Logf("Response content (first 500 chars): %s", content[:min(len(content), 500)])
	} else {
		t.Log("✅ Database connection working - classes page loaded")
	}
}

// TestHealthEndpoints tests health check endpoints
func TestHealthEndpoints(t *testing.T) {
	endpoints := []struct {
		name string
		url  string
	}{
		{"Editor Health", "http://localhost:8000/health"},
		{"Parser Health", "http://localhost:8100/health"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(endpoint.url)
			if err != nil {
				t.Fatalf("Failed to access %s: %v", endpoint.name, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("%s returned status %d, expected 200", endpoint.name, resp.StatusCode)
			}

			var healthData map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&healthData); err != nil {
				t.Fatalf("Failed to decode health response: %v", err)
			}

			// Check required fields
			if status, ok := healthData["status"].(string); !ok || status != "healthy" {
				t.Errorf("%s does not report healthy status", endpoint.name)
			}

			if version, ok := healthData["version"].(string); !ok || version == "" {
				t.Errorf("%s does not report version", endpoint.name)
			}

			// Check for Go-specific improvements
			if strings.Contains(endpoint.url, "8000") { // Editor
				if arch, ok := healthData["architecture"].(string); !ok || arch != "hexagonal" {
					t.Errorf("Editor does not report hexagonal architecture")
				}
			}

			t.Logf("✅ %s is healthy", endpoint.name)
		})
	}
}

// TestAPIEndpoints tests key API endpoints
func TestAPIEndpoints(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	endpoints := []struct {
		name          string
		url           string
		expectStatus  int
		expectContent string
	}{
		{"Home Page", "http://localhost:8000/", 200, "D&D 5e SRD"},
		{"Classes Collection", "http://localhost:8000/c/classi", 200, "classi"},
		{"Spells Collection", "http://localhost:8000/c/incantesimi", 200, "incantesimi"},
		{"Search Page", "http://localhost:8000/search", 200, "cerca"},
		{"Admin Page", "http://localhost:8000/admin", 200, "admin"},
		{"Parser Home", "http://localhost:8100/", 200, "parser"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			resp, err := client.Get(endpoint.url)
			if err != nil {
				t.Fatalf("Failed to access %s: %v", endpoint.name, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != endpoint.expectStatus {
				t.Fatalf("%s returned status %d, expected %d", endpoint.name, resp.StatusCode, endpoint.expectStatus)
			}

			buf := make([]byte, 2048)
			n, _ := resp.Body.Read(buf)
			content := strings.ToLower(string(buf[:n]))

			if !strings.Contains(content, strings.ToLower(endpoint.expectContent)) {
				t.Errorf("%s does not contain expected content '%s'", endpoint.name, endpoint.expectContent)
				t.Logf("Response content (first 300 chars): %s", content[:min(len(content), 300)])
			} else {
				t.Logf("✅ %s working correctly", endpoint.name)
			}
		})
	}
}

// TestDatabaseDirect tests direct MongoDB connection
func TestDatabaseDirect(t *testing.T) {
	// Connect to MongoDB directly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := "mongodb://admin:password@localhost:27017/?authSource=admin"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Test ping
	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("Failed to ping MongoDB: %v", err)
	}
	t.Log("✅ Direct MongoDB connection successful")

	// Test database and collections
	db := client.Database("dnd")
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		t.Fatalf("Failed to list collections: %v", err)
	}

	expectedCollections := []string{"incantesimi", "mostri", "classi", "backgrounds"}
	foundCollections := make(map[string]bool)
	for _, col := range collections {
		foundCollections[col] = true
	}

	for _, expected := range expectedCollections {
		if foundCollections[expected] {
			t.Logf("✅ Found collection: %s", expected)
		} else {
			t.Errorf("❌ Missing collection: %s", expected)
		}
	}

	// Test document count in a key collection
	incantesimiCount, err := db.Collection("incantesimi").CountDocuments(ctx, bson.D{})
	if err != nil {
		t.Fatalf("Failed to count documents in incantesimi: %v", err)
	}

	if incantesimiCount > 0 {
		t.Logf("✅ Found %d documents in incantesimi collection", incantesimiCount)
	} else {
		t.Error("❌ No documents found in incantesimi collection")
	}
}

// TestPerformanceMetrics tests performance monitoring
func TestPerformanceMetrics(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}

	// Make several requests to generate metrics
	urls := []string{
		"http://localhost:8000/",
		"http://localhost:8000/c/classi",
		"http://localhost:8000/health",
	}

	for i := 0; i < 5; i++ {
		for _, url := range urls {
			resp, err := client.Get(url)
			if err != nil {
				continue
			}
			resp.Body.Close()
		}
	}

	// Check health endpoint for performance metrics
	resp, err := client.Get("http://localhost:8000/health")
	if err != nil {
		t.Fatalf("Failed to access health endpoint: %v", err)
	}
	defer resp.Body.Close()

	var healthData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&healthData); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	// Check for performance metrics
	if perfData, ok := healthData["performance"].(map[string]interface{}); ok {
		if requestCount, ok := perfData["request_count"].(float64); ok && requestCount > 0 {
			t.Logf("✅ Performance metrics working - %d requests recorded", int(requestCount))
		} else {
			t.Error("❌ No request count in performance metrics")
		}

		if avgResponse, ok := perfData["average_response"].(string); ok && avgResponse != "" {
			t.Logf("✅ Average response time: %s", avgResponse)
		} else {
			t.Error("❌ No average response time in performance metrics")
		}

		if memUsage, ok := perfData["memory_usage_mb"].(float64); ok && memUsage > 0 {
			t.Logf("✅ Memory usage: %.2f MB", memUsage)
		} else {
			t.Error("❌ No memory usage in performance metrics")
		}
	} else {
		t.Error("❌ No performance metrics in health response")
	}
}

// TestCacheSystem tests the caching system
func TestCacheSystem(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	// Access the same item multiple times to test caching
	itemURL := "http://localhost:8000/c/classi" // This should use caching

	times := make([]time.Duration, 5)
	for i := 0; i < 5; i++ {
		start := time.Now()
		resp, err := client.Get(itemURL)
		if err != nil {
			t.Fatalf("Failed to access item: %v", err)
		}
		resp.Body.Close()
		times[i] = time.Since(start)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Item request returned status %d", resp.StatusCode)
		}
	}

	// Check that subsequent requests are generally faster (due to caching)
	firstRequest := times[0]
	avgSubsequent := (times[1] + times[2] + times[3] + times[4]) / 4

	t.Logf("First request: %v, Average subsequent: %v", firstRequest, avgSubsequent)

	if avgSubsequent < firstRequest*2 { // Allow some variance
		t.Log("✅ Caching appears to be working - subsequent requests are reasonably fast")
	} else {
		t.Log("ℹ️ Cache performance not clearly measurable in test environment")
	}
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestDataMigrationIntegrity tests that data migrated correctly
func TestDataMigrationIntegrity(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	// Test that we can access key items that should exist
	testCases := []struct {
		name        string
		collection  string
		expectItems bool
	}{
		{"Spells", "incantesimi", true},
		{"Monsters", "mostri", true},
		{"Classes", "classi", true},
		{"Backgrounds", "backgrounds", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			url := fmt.Sprintf("http://localhost:8000/c/%s", testCase.collection)
			resp, err := client.Get(url)
			if err != nil {
				t.Fatalf("Failed to access %s collection: %v", testCase.name, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("%s collection returned status %d", testCase.name, resp.StatusCode)
			}

			buf := make([]byte, 4096)
			n, _ := resp.Body.Read(buf)
			content := string(buf[:n])

			if testCase.expectItems {
				// Look for item indicators (like "elemento" or item structure)
				hasItems := strings.Contains(content, "notion-list-item") ||
					strings.Contains(content, "collection") ||
					strings.Contains(strings.ToLower(content), "elementi")

				if hasItems {
					t.Logf("✅ %s collection contains items", testCase.name)
				} else {
					t.Errorf("❌ %s collection appears empty", testCase.name)
					t.Logf("Response preview: %s", content[:min(len(content), 300)])
				}
			}
		})
	}
}

// TestAPIResponseFormat tests that API responses are properly formatted
func TestAPIResponseFormat(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}

	// Test health endpoint JSON format
	resp, err := client.Get("http://localhost:8000/health")
	if err != nil {
		t.Fatalf("Failed to access health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Health endpoint should return JSON, got: %s", resp.Header.Get("Content-Type"))
	}

	var healthData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&healthData); err != nil {
		t.Fatalf("Health endpoint returned invalid JSON: %v", err)
	}

	requiredFields := []string{"status", "version", "architecture"}
	for _, field := range requiredFields {
		if _, ok := healthData[field]; !ok {
			t.Errorf("Health response missing required field: %s", field)
		}
	}

	t.Log("✅ API response format is correct")
}
