#!/bin/sh

xtrap() {
    echo "Getting SIGNAL $1. Exiting"
    exit 0
}

trap "xtrap SIGINT" 2 # graceful shutdown handler

echo "Arguments: $@"

(
    # pretending someone kills process
    echo "Sleeping..."
    sleep 1
    echo "Going to kill parent process..."
    kill -2 $$
) &

wait
