#!/bin/bash
set -e

echo "Running go vet..."
go vet ./...

if command -v staticcheck &> /dev/null
then
    echo "Running staticcheck..."
    staticcheck ./...
else
    echo "staticcheck could not be found, skipping..."
fi

echo "Running go test..."
go test -v ./...

echo "Running go build..."
go build -v ./...

echo "CI checks passed!"
