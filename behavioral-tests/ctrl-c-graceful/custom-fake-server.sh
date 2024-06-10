#!/bin/sh

child_pid=

xtrap() {
    echo "Getting SIGNAL $1. Exiting"
    kill -9 $child_pid
    exit
}

trap "xtrap SIGINT" 2 # graceful shutdown handler

echo "Arguments: $@"

(
    # pretending someone kills process
    for i in "$@"; do
      echo "Working hard on ${i}..."
      sleep 5
    done
) &

child_pid=$!

wait
