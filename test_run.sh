
basedir="$(pwd)/${1:-xxx}"
echo $basedir
mkdir -p "$basedir"
./cmd/slgeomag/slgeomag -base=xxx \
                         -streams=NZ_APIM,NZ_SBAM,NZ_EYWM,NZ_SMHS \
                         -selectors=5?L?? \
                         -statefile="$basedir"/slgeomag.state \
                         -lockfile="$basedir"/slgeomag.lock \
                         -verbose \
                         link.geonet.org.nz