#!/bin/bash

set -e

# Run this from the repository root, not from scripts!
export GOCOVERDIR=$(pwd)/coverage

echo "Running tests with coverage written to $GOCOVERDIR..."
rm -rf $GOCOVERDIR
mkdir $GOCOVERDIR
go test ./... -count=1

echo "Converting coverage formats..."
go tool covdata textfmt -i=$GOCOVERDIR -o $GOCOVERDIR/coverage.txt

echo "Opening coverage..."
go tool cover -html=$GOCOVERDIR/coverage.txt
