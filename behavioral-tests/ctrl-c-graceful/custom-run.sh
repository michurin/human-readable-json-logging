#!/bin/sh

# emulate terminal: run command in separate process group
set -m
(
  go run ../../cmd/... ./custom-fake-server.sh ARG1 ARG2 ARG3 > output.log
) &
set +m
child_pgid=${!}

sleep 1 # allow to work a little
kill -2 -${child_pgid} # emulate user's ctrl-C

wait ${child_pgid}

echo "ok"
