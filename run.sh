#!/bin/bash

set -eu -o pipefail

# Get the directory that this script file is in
THIS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

cd "$THIS_DIR"

function _default() {
	# shellcheck disable=SC2119
	api
}

function _die() {
	echo >&2 "Fatal: ${*}"
	exit 1
}

function _installed() {
	hash "$1" >/dev/null 2>&1
}

function _git-xargs() {
	local PATTERN=$1
	shift
	git ls-files --exclude="$PATTERN" -ciz | xargs -0 -I _ "$@"
}

function help() {
	local SCRIPT=$0
	cat <<EOF
Usage

	$SCRIPT <task> <args>

Tasks:

EOF
	compgen -A function | grep -e '^_' -v | sort | xargs printf ' - %s\n'
	exit 2
}

function sql() {
	set -x
	format:sql
	sql:sqlc
	set +x
}

function sql:sqlc() {
	_installed sqlc || _die "sqlc not installed"
	{
		cd sql
		sqlc generate
		sqlc compile
		sqlc vet
	}
}

function db:migrate() {
	cd sql/schema
	tern migrate "$@"
}

function db:migrate:prod() {
	cd sql/schema
	tern migrate -c prod.conf "$@"
}

function build:frontend() {
	yarn run build
}

function build:backend() {
	go version
	set -x
	echo "${DEPLOY_PRIME_URL:-http://local.dev}" >pkg/almanack/deploy-url.txt
	GOBIN=$THIS_DIR/functions go install ./funcs/...
	cp "$THIS_DIR/functions/almanack-api" "$THIS_DIR/functions/almanack-api-background"
	set +x
}

function build:prod() {
	build:backend
	build:frontend
}

function test() {
	set -x
	test:backend
	test:frontend
	test:misc
	set +x
}

function test:frontend() {
	yarn run test
}

function test:backend() {
	go test -race -v ./...
}

function test:db() {
	ALMANACK_POSTGRES=$PG_LOCAL_URL go test "$@" ./internal/db
}

function test:misc() {
	_git-xargs '*.sh' shellcheck _
	go mod tidy -diff
}

function format() {
	set -x
	format:js
	format:go
	format:sh
	format:sql
	set +x
}

function format:js() {
	yarn run lint
}

function format:go() {
	gofmt -s -w .
}

function format:sh() {
	_git-xargs '*.sh' shfmt -w _
}

function format:sql() {
	_git-xargs '*.sql' pg_format -w 80 -s 2 -i _
}

function db:copy-prod() {
	local DUMP_FILE
	local DATE_NAME
	echo "Using $PG_BIN"
	set -x
	DATE_NAME=$(date -u +"%Y-%m-%dT%H:%M:%S")
	DUMP_FILE="$(mktemp -d)/dump-$DATE_NAME.sql.tar"
	db:dump-prod "$DUMP_FILE"
	db:load-dump "$DUMP_FILE"
	set +x
}

function db:dump-prod() {
	local DUMP_FILE=$1
	"${PG_BIN}pg_dump" \
		-d "$PG_PROD_URL" \
		--verbose \
		--no-owner \
		--format=tar \
		--file="$DUMP_FILE"
}

function db:load-dump() {
	local DUMP_FILE=$1
	"${PG_BIN}pg_restore" \
		-d "$PG_LOCAL_URL" \
		--verbose \
		--clean \
		--no-owner \
		"$DUMP_FILE"
}

# shellcheck disable=SC2120
function api() {
	# shellcheck disable=SC1091
	[[ -f .env ]] && echo "Using .env file" && source .env
	${GO_EXEC:-go} run ./funcs/almanack-api "$@"
}

function frontend() {
	yarn run serve
}

function check-deps() {
	_installed shellcheck || echo "install https://www.shellcheck.net"
	_installed shfmt || echo "install https://github.com/mvdan/sh"
	_installed sqlc || echo "install https://sqlc.dev"
	_installed tern || echo "install https://github.com/jackc/tern"
}

TIMEFORMAT="Task completed in %1lR"
time "${@:-_default}"
