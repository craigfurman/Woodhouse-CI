#!/bin/bash -e

project=$(dirname $0)/..

bundle exec compass compile --sass-dir $project/web/scss --css-dir $project/web/assets/stylesheets 
