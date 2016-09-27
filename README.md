# Torus CLI and Daemon

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
2. Once setup, build the development docker container for the CLI
   using `make container` inside the `$CLI_REPO` folder.
3. Now, you can build the daemon using `make docker-build`  or `make`inside the
   `$CLI_REPO` folder.
4. Override the host name to communicate with your local registry using
  `./torus prefs set core.registry_uri https://arigato.tools`
4. Finally, before you can start working with the daemon and cli you must
   override the torus root key with your local development keys. Using
   `./torus prefs set core.public_key_file $REGISTRY_REPO/keys/offline-pub.json`
5. Now you should be able to begin interacting with the CLI.


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
- Tag master (e.g. `git tag v0.5.0-rc`)
- Push master and the tag to github (`git push --tags origin maser`)
- Run `$CLI_REPO/scripts/release.sh v0.5.0-rc [environment]` where
  environment is the targeted stack this build of the CLI will be used against.
- Only publish to NPM if the tag is a full release.

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

#### Docker

A docker container is provided for convenience and reproducability. It can be
used for both development and continuous integration.

**To build the container:**

`make container` inside the CLI directory.

**To build the artifacts:**

`make docker-release-all` inside the CLI directory.

**To run the tests (which also creates a local build):**

`make docker-test` inside the CLI directory.

#### Troubleshooting TL;DR

1. Are you up to date with master in both registry, and cli?
3. Did you kill your daemon process after rebuild?
4. Does your `~/.torusrc` point to the offline signing public key?
 - Under `[core]` set `public_key_file=$REGISTRY_REPO/keys/offline-pub.json` where `$REGISTRY_REPO` is the path to your repo on disk.
5. What does `tail -f ~/.torus/daemon.log` say?
