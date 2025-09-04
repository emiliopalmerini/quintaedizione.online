#!/bin/bash
# Integration test for Go migration using curl and Go test tools

echo "🚀 Starting Go Migration Integration Tests"
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Helper function to run a test
run_test() {
    local test_name="$1"
    local url="$2"
    local expected_text="$3"
    local timeout="${4:-10}"
    
    echo ""
    echo "🧪 Testing: $test_name"
    echo "URL: $url"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Make request with timeout
    response=$(curl -s --max-time "$timeout" "$url" 2>/dev/null)
    curl_exit_code=$?
    
    if [ $curl_exit_code -eq 0 ]; then
        if [[ "$response" == *"$expected_text"* ]]; then
            echo -e "${GREEN}✅ PASSED${NC} - $test_name"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        else
            echo -e "${RED}❌ FAILED${NC} - $test_name (content check failed)"
            echo "Expected to find: '$expected_text'"
            echo "Response preview (first 200 chars):"
            echo "$response" | head -c 200
            echo ""
            FAILED_TESTS=$((FAILED_TESTS + 1))
            return 1
        fi
    else
        echo -e "${RED}❌ FAILED${NC} - $test_name (connection failed, exit code: $curl_exit_code)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Helper function to run Go tests
run_go_test() {
    local test_name="$1"
    local test_package="$2"
    local test_function="$3"
    
    echo ""
    echo "🔧 Running Go Test: $test_name"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ -n "$test_function" ]; then
        go test -v "$test_package" -run "$test_function" > /tmp/go_test_output 2>&1
    else
        go test -v "$test_package" > /tmp/go_test_output 2>&1
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ PASSED${NC} - $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}❌ FAILED${NC} - $test_name"
        echo "Test output:"
        cat /tmp/go_test_output | tail -10
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Helper function to run benchmarks
run_benchmark() {
    local bench_name="$1"
    local test_package="$2"
    local bench_pattern="$3"
    
    echo ""
    echo "⚡ Running Benchmark: $bench_name"
    
    go test -bench="$bench_pattern" -benchmem "$test_package" > /tmp/bench_output 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ BENCHMARK COMPLETED${NC} - $bench_name"
        echo "Results:"
        cat /tmp/bench_output | grep -E "(Benchmark|PASS|FAIL)" | tail -5
        return 0
    else
        echo -e "${RED}❌ BENCHMARK FAILED${NC} - $bench_name"
        cat /tmp/bench_output | tail -5
        return 1
    fi
}

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
for i in {1..30}; do
    if curl -s http://localhost:8000/health > /dev/null && curl -s http://localhost:8100/health > /dev/null; then
        echo -e "${GREEN}✅ Services are ready${NC}"
        break
    fi
    echo "Waiting... ($i/30)"
    sleep 1
done

echo ""
echo -e "${BLUE}📋 Phase 1: Basic Service Tests${NC}"
echo "================================"

# Test Go Editor service
run_test "Go Editor Homepage" "http://localhost:8000/" "D&D 5e SRD"
run_test "Go Editor Health" "http://localhost:8000/health" "healthy"
run_test "Go Editor Collections" "http://localhost:8000/c/classi" "classi"
run_test "Go Editor Search" "http://localhost:8000/search" "cerca"
run_test "Go Editor Admin" "http://localhost:8000/admin" "admin"

# Test Go Parser service
run_test "Go Parser Homepage" "http://localhost:8100/" "parser"
run_test "Go Parser Health" "http://localhost:8100/health" "healthy"

echo ""
echo -e "${BLUE}📋 Phase 2: Performance & Architecture Tests${NC}"
echo "=============================================="

# Test performance metrics in health endpoint
response=$(curl -s http://localhost:8000/health)
if echo "$response" | grep -q "performance"; then
    echo -e "${GREEN}✅ PASSED${NC} - Performance metrics available in health endpoint"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}❌ FAILED${NC} - No performance metrics in health endpoint"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# Test response headers for performance tracking
response_headers=$(curl -I -s http://localhost:8000/)
if echo "$response_headers" | grep -q "X-Response-Time"; then
    echo -e "${GREEN}✅ PASSED${NC} - Performance headers present"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}❌ FAILED${NC} - No performance headers found"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

echo ""
echo -e "${BLUE}📋 Phase 3: Go Integration Tests${NC}"
echo "================================="

# Run Go integration tests
if [ -f "tests/integration_test.go" ]; then
    run_go_test "Service Availability" "./tests" "TestServiceAvailability"
    run_go_test "Database Connection" "./tests" "TestDatabaseConnection" 
    run_go_test "Health Endpoints" "./tests" "TestHealthEndpoints"
    run_go_test "API Endpoints" "./tests" "TestAPIEndpoints"
    run_go_test "Performance Metrics" "./tests" "TestPerformanceMetrics"
    run_go_test "Cache System" "./tests" "TestCacheSystem"
else
    echo -e "${YELLOW}⚠️ SKIPPED${NC} - Go integration tests not found"
fi

echo ""
echo -e "${BLUE}📋 Phase 4: Performance Benchmarks${NC}"
echo "=================================="

# Run performance benchmarks
if [ -f "tests/benchmarks/performance_test.go" ]; then
    run_benchmark "Health Endpoint Benchmark" "./tests/benchmarks" "BenchmarkHealthEndpoint"
    run_benchmark "Cache Operations Benchmark" "./tests/benchmarks" "BenchmarkCacheOperations"
    run_benchmark "JSON Serialization Benchmark" "./tests/benchmarks" "BenchmarkJSONSerialization"
    run_benchmark "Concurrent Requests Benchmark" "./tests/benchmarks" "BenchmarkConcurrentRequests"
else
    echo -e "${YELLOW}⚠️ SKIPPED${NC} - Performance benchmarks not found"
fi

echo ""
echo -e "${BLUE}📋 Phase 5: Migration Verification${NC}"
echo "=================================="

# Test data migration integrity
run_test "Spells Collection Data" "http://localhost:8000/c/incantesimi" "incantesimi"
run_test "Monsters Collection Data" "http://localhost:8000/c/mostri" "mostri"
run_test "Classes Collection Data" "http://localhost:8000/c/classi" "classi"
run_test "Backgrounds Collection Data" "http://localhost:8000/c/backgrounds" "backgrounds"

# Test specific item access (if available)
run_test "Item Detail Access" "http://localhost:8000/c/classi?page=1" "classi" 15

echo ""
echo -e "${BLUE}📋 Phase 6: HTMX and Frontend Tests${NC}"
echo "===================================="

# Test HTMX functionality by checking for HTMX attributes
response=$(curl -s http://localhost:8000/)
if echo "$response" | grep -q "hx-"; then
    echo -e "${GREEN}✅ PASSED${NC} - HTMX attributes found in HTML"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}❌ FAILED${NC} - No HTMX attributes found"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# Test CSS loading
response=$(curl -s http://localhost:8000/static/style.css)
if [ ${#response} -gt 1000 ]; then
    echo -e "${GREEN}✅ PASSED${NC} - CSS file loads correctly"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}❌ FAILED${NC} - CSS file not loading or too small"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

echo ""
echo -e "${BLUE}📋 Phase 7: Error Handling Tests${NC}"
echo "=================================="

# Test 404 handling
response_code=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8000/nonexistent)
if [ "$response_code" == "404" ]; then
    echo -e "${GREEN}✅ PASSED${NC} - 404 errors handled correctly"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}❌ FAILED${NC} - Expected 404, got $response_code"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# Test invalid collection
response_code=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8000/c/invalid_collection)
if [ "$response_code" == "404" ] || [ "$response_code" == "400" ]; then
    echo -e "${GREEN}✅ PASSED${NC} - Invalid collection handled correctly"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}❌ FAILED${NC} - Invalid collection not handled properly"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

echo ""
echo "=============================================="
echo -e "${BLUE}📊 FINAL RESULTS${NC}"
echo "=============================================="
echo "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 ALL TESTS PASSED! Go migration is working correctly.${NC}"
    echo -e "${GREEN}✅ Editor service: Functional${NC}"
    echo -e "${GREEN}✅ Parser service: Functional${NC}"
    echo -e "${GREEN}✅ Database integration: Working${NC}"
    echo -e "${GREEN}✅ Performance monitoring: Active${NC}"
    echo -e "${GREEN}✅ HTMX functionality: Preserved${NC}"
    echo -e "${GREEN}✅ Error handling: Proper${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}❌ SOME TESTS FAILED. Please check the Go migration.${NC}"
    echo -e "${YELLOW}💡 Consider checking:${NC}"
    echo "   - Service startup logs"
    echo "   - Database connectivity"
    echo "   - Template rendering"
    echo "   - Static asset serving"
    exit 1
fi