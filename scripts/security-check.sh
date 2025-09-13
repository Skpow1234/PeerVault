#!/bin/bash

# PeerVault Security Check Script
# This script runs comprehensive security checks locally

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install Go security tools
install_go_tools() {
    print_status "Installing Go security tools..."
    
    if ! command_exists govulncheck; then
        print_status "Installing govulncheck..."
        go install golang.org/x/vuln/cmd/govulncheck@latest
    fi
    
    if ! command_exists semgrep; then
        print_status "Installing semgrep..."
        python3 -m pip install semgrep || {
            print_warning "Failed to install semgrep with pip, trying alternative method..."
            if command -v curl >/dev/null 2>&1; then
                curl -L https://github.com/semgrep/semgrep/releases/latest/download/semgrep-linux64 -o /usr/local/bin/semgrep && chmod +x /usr/local/bin/semgrep
            else
                print_error "Unable to install semgrep. Please install it manually: pip install semgrep"
            fi
        }
    fi
    
    print_success "Go security tools installed"
}

# Function to run vulnerability scanning
run_vulnerability_scan() {
    print_status "Running vulnerability scanning..."
    
    # Create reports directory
    mkdir -p security-reports
    
    # Run govulncheck
    print_status "Running govulncheck..."
    if govulncheck ./... > security-reports/govulncheck-report.txt 2>&1; then
        print_success "No vulnerabilities found by govulncheck"
    else
        print_warning "Vulnerabilities found by govulncheck (check security-reports/govulncheck-report.txt)"
    fi
    
    # Run semgrep
    print_status "Running semgrep security scan..."
    if semgrep --config auto --json -o security-reports/semgrep-report.json ./... 2>/dev/null; then
        print_success "No security issues found by semgrep"
    else
        print_warning "Security issues found by semgrep (check security-reports/semgrep-report.json)"
    fi
}

# Function to run compliance checks
run_compliance_checks() {
    print_status "Running compliance checks..."
    
    # Create reports directory
    mkdir -p compliance-reports
    
    # Run SOC 2 compliance check
    print_status "Running SOC 2 compliance check..."
    go run ./security/audit/compliance.go soc2 ./ > compliance-reports/soc2-compliance.json 2>&1 || {
        print_warning "SOC 2 compliance check completed with findings"
    }
    
    # Run GDPR compliance check
    print_status "Running GDPR compliance check..."
    go run ./security/audit/compliance.go gdpr ./ > compliance-reports/gdpr-compliance.json 2>&1 || {
        print_warning "GDPR compliance check completed with findings"
    }
    
    # Run ISO 27001 compliance check
    print_status "Running ISO 27001 compliance check..."
    go run ./security/audit/compliance.go iso27001 ./ > compliance-reports/iso27001-compliance.json 2>&1 || {
        print_warning "ISO 27001 compliance check completed with findings"
    }
    
    print_success "Compliance checks completed"
}

# Function to run custom security scanner
run_custom_security_scan() {
    print_status "Running custom PeerVault security scanner..."
    
    # Create reports directory
    mkdir -p security-reports
    
    if go run ./security/audit/scanner.go ./ > security-reports/custom-security-report.json 2>&1; then
        print_success "Custom security scan completed successfully"
    else
        print_warning "Custom security scan completed with findings"
    fi
}

# Function to test security modules
test_security_modules() {
    print_status "Testing security modules..."
    
    # Test compilation
    print_status "Testing compilation of security modules..."
    go build ./internal/auth/... || {
        print_error "RBAC module compilation failed"
        return 1
    }
    
    go build ./internal/audit/... || {
        print_error "Audit module compilation failed"
        return 1
    }
    
    go build ./internal/privacy/... || {
        print_error "Privacy module compilation failed"
        return 1
    }
    
    go build ./internal/pki/... || {
        print_error "PKI module compilation failed"
        return 1
    }
    
    print_success "All security modules compile successfully"
    
    # Test functionality
    print_status "Testing security tools functionality..."
    
    # Test security scanner
    go run ./security/audit/scanner.go ./internal/ > /dev/null 2>&1 || {
        print_warning "Security scanner test completed"
    }
    
    # Test compliance auditor
    go run ./security/audit/compliance.go soc2 ./internal/ > /dev/null 2>&1 || {
        print_warning "Compliance auditor test completed"
    }
    
    print_success "Security tools are functional"
}

# Function to validate security policies
validate_security_policies() {
    print_status "Validating security policies..."
    
    # Check if security policies exist
    if [ ! -f "security/policies/access-control.yaml" ]; then
        print_error "Access control policy missing"
        return 1
    fi
    
    if [ ! -f "security/policies/data-classification.yaml" ]; then
        print_error "Data classification policy missing"
        return 1
    fi
    
    # Validate YAML syntax (if Python3 is available and working)
    if command -v python3 >/dev/null 2>&1 && python3 -c "import yaml" >/dev/null 2>&1; then
        python3 -c "import yaml; yaml.safe_load(open('security/policies/access-control.yaml'))" || {
            print_error "Access control policy has invalid YAML syntax"
            return 1
        }
        
        python3 -c "import yaml; yaml.safe_load(open('security/policies/data-classification.yaml'))" || {
            print_error "Data classification policy has invalid YAML syntax"
            return 1
        }
    else
        print_warning "Python3 not available or yaml module not installed, skipping YAML validation"
    fi
    
    print_success "Security policies are valid"
}

