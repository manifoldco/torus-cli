#!/usr/bin/env bash

#!/usr/bin/env bash
set -eE
set -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SRC=$DIR/..

pushd $SRC > /dev/null
  echo "Building docker container"
  docker build -t arigato/cli:latest .
popd > /dev/null
