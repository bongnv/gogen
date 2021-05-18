#!/usr/bin/env sh

cd "$(git rev-parse --show-toplevel)"

RUNNING="\x1b[1mRunning: "
PASS="\x1b[32mPASS"
FAIL="\x1b[31mFAIL"
RESET="\x1b[0m"

for f in examples/*; do
    if [ -d "$f" ]; then
        echo "${RUNNING}$f${RESET}"
        sh ${f}/generate.sh
    fi
done

echo "${RUNNING}git diff${RESET}"
RET_DIFF=$(git diff --no-prefix HEAD 2>&1)
if [ ! -z "$RET_DIFF" ]; then echo "$RET_DIFF"; echo; fi

echo "${RUNNING}git ls-files${RESET}"
RET_FILES=$(git ls-files --others --exclude-standard 2>&1)
if [ ! -z "$RET_FILES" ]; then echo "$RET_FILES"; echo; fi

if [ ! -z "$RET_DIFF" ] || [ ! -z "$RET_FILES" ]; then
	echo "${FAIL}${RESET}"; exit 1
else
	echo "${PASS}${RESET}"; exit 0
fi
