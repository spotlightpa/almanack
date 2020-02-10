#!/bin/bash

set -eux -o pipefail

# Get the directory that this script file is in
THIS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

cd "$THIS_DIR"

[[ -n "$(command -v gimme)" ]] && eval "$(gimme stable)"
go version
GOBIN=$THIS_DIR/functions go install ./funcs/...
yarn run build
