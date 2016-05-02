#!/usr/bin/env bash
set -eE
set -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

CMD=$1
GOOS=${GOOS:=darwin}
GOARCH=${GOARCH:=amd64}

function usage {
  echo -e "CLI Docker Container"
  echo -e "Used for building binaries and testing code in a repeatable environment\n"
  echo -e "Available Commands:\n"
  echo -e "\tbuild\t\t\t\tBuild the cli; npm install and go build"
  echo -e "\ttest\t\t\t\tBuild and then run all tests on the current build\n"
}

function build {
  echo ""
  echo "Target OS: $GOOS"
  echo "Target Architecture: $GOARCH"
  echo "Build Directory: $BUILD_TARGET_DIR"
  echo ""

  pushd "$DIR/../daemon" > /dev/null
    echo "Building daemon binary"
    glide install
    GOOS=$GOOS GOARCH=$GOARCH make
  popd > /dev/null

  # TODO: Bring npm install inside the container by copying npm token
  # or ssh key for use
  echo "Copying ag-daemon into $DIR/../cli/bin"
  cp $DIR/../daemon/ag-daemon $DIR/../cli/bin/ag-daemon
  chmod +x $DIR/../cli/bin/ag-daemon

  echo "Success, build complete!"
}

function run_tests {
  echo "Running daemon tests"
  pushd "$DIR/../daemon" >/dev/null
    fmt
    make vet
    make test
  popd
  echo "Daemon tests have passed"

  echo "Running cli tests"
  pushd "$DIR/../cli" > /dev/null
    gulp test
  popd

  echo "All tests have passed!"
}

function fmt {
  fmt="$(find . ! \( -path './vendor' -prune \) -type f -name '*.go' -print0 | xargs -0 gofmt -l )"

  if [ -n "$fmt" ]; then
    echo "Unformatted go source code!"
    echo "$fmt"
    exit 1
  fi
}

go version
echo "node version `node -v`"

case "$CMD" in
  "build")
    build
    ;;

  "test")
    build
    run_tests
    ;;

  "all")
    build
    run_tests
    ;;
  *)
    usage
    ;;
esac
