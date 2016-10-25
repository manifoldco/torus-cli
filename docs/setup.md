# Development environment

## Codebase

This repository contains the code for both the torus daemon and cli. It's
also the home of all torus issues (server and client).

The daemon is responsible for:
- holding sensitive data (the user's password and access tokens).
- more complex logic (adding a new crediential, including a keyring).
- all cryptographic actions.

The CLI handles interacting with, and validating input from, the user. It
communicates with the daemon over a unix domain socket and HTTP. Some trivial
actions are simply proxied by the daemon.

### Docker

A docker container is provided for convenience and reproducability. It can be
used for both development and continuous integration.

**To build the container:**

`make container` inside the CLI directory.

**To build the artifacts:**

`make docker-release-all` inside the CLI directory.

**To run the tests (which also creates a local build):**

`make docker-test` inside the CLI directory.

### Troubleshooting TL;DR

1. Are you up to date with master?
2. Did you kill your daemon process after rebuild? `torus daemon stop`
3. What does `tail -f ~/.torus/daemon.log` say?
4. Open an issue with the output from `torus debug`
