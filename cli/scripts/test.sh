#!/usr/bin/env bash
set -eE
set -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SRC=$DIR/../..

pushd $SRC > /dev/null
    # TODO: Derive the right values for GOOS and GOARCH based on `uname -a` and
    # pass them into the build process.
    docker run --name cli --rm \
        -v $SRC:/go/src/github.com/arigatomachine/cli -v $SRC/builds:/builds \
        arigato/cli:latest test
popd > /dev/null
