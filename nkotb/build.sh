#!/bin/bash

set -eu -o pipefail

# Get the directory that this script file is in
THIS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

cd "$THIS_DIR"

git rev-parse --short HEAD >build/rev.txt
echo "$DEPLOY_PRIME_URL" >build/url.txt
if [[ "$CONTEXT" == "production" ]]; then
	echo "$URL" >build/url.txt
fi
GOBIN=$THIS_DIR/functions go install ./cmd/...
