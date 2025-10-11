# PeerVault CLI Demo Script
# This script demonstrates the CLI functionality

Write-Host "🚀 PeerVault CLI Demo" -ForegroundColor Green
Write-Host "====================" -ForegroundColor Green
Write-Host ""

# Build the CLI if it doesn't exist
if (-not (Test-Path "bin/peervault-cli.exe")) {
    Write-Host "Building CLI..." -ForegroundColor Yellow
    go build -o bin/peervault-cli.exe ./cmd/peervault-cli
    Write-Host "✓ CLI built successfully" -ForegroundColor Green
}

Write-Host "Starting PeerVault CLI Demo..." -ForegroundColor Blue
Write-Host ""

# Create a demo script file
$demoScript = @"
help
peers list
health
metrics
status
history
exit
"@

# Write demo script to temp file
$tempFile = [System.IO.Path]::GetTempFileName()
$demoScript | Out-File -FilePath $tempFile -Encoding UTF8

Write-Host "Running CLI with demo commands..." -ForegroundColor Cyan
Write-Host ""

# Run the CLI with the demo script
Get-Content $tempFile | ./bin/peervault-cli.exe

# Clean up
Remove-Item $tempFile

Write-Host ""
Write-Host "Demo completed!" -ForegroundColor Green
Write-Host ""
Write-Host "To run the CLI interactively, use:" -ForegroundColor Yellow
Write-Host "  ./bin/peervault-cli.exe" -ForegroundColor White
