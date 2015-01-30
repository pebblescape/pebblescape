#!/bin/bash

./deps.sh

echo "-----> Building runner"
docker build -t pebbles/pebblerunner ../pebblerunner

echo "-----> Building Mike"
cd ..
ROOT=$(pwd)
cd mike
docker rm -f mike-build > /dev/null
git archive master | docker run -i -a stdin --name mike-build -e CURL_TIMEOUT=600 -e DBNAME=postgres -e DBUSER=postgres -e DBPASS=dbpass --link mike-redis:redis --link mike-postgresql:db pebbles/pebblerunner build > /dev/null
docker logs -f mike-build
docker commit mike-build pebbles/mike
docker rm mike-build
cd $ROOT/pebblescape

echo "-----> Building receiver"
docker build -t pebbles/receiver ../receiver

