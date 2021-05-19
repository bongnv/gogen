#!/usr/bin/env sh

BASEDIR=$(dirname "$0")
ROOT="$(git rev-parse --show-toplevel)"
GOGENMAIN=${ROOT}/cmd/gogen/main.go
TEMPLATE=${BASEDIR}/name-only.tmpl
OUTPUT=${BASEDIR}/output.txt
go run ${GOGENMAIN} Example -t ${TEMPLATE} -o ${OUTPUT}