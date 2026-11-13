#!/usr/bin/env bash

set -eo pipefail

GOTESTSUM_VERSION="${GOTESTSUM_VERSION:-v1.13.0}"

if gotestsum --version 2>/dev/null | grep -F "$GOTESTSUM_VERSION" >/dev/null; then
    echo "gotestsum ${GOTESTSUM_VERSION} already installed"
    exit 0
fi

go install "gotest.tools/gotestsum@${GOTESTSUM_VERSION}"

gotestsum --version
