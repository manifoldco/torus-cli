#!/usr/bin/env bash
set -eE
set -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/.."

CMD=$1
GOOS=${GOOS:=darwin}
GOARCH=${GOARCH:=amd64}
GOVERSION=${GOVERSION:="go1.7.1"}

function usage {
  echo -e "CLI Docker Container"
  echo -e "Used for building binaries and testing code in a repeatable environment\n"
  echo -e "Available Commands:\n"
  echo -e "\tbuild\t\t\t\tBuild the cli; npm install and go build"
  echo -e "\ttest\t\t\t\tBuild and then run all tests on the current build\n"
}

function build_binary {
  os="$1"
  arch="$2"

  if [ "$os" != "darwin" -a "$os" != "linux" ]; then
    echo "Unknown or unsupported operating system: $os"
    exit 1
  fi

  if [ "$arch" != "amd64" ]; then
    echo "Unknown or unsupported architecture: $arch"
    exit 1
  fi

  if ! go version | grep --quiet "$GOVERSION"; then
    echo ""
    echo "We require ag to be built with $GOVERSION"
    echo ""
    exit 1
  fi

  pushd "$DIR" > /dev/null
    echo "Building ag"
    echo "Target OS: $os"
    echo "Target Arch: $arch"

    glide install
    GOOS=$os GOARCH=$arch make
  popd > /dev/null

  bin="ag-$os-$arch"
  cp $DIR/ag $DIR/cli/bin/$bin
  chmod +x $DIR/cli/bin/$bin

  echo "Copied $bin to $DIR/cli/bin/$bin"
}

function build {
  echo "Building for development (darwin/amd64 only)"
  build_binary $GOOS $GOARCH
  echo "Success, build complete!"
}

function build_release {
  echo "Building for release (darwin/amd64 and linux/amd64)"
  build_binary darwin amd64
  build_binary linux amd64
  echo "Success, build for release complete!"
}

function run_tests {
  echo "Running tests"
  pushd "$DIR" >/dev/null
    make fmtcheck
    make vet
    make lint
    make test
  popd
  echo "All tests have passed!"
}

go version
echo "node version `node -v`"

case "$CMD" in
  "build")
    build
    ;;

  "release")
    build_release
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
