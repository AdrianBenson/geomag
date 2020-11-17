#!/usr/bin/env bash

basedir="$(pwd)/${1:-xxx}"
mkdir -p "$basedir"
chmod g+s "$basedir"
./cmd/slgeomag/slgeomag -base="$basedir" \
                         -streams=NZ_APIM,NZ_SBAM,NZ_EYWM,NZ_SMHS \
                         -selectors=5?L?? \
                         -statefile="$basedir"/slgeomag.state \
                         -lockfile="$basedir"/slgeomag.lock \
                         -verbose \
                         link.geonet.org.nz