# Arigato CLI and Daemon

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

### Structure

The top level is responsible for packaging and releasing the cli and daemon.
All component specific code and dependencies live in their respective folders.

### Docker

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

`npm run build-container` inside the CLI directory.

The node version used inside the container will be the same version as your
HOST. If you change your version of node you will need to re-build the
container.

**To build the artifacts:**

`npm run build` inside the CLI directory.

The compiled binary will be placed inside `cli/bin` which `cli/bin/arigato`
will look for when trying to start the daemon.

**To run the tests (which also creates a local build):**

`npm run test` inside the CLI directory.

### Local Module Development Work Around

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

### Travis

Travis builds and runs the docker container and uses a matrix of node versions
for running tests which are defined in the `.travis.yml` file.

The ssh key used by travis (which is included encrypted in `id_rsa.enc`) has
been assigned to the `arigato-automated` github account which allows us to
bring in private dependencies.
