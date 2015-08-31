#!/bin/bash -e

sass_cmd="compile"
if [ -n "$1" ]
then
  sass_cmd=$1
  shift
fi

project=$(dirname $0)/..

# Need to supply relative path for --fonts-dir
pushd $project
bundle exec compass $sass_cmd \
  --sass-dir $project/web/scss \
  --css-dir $project/web/assets/stylesheets \
  --fonts-dir web/assets/fonts
popd
