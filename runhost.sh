#!/bin/bash

if [ ! -f ./dev/hostkey ]; then
    ssh-keygen -t rsa -f ./dev/hostkey
fi
# go run host/*.go daemon --git-port=2341 --git-keys=./hostkey --state=test.db
go run host/*.go daemon $@
