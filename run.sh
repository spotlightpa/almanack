#!/bin/bash

set -eu -o pipefail

# Get the directory that this script file is in
THIS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

cd "$THIS_DIR"

function _default() {
	api -src-feed "$ARC_FEED" -cache
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
	git ls-files --exclude="$PATTERN" --ignored -z | xargs -0 -I _ "$@"
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

function redis-start() {
	docker run --name redis-container --rm -p 6379:6379 redis:latest
}

function redis-cli() {
	docker run --name redis-cli -it --rm \
		--link redis-container:redis \
		redis:latest \
		redis-cli -h redis -p 6379
}

function sql() {
	_installed sqlc || _die "sqlc not installed"
	go generate ./...
}

function migrate() {
	cd sql/schema
	tern migrate
}

function migrate:prod() {
	cd sql/schema
	tern migrate -c prod.conf
}

function build:frontend() {
	yarn run build
}

function build:backend() {
	go version
	set -x
	echo "$(git rev-parse --short HEAD)" >pkg/almanack/build-version.txt
	echo "${DEPLOY_PRIME_URL:-http://local.dev}" >pkg/almanack/deploy-url.txt
	GOBIN=$THIS_DIR/functions go install ./funcs/...
}

function build:prod() {
	build:backend
	build:frontend
}

function test() {
	test:backend
	test:frontend
	test:misc
}

function test:frontend() {
	yarn run test
}

function test:backend() {
	go test ./... -v
}

function test:misc() {
	_git-xargs '*.sh' shellcheck _
}

function format() {
	yarn run lint
	gofmt -s -w .
	format:misc
}

function format:misc() {
	_git-xargs '*.sh' shfmt -w _
	_git-xargs '*.sql' pg_format -w 80 -s 2 _ -o _
}

function db:copy-prod() {
	local DUMP_FILE
	local DATE_NAME
	echo "Using $PG_BIN"
	set -x
	heroku pg:backups:capture
	DATE_NAME=$(date -u +"%Y-%m-%dT%H:%M:%S")
	DUMP_FILE="$(mktemp -d)/dump-$DATE_NAME.sql"
	heroku pg:backups:download --output "$DUMP_FILE"
	"$PG_BIN"/pg_restore \
		-d 'postgres://localhost/almanack?sslmode=disable' \
		--clean \
		--no-owner \
		"$DUMP_FILE"
}

function api() {
	go run ./funcs/almanack-api "$@"
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
