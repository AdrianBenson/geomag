#!/bin/bash -ex

# Builds Docker images for the arg list with cvendor support.  These must be
# project directories where this script is executed.
#
# Builds a statically linked executable and adds it to the container.
# Adds the assets dir from each project to the container e.g., origin/assets
# It is not an error for the assets dir to not exist.
# Any assets needed by the application should be read from the assets dir
# relative to the executable. 
#
# usage: ./build.sh project [project]

if [ $# -eq 0 ]; then
  echo Error: please supply a project to build with cvendor support. Usage: ./build-c.sh project [project]
  exit 1
fi

# Go Version
GO_VERSION=${TRAVIS_GO_VERSION:-1.8.1}
# code will be compiled in this container
BUILD_CONTAINER=golang:${GO_VERSION}-alpine
GH_USER=${GH_USER:-ozym}
REPO_BASE=${REPO_BASE:-quay.io/geonet}

# Where to find c vendored code
CVENDORS="internal/cvendor vendor/github.com/GeoNet/kit/cvendor"

DOCKER_TMP=docker-build-tmp

mkdir -p $DOCKER_TMP
chmod +s $DOCKER_TMP

rm -rf $DOCKER_TMP/*

VERSION='git-'`git rev-parse --short HEAD`

# The current working dir to use in GOBIN etc e.g., geonet-web
CWD=${PWD##*/}

mkdir -p ${DOCKER_TMP}/etc

# Assemble common resource for user.
echo "nobody:x:65534:65534:Nobody:/:" > ${DOCKER_TMP}/etc/passwd

for i in "$@"
do
  cmd=${i##*/}

  docker run --rm \
    -e "GOBIN=/usr/src/go/src/github.com/${GH_USER}/${CWD}/${DOCKER_TMP}" -e "GOPATH=/usr/src/go" -e "CGO_ENABLED=1" -e "GOOS=linux" -e "BUILD=$BUILD" \
    -v "$PWD":/usr/src/go/src/github.com/${GH_USER}/${CWD} \
    -w /usr/src/go/src/github.com/${GH_USER}/${CWD} --entrypoint /bin/sh ${BUILD_CONTAINER} \
    -c "apk add --no-cache gcc libc-dev make; for c in ${CVENDORS}; do for d in \$c/*; do (cd \$d; make clean; make); done; done; go install -a -ldflags \"-X main.Prefix=${i}/${VERSION}\" -installsuffix cgo ./${i}"

  rm -f $DOCKER_TMP/Dockerfile

  echo "FROM alpine:3.5" > ${DOCKER_TMP}/Dockerfile
  echo "RUN apk add --no-cache ca-certificates tzdata" >> ${DOCKER_TMP}/Dockerfile
  echo "ADD ./${cmd} /" >> ${DOCKER_TMP}/Dockerfile
  echo "USER nobody" >> ${DOCKER_TMP}/Dockerfile
  echo "ENTRYPOINT [\"/${cmd}\"]" >> ${DOCKER_TMP}/Dockerfile

  docker build -t ${REPO_BASE}/${cmd}:${VERSION} -f ${DOCKER_TMP}/Dockerfile ${DOCKER_TMP}
  # tag latest.  Makes it easier to test with compose.
  docker tag ${REPO_BASE}/${cmd}:${VERSION} ${REPO_BASE}/${cmd}:latest

done

rm -rf ${DOCKER_TMP}

# vim: set ts=2 sw=2 tw=0 et:
