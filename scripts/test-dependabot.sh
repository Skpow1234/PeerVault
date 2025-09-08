#!/bin/bash

# Test Dependabot Configuration Script
# This script validates the Dependabot configuration and checks for common issues

set -e

echo "🔍 Testing Dependabot Configuration"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "success")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "warning")
            echo -e "${YELLOW}⚠️  $message${NC}"
            ;;
        "error")
            echo -e "${RED}❌ $message${NC}"
            ;;
        "info")
            echo -e "ℹ️  $message"
            ;;
    esac
}

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    print_status "error" "Not in a git repository"
    exit 1
fi

# Check for Dependabot configuration files
print_status "info" "Checking Dependabot configuration files..."

if [ -f ".github/dependabot.yml" ]; then
    print_status "success" "Main Dependabot configuration found"
    
    # Validate YAML syntax
    if command -v yq &> /dev/null; then
        if yq eval '.version' .github/dependabot.yml > /dev/null 2>&1; then
            print_status "success" "Dependabot YAML syntax is valid"
        else
            print_status "error" "Dependabot YAML syntax is invalid"
        fi
    else
        print_status "warning" "yq not installed, skipping YAML validation"
    fi
else
    print_status "error" "Main Dependabot configuration not found"
fi

if [ -f ".github/dependabot-security.yml" ]; then
    print_status "success" "Security Dependabot configuration found"
else
    print_status "warning" "Security Dependabot configuration not found"
fi

# Check for GitHub Actions workflow
if [ -f ".github/workflows/dependabot-security.yml" ]; then
    print_status "success" "Dependabot security workflow found"
else
    print_status "warning" "Dependabot security workflow not found"
fi

# Check for issue template
if [ -f ".github/ISSUE_TEMPLATE/dependabot-security.md" ]; then
    print_status "success" "Dependabot security issue template found"
else
    print_status "warning" "Dependabot security issue template not found"
fi

# Check for security configuration
if [ -f ".github/security.yml" ]; then
    print_status "success" "GitHub security configuration found"
else
    print_status "warning" "GitHub security configuration not found"
fi

# Check for Go modules
print_status "info" "Checking Go modules..."

if [ -f "go.mod" ]; then
    print_status "success" "Main go.mod found"
    
    # Check for outdated dependencies
    if command -v go &> /dev/null; then
        print_status "info" "Checking for outdated Go dependencies..."
        outdated=$(go list -u -m all 2>/dev/null | grep '\[' | wc -l)
        if [ "$outdated" -gt 0 ]; then
            print_status "warning" "Found $outdated outdated Go dependencies"
            echo "Outdated dependencies:"
            go list -u -m all 2>/dev/null | grep '\[' | head -10
        else
            print_status "success" "No outdated Go dependencies found"
        fi
    else
        print_status "warning" "Go not installed, skipping dependency check"
    fi
else
    print_status "error" "go.mod not found"
fi

# Check for plugin Go modules
if [ -f "plugins/s3-storage/go.mod" ]; then
    print_status "success" "Plugin go.mod found"
else
    print_status "warning" "Plugin go.mod not found"
fi

# Check for proto Go modules
if [ -f "proto/go.mod" ]; then
    print_status "success" "Proto go.mod found"
else
    print_status "warning" "Proto go.mod not found"
fi

# Check for Docker files
print_status "info" "Checking Docker files..."

docker_files=$(find . -name "Dockerfile*" -o -name "docker-compose*.yml" | wc -l)
if [ "$docker_files" -gt 0 ]; then
    print_status "success" "Found $docker_files Docker files"
else
    print_status "warning" "No Docker files found"
fi

# Check for GitHub Actions
print_status "info" "Checking GitHub Actions..."

action_files=$(find .github/workflows -name "*.yml" -o -name "*.yaml" 2>/dev/null | wc -l)
if [ "$action_files" -gt 0 ]; then
    print_status "success" "Found $action_files GitHub Actions workflows"
else
    print_status "warning" "No GitHub Actions workflows found"
fi

# Check for package.json (npm)
if [ -f "package.json" ]; then
    print_status "success" "package.json found"
else
    print_status "info" "No package.json found (npm ecosystem not used)"
fi

# Check for requirements.txt (pip)
if [ -f "requirements.txt" ]; then
    print_status "success" "requirements.txt found"
else
    print_status "info" "No requirements.txt found (pip ecosystem not used)"
fi

# Check repository settings (if GitHub CLI is available)
if command -v gh &> /dev/null; then
    print_status "info" "Checking repository settings..."
    
    # Check if Dependabot is enabled
    if gh api repos/:owner/:repo/dependabot/alerts &> /dev/null; then
        print_status "success" "Dependabot alerts are accessible"
    else
        print_status "warning" "Cannot access Dependabot alerts (may not be enabled)"
    fi
    
    # Check security features
    if gh api repos/:owner/:repo/vulnerability-alerts &> /dev/null; then
        print_status "success" "Vulnerability alerts are enabled"
    else
        print_status "warning" "Vulnerability alerts may not be enabled"
    fi
else
    print_status "warning" "GitHub CLI not installed, skipping repository checks"
fi

# Summary
echo ""
echo "📊 Summary"
echo "=========="
echo "Dependabot configuration test completed."
echo ""
echo "Next steps:"
echo "1. Review any warnings or errors above"
echo "2. Ensure Dependabot is enabled in repository settings"
echo "3. Check GitHub Security tab for alerts"
echo "4. Monitor Dependabot pull requests and issues"
echo ""
echo "For more information, see: .github/DEPENDABOT_README.md"
