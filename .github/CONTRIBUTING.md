# Contributing Guidelines

Contributions are always welcome; however, please read this document in its entirety before submitting a Pull Request or Reporting a bug.

### Table of Contents

- [Reporting a bug](#reporting-a-bug)
 - [Security disclosure](#security-disclosure)
- [Creating an issue](#creating-an-issue)
- [Feature requests](#feature-requests)
- [Opening a pull request](#opening-a-pull-request)
- [Code of Conduct](#code-of-conduct)
- [License](#license)
- [Contributor license agreement](#contributor-license-agreement)

---------------

# Reporting a Bug

Think you've found a bug? Let us know!

### Security disclosure

Security is a top priority for us. If you have encountered a security issue please responsibly disclose it by following our [security disclosure](../docs/security.md) document.

# Creating an Issue

Your issue must follow these guidelines for it to be considered:

#### Before submitting

- Check you’re on the latest version, we may have already fixed your bug!
- [Search our issue tracker](https://github.com/manifoldco/torus-cli/issues/search&type=issues) for your problem, someone may have already reported it

#### Things to include
 - Clear, reproducible steps to encounter the bug
 - Current system state from `torus debug`
 - Relevant logs from `~/.torus/daemon.log`

#### Labels

- Apply the `bug` label
- Apply (to the best of your ability) the appropriate `component/` label
  - If you are uncertain simply use `component/cli`


# Feature Requests

Torus is ever-evolving, and that's thanks to user feedback. If you have an idea for a feature, let us know!

#### Things to include

- How the tool operates today.
- Why current functionality is problematic.
- Use cases for the new functionality
- A detailed proposal (or pull request) that demonstrates how you feel the problem could be solved.

#### Labels

- Apply the `discussion/feature` label
- Apply (to the best of your ability) the appropriate `component/` label
  - If you are uncertain simply use `component/cli`


# Opening a Pull Request

To contribute, [fork](https://help.github.com/articles/fork-a-repo/) `torus-cli`, commit your changes, and [open a pull request](https://help.github.com/articles/using-pull-requests/).

Your request will be reviewed once the appropriate labels have been added. You may be asked to make changes to your submission during the review process.

#### Before submitting

- Run the tests using `make test`
- Review the [manual QA guide](../docs/qa.md) (until our test coverage improves)
- Test your change thoroughly

#### Body

- Supply examples of command behaviour (command output, daemon logs, etc)
- Explain why your particular changes were made the way they are
- Reference the issue your request closes in the body of the PR with `Closes #`

#### Labels

- Apply the `meta/in-review` label once your code is ready for review
 - If you are opening before it is finished, please use `meta/in-progress`
- Apply (to the best of your ability) the appropriate `component/` label
  - If you are uncertain simply use `component/cli`


#### Code Guidelines

Go uses `gofmt` to automatically format your code, it's best to integrate this into your editor.

Check your code's formatting using `make fmtcheck`.


# Code of Conduct

All community members are expected to adhere to our [code of conduct](./CONDUCT.md).


# License

Manifold's torus-cli is released under the [BSD 3-Clause License](../LICENSE.md).


# Contributor license agreement

For legal purposes all contributors must sign a [contributor license agreement](https://cla-assistant.io/manifoldco/torus-cli), either for an individual or corporation, before a pull request can be accepted.

You will be prompted to sign the agreement by CLA Assistant (bot) when you open a Pull Request for the first time.
