#!/bin/bash

set -e

# Install Quickhook in this repository using the locally-built executable.
go build
QUICKHOOK=$(pwd)/quickhook
$QUICKHOOK install --bin=$QUICKHOOK --yes