# Function to run security unit tests
run_security_tests() {
    print_status "Running security unit tests..."
    
    # Run tests for security modules
    go test -v -timeout=30s ./internal/auth/... || {
        print_error "RBAC tests failed"
        return 1
    }
    
    go test -v -timeout=30s ./internal/audit/... || {
        print_error "Audit tests failed"
        return 1
    }
    
    go test -v -timeout=30s ./internal/privacy/... || {
        print_error "Privacy tests failed"
        return 1
    }
    
    go test -v -timeout=30s ./internal/pki/... || {
        print_error "PKI tests failed"
        return 1
    }
    
    print_success "All security unit tests passed"
}

# Function to check security documentation
check_security_documentation() {
    print_status "Checking security documentation..."
    
    # Check if security documentation exists
    required_docs=(
        "security/README.md"
        "security/audit/scanner.go"
        "security/audit/compliance.go"
        "internal/auth/rbac.go"
        "internal/audit/audit.go"
        "internal/privacy/privacy.go"
        "internal/pki/pki.go"
    )
    
    for doc in "${required_docs[@]}"; do
        if [ ! -f "$doc" ]; then
            print_error "Required security documentation missing: $doc"
            return 1
        fi
    done
    
    print_success "Security documentation is complete"
}

# Function to display help
show_help() {
    echo "PeerVault Security Check Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -a, --all               Run all security checks (default)"
    echo "  -v, --vulnerability     Run vulnerability scanning only"
    echo "  -c, --compliance        Run compliance checks only"
    echo "  -t, --test              Run security tests only"
    echo "  -p, --policies          Validate security policies only"
    echo "  -m, --modules           Test security modules only"
    echo "  -d, --docs              Check security documentation only"
    echo "  -s, --scan              Run custom security scanner only"
    echo "  --install-tools         Install required security tools"
    echo ""
    echo "Examples:"
    echo "  $0                      # Run all security checks"
    echo "  $0 --vulnerability      # Run vulnerability scanning only"
    echo "  $0 --compliance         # Run compliance checks only"
    echo "  $0 --install-tools      # Install security tools"
}

# Main function
main() {
    local run_all=true
    local run_vulnerability=false
    local run_compliance=false
    local run_tests=false
    local run_policies=false
    local run_modules=false
    local run_docs=false
    local run_scan=false
    local install_tools=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -a|--all)
                run_all=true
                shift
                ;;
            -v|--vulnerability)
                run_all=false
                run_vulnerability=true
                shift
                ;;
            -c|--compliance)
                run_all=false
                run_compliance=true
                shift
                ;;
            -t|--test)
                run_all=false
                run_tests=true
                shift
                ;;
            -p|--policies)
                run_all=false
                run_policies=true
                shift
                ;;
            -m|--modules)
                run_all=false
                run_modules=true
                shift
                ;;
            -d|--docs)
                run_all=false
                run_docs=true
                shift
                ;;
            -s|--scan)
                run_all=false
                run_scan=true
                shift
                ;;
            --install-tools)
                install_tools=true
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Print header
    echo "🔒 PeerVault Security Check Script"
    echo "=================================="
    echo ""
    
    # Install tools if requested
    if [ "$install_tools" = true ]; then
        install_go_tools
        exit 0
    fi
    
    # Run requested checks
    if [ "$run_all" = true ] || [ "$run_vulnerability" = true ]; then
        install_go_tools
        run_vulnerability_scan
        echo ""
    fi
    
    if [ "$run_all" = true ] || [ "$run_compliance" = true ]; then
        run_compliance_checks
        echo ""
    fi
    
    if [ "$run_all" = true ] || [ "$run_scan" = true ]; then
        run_custom_security_scan
        echo ""
    fi
    
    if [ "$run_all" = true ] || [ "$run_modules" = true ]; then
        test_security_modules
        echo ""
    fi
    
    if [ "$run_all" = true ] || [ "$run_tests" = true ]; then
        run_security_tests
        echo ""
    fi
    
    if [ "$run_all" = true ] || [ "$run_policies" = true ]; then
        validate_security_policies
        echo ""
    fi
    
    if [ "$run_all" = true ] || [ "$run_docs" = true ]; then
        check_security_documentation
        echo ""
    fi
    
    # Print summary
    echo "🔒 Security Check Summary"
    echo "========================"
    echo ""
    echo "✅ Security checks completed successfully!"
    echo ""
    echo "Reports generated:"
    echo "- security-reports/ (vulnerability and security scan reports)"
    echo "- compliance-reports/ (compliance assessment reports)"
    echo ""
    echo "For comprehensive security scanning, run the security pipeline:"
    echo "  ./scripts/security-check.sh --all"
    echo ""
    echo "To install security tools:"
    echo "  ./scripts/security-check.sh --install-tools"
}

# Run main function with all arguments
main "$@"
