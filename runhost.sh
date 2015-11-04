#!/bin/bash

mkdir -p ./dev/db
chmod -R 777 ./dev/db
go run host/*.go daemon --home=./dev --dev $@
