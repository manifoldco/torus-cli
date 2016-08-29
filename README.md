# Arigato CLI and Daemon

## Issues and Process

The
[process](https://docs.google.com/document/d/1IejfO1_bJ0einojZOALeN3vEr5XkyhqttpUOHuBQRdo/edit#)
we follow for managing issues, developing product, and implementation is
available in our shared [Google Docs
Drive](https://drive.google.com/drive/u/0/folders/0Bx72T5vLCOgmeVlQbjVlUVVQRDg).

When new functionality is added to the CLI interface, ensure the appropriate
manual testing steps are added to the `./docs/qa.md` checklist.

## Setup

There are several steps required to get up and running locally with the
daemon/cli and [registry](https://github.com/arigatomachine/registry).

1. Ensure you have a working local development environment of the
   [registry](https://github.com/arigatomachine/registry#setup). Make sure
   you've run the migration and seed scripts!
2. Once setup, build the development docker container for the CLI and Daemon
   using `$CLI_REPO/scripts/build-container.sh` inside the `$CLI_REPO` folder.
3. Now, you can build the daemon using `npm run build` inside the
   `$CLI_REPO/cli` folder.
4. Override the host name to communicate with your local registry using
  `bin/arigato prefs:set core.registry_uri https://arigato.tools`
4. Finally, before you can start working with the daemon and cli you must
   override the arigato root key with your local development keys. Using
   `bin/arigato prefs:set core.public_key_file $REGISTRY_REPO/keys/offline-pub.json`
5. Now you should be able to begin interacting with the CLI and Daemon.


## Releasing

Releases are coordinated using a "release issue" which tracks the current RC
and the state of all manual qa. Our desired state is for the role of "release
manager" to pass between maintainers ensuring *everyone* is capable of
releasing.

A release manager is responsible for:

- Creating the release issue
- Coordinating the tagging of release candidates and deployment
- Tracking and triaging bugs; while coordinating fixes for blockers
- Curating the change log
- Signing and pubishing the release.

The release manager role rotates on a per-release basis.

**The Flow**:

1. Create a release issue containing the targeted semver versions *and* current
   RC status using the [release template](./docs/release-issue.md).
2. Curate a changelist against the current "stable" version using the
   [changelog template](./docs/changelog.md).
3. Tag release candidates for the targeted components (registry, cli, etc) and
   make them available (deploy/distribute).
4. Execute manual qa checklist. If bugs are found, track bugs in the checklist
   by linking to the bug issue. Repeat step 2-4 untl checklist passes.
5. Tag production releases and deploy to hosted registry. Build the CLI for
   production and publish to npm with an appropriate tag.

**Releasing the CLI**

The CLI and Daemon are packaged together using the
`$CLI_REPO/scripts/release.sh` into the `cli` folder.

Example:

The following command will build the v0.2.0 tag for production. You will be
prompted to upload the release to s3 or npm.

```
./scripts/release.sh v0.2.0 production
```

You will need the following:

- AWS SDK installed locally (e.g. `aws-cli/1.10.56 Python/2.7.10 Darwin/15.6.0 botocore/1.4.46`)
- Correct AWS environment variables set (`aws iam get-user` is successful)
- You belong to the CLIDevelopers group on AWS

The steps for packaging the CLI:

- Make sure you've pulled latest master
- Update the package.json with the version (e.g. `0.5.0-preview`).
- Commit the change and tag it (e.g. `v0.5.0-preview`)
- Push master and the tag to github (`git push origin maser; git push origin
  v0.5.0-preview`)
- Run `$CLI_REPO/scripts/release.sh v0.5.0-preview [environment]` where
  environment is the targeted stack this build of the CLI will be used against.
- Only publish to NPM if the tag is a full release.

## Codebase

This repository contains the code for both the arigato daemon and cli. It's
also the home of all arigato issues (server and client).

The daemon is written in Go and is responsible for managing access tokens and
user plain-text passwords.

The CLI is written using Node.js and communicates with the deamon using a local
unix socket only accessible the user running the daemon. It's the nice fluffly
front-end that is responsible for doing fun things like auto completion and
ensuring a great user experience.

Right now the CLI has all the business logic. Over time as we figure things out
we'll push more code into the Daemon and eventually rewrite the CLI into Go.

#### Structure

The top level is responsible for packaging and releasing the cli and daemon.
All component specific code and dependencies live in their respective folders.

#### Docker

A docker container is provided for building and testing the daemon and cli
together. Its used for both development and continuous integration.

It takes a single build argument `NODEJS_VERSION` which defines the version of
Node to use inside the container. It also relies on two environment variables
at run time for building the Go binary: `GOOS` and `GOARCH`.

Inside the container these values default to `darwin` and `amd64` by default.
They can be provided at run-time using `-e GOOS=linux -e GOARCH=amd64` options
with `docker run`.

**The container does not install *any* npm modules at the moment. You must
install them locally on your host using `npm install`. Dependencies that are
not cross-platform (compiling C library etc) are not supported.**

**To build the container:**

`$CLI_HOME/scripts/build-container.sh` inside the CLI directory.

The node version used inside the container will be the same version as your
HOST. If you change your version of node you will need to re-build the
container.

**To build the artifacts:**

`$CLI_HOME/scripts/build.sh` inside the CLI directory.

The compiled binary will be placed inside `cli/bin` which `cli/bin/arigato`
will look for when trying to start the daemon.

**To run the tests (which also creates a local build):**

`$CLI_HOME/scripts/test.sh` inside the CLI directory.

#### Local Module Development Work Around

The CLI currently relies on a shared node module
([common](https://github.com/arigatomachine/common)). At times its desirable to
be able to develop the module without having to release a new version to npm.

This workflow is not supported by the Docker container (using a symlink inside
`node_modules`) as links outside the volume are not supported by Docker Volumes
(since they are not repeatable).

In order to enable this workflow you will need to run your tests outside the
container using `gulp test` directly. All other uses of the container will
still work.

The container is still used for building the Go binary.

#### Travis

Travis builds and runs the docker container and uses a matrix of node versions
for running tests which are defined in the `.travis.yml` file.

The ssh key used by travis (which is included encrypted in `id_rsa.enc`) has
been assigned to the `arigato-automated` github account which allows us to
bring in private dependencies.

#### Troubleshooting TL;DR

1. Are you up to date with master in both registry, and cli?
2. Did you rebuild your daemon with `$CLI_HOME/scripts/build.sh`?
3. Did you kill your daemon process after rebuild?
4. Does your `~/.arigatorc` point to the offline signing public key?
 - Under `[core]` set `public_key_file=$REGISTRY_REPO/keys/offline-pub.json` where `$REGISTRY_REPO` is the path to your repo on disk.
5. What does `tail -f ~/.arigato/daemon.log` say?
