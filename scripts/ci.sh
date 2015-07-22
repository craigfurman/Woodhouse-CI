#!/bin/bash -ex

go install github.com/onsi/ginkgo/ginkgo
HEADLESS=true ginkgo -r
