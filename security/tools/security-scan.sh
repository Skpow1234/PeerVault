#!/bin/bash

# Security Scanning Script for PeerVault
# This script performs comprehensive security scans of the codebase

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCAN_DIR="$PROJECT_ROOT"
OUTPUT_DIR="$PROJECT_ROOT/security/audit/reports"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo -e "${BLUE}🔍 Starting PeerVault Security Scan${NC}"
echo "Project Root: $PROJECT_ROOT"
echo "Scan Directory: $SCAN_DIR"
echo "Output Directory: $OUTPUT_DIR"
echo "Timestamp: $TIMESTAMP"
echo ""

# Function to print section headers
print_section() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 1. Static Code Analysis
print_section "Static Code Analysis"

if command_exists semgrep; then
    echo -e "${GREEN}✓ Running semgrep security scanner${NC}"
    semgrep --config auto --json -o "$OUTPUT_DIR/semgrep_$TIMESTAMP.json" "$SCAN_DIR" || true
    semgrep --config auto --sarif -o "$OUTPUT_DIR/semgrep_$TIMESTAMP.sarif" "$SCAN_DIR" || true
else
    echo -e "${YELLOW}⚠ semgrep not found, skipping static security scan${NC}"
    echo "Install with: pip install semgrep"
fi


# 2. Dependency Scanning
print_section "Dependency Scanning"

if command_exists govulncheck; then
    echo -e "${GREEN}✓ Running Go vulnerability check${NC}"
    govulncheck -json "$SCAN_DIR" > "$OUTPUT_DIR/govulncheck_$TIMESTAMP.json" 2>&1 || true
else
    echo -e "${YELLOW}⚠ govulncheck not found, skipping Go vulnerability check${NC}"
    echo "Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"
fi

if command_exists npm; then
    echo -e "${GREEN}✓ Running npm audit${NC}"
    cd "$SCAN_DIR" && npm audit --json > "$OUTPUT_DIR/npm_audit_$TIMESTAMP.json" 2>&1 || true
else
    echo -e "${YELLOW}⚠ npm not found, skipping npm audit${NC}"
fi

# 3. Container Security
print_section "Container Security"

if command_exists trivy; then
    echo -e "${GREEN}✓ Running Trivy container scan${NC}"
    if [ -f "$SCAN_DIR/Dockerfile" ]; then
        trivy image --format json --output "$OUTPUT_DIR/trivy_dockerfile_$TIMESTAMP.json" "$SCAN_DIR" || true
    fi
else
    echo -e "${YELLOW}⚠ Trivy not found, skipping container security scan${NC}"
    echo "Install with: https://aquasecurity.github.io/trivy/"
fi

# 4. Secrets Detection
print_section "Secrets Detection"

if command_exists detect-secrets; then
    echo -e "${GREEN}✓ Running detect-secrets${NC}"
    detect-secrets scan --all-files --baseline "$OUTPUT_DIR/secrets_baseline_$TIMESTAMP.json" "$SCAN_DIR" || true
else
    echo -e "${YELLOW}⚠ detect-secrets not found, skipping secrets detection${NC}"
    echo "Install with: pip install detect-secrets"
fi

if command_exists trufflehog; then
    echo -e "${GREEN}✓ Running TruffleHog${NC}"
    trufflehog filesystem --directory "$SCAN_DIR" --json --output "$OUTPUT_DIR/trufflehog_$TIMESTAMP.json" || true
else
    echo -e "${YELLOW}⚠ TruffleHog not found, skipping secrets detection${NC}"
    echo "Install with: go install github.com/trufflesecurity/trufflehog/v3@latest"
fi

# 5. License Compliance
print_section "License Compliance"

if command_exists go-licenses; then
    echo -e "${GREEN}✓ Running go-licenses check${NC}"
    go-licenses report "$SCAN_DIR" > "$OUTPUT_DIR/go_licenses_$TIMESTAMP.txt" 2>&1 || true
else
    echo -e "${YELLOW}⚠ go-licenses not found, skipping license check${NC}"
    echo "Install with: go install github.com/google/go-licenses@latest"
