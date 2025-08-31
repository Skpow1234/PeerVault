# PeerVault Test Script for Windows
# Usage: .\scripts\test.ps1 [type]

param(
    [Parameter(Position=0)]
    [ValidateSet("all", "unit", "integration", "race", "fuzz")]
    [string]$Type = "all"
)

Write-Host "PeerVault Test Script for Windows" -ForegroundColor Green
Write-Host "Test Type: $Type" -ForegroundColor Yellow

switch ($Type) {
    "unit" {
        Write-Host "Running unit tests..." -ForegroundColor Blue
        go test -v .\internal\...
    }
    "integration" {
        Write-Host "Running integration tests..." -ForegroundColor Blue
        if (Test-Path ".\tests\integration") {
            go test -v .\tests\integration\...
        } else {
            Write-Host "Integration tests directory not found" -ForegroundColor Yellow
        }
    }
    "race" {
        Write-Host "Running tests with race detector..." -ForegroundColor Blue
        go test -race -v .\...
    }
    "fuzz" {
        Write-Host "Running fuzz tests..." -ForegroundColor Blue
        go test -run ^$ -fuzz=Fuzz -fuzztime=30s .\internal\transport\p2p
    }
    "all" {
        Write-Host "Running all tests..." -ForegroundColor Blue
        go test -v .\...
    }
}

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ All tests passed!" -ForegroundColor Green
} else {
    Write-Host "✗ Some tests failed" -ForegroundColor Red
    exit $LASTEXITCODE
}
