#!/usr/bin/env bash

# Usage: test.sh {linux|darwin}

set -eo pipefail

gotestsum --junitfile test-"$1"-race.xml --format testname -- -race -count=1 -p=1 -timeout 10m ./tests
