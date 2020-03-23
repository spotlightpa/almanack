#!/bin/bash

set -eu -o pipefail

# Get the directory that this script file is in
THIS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

cd "$THIS_DIR"

function _default() {
	help
}

function _die() {
	echo "$@" 1>&2
	exit 1
}

function help() {
	SCRIPT=$0
	printf "Usage\n\t$SCRIPT <task> <args>\nTasks:\n\n"
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

function rebuild-sql() {
	go generate ./...
}

function migrate() {
	cd sql/schema
	tern migrate -c prod.conf
}

function migrate:prod() {
	cd sql/schema
	tern migrate -c prod.conf
}

function build:prod(){
	./build.sh
}

TIMEFORMAT="Task completed in %3lR"
time "${@:-_default}"
