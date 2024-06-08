#!/bin/sh

../fake-server.sh |
    go run ../../cmd/... |
    tee output.log
