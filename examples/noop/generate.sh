#!/usr/bin/env sh

BASEDIR=$(dirname "$0")
ROOT="$(git rev-parse --show-toplevel)"
GOGENMAIN=${ROOT}/cmd/gogen/main.go
TEMPLATE=${BASEDIR}/noop.tmpl
OUTPUT=${BASEDIR}/output.txt
go run ${GOGENMAIN} Example -t ${TEMPLATE} -o ${OUTPUT} -f -d ${BASEDIR}
