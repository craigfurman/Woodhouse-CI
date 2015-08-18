#!/bin/bash -e

ginkgo -keepGoing -randomizeAllSpecs -randomizeSuites "$@"
