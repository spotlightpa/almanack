#!/bin/bash

set -eux -o pipefail

# Get the directory that this script file is in
THIS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

cd "$THIS_DIR"

go version
GOOS=linux go build -o $THIS_DIR/functions/ ./cmd/...
yarn run build
