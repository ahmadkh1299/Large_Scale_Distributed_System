#!/bin/bash
echo "Building the Go project..."
# Build the Go project and output the binary to the specified location
go build -buildvcs=false -o ./output/large-scale-workshop