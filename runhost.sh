#!/bin/bash

if [ ! -f hostkey ]; then
    ssh-keygen -t rsa -f hostkey
fi
go run host/*.go daemon --git-port=2341 --git-keys=./hostkey --state=test.db
