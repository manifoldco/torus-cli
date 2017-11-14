# Contributing Guidelines

Contributions are always welcome; however, it's helpful to read this document in its entirety before submitting a Pull Request, reporting a bug, or submitting a feature request.

### Table of Contents

- [Getting Started](#getting-started)
- [Reporting a bug](#reporting-a-bug)
 - [Security disclosure](#security-disclosure)
- [Feature requests](#feature-requests)
- [Opening a pull request](#opening-a-pull-request)
- [Code of Conduct](#code-of-conduct)
- [License](#license)
- [Contributor license agreement](#contributor-license-agreement)

---------------

# Getting Started

Torus' goal is to make it simple for any developers to securely store, share, and organize application secrets and configuration without having to understand cryptography, private key management, or other setup and configure complicated on-premise solutions.

We're always looking for feature ideas and contributers! Please, don't hesitate to open an issue or pull request, we'd love to work with you to better Torus.

### Development Workflow

To get started as a contributor you'll need to have the following installed and configured:

- The latest version of Go
- An appropriately configured `$GOPATH` (e.g. `~/go`)
- A clone of the `torus-cli` repository at `$GOPATH/src/github.com/manifoldco/torus-cli`
- The `make` utility available on your system

Once done, you can install all of the required bootstrapping utilies required to build and test Torus using the `make bootstrap` command inside your `torus-cli` checkout.

Now, you can build the torus binary by running `make binary`, the tests and linters with `make ci`, and just the test with `make test`.

### Getting Involved

Looking to contribute to Torus but not sure what to tackle? We've labelled all of the ready issues with [help wanted](https://github.com/manifoldco/torus-cli/labels/help%20wanted). For those just looking to get their feet wet we've labelled bite sized issues with [good first issue](https://github.com/manifoldco/torus-cli/labels/help%20wanted).

Have an idea but an issue doesn't exist or it's not labelled with `help wanted`? No problem, feel free to kick off the conversation by suggesting a possible solution or by opening a pull request!

If you ever need help, don't hesitate to ask!

# Reporting a Bug

Think you've found a bug? Let us know by opening an [issue in our tracker](https://github.com/manifoldco/torus-cli/issues) and applying the "bug" label!

### Security disclosure

Security is a top priority for us. If you have encountered a security issue please responsibly disclose it by following our [security disclosure](../docs/security.md) document.

#### Before submitting

Make sure you're on the latest version, your bug may have already been fixed! You can consult our [change log](../CHANGELOG.md) for high-level notes on recent versions.

Finally, [search our issue tracker](https://github.com/manifoldco/torus-cli/issues/search&type=issues) for your problem, someone may have already reported it. If so, please add your unique situation as a comment, any and all information helps us resolve it as quickly as possible!

#### Things to include

 - Clear, reproducible steps to encounter the bug
 - Current system state from `torus debug`
 - Relevant logs from `~/.torus/daemon.log`

#### Labels

If you've encountered a bug please ensure to assign the `bug` label!

# Feature Requests

Torus is ever-evolving, and that's thanks to user feedback. If you have an idea for a feature, let us know by opening an issue or pull request!

#### Things to include

- How the tool operates today.
- Why the current functionality is problematic
- Use cases for the new functionality
- If possible a proposal (or pull request :)) for how the problem could be solved to kick off discussing solutions!

# Opening a Pull Request

To contribute, [fork](https://help.github.com/articles/fork-a-repo/) `torus-cli`, commit your changes, and [open a pull request](https://help.github.com/articles/using-pull-requests/).

You may be asked to make changes to your submission during the review process, we'll work with you on figuring out how to get your pull request ready to be merged and if any changes need to be made.

#### Before submitting

- Run the tests using `make test` and linters using `make lint`
- Review the [manual QA guide](../docs/qa.md)
- Test your change thoroughly with unit tests where appropriate
- If you're adding a new command or modifying the behaviour of a command please ensure to update the relevant documentation in the `docs` folder.
- Don't forget to add a relevant line to the [CHANGELOG.md](../CHANGELOG.md)!

#### Continuous Integration

All pull requests are run against our continuous integration suite on [Travis CI](https://travis-ci.org/manifoldco/torus-cli) through a docker container. You can build this docker container locally via `make docker` and then run the full CI suite inside the container using `make docker-test`.

#### Body

- Supply examples of command behaviour (command output, daemon logs, etc)
- Explain why your particular changes were made the way they are
- Reference the issue your request closes in the body of the PR with `Closes #`

# Code of Conduct

All community members are expected to adhere to our [code of conduct](./CODE_OF_CONDUCT.md).

# License

Manifold's torus-cli is released under the [BSD 3-Clause License](../LICENSE.md).

# Contributor license agreement

For legal purposes all contributors must sign a [contributor license agreement](https://cla-assistant.io/manifoldco/torus-cli), either for an individual or corporation, before a pull request can be accepted.

You will be prompted to sign the agreement by CLA Assistant (bot) when you open a Pull Request for the first time.
