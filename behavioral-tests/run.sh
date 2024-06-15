#!/bin/sh

set -x # debug: display commands
set -e # exit immediately on any error

cleanup() {
    test "$?" = 0 && echo "OK" || echo "FAIL"
}

trap cleanup EXIT

cd "$(dirname $0)"

for cs in $(find . -mindepth 1 -maxdepth 1 -type d | sort)
do
    (
    cd $cs
    if test -x ./custom-run.sh
    then
        ./custom-run.sh
    else
        go run ../../cmd/... ../fake-server.sh | tee output.log
    fi
    diff expected.log output.log # will cause interruption due to -e
    rm output.log
    echo "OK: $cs"
    )
done