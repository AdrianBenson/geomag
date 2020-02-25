#!/bin/bash -e

# Builds and pushes Docker images for the arg list.
#
# usage: ./build-push.sh project [project]

./build-c.sh $@

REPO_BASE=${REPO_BASE:-quay.io/geonet}

VERSION='git-'`git rev-parse --short HEAD`

for i in "$@"
do
    cmd=${i##*/}

    docker push ${REPO_BASE}/${cmd}:${VERSION}
    docker push ${REPO_BASE}/${cmd}:latest
done
# vim: set ts=4 sw=4 tw=0 et:
