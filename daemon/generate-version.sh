#!/bin/bash

set -e

# Default computed version
COMPUTED_VERSION=$(git describe --tag --match 'v*.*.*' 2>/dev/null || true)

# Compute the version if the previous command has failed
if [ -z "${COMPUTED_VERSION}" ]; then
    COMMIT_COUNT=$(git rev-list --all --count)
    COMPUTED_VERSION="0.0.0-${COMMIT_COUNT}-$(git rev-parse --short HEAD)"
fi

echo ${COMPUTED_VERSION}
