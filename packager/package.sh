#!/bin/bash
set -e

if [ ! -d packager ]
then
  echo "You have to run this script from the base directory of the package (i.e. as packager/package.sh)"
  exit 1
fi

if [ "$1" == "" ]
then
  echo "Pass arch as first argument. One of 386 or amd64."
  exit 1
fi

if [ "$1" != '386' -a "$1" != 'amd64' ]
then
  echo "Arch must be one of 386 or amd64."
  exit 1
fi


arch=$1
outdir=nest-scrape

case $arch in
  386)
    firefox=packager/firefox-64.0.2-i386.tar.bz2
    ;;
  amd64)
    firefox=packager/firefox-64.0.2-amd64.tar.bz2
    ;;
  *)
    echo "Arch must be one of 386 or amd64."
    exit 1
    ;;
esac

if [ ! -f "$firefox" ]
then
  echo "This packaging script requires the firefox Linux installer downloaded into"
  echo "the packager/ directory, and renamed to include the arch it was downloaded for." 
  echo "It's not kept in the git repo because it's too big." 
  echo 
  echo "Currently the script is looking for $firefox. This is also the supported FF version."
  exit 1
fi

echo "== Initializing output dir"
rm -rf $outdir
mkdir $outdir

# Get the last version tag
version=$(git describe --tags --long)

export GOARCH=$arch
echo "== Building binary"
go build -o $outdir/nest-scrape -ldflags="-X main.version=$version" 

echo "== Unpacking firefox"
tar -C $outdir -xjf $firefox 

echo "== Generating sample config"
pushd . 2>&1 > /dev/null
cd $outdir 
./nest-scrape -g
chmod go-wrx nest.yaml
popd 2>&1 > /dev/null
sed -i -e 's|^\(browserpath:\).*|\1 firefox/firefox|' $outdir/nest.yaml
sed -i -e 's|^\(browserprofiledir:\).*|\1 ff-profile|' $outdir/nest.yaml

echo "== Generating Readme"
cat > $outdir/README.txt <<HERE
The nest-scrape tool is used to log into to the Nest website and retrieve the thermostat,
temperature sensor, and humidity measurements, and the external temperature.

This directory contains a pre-packaged binary version that includes the firefox browser that
the tool needs. To run the script, first edit nest.yaml and set the login and password to
the ones you use to log into the website. Then run the program from this directory, like:

    ./nest-scrape

Use the --help option for help. The -s option is useful for troubleshooting problems.
HERE

echo "== Archiving"
archive="nest-scrape-$arch-$version.tar.gz"
tar czf $archive $outdir

echo "== Done packaging. Result is $archive"
