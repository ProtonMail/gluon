#!/usr/bin/env bash

# Usage: test.sh {linux|windows|darwin}

set -eo pipefail

gotestsum --rerun-fails=2 --packages="./..." --junitfile test-"$1".xml --format testname -- -count=1 -p=1 -timeout 15m ./...