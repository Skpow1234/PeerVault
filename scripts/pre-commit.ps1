# Pre-commit script to fix common formatting issues (PowerShell version)
# Run this before committing: .\scripts\pre-commit.ps1

param(
    [switch]$SkipTests
)

Write-Host "Running pre-commit checks..." -ForegroundColor Green

# Check if we're in a git repository
try {
    git rev-parse --git-dir | Out-Null
} catch {
    Write-Host "Not in a git repository" -ForegroundColor Red
    exit 1
}

# Format Go code
Write-Host "Formatting Go code..." -ForegroundColor Yellow
go fmt ./...
goimports -w .

# Check for trailing whitespace in code files
Write-Host "Checking for trailing whitespace in code files..." -ForegroundColor Yellow
$filesWithTrailingWhitespace = Get-ChildItem -Recurse -Include "*.go", "*.yml", "*.yaml" | ForEach-Object { 
    $content = Get-Content $_.FullName -Raw
    if ($content -match '\s+$') { $_.FullName }
}

if ($filesWithTrailingWhitespace) {
    Write-Host "Found trailing whitespace in code files:" -ForegroundColor Red
    $filesWithTrailingWhitespace | ForEach-Object { Write-Host "  $_" -ForegroundColor Red }
    Write-Host "Run: Get-ChildItem -Recurse -Include '*.go', '*.yml', '*.yaml' | ForEach-Object { `$content = Get-Content `$_.FullName -Raw; `$cleanContent = `$content -replace '\s+$', ''; if (`$content -ne `$cleanContent) { Set-Content `$_.FullName `$cleanContent -NoNewline } }" -ForegroundColor Yellow
    exit 1
}

# Run linter
Write-Host "Running linter..." -ForegroundColor Yellow
try {
    golangci-lint run ./...
} catch {
    Write-Host "Warning: golangci-lint not found. Skipping linting." -ForegroundColor Yellow
}

# Run tests (unless skipped)
if (-not $SkipTests) {
    Write-Host "Running tests..." -ForegroundColor Yellow
    go test ./tests/unit/...
    go test ./internal/...
} else {
    Write-Host "Skipping tests (--SkipTests flag used)" -ForegroundColor Yellow
}

Write-Host "Pre-commit checks passed!" -ForegroundColor Green
Write-Host "You can now commit your changes." -ForegroundColor Green
