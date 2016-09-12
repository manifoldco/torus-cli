#!/usr/bin/env bash
set -eE
set -o pipefail

# Figure out where the script lives
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SRC=$DIR/..

# These are our input variables
TARGET_REF=$1
ENVIRONMENT=$2

USAGE_STR="Usage: bash release.sh <target-ref> <environment>\n"
USAGE_STR="$USAGE_STR\nenvironment must be 'staging' or 'production'"
if [ -z "$TARGET_REF" -o -z "$ENVIRONMENT" ]; then
    echo "$USAGE_STR"
    exit 1
fi

if [ "$ENVIRONMENT" != "staging" -a "$ENVIRONMENT" != "production" ]; then
    echo "$USAGE_STR"
    exit 1
fi

GIT_REPOSITORY="git@github.com:arigatomachine/cli.git"
RELEASE_STAMP=`date -u +"%Y-%m-%dT%H-%M-%SZ"`
BUILD_DIRECTORY=$HOME/build
RELEASE_DIRECTORY="$BUILD_DIRECTORY/$RELEASE_STAMP"
RELEASE_BUCKET="releases.arigato.sh"

if [ ! -d "$BUILD_DIRECTORY" ]; then
    echo "Build directory $BUILD_DIRECTORY does not exist; creating."
    mkdir -p $BUILD_DIRECTORY
fi

if ! aws iam get-user > /dev/null; then
    echo "You must be logged into the aws cli to publish a release"
    exit 1
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
echo "Installing NPM Modules"
echo ""
pushd cli > /dev/null
    npm install
popd > /dev/null

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
    -e PUBLIC_KEY=data/keys/$ENVIRONMENT.json \
    arigato/cli:$TARGET_SHA \
    release

# Remove the node modules; they'll get installed via npm on the way down.
echo ""
echo "Removing Node Modules"
echo ""
rm -rf cli/node_modules

TAR_FILENAME="$TARGET_REF"
if [ "$ENVIRONMENT" == "staging" ]; then
  TAR_FILENAME="$TAR_FILENAME+staging"
fi
TAR_FILENAME="$TAR_FILENAME.tar.gz"

echo ""
echo "Creating Distributable ($RELEASE_STAMP.tar.gz) in $RELEASE_DIRECTORY"
echo ""
tar czf "$RELEASE_STAMP.tar.gz" cli/

read -p "Do you wish to publish the release to s3? [yn]" s3_publish
case $s3_publish in
    [Yy]*)
        echo "Uploading $RELEASE_STAMP to https://s3.amazonaws.com/$RELEASE_BUCKET/${TAR_FILENAME/+/%2B}"
        aws s3api put-object --bucket $RELEASE_BUCKET --key "$TAR_FILENAME" \
            --body "$RELEASE_STAMP.tar.gz"
        ;;
    *)
        echo "Not publishing to s3; exiting.. cannot publish to other sources"
        exit 1
        ;;
esac

if [ "$ENVIRONMENT" == "production" ]; then
    read -p "Do you wish to publish to npm? [yn] " npm_publish
    case $npm_publish in
        [Yy]* )
            npm publish $RELEASE_STAMP.tar.gz
            ;;
        * )
            echo "Not publishing to npm"
            ;;
    esac
fi

echo ""
echo "Done; package $RELEASE_DIRECTORY/$RELEASE_STAMP.tar.gz"
echo ""
