#!/bin/sh

go run ../../cmd/... -c ./custom-fake-server.sh ARG1 ARG2 | tee output.log