fi

# 6. Custom Security Checks
print_section "Custom Security Checks"

echo -e "${GREEN}✓ Running custom PeerVault security scanner${NC}"
cd "$PROJECT_ROOT"
go run ./security/audit/scanner.go "$SCAN_DIR" > "$OUTPUT_DIR/custom_scan_$TIMESTAMP.json" 2>&1 || true

# 7. Generate Summary Report
print_section "Generating Summary Report"

SUMMARY_FILE="$OUTPUT_DIR/security_scan_summary_$TIMESTAMP.md"

cat > "$SUMMARY_FILE" << EOF
# Security Scan Summary

**Scan Date:** $(date)
**Project:** PeerVault
**Scan Directory:** $SCAN_DIR

## Scan Results

### Static Code Analysis
- **semgrep:** $([ -f "$OUTPUT_DIR/semgrep_$TIMESTAMP.json" ] && echo "✅ Completed" || echo "❌ Failed/Skipped")

### Dependency Scanning
- **govulncheck:** $([ -f "$OUTPUT_DIR/govulncheck_$TIMESTAMP.json" ] && echo "✅ Completed" || echo "❌ Failed/Skipped")
- **npm audit:** $([ -f "$OUTPUT_DIR/npm_audit_$TIMESTAMP.json" ] && echo "✅ Completed" || echo "❌ Failed/Skipped")

### Container Security
- **Trivy:** $([ -f "$OUTPUT_DIR/trivy_dockerfile_$TIMESTAMP.json" ] && echo "✅ Completed" || echo "❌ Failed/Skipped")

### Secrets Detection
- **detect-secrets:** $([ -f "$OUTPUT_DIR/secrets_baseline_$TIMESTAMP.json" ] && echo "✅ Completed" || echo "❌ Failed/Skipped")
- **TruffleHog:** $([ -f "$OUTPUT_DIR/trufflehog_$TIMESTAMP.json" ] && echo "✅ Completed" || echo "❌ Failed/Skipped")

### License Compliance
- **go-licenses:** $([ -f "$OUTPUT_DIR/go_licenses_$TIMESTAMP.txt" ] && echo "✅ Completed" || echo "❌ Failed/Skipped")

### Custom Security Checks
- **PeerVault Scanner:** $([ -f "$OUTPUT_DIR/custom_scan_$TIMESTAMP.json" ] && echo "✅ Completed" || echo "❌ Failed/Skipped")

## Files Generated

EOF

# List generated files
for file in "$OUTPUT_DIR"/*_$TIMESTAMP.*; do
    if [ -f "$file" ]; then
        echo "- $(basename "$file")" >> "$SUMMARY_FILE"
    fi
done

cat >> "$SUMMARY_FILE" << EOF

## Next Steps

1. Review all generated reports
2. Address high and critical severity issues
3. Update security policies as needed
4. Schedule regular security scans
5. Integrate security scanning into CI/CD pipeline

## Security Tools Installation

If any tools were missing, install them using:

\`\`\`bash
# Go security tools
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/google/go-licenses@latest

# Python tools
pip install semgrep detect-secrets

# Other tools
# Trivy: https://aquasecurity.github.io/trivy/
# TruffleHog: go install github.com/trufflesecurity/trufflehog/v3@latest
\`\`\`

EOF

echo -e "${GREEN}✅ Security scan completed successfully!${NC}"
echo -e "${BLUE}📊 Summary report: $SUMMARY_FILE${NC}"
echo -e "${BLUE}📁 All reports saved to: $OUTPUT_DIR${NC}"

# Display summary
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}SCAN SUMMARY${NC}"
echo -e "${BLUE}========================================${NC}"
echo "Total files generated: $(ls -1 "$OUTPUT_DIR"/*_$TIMESTAMP.* 2>/dev/null | wc -l)"
echo "Summary report: $(basename "$SUMMARY_FILE")"
echo "Output directory: $OUTPUT_DIR"
echo ""
echo -e "${GREEN}🎉 Security scan completed!${NC}"
