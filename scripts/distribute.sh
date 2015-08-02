#!/bin/bash -e

function cleanup() {
	echo "Cleaning up temporary directory $tmpdir"
	rm -r $tmpdir
}

tmpdir=$(mktemp -d -t tmp.XXXXXXXX)

trap cleanup EXIT

echo "Created temporary directory at $tmpdir"

pushd $(dirname $0)/..

binDir=$tmpdir/Woodhouse-CI/bin
mkdir -p $binDir
mkdir -p $tmpdir/Woodhouse-CI/web
mkdir $tmpdir/Woodhouse-CI/db
mkdir -p out

echo "Compiling binaries"
go build -o $binDir/goose bitbucket.org/liamstask/goose/cmd/goose
go build -o $binDir/woodhouse-ci

echo "Compiling stylesheets"
bundle install
./scripts/compile_stylesheets.sh

echo "Copying web assets"
cp -r web/templates $tmpdir/Woodhouse-CI/web
cp -r web/assets $tmpdir/Woodhouse-CI/web

cp db/dbconf.yml $tmpdir/Woodhouse-CI/db
cp -r db/migrations $tmpdir/Woodhouse-CI/db

echo "Creating compressed archive"
pushd $tmpdir
tar -cvzf Woodhouse-CI.tar.gz Woodhouse-CI
popd
mv $tmpdir/Woodhouse-CI.tar.gz out/

popd
