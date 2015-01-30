#!/bin/bash

echo "-----> Starting Mike dependencies"
docker run -d --name mike-redis redis > /dev/null
docker run -d --name mike-postgresql -e POSTGRES_PASSWORD=dbpass postgres > /dev/null
docker run -d --name etcd -p 4001:4001 -p 7001:7001 coreos/etcd > /dev/null
docker start mike-redis mike-postgresql etcd > /dev/null
