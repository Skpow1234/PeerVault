#!/bin/bash

# PeerVault API Testing Script
# This script runs comprehensive API tests including interactive, contract, performance, and security tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL=${BASE_URL:-"http://localhost:3000"}
MOCK_URL=${MOCK_URL:-"http://localhost:3001"}
TEST_TIMEOUT=${TEST_TIMEOUT:-"30s"}
VERBOSE=${VERBOSE:-false}

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if a service is running
check_service() {
    local url=$1
    local name=$2
    
    print_status "Checking if $name is running at $url..."
    
    if curl -s --max-time 5 "$url/health" > /dev/null 2>&1; then
        print_success "$name is running"
        return 0
    else
        print_warning "$name is not running at $url"
        return 1
    fi
}

# Function to start mock server
start_mock_server() {
    print_status "Starting mock server..."
    
    if [ -f "cmd/peervault-mock/main.go" ]; then
        go run cmd/peervault-mock/main.go --port 3001 --host localhost &
        MOCK_PID=$!
        sleep 2
        
        if check_service "$MOCK_URL" "Mock Server"; then
            print_success "Mock server started successfully"
        else
            print_error "Failed to start mock server"
            return 1
        fi
    else
        print_warning "Mock server not found, skipping mock tests"
        return 1
    fi
}

# Function to stop mock server
stop_mock_server() {
    if [ ! -z "$MOCK_PID" ]; then
        print_status "Stopping mock server..."
        kill $MOCK_PID 2>/dev/null || true
        print_success "Mock server stopped"
    fi
}

# Function to run unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    
    if go test -v ./tests/contracts/... ./tests/performance/... ./tests/security/...; then
        print_success "Unit tests passed"
    else
        print_error "Unit tests failed"
        return 1
    fi
}

# Function to run contract tests
run_contract_tests() {
    print_status "Running contract tests..."
    
    if [ -d "tests/contracts" ]; then
        # Run Go contract tests
        if go test -v ./tests/contracts/...; then
            print_success "Contract tests passed"
        else
            print_error "Contract tests failed"
            return 1
        fi
        
        # Run Pact tests if available
        if command -v pact-verifier >/dev/null 2>&1; then
            print_status "Running Pact contract verification..."
            if pact-verifier --provider-base-url="$BASE_URL" tests/contracts/; then
                print_success "Pact contract verification passed"
            else
                print_warning "Pact contract verification failed"
            fi
        else
            print_warning "Pact not installed, skipping Pact tests"
        fi
    else
        print_warning "Contract tests not found"
    fi
}

# Function to run performance tests
run_performance_tests() {
    print_status "Running performance tests..."
    
    if [ -d "tests/performance" ]; then
        # Run Go performance tests
        if go test -v -bench=. ./tests/performance/...; then
            print_success "Performance tests passed"
        else
            print_error "Performance tests failed"
            return 1
        fi
        
        # Run k6 tests if available
        if command -v k6 >/dev/null 2>&1; then
            print_status "Running k6 load tests..."
            if [ -f "tests/performance/load-test.js" ]; then
                if k6 run tests/performance/load-test.js; then
                    print_success "k6 load tests passed"
                else
                    print_warning "k6 load tests failed"
                fi
            else
                print_warning "k6 test script not found"
            fi
        else
            print_warning "k6 not installed, skipping k6 tests"
        fi
    else
        print_warning "Performance tests not found"
    fi
}

# Function to run security tests
run_security_tests() {
    print_status "Running security tests..."
    
    if [ -d "tests/security" ]; then
        # Run Go security tests
        if go test -v ./tests/security/...; then
            print_success "Security tests passed"
        else
            print_error "Security tests failed"
            return 1
        fi
        
        # Run OWASP ZAP if available
        if command -v zap-baseline.py >/dev/null 2>&1; then
            print_status "Running OWASP ZAP security scan..."
            if zap-baseline.py -t "$BASE_URL"; then
                print_success "OWASP ZAP scan passed"
            else
                print_warning "OWASP ZAP scan found issues"
            fi
        else
            print_warning "OWASP ZAP not installed, skipping ZAP tests"
        fi
    else
        print_warning "Security tests not found"
    fi
}

# Function to run Postman tests
run_postman_tests() {
    print_status "Running Postman tests..."
    
    if [ -f "tests/api/collections/peervault-postman.json" ]; then
        if command -v newman >/dev/null 2>&1; then
            print_status "Running Newman (Postman CLI) tests..."
            if newman run tests/api/collections/peervault-postman.json \
                --environment-var "baseUrl=$BASE_URL" \
                --environment-var "mockUrl=$MOCK_URL" \
                --timeout-request 30000; then
                print_success "Postman tests passed"
            else
                print_error "Postman tests failed"
                return 1
            fi
        else
            print_warning "Newman not installed, skipping Postman tests"
            print_status "To install Newman: npm install -g newman"
        fi
    else
        print_warning "Postman collection not found"
    fi
}

# Function to generate test report
generate_report() {
    print_status "Generating test report..."
    
    local report_file="test-report-$(date +%Y%m%d-%H%M%S).md"
    
    cat > "$report_file" << EOF
# PeerVault API Test Report

**Generated:** $(date)
**Base URL:** $BASE_URL
**Mock URL:** $MOCK_URL

## Test Results

### Unit Tests
- Status: $([ $? -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")

### Contract Tests
- Status: $([ $? -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")

### Performance Tests
- Status: $([ $? -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")

### Security Tests
- Status: $([ $? -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")

### Postman Tests
- Status: $([ $? -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")

## Recommendations

1. Ensure all services are running before running tests
2. Install required tools: newman, k6, pact-verifier, zap-baseline.py
3. Review failed tests and fix issues
4. Run tests regularly in CI/CD pipeline

EOF

    print_success "Test report generated: $report_file"
}

# Main function
main() {
    print_status "Starting PeerVault API Testing..."
    print_status "Base URL: $BASE_URL"
    print_status "Mock URL: $MOCK_URL"
    
    # Check if main service is running
    if ! check_service "$BASE_URL" "PeerVault API"; then
        print_error "PeerVault API is not running. Please start the service first."
        exit 1
    fi
    
    # Start mock server
    start_mock_server
    
    # Set up cleanup
    trap stop_mock_server EXIT
    
    local failed_tests=0
    
    # Run tests
    print_status "Running comprehensive API tests..."
    
    if ! run_unit_tests; then
        ((failed_tests++))
    fi
    
    if ! run_contract_tests; then
        ((failed_tests++))
    fi
    
    if ! run_performance_tests; then
        ((failed_tests++))
    fi
    
    if ! run_security_tests; then
        ((failed_tests++))
    fi
    
    if ! run_postman_tests; then
        ((failed_tests++))
    fi
    
    # Generate report
    generate_report
    
    # Summary
    if [ $failed_tests -eq 0 ]; then
        print_success "All tests completed successfully! 🎉"
        exit 0
    else
        print_error "$failed_tests test suite(s) failed"
        exit 1
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --base-url)
            BASE_URL="$2"
            shift 2
            ;;
        --mock-url)
            MOCK_URL="$2"
            shift 2
            ;;
        --timeout)
            TEST_TIMEOUT="$2"
            shift 2
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --base-url URL     Base URL for API (default: http://localhost:3000)"
            echo "  --mock-url URL     Mock server URL (default: http://localhost:3001)"
            echo "  --timeout DURATION Test timeout (default: 30s)"
            echo "  --verbose          Enable verbose output"
            echo "  --help             Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run main function
main
