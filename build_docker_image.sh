#!/bin/bash

PRE_VERSION=0.0.1
REPO_URL=duyunzhi1

VERSION=$PRE_VERSION

if [ -n "$1" ]; then
    VERSION=${PRE_VERSION}_$1
fi

IMAGE_NAME=pdh-relay

# build image
docker build -f deploy/docker/Dockerfile -t ${REPO_URL}/${IMAGE_NAME}:"${VERSION}" .

# push to repo
docker push ${REPO_URL}/${IMAGE_NAME}:"${VERSION}"

# remove builder intermediate image
sleep 5s
docker image prune --force
