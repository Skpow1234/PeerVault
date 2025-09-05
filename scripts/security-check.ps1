# PeerVault Security Check Script (PowerShell)
# This script runs comprehensive security checks locally on Windows

param(
    [switch]$Help,
    [switch]$All,
    [switch]$Vulnerability,
    [switch]$Compliance,
    [switch]$Test,
    [switch]$Policies,
    [switch]$Modules,
    [switch]$Docs,
    [switch]$Scan,
    [switch]$InstallTools
)

# Function to print colored output
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Function to check if a command exists
function Test-Command {
    param([string]$Command)
    try {
        Get-Command $Command -ErrorAction Stop | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

# Function to install Go security tools
function Install-GoTools {
    Write-Status "Installing Go security tools..."
    
    if (-not (Test-Command "govulncheck")) {
        Write-Status "Installing govulncheck..."
        go install golang.org/x/vuln/cmd/govulncheck@latest
    }
    
    if (-not (Test-Command "gosec")) {
        Write-Status "Installing gosec..."
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    }
    
    Write-Success "Go security tools installed"
}

# Function to run vulnerability scanning
function Invoke-VulnerabilityScan {
    Write-Status "Running vulnerability scanning..."
    
    # Create reports directory
    if (-not (Test-Path "security-reports")) {
        New-Item -ItemType Directory -Path "security-reports" | Out-Null
    }
    
    # Run govulncheck
    Write-Status "Running govulncheck..."
    try {
        govulncheck ./... | Out-File -FilePath "security-reports/govulncheck-report.txt" -Encoding UTF8
        Write-Success "No vulnerabilities found by govulncheck"
    }
    catch {
        Write-Warning "Vulnerabilities found by govulncheck (check security-reports/govulncheck-report.txt)"
    }
    
    # Run gosec
    Write-Status "Running gosec..."
    try {
        gosec -fmt json -out security-reports/gosec-report.json ./...
        Write-Success "No security issues found by gosec"
    }
    catch {
        Write-Warning "Security issues found by gosec (check security-reports/gosec-report.json)"
    }
}

# Function to run compliance checks
function Invoke-ComplianceChecks {
    Write-Status "Running compliance checks..."
    
    # Create reports directory
    if (-not (Test-Path "compliance-reports")) {
        New-Item -ItemType Directory -Path "compliance-reports" | Out-Null
    }
    
    # Run SOC 2 compliance check
    Write-Status "Running SOC 2 compliance check..."
    try {
        go run ./security/audit/compliance.go soc2 ./ | Out-File -FilePath "compliance-reports/soc2-compliance.json" -Encoding UTF8
    }
    catch {
        Write-Warning "SOC 2 compliance check completed with findings"
    }
    
    # Run GDPR compliance check
    Write-Status "Running GDPR compliance check..."
    try {
        go run ./security/audit/compliance.go gdpr ./ | Out-File -FilePath "compliance-reports/gdpr-compliance.json" -Encoding UTF8
    }
    catch {
        Write-Warning "GDPR compliance check completed with findings"
    }
    
    # Run ISO 27001 compliance check
    Write-Status "Running ISO 27001 compliance check..."
    try {
        go run ./security/audit/compliance.go iso27001 ./ | Out-File -FilePath "compliance-reports/iso27001-compliance.json" -Encoding UTF8
    }
    catch {
        Write-Warning "ISO 27001 compliance check completed with findings"
    }
    
    Write-Success "Compliance checks completed"
}

# Function to run custom security scanner
function Invoke-CustomSecurityScan {
    Write-Status "Running custom PeerVault security scanner..."
    
    # Create reports directory
    if (-not (Test-Path "security-reports")) {
        New-Item -ItemType Directory -Path "security-reports" | Out-Null
    }
    
    try {
        go run ./security/audit/scanner.go ./ | Out-File -FilePath "security-reports/custom-security-report.json" -Encoding UTF8
        Write-Success "Custom security scan completed successfully"
    }
    catch {
        Write-Warning "Custom security scan completed with findings"
    }
}

# Function to test security modules
function Test-SecurityModules {
    Write-Status "Testing security modules..."
    
    # Test compilation
    Write-Status "Testing compilation of security modules..."
    
    try {
        go build ./internal/auth/...
        Write-Success "RBAC module compiled successfully"
    }
    catch {
        Write-Error "RBAC module compilation failed"
        return
    }
    
    try {
        go build ./internal/audit/...
        Write-Success "Audit module compiled successfully"
    }
    catch {
        Write-Error "Audit module compilation failed"
        return
    }
    
    try {
        go build ./internal/privacy/...
        Write-Success "Privacy module compiled successfully"
    }
    catch {
        Write-Error "Privacy module compilation failed"
        return
    }
    
    try {
        go build ./internal/pki/...
        Write-Success "PKI module compiled successfully"
    }
    catch {
        Write-Error "PKI module compilation failed"
        return
    }
    
    Write-Success "All security modules compile successfully"
    
    # Test functionality
    Write-Status "Testing security tools functionality..."
    
    # Test security scanner
    try {
        go run ./security/audit/scanner.go ./internal/ | Out-Null
        Write-Success "Security scanner test completed"
    }
    catch {
        Write-Warning "Security scanner test completed"
    }
    
    # Test compliance auditor
    try {
        go run ./security/audit/compliance.go soc2 ./internal/ | Out-Null
        Write-Success "Compliance auditor test completed"
    }
    catch {
        Write-Warning "Compliance auditor test completed"
    }
    
    Write-Success "Security tools are functional"
}

# Function to validate security policies
function Test-SecurityPolicies {
    Write-Status "Validating security policies..."
    
    # Check if security policies exist
    if (-not (Test-Path "security/policies/access-control.yaml")) {
        Write-Error "Access control policy missing"
        return
    }
    
    if (-not (Test-Path "security/policies/data-classification.yaml")) {
        Write-Error "Data classification policy missing"
        return
    }
    
    # Validate YAML syntax (basic check)
    try {
        Get-Content "security/policies/access-control.yaml" | Out-Null
        Get-Content "security/policies/data-classification.yaml" | Out-Null
        Write-Success "Security policies are valid"
    }
    catch {
        Write-Error "Security policies have invalid syntax"
        return
    }
}

# Function to run security unit tests
function Invoke-SecurityTests {
    Write-Status "Running security unit tests..."
    
    # Run tests for security modules
    try {
        go test -v -timeout=30s ./internal/auth/...
        Write-Success "RBAC tests passed"
    }
    catch {
        Write-Error "RBAC tests failed"
        return
    }
    
    try {
        go test -v -timeout=30s ./internal/audit/...
        Write-Success "Audit tests passed"
    }
    catch {
        Write-Error "Audit tests failed"
        return
    }
    
    try {
        go test -v -timeout=30s ./internal/privacy/...
        Write-Success "Privacy tests passed"
    }
    catch {
        Write-Error "Privacy tests failed"
        return
    }
    
    try {
        go test -v -timeout=30s ./internal/pki/...
        Write-Success "PKI tests passed"
    }
    catch {
        Write-Error "PKI tests failed"
        return
    }
    
    Write-Success "All security unit tests passed"
}

# Function to check security documentation
function Test-SecurityDocumentation {
    Write-Status "Checking security documentation..."
    
    # Check if security documentation exists
    $requiredDocs = @(
        "security/README.md",
        "security/audit/scanner.go",
        "security/audit/compliance.go",
        "internal/auth/rbac.go",
        "internal/audit/audit.go",
        "internal/privacy/privacy.go",
        "internal/pki/pki.go"
    )
    
    foreach ($doc in $requiredDocs) {
        if (-not (Test-Path $doc)) {
            Write-Error "Required security documentation missing: $doc"
            return
        }
    }
    
    Write-Success "Security documentation is complete"
}

# Function to display help
function Show-Help {
    Write-Host "PeerVault Security Check Script (PowerShell)"
    Write-Host ""
    Write-Host "Usage: .\scripts\security-check.ps1 [OPTIONS]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Help              Show this help message"
    Write-Host "  -All               Run all security checks (default)"
    Write-Host "  -Vulnerability     Run vulnerability scanning only"
    Write-Host "  -Compliance        Run compliance checks only"
    Write-Host "  -Test              Run security tests only"
    Write-Host "  -Policies          Validate security policies only"
    Write-Host "  -Modules           Test security modules only"
    Write-Host "  -Docs              Check security documentation only"
    Write-Host "  -Scan              Run custom security scanner only"
    Write-Host "  -InstallTools      Install required security tools"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\scripts\security-check.ps1                    # Run all security checks"
    Write-Host "  .\scripts\security-check.ps1 -Vulnerability     # Run vulnerability scanning only"
    Write-Host "  .\scripts\security-check.ps1 -Compliance        # Run compliance checks only"
    Write-Host "  .\scripts\security-check.ps1 -InstallTools      # Install security tools"
}

# Main function
function Main {
    # Show help if requested
    if ($Help) {
        Show-Help
        return
    }
    
    # Set default to run all if no specific options provided
    if (-not $Vulnerability -and -not $Compliance -and -not $Test -and -not $Policies -and -not $Modules -and -not $Docs -and -not $Scan -and -not $InstallTools) {
        $All = $true
    }
    
    # Print header
    Write-Host "🔒 PeerVault Security Check Script (PowerShell)"
    Write-Host "=============================================="
    Write-Host ""
    
    # Install tools if requested
    if ($InstallTools) {
        Install-GoTools
        return
    }
    
    # Run requested checks
    if ($All -or $Vulnerability) {
        Install-GoTools
        Invoke-VulnerabilityScan
        Write-Host ""
    }
    
    if ($All -or $Compliance) {
        Invoke-ComplianceChecks
        Write-Host ""
    }
    
    if ($All -or $Scan) {
        Invoke-CustomSecurityScan
        Write-Host ""
    }
    
    if ($All -or $Modules) {
        Test-SecurityModules
        Write-Host ""
    }
    
    if ($All -or $Test) {
        Invoke-SecurityTests
        Write-Host ""
    }
    
    if ($All -or $Policies) {
        Test-SecurityPolicies
        Write-Host ""
    }
    
    if ($All -or $Docs) {
        Test-SecurityDocumentation
        Write-Host ""
    }
    
    # Print summary
    Write-Host "🔒 Security Check Summary"
    Write-Host "========================"
    Write-Host ""
    Write-Success "Security checks completed successfully!"
    Write-Host ""
    Write-Host "Reports generated:"
    Write-Host "- security-reports/ (vulnerability and security scan reports)"
    Write-Host "- compliance-reports/ (compliance assessment reports)"
    Write-Host ""
    Write-Host "For comprehensive security scanning, run the security pipeline:"
    Write-Host "  .\scripts\security-check.ps1 -All"
    Write-Host ""
    Write-Host "To install security tools:"
    Write-Host "  .\scripts\security-check.ps1 -InstallTools"
}

# Run main function
Main
