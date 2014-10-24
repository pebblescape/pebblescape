#!/bin/bash

echo "-----> Starting Mike dependencies"
docker run -d --name mike-redis pebbles/redis 2&> /dev/null
docker run -d --name mike-postgresql pebbles/postgresql 2&> /dev/null
docker start mike-redis mike-postgresql 2&> /dev/null
# start etcd here