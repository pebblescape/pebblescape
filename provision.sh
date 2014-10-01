#!/bin/bash

if [ -e "/etc/.provisioned" ] ; then
  echo "VM already provisioned.  Remove /etc/.provisioned to force"
  exit 0
fi

## Temporarily disable dpkg fsync to make building faster.
echo force-unsafe-io > /etc/dpkg/dpkg.cfg.d/02apt-speedup

## Install packages
sed -i 's/^#\s*\(deb.*universe\)$/\1/g' /etc/apt/sources.list
sed -i 's/^#\s*\(deb.*multiverse\)$/\1/g' /etc/apt/sources.list
echo deb https://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list
echo deb http://ppa.launchpad.net/brightbox/ruby-ng/ubuntu trusty main > /etc/apt/sources.list.d/ruby.list
apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 36A1D7869245C8950F966E92D8576A8BA88D21E9
apt-get update

apt-get install -y --force-yes --no-install-recommends linux-image-extra-`uname -r` lxc wget nano htop git-core libxml2-dev ruby2.1 lxc-docker

## Docker
service docker restart
usermod -a -G docker vagrant

## Mike

USER_DIR=/home/pebbles
APP_DIR=/home/pebbles/mike

useradd -d $USER_DIR -G sudo,docker -U pebbles
mkdir -p $USER_DIR/data
mkdir -p $APP_DIR
chown -R pebbles:pebbles $APP_DIR

docker run -d --name mike-redis -v /home/pebbles/data/redis:/var/lib/redis pebbles/redis
docker run -d --name mike-postgresql -v /home/pebbles/data/postgres:/data/main pebbles/postgresql

## Cleanup
apt-get clean
rm -rf /build
rm -rf /tmp/* /var/tmp/*
rm -rf /var/lib/apt/lists/*
rm -f /etc/dpkg/dpkg.cfg.d/02apt-speedup

touch /etc/.provisioned
