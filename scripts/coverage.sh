#!/bin/bash

set -e

# Run this from the repository root, not from scripts!

COVERAGE=$(pwd)/coverage
INTEGRATION_COVERAGE=$COVERAGE/integration

echo "Rebuilding with coverage enabled..."
go clean
go build -cover

rm -rf $COVERAGE
mkdir -p $INTEGRATION_COVERAGE

echo "Running integration tests with coverage written to $INTEGRATION_COVERAGE..."
GOCOVERDIR=$INTEGRATION_COVERAGE go test ./... -count=1

echo "Running unit tests with coverage written to $UNIT_COVERAGE..."
go test ./... -count=1 -coverprofile=$COVERAGE/unit-coverage.txt

echo "Converting integration coverage format..."
go tool covdata textfmt -i=$INTEGRATION_COVERAGE -o $COVERAGE/integration-coverage.txt

echo "Merging coverage..."
cp $COVERAGE/integration-coverage.txt $COVERAGE/coverage.txt
tail -n +2 $COVERAGE/unit-coverage.txt >> $COVERAGE/coverage.txt

echo "Opening coverage..."
go tool cover -html=$COVERAGE/coverage.txt
