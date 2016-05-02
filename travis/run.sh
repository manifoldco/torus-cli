#!/usr/bin/env bash
set -eE

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

pushd "$DIR/../cli" > /dev/null
    echo "Running tests in $PWD"
    gulp test
popd > /dev/null
