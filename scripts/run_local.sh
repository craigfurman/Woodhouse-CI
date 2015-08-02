#!/bin/bash -e

go run $(dirname $0)/../main.go -port 3000 -templateDir $PWD/web/templates -storeDir $PWD/db -buildsDir $PWD/build_output -gooseCmd goose
