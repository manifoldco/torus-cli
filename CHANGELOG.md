# CHANGELOG

## v0.14.0

**Notable Changes**

- License changed to BSD 3-Clause
- Repository open-sourced, updated README

## v0.13.0

_2016-10-26_

**Notable Changes**

- The alpha waitlist has been removed, any user can now signup for their free account. We've introduced the `torus verify` command allowing users to verify their email addresses if they interrupt the signup flow
- Introduced the `torus ls` command for navigating through all of the organizations, projects, services, environments, and the secrets you have access too.
- Improved error messages across the product, including input validation.

**Fixes**

- Fix to `torus keypairs generate` when used with `--all`
- Fixed invite approval for orgs with secrets set using an or operation (e.g. `ag set -e production -e development secret mysupersecret`).
- Fix to prevent display of credentials which had been unset.

## v0.12.0

_2016-10-18_

**Notable Changes**

- Introduce new `orgs remove <username>` command, to remove a user from an
  org, including their team memberships and secret access.
- Introduce keyring versioning. After a user has been removed from a keyring,
  we increment the keyring version, creating a logical access boundary. New
  secrets are added to the new keyring version, and old secrets can be
  called out as needing to be rotated.
- Introduce the `worklog` command. `worklog` discovers and tracks important
  tasks to do within `torus`. The first type of item it tracks are secrets
  that should be rotated due to users being removed from an organization.

**Fixes**

- Assorted fixes for help text.
- Fixed a bug with `orgs invite send` which prevented a user from being invited
  if any teams were specified.
- `teams remove` no longer panics if a bad team name is supplied.
- `policies detach` no longer panics if too few arguments are supplied.
- Adding an admin or owner to a team with a deny no longer removes their access.

## v0.11.0

_Published: 2016-10-12_

**Breaking**

- The credential schema revision has changed to make unset credentials
  easier to identify. This change is backwards incompatible; `torus` clients
  before `v0.11.0` will error when trying to read credentials set or unset by
  `v0.11.0`+ clients.
- An API change to the server breaks compatability with `torus` clients with
  versions at or above `v0.10.0` and before this version (`v0.11.0`).

**Fixes**

- Grammar fixes in command output.
- The username displayed in `invites list` now has its own column.
- Fix a panic in `keypairs generate` when the supplied org is not found.

**Notable Changes**

- Defaults for the `instance` value have been cleaned up. During `set` and
  `unset`, `instance` defaults to `*` (all instances of a service run by an
  identity). During `view`, `run`, and `status`, it defaults to `1`.
- `torus` now ships with the final production root signing key.

## v0.10.1

_Published: 2016-09-29_

**Fixes**

- Credential names are case insensitive, normalized to lower case. Teach the
  cli to do this before sending credentials to the server.

## v0.10.0

_Published: 2016-09-28_

**Deprecation**:

- All previous versions of `ag` are deprecated and support will cease as of
  October 24th 2016. Please switch to using `torus` by that date.

**Breaking**

- The command line utility has been renamed to `torus` from `ag`.
- All `.arigatorc` and `.arigato.json` files will need to be renamed to
  `.torusrc` and `.torus.json`.
- The arigato root directory has been renamed to `~/.torus` from `~/.arigato`

**Upgrade Instructions**

- If you already have `ag` installed, stop the daemon using `ag daemon stop`
- Uninstall the `ag` using `npm uninstall -g ag`
- Install the new version using `npm install -g torus-cli`
- Rename `.arigato.json` and `.arigatorc` to `.torus.json` and `.torusrc`
  respectively

**Notable Changes**

- All environment variables are now prefixed with `TORUS_` instead of `AG_`
- `torus link` will now generate a `.torus.json` file instead of `.arigato.json`
- `torus prefs` will now read and write to a global `.torusrc` file

**Fixes**

- A secret can no longer be set for a non-existent service, environment, or
  user.

## v0.9.0

_Published: 2016-09-20_

**Notable Changes**

- Command added for viewing policy statements: `ag policies view`

**Fixes**

- Corrected the help message for `ag invites approve`

## v0.8.0

_Published: 2016-09-13_

**Notable Changes**

- The conversion to Go is complete.
- Required external files are now bundled into the Go binary.

## v0.7.0

_Published: 2016-09-08_

**Notable Changes**

- Five commands converted from Node.js to Go (run, view, invites accept, teams create, teams remove`).

**Performance Improvements**

- Significant performance improvements to the run and view commands (60% reduction in execution time).

**Fixes**

- Fixed an issue introduced in v0.6.0 that prevented alpha users from accepting invitations.


## v0.6.1

_Published: 2016-09-08_

**Fixes**

- ag org invites approve, approved the first invite in the list instead of the invite for the supplied email.


## v0.6.0

_Published: 2016-09-07_

**Notable Changes**

- Support for specifying org, project, user, and instance flags using AG_ORG, AG_PROJECT, AG_USER, and AG_INSTANCE environment variables.
- Improved output for listing subcommand help (e.g. ag help orgs)
- UI improvements for all list commands converted to Go
- All prompts now provide inline feedback on input validity
- When creating a service, environment, or project you can now create a parent object in one flow (e.g. while creating a new service you can also create a new org and project).
- Significant performance improvement for all commands converted from Node to Go.

We've converted 29 of 41 total commands from Node.js to Go since our last release (v0.5.0). The 12 remaining commands are listed below.

- ag view
- ag run
- ag allow
- ag deny
- ag invites accept
- ag policies detach
- ag set
- ag teams add
- ag teams create
- ag teams remove
- ag unset
- ag verify


## v0.5.0

_Published: 2016-08-29_

This release marks the first stage of our conversion to go. As such, many changes are structural, and not visible (but they're all still great!)

**Breaking Changes**

- Subcommand structure has changed:
 - Subcommands were previously delimited with a colon (ie `ag envs:create`). They are now delimited with a space (ie `ag envs create`).
 - Top-level commands containing subcommands are now `list` subcommands of the top-level command. For example, the old `ag orgs` is now `ag orgs list`.
 - For more details of the new command structure, please see `ag help` to view all top level commands, `ag <command> --help` to view the subcommands within a top-level command, and `ag <command> <subcommand> --help` to see the help for an individual subcommand.

**Notable Changes**

- `ag run` reads `environment` and `service` from environment variables (`AG_ENVIRONMENT` and `AG_SERVICE`).
- New command: `ag daemon` can display the session daemon's status, and start or stop it.
- `ag login` provides validation feedback while entering email and password.

**Performance Improvements**

- Help output is noticeably faster.
- Server-side performance improvements will speed up most commands.


## v0.4.0

_Published: 2016-08-22_

**Breaking**

- Generating policies via. allow/deny will require  >= v0.4.0.

**Notable**

- Added feedback messages when generating a keypair or encrypting a secret.
- Added the ability to view members of a team and to remove them using ag teams:members and ag teams:remove.

**Fixes**

- If the CLI cancels mid-operation the daemon now cancels its on-going crypto operations.
- The CLI no longer checks the file permissions of the .arigato.json file


## v0.3.0

_Published: 2016-08-17_

**Notable Changes**

- `ag run` now accepts an email and password variables (e.g. AG_EMAIL=my@email.com AG_PASSWORD=my_password). This allows you to automate the login process!
- Listing services via. `ag services` or `ag environments` now takes your context into consideration. To list all projects or environments just use -a, --all.

**Fixes**

- The daemon is now compiled using go 1.7 fixing crashes on MacOS X Sierra.
- `ag run` did not start the process or pass parameters to the child properly, this has been fixed.
