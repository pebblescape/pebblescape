#!/bin/bash

echo "-----> Building runner"
docker build -t pebbles/pebblerunner /pebblescape/pebblerunner

echo "-----> Building Mike"

cd /pebblescape/mike/app
docker rm -f mike-build > /dev/null
docker start mike-redis mike-postgresql > /dev/null
git archive master | docker run -i -a stdin -a stdout --name mike-build -e DBNAME=docker -e DBUSER=docker -e DBPASS=docker --link mike-redis:redis --link mike-postgresql:db -v /tmp/mike-cache:/tmp/cache:rw pebbles/pebblerunner build
docker commit mike-build pebbles/mike
docker rm mike-build

echo "-----> Building receiver"
# docker build -t pebbles/receiver /pebblescape/receiver

