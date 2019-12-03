#!/bin/bash

set -eux -o pipefail

# Get the directory that this script file is in
THIS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

cd "$THIS_DIR"

echo "GIMME_GO_VERSION: ${GIMME_GO_VERSION:-Not set}"
GIMME_GO_VERSION=current gimme
gimme list
go version
GOBIN=$THIS_DIR/functions go install ./cmd/...
yarn run build
