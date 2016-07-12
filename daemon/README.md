# Arigato Daemon

This folder contains the `Arigato Daemon` which is responsible for performing
all cryptographic operations on behalf of the user (through the CLI).

More details are available in the [High-Level Architecture Document](https://docs.google.com/document/d/1t_u3Xk3THQ2TABcEwz8ioIdpSBZHr5iog0lfiGPPXHo/edit).

## Building and Linting

You can build the daemon using the `arigato/cli` Docker container through the
`$REPO_HOME/scripts/build.sh` script. To build the container use
`$REPO_HOME/scripts/build-container.sh`.

You can develop outside the container using `make build` after `glide install`.
However, you will manually have to place the binary in
`$REPO_HOME/cli/bin/ag-daemon`.

## Dependencies

* Go 1.6
* [Glide](https://github.com/Masterminds/glide) for package vendoring

