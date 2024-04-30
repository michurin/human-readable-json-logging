#!/bin/sh

set -x # debug: display commands
set -e # exit immediately on any error

cleanup() {
    test "$?" = 0 && echo "OK" || echo "FAIL"
}

trap cleanup EXIT

cd "$(dirname $0)"

for cs in $(find . -mindepth 1 -maxdepth 1 -type d)
do
    (
    cd $cs
    go run ../../cmd/... cat ../input.log | tee output.log
    diff expected.log output.log # will cause interruption due to -e
    rm output.log
    )
done