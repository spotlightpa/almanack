#!/bin/bash

set -eux -o pipefail

# Get the directory that this script file is in
THIS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

cd "$THIS_DIR"

GO_CMD="$(gimme stable)"
gimme list
go version
GOBIN=$THIS_DIR/functions $GO_CMD install ./cmd/...
yarn run build
