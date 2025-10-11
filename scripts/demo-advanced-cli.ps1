# PeerVault CLI Advanced Features Demo
# This script demonstrates the enhanced interactive features

Write-Host "🚀 PeerVault CLI - Advanced Features Demo" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host

# Build the CLI
Write-Host "📦 Building CLI with advanced features..." -ForegroundColor Yellow
& make build-cli
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Build successful" -ForegroundColor Green
Write-Host

# Create demo files
Write-Host "📁 Creating demo files..." -ForegroundColor Yellow
New-Item -ItemType Directory -Path "demo-files" -Force | Out-Null
"This is a demo document" | Out-File -FilePath "demo-files/document.txt" -Encoding UTF8
"Demo image data" | Out-File -FilePath "demo-files/image.jpg" -Encoding UTF8
"Demo JSON data" | Out-File -FilePath "demo-files/data.json" -Encoding UTF8
Write-Host "✅ Demo files created" -ForegroundColor Green
Write-Host

# Create demo commands file
$demoCommands = @"
# PeerVault CLI Advanced Features Demo
# This demonstrates the enhanced interactive features

# Show welcome message
help

# Demonstrate tab completion (these will be completed automatically)
st
ge
li
pe
he
me

# Demonstrate command history navigation
# Use ↑ and ↓ arrows to navigate through history

# Demonstrate file operations with completion
store demo-files/document.txt
store demo-files/image.jpg
store demo-files/data.json

# List files with rich formatting
list

# Demonstrate peer management
peers list
peers add localhost:3000
peers add localhost:8080

# Demonstrate system monitoring
health
metrics
status

# Demonstrate configuration
config show
set output_format json
set verbose true

# Demonstrate different output formats
list
peers list
health

# Demonstrate help system
help store
help peers
help health

# Show command history
history

# Clear screen
clear

# Show final status
status
health
metrics

# Exit
exit
"@

$demoCommands | Out-File -FilePath "demo-commands.txt" -Encoding UTF8

Write-Host "📝 Demo commands file created" -ForegroundColor Green
Write-Host

# Run the demo
Write-Host "🎬 Running CLI demo with advanced features..." -ForegroundColor Yellow
Write-Host "Features demonstrated:" -ForegroundColor Cyan
Write-Host "  • Tab completion for commands and arguments" -ForegroundColor White
Write-Host "  • Arrow key navigation for command history" -ForegroundColor White
Write-Host "  • Rich terminal UI with colors and formatting" -ForegroundColor White
Write-Host "  • Progress indicators and spinners" -ForegroundColor White
Write-Host "  • Context-aware prompts" -ForegroundColor White
Write-Host "  • Multi-format output (table, JSON, YAML)" -ForegroundColor White
Write-Host "  • Syntax highlighting" -ForegroundColor White
Write-Host "  • Cross-platform terminal support" -ForegroundColor White
Write-Host

# Run the CLI with demo commands
Get-Content "demo-commands.txt" | & "./bin/peervault-cli.exe"

Write-Host
Write-Host "🎉 Demo completed!" -ForegroundColor Green
Write-Host
Write-Host "To try the interactive features manually:" -ForegroundColor Cyan
Write-Host "  ./bin/peervault-cli.exe" -ForegroundColor White
Write-Host
Write-Host "Interactive features to try:" -ForegroundColor Cyan
Write-Host "  • Type 'st' and press Tab for completion" -ForegroundColor White
Write-Host "  • Use ↑↓ arrows to navigate command history" -ForegroundColor White
Write-Host "  • Try 'help <command>' for detailed help" -ForegroundColor White
Write-Host "  • Use 'set output_format json' for JSON output" -ForegroundColor White
Write-Host "  • Try 'peers list' to see rich table formatting" -ForegroundColor White
Write-Host

# Cleanup
Write-Host "🧹 Cleaning up demo files..." -ForegroundColor Yellow
Remove-Item -Path "demo-files" -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -Path "demo-commands.txt" -Force -ErrorAction SilentlyContinue
Write-Host "✅ Cleanup complete" -ForegroundColor Green
