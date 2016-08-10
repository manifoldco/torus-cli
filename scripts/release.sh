#!/usr/bin/env bash
set -eE
set -o pipefail

# Figure out where the script lives
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SRC=$DIR/..

# These are our input variables
TARGET_REF=$1
ENVIRONMENT=$2

USAGE_STR="Usage: bash release.sh <target-ref> <environment>"
if [ -z "$TARGET_REF" -o -z "$ENVIRONMENT" ]; then
    echo "Usage: bash release.sh <target-ref> <environment>"
    exit 1
fi

GIT_REPOSITORY="git@github.com:arigatomachine/cli.git"
KEY_FILE="$SRC/keys/$ENVIRONMENT.json"
CA_BUNDLE="$SRC/daemon/ca_bundle.pem"
RELEASE_STAMP=`date -u +"%Y-%m-%dT%H-%M-%SZ"`
BUILD_DIRECTORY=$HOME/build
RELEASE_DIRECTORY="$BUILD_DIRECTORY/$RELEASE_STAMP"

if [ ! -f "$KEY_FILE" ]; then
    echo "Cannot find offline public key file: $KEY_FILE"
    exit 1
fi

if [ ! -d "$BUILD_DIRECTORY" ]; then
    echo "Build directory $BUILD_DIRECTORY does not exist; creating."
    mkdir -p $BUILD_DIRECTORY
fi

# Fetch the latest greatest and then figure out the pointer of the tag
echo "Attempting to resolve $TARGET_REF to SHA"
git fetch
TARGET_SHA=`git rev-list -n 1 $TARGET_REF`
echo "Resoled $TARGET_REF to $TARGET_SHA"

echo "Changing CWD to $BUILD_DIRECTORY"
cd $BUILD_DIRECTORY

echo "Cloning repository into $BUILD_DIRECTORY/$RELEASE_STAMP"
git clone --reference $SRC $GIT_REPOSITORY $RELEASE_STAMP

echo "Switching into $RELEASE_DIRECTORY"
cd "$RELEASE_DIRECTORY"
git checkout $TARGET_SHA

echo ""
echo "Building Docker Container"
echo ""
docker build -t arigato/cli:$TARGET_SHA .

echo ""
echo "Building Daemon"
echo ""
docker run --name relase-builder --rm \
    -v $RELEASE_DIRECTORY:/go/src/github.com/arigatomachine/cli \
    -v $RELEASE_DIRECTORY/builds:/builds \
    arigato/cli:$TARGET_SHA \
    release

echo ""
echo "Copying Key File"
echo ""
cp $KEY_FILE cli/public_key.json
cp $CA_BUNDLE cli/ca_bundle.pem

echo ""
echo "Creating Distributable ($RELEASE_STAMP.tar.gz) in $RELEASE_DIRECTORY"
echo ""
tar czf "$RELEASE_STAMP.tar.gz" cli/

read -p "Do you wish to publish to npm? [yn] " publish
case $publish in
    [Yy]* )
        npm publish $RELEASE_STAMP.tar.gz
        ;;
    * )
        echo "Not publishing"
        ;;
esac

echo ""
echo "Done; package $RELEASE_DIRECTORY/$RELEASE_STAMP.tar.gz"
echo ""
