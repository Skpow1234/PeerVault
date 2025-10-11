#!/bin/bash

# PeerVault CLI Advanced Features Demo
# This script demonstrates the enhanced interactive features

echo "🚀 PeerVault CLI - Advanced Features Demo"
echo "=========================================="
echo

# Build the CLI
echo "📦 Building CLI with advanced features..."
make build-cli
if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi
echo "✅ Build successful"
echo

# Create demo files
echo "📁 Creating demo files..."
mkdir -p demo-files
echo "This is a demo document" > demo-files/document.txt
echo "Demo image data" > demo-files/image.jpg
echo "Demo JSON data" > demo-files/data.json
echo "✅ Demo files created"
echo

# Create demo commands file
cat > demo-commands.txt << 'EOF'
# PeerVault CLI Advanced Features Demo
# This demonstrates the enhanced interactive features

# Show welcome message
help

# Demonstrate tab completion (these will be completed automatically)
st<TAB>
ge<TAB>
li<TAB>
pe<TAB>
he<TAB>
me<TAB>

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
EOF

echo "📝 Demo commands file created"
echo

# Run the demo
echo "🎬 Running CLI demo with advanced features..."
echo "Features demonstrated:"
echo "  • Tab completion for commands and arguments"
echo "  • Arrow key navigation for command history"
echo "  • Rich terminal UI with colors and formatting"
echo "  • Progress indicators and spinners"
echo "  • Context-aware prompts"
echo "  • Multi-format output (table, JSON, YAML)"
echo "  • Syntax highlighting"
echo "  • Cross-platform terminal support"
echo

# Run the CLI with demo commands
./bin/peervault-cli < demo-commands.txt

echo
echo "🎉 Demo completed!"
echo
echo "To try the interactive features manually:"
echo "  ./bin/peervault-cli"
echo
echo "Interactive features to try:"
echo "  • Type 'st' and press Tab for completion"
echo "  • Use ↑↓ arrows to navigate command history"
echo "  • Try 'help <command>' for detailed help"
echo "  • Use 'set output_format json' for JSON output"
echo "  • Try 'peers list' to see rich table formatting"
echo

# Cleanup
echo "🧹 Cleaning up demo files..."
rm -rf demo-files
rm -f demo-commands.txt
echo "✅ Cleanup complete"
