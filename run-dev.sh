#!/bin/bash

echo "Starting Swarm Browser in development mode..."
echo ""

# Run with the dev clusters configuration
go run main.go --clusters dev-clusters.yaml
