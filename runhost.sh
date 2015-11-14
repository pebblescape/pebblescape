#!/bin/bash

mkdir -p ./dev/db
chmod -R 777 ./dev/db
go run host/*.go --config=./dev/host.json daemon --home=./dev --dev $@
