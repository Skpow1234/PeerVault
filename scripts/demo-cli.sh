#!/bin/bash

# Shugo CLI Demo Script
# This script demonstrates the CLI functionality

echo "Shugo CLI Demo"
echo "===================="
echo ""

# Build the CLI if it doesn't exist
if [ ! -f "bin/shugo-cli" ]; then
    echo "Building CLI..."
    go build -o bin/shugo-cli ./cmd/shugo-cli
    echo "CLI built successfully"
fi

echo "Starting Shugo CLI Demo..."
echo ""

# Create a demo script file
cat > /tmp/shugo-demo.txt << EOF
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
cat /tmp/shugo-demo.txt | ./bin/shugo-cli

# Clean up
rm /tmp/shugo-demo.txt

echo ""
echo "Demo completed!"
echo ""
echo "To run the CLI interactively, use:"
echo "  ./bin/shugo-cli"
