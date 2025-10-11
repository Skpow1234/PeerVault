#!/bin/bash

# PeerVault CLI Demo Script
# This script demonstrates the CLI functionality

echo "🚀 PeerVault CLI Demo"
echo "===================="
echo ""

# Build the CLI if it doesn't exist
if [ ! -f "bin/peervault-cli" ]; then
    echo "Building CLI..."
    go build -o bin/peervault-cli ./cmd/peervault-cli
    echo "✓ CLI built successfully"
fi

echo "Starting PeerVault CLI Demo..."
echo ""

# Create a demo script file
cat > /tmp/peervault-demo.txt << EOF
help
peers list
health
metrics
status
history
exit
EOF

echo "Running CLI with demo commands..."
echo ""

# Run the CLI with the demo script
cat /tmp/peervault-demo.txt | ./bin/peervault-cli

# Clean up
rm /tmp/peervault-demo.txt

echo ""
echo "Demo completed!"
echo ""
echo "To run the CLI interactively, use:"
echo "  ./bin/peervault-cli"
