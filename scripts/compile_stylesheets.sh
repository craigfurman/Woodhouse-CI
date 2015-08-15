#!/bin/bash -e

project=$(dirname $0)/..

# Need to supply relative path for --fonts-dir
pushd $project
bundle exec compass compile \
  --sass-dir $project/web/scss \
  --css-dir $project/web/assets/stylesheets \
  --fonts-dir web/assets/fonts
popd
