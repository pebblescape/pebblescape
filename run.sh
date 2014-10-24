#!/bin/bash

./deps.sh

echo "-----> Cleanup"
docker rm -f mike receiver 2&> /dev/null

echo "-----> Starting Mike"
docker run -d --name mike -e DBNAME=docker -e DBUSER=docker -e DBPASS=docker -e PORT=5000 -e WORKERS=3 -p 5000:5000 --link mike-redis:redis --link mike-postgresql:db pebbles/mike start web

echo "-----> Starting receiver"
if [ -f /etc/ssh_host_rsa_key ];
then
  export SSH_PRIVATE_KEYS=`sudo cat /etc/ssh_host_rsa_key`
else
  export SSH_PRIVATE_KEYS=`sudo cat /etc/ssh/ssh_host_rsa_key`
fi
docker run -d --name receiver -p 2341:22 -e SSH_PRIVATE_KEYS="$SSH_PRIVATE_KEYS" -e MIKE_AUTH_KEY="somesuch" -v /var/run/docker.sock:/var/run/docker.sock -v /tmp/pebble-repos:/tmp/pebble-repos:rw -v /tmp/pebble-cache:/tmp/pebble-cache:rw --link mike:mike pebbles/receiver
