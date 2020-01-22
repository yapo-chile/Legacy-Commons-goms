#!/usr/bin/env bash

echoTitle "Creating needed networks"
for network in ${DOCKER_COMPOSE_NETWORKS}; do
    networkId=`docker network ls -q -f name=${network}`
    if [ -z "$networkId" ];
    then
        echo "Creating network ${network}"
        docker network create ${network}
    fi
done

echoTitle "Starting containers"
docker-compose -f docker/docker-compose.yml -p ${APPNAME} up -d

echoTitle "Done"
