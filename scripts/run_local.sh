#!/bin/bash -e

project=$(dirname $0)/..

go run $project/main.go -port 3000 \
    -templateDir $project/web/templates \
    -storeDir $project/db \
    -buildsDir $project/build_output \
    -gooseCmd goose \
    -assetsDir $project/web/assets \
    -debugMode true
