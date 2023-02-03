#!/bin/bash

PRE_VERSION=0.1.2
REPO_URL=duyunis

VERSION=$PRE_VERSION

if [ -n "$1" ]; then
    VERSION=${PRE_VERSION}_$1
fi

IMAGE_NAME=pdh-relay

# build image
docker build -f deploy/docker/Dockerfile -t ${REPO_URL}/${IMAGE_NAME}:"${VERSION}" .

# tag latest image
docker tag ${REPO_URL}/${IMAGE_NAME}:"${VERSION}" ${REPO_URL}/${IMAGE_NAME}:latest

# push to repo
docker push ${REPO_URL}/${IMAGE_NAME}:"${VERSION}"
docker push ${REPO_URL}/${IMAGE_NAME}:latest

# remove builder intermediate image
sleep 5s
docker image prune --force
