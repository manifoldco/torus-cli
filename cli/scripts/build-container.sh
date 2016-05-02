#!/usr/bin/env bash

#!/usr/bin/env bash
set -eE
set -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SRC=$DIR/../..
NODEJS_VERSION=`node --version`

pushd $SRC > /dev/null
  echo "Building docker container with node version: $NODEJS_VERSION"
  docker build -t arigato/cli:latest --build-arg NODEJS_VERSION="$NODEJS_VERSION" .   
popd > /dev/null
