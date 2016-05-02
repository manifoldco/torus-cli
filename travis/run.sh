#!/usr/bin/env bash
set -eE

PWD=$(pwd)
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
NODEJS_VERSION=`node -v`

pushd "$DIR/../cli" > /dev/null
  echo "Installing npm dependencies"
  npm install
popd > /dev/null

echo "Logging into docker as $DOCKER_USERNAME"
docker login -u=$DOCKER_USERNAME -p=$DOCKER_PASSWORD

./cli/scripts/build-container.sh
./cli/scripts/test.sh
