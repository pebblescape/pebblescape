#!/bin/bash

echo "-----> Cleanup"
docker rm -f mike
docker rm -f receiver

echo "-----> Starting Mike dependencies"
docker run -d --name mike-redis -v /home/pebbles/data/redis:/var/lib/redis pebbles/redis
docker run -d --name mike-postgresql -v /home/pebbles/data/postgres:/data/main pebbles/postgresql
# start etcd here

echo "-----> Starting Mike"
docker run -d --name mike -e DBNAME=docker -e DBUSER=docker -e DBPASS=docker -e PORT=5000 -e WORKERS=3 -p 5000:5000 --link mike-redis:redis --link mike-postgresql:db pebbles/mike start web

echo "-----> Starting receiver"
export SSH_PRIVATE_KEYS=`sudo cat /etc/ssh/ssh_host_rsa_key`
docker run -d --name receiver -p 2341:22 -e SSH_PRIVATE_KEYS="$SSH_PRIVATE_KEYS" -e MIKE_AUTH_KEY="somesuch" -v /var/run/docker.sock:/var/run/docker.sock -v /tmp/pebble-repos:/tmp/pebble-repos:rw -v /tmp/pebble-cache:/tmp/pebble-cache:rw --link mike:mike pebbles/receiver
