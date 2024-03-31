#!/bin/sh

set -x
set -e

cd "`dirname $0`"

for cs in `find . -mindepth 1 -maxdepth 1 -type d`
do
    (
    cd $cs
    go run ../../cmd/... cat ../input.log | tee output.log
    diff expected.log output.log
    rm output.log
    )
done