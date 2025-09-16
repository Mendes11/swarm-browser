#!/bin/bash

echo "Starting Swarm Browser in development mode..."
echo ""

# Run in development mode
go run cmd/swarm-tui/main.go -dev -cluster dev-local
