# CHANGELOG

## Unreleased

_Date_

**Breaking Changes**

The concept of instance and identity as a part of the secret path has been
deprecated. Existing secrets set with non `*` identity and instance values can
still be set and unset by providing the full 7 segment path (e.g. `torus set
/org/project/environment/service/identity/instance/secret`).

Torus will only display the full 7 segment path if identity or instance is a
non `*` value (e.g. `/org/project/environment/service/machine-api/1/secret`).
In all other cases, a 5 segment path will be displayed (e.g.
`/org/project/environment/service/secret`).

**Notable Changes**

- The status of your account is now displayed via `torus profile view`
- When updating your email address via `torus profile update` you are now
  prompted for the verification code to re-verify your account.
- More precise errors are returned to users when invalid usernames, org,
  project, team, or policy names are submitted to a ui prompt.
- The experimental and hidden `policies test` command has been removed.
- Added spinners to represent progress. This means fewer lasting print-outs
  for certain commands.
- The `user`, `machine`, and `instance` flags have been removed from `torus
  set`, `torus unset`, `torus import`, `torus export`, and `torus view`.
- Instance and identity values are no longer displayed via `torus status`.
- `torus allow` and `torus deny` now accept a 5 segment path along with the
  deprecated 7 path version (e.g. `torus allow crudl
  /org/project/env/service/secret <team>`).
- `torus policies view` will only display the full 7 segment path if the
  `identity` or `identity` components are not a `*`.
- `torus view` and `torus list` will only display the full 7 segment path in
  verbose mode if the `instance` and `identity` components are not a `*`.
- The current org, project, environment, and service is now injected into the
  process started by `torus run`.

**Fixes**

- `torus list` did not display secrets which were not set with an instance of `*`.

## v0.29.0

_2017-12-22_

**Notable Changes**

- Redesigned the output of `view` and removed the `--format` flag.
- Updated formatting and flags of `envs list`, `services list`, `invites list`,
  `machines list`, `machines view`, `policies list` and `policies view`.
- Added styling, color and org/project prompts throughout commands. Colors can
  be disable by running `torus prefs set core.colors false`.

## v0.28.1

_2017-12-17_

**Fixes**

- Torus will only print out secondary information such as when it's attempting
  to authenticate using credentials from `TORUS_EMAIL`, `TORUS_PASSWORD`,
  `TORUS_TOKEN_ID`, and `TORUS_TOKEN_SECRET` if stdout is attached to a
  terminal window.

## v0.28.0

_2017-12-05_

**Notable Changes**

- Introduced command `orgs members --org ORG` to list all members within an
  organization.
- Changed the output style of `teams members` to match the output style of
  `orgs members --org ORG`.
- Introduced the `torus export` command making it easy to export secrets from a
  specific environment and service. As a result the `torus view` `--format, -f`
  flas has been deprecated and will be removed on December 31st 2017.
- Using `torus export`, you can now export secrets to [terraform's](https://terraform.io) `tfvars` file format.
- Encryption keys, user passwords, and machine secret tokens are now stored in
  secure and guarded memory making it more difficult to extract data from a
  running process.
- Replaced `torus ls` with `torus list` making it easy to list and search for secrets within a project. Listing secrets is now twice as fast as `torus ls`.

**Fixes**

- Fixed a bug preventing old credential values from being decrypted.
- Previously, expiration of a key was set to be one year, instead, we've set it
  to be three. This fixes a bug which prevented users from setting secrets as
  their keys had expired. A corresponding change was made to the torus server.

## v0.27.0

_2017-11-08_

**Notable Changes**

- Significant performance increases when secrets are sourced from multiple
  keyrings. For example if a secret is brought in from `dev-*` and `dev-user`
  Torus will no longer unseal the private encryption key twice which leads to a
  signficiant reduction in decryption time. Users should notice this
  improvement when using `torus view` and `torus run`.
- Significant reduction in the number of round trips made to the Torus server
  to fetch an organizations claimtree (web of trust) when decrypting or
  encrypting secrets. Users should notice this improvement when many different
  users are contributing secrets to the same keyrings when using `torus view`,
  `torus run`, `torus set`, or `torus unset`.
- Parallelized fetching of keypairs and an orgs claimtree during secret
  decryption. Users should notice a modest improvement when using `torus view`
  and `torus run`.

**Fixes**

- Request timeout to the server has been increased from 6s to 60s.
- Fixed a bug preventing a user's invitation from being approved after a user
  was removed from the organization.

## v0.26.1

_2017-10-28_

**Fixes**

- `torus import` will now use the provided key instead of the value
- Users are now reminded to keep their password safe and secure during `torus signup`.
- No longer prompt to enable or disable hints during signup, instead always enable them.
- Actually set the default value of the `core.check_updates` preference to `true`
- Add a tip after signup to generate a `.torus.json` file using `torus link`

**Thanks**

- PatDuJour

## v0.26.0

_2017-10-27_

**Notable Changes**

- Introduced `torus policies attach` allowing a user to attach a policy to
  multiple teams or machine roles.
- Introduced `torus policies delete` allowing a user to delete a policy and all
  of it's attachment from an org. System policies cannot be deleted.
- When generating a policy using `torus allow` or `torus deny` you can now
  specify it's name and description using the `--name` and `--description`
  flags. If no description is provided, one will be generated.

**Fixes**

- Clarify the behaviour of the `--environment`, `--service`, `--instance`,
  `--user`, and `--machine` flags when reading or writing secrets.

## v0.25.2

_2017-10-19_

**Fixes**

- Fixed a bug preventing Torus from being used once installed via npm on win32.

## v0.25.1

_2017-10-16_

**Fixes**

- Fixed a bug preventing Torus being installed from a Brew formula

## v0.25.0

_2017-10-13_

**Notable Changes**

- You can now install the windows client via `npm` (e.g. `npm install -g torus-cli`).
- Multiple secrets can be imported at once from a `.env` file using `torus
  import` (e.g. `torus import .env`).

**Fixes**

- Torus can now be installed on Mac OS X via `brew`.
- `torus signup` will no longer error unexpectedly if you provide name with
  less than 3 characters.
- Changing your password using `torus profile update` will no longer lock you
  out of your account.
- The daemon will no longer crash if it cannot reach `get.torus.sh` during
  version checking.
- New version checking has been re-enabled after being disabled in `v0.24.2`
  whcih will be checked at startup of the daemon and every day at 6am.
- Torus is now compiled using go1.9.1

**Thanks**

- Luiz Branco

## v0.24.2

_2017-09-25_

**Fixes**

- Disabled version checking against `get.torus.sh` as a temporary work around
  to torus DNS outage.
- Disabled update checking by default if a `~/.torusrc` does not already exist.

## v0.24.1

_2017-05-31_

**Fixes**

- Hints will no longer be displayed if stdout is not a terminal.
- The CLI will now wait indefinitely for a request to be completed by the daemon.

## v0.24.0

_2017-05-24_

**Notable Changes**

- `torus set` now supports `name=path` syntax (e.g. `torus set foo=bar` or
  `torus set /org/project/env/*/*/*/foo=bar`)
- We now refer to `Name` as `Full Name` to differentiate between a user's full
  name and username.

**Thanks**

- Luiz Branco

## v0.23.0

_2017-05-17_

**Notable Changes**

- `keyring` type worklog items are now organized by user, not keyring. Keyrings
  are internal structures that hold secrets; they shouild rarely appear in the
  UI. Focusing on users that are missing access they should have is much more
  understandable.
- Torus now checks for available updates to itself, and reports on them during
  the `login` and `version` commands. This behaviour can be disabled with
  `torus prefs set core.check_updates false`.
- Exciting new worklog ui:
  - Items are grouped by type, making the display more compact and usable.
  - Lots of color and formatting!
  - Each worklog item includes details visible with `view`. For example, secret
    rotation items include which users caused the need for rotation, and why
    (i.e. 'james was removed from the org.').
- A beta version of the windows client is now available on
  [get.torus.sh](https://get.torus.sh)!

**Fixes**

- Correct the help message for `invites accept`'s org flag.
- Fixed a problem where machine's with a name containing `machine-` (but as a
  prefix) could not interact with credential.

**Security**

- Added documentation to the README.md regarding the default security profile
  of Torus on Windows.

**Thanks**

- Federico Ruggi
- Jelmer Snoeck

## v0.22.0

_2017-01-17_

**Notable Changes**

- Publish release details to GitHub as proper releases.
- Show more details in the summary of invite and keypairs worklog items.
- Passphrase derived public key authentication (PDPKA) is now used to
  authenticate users. Old users will be upgraded to support this auth method on
  their next login once they've upgraded to the latest version of Torus. New
  users will support PDPKA out of the box. Once a user has upgraded to support
  PDPKA, HMAC authentication is no longer supported.
- When creating a project, a `default` service is always created as well. As a
  result, the `--bare` option has been removed from `torus link`.

**Fixes**

- If a user is missing access to a keyring, but they do not yet have a valid
  keypair, don't alert other users to add this user to the keyring; they won't
  be able to!
- Removed forgotten debug logs from appearing in `~/.torus/daemon.log`

## v0.21.1

_2016-12-20_

**Security**

- Resolved information leak to daemon log file during machine login.

## v0.21.0

_2016-12-16_

**Notable Changes**

- Support Ubuntu 16.10 (Yakkety Yak) for deb packages.
- Secrets set on the command line are now always treated as strings. Previously,
  We would attempt to convert to ints or floats. Torus doesn't know if
  you want `-007` to be a string suffix for your spy identifiers, or the number
  `-7`; so no longer guess, and use the provided value.
  This change will affect newly set values, but not existing ones.

**Fixes**

- Ensure `keypairs generate` does not panic when used against an org that has
  existing keypairs.
- Teach `keypairs list` to display the real validity state of a key, not just
  always "YES".
- Under NPM/Node.js, run via a passthrough script that will select the right
  binary. This replaces the previous install time symlinking script, which
  was error prone and unusable with sudo installs in some cases.
- Skip over users without encryption keys when storing secrets, instead of
  erroring out, allowing other users to still access the secrets.
- Teach the `keypairs` `worklog` item how to handle users that have been
  removed from a keyring (or had their keys revoked), and then subsequently
  re-added: The old secret values still require rotation, but the user can be
  given access to the secrets once again.
- Allow non-admin users to run worklog list, by continuing passed unauthorized
  requests when looking at invites. Only admins can view invites.

## v0.20.0

_2016-12-13_

**Notable Changes**

- Update the style of selection lists for improved readability.
- Added hint output to core commands, prompting the user during signup if they
  wish to enable them.
- Confirm dialogues now show default value as uppercase
- Teach `worklog` how to identify and fix cases where users or machines
  haven't been included in a keyring for secrets access when they should be.

**Fixes**

- Resolved possible race condition in the progress notification code.
- Ensure the user is logged in when trying to create an org.

**Thanks**

- Jelmer Snoeck

## v0.19.0

_2016-12-08_

**Notable Changes**

- Add preferences for `core.disable_progress` and `core.disable_hints` to
  control levels of output in preparation for guided on-boarding.
- Support vim movement bindings for interactive inputs. This can be enabled
  with `torus prefs set core.vim true`.
- Support `**` in path expressions passed to commands.  `torus set
  /org/project/**/port 5000` is equivalent to `torus set
  /org/project/*/*/*/*/port`
- `torus ls` behaviour changed to follow system `ls` more closely, no longer
  supporting context or command flags (e.g. `--org, --project, etc`).
- `torus worklog list` now displays a friendly message if no actions need to be
  taken.
- `torus prefs list` now displays the default values for preferences if no
  override has been set by the user.
- Added directory styles to [get.torus.sh](https://get.torus.sh)
- Updated validation for `torus allow` and `torus deny` to catch when secret name is missing
- `torus ls` no longer filters out credentials with the same name based on specificity.

**Fixes**

- Ensure that `torus version` will always return, even if the upstream server
  is misconfigured.
- Fixed an issue where the wrong version of a credential would be used after a
  user was removed from an org.
- Fixed an issue where the wrong version of a credential would be displayed if
  more than two credentials of the same name existed inside the same keyring.

**Docs**

- Added documentation for `torus worklog resolve`

**Thanks**

- Ben Tranter

## v0.18.0

_2016-11-30_

**Notable Changes**

- Re-organization of commands and editing of help output.
- Include a systemd service unit with deb packaging, to run the torus daemon in
  a system wide machine mode. When the unit is running, users in the `torus`
  group can access it. To run the unit, both `TORUS_TOKEN_ID` and
  `TORUS_TOKEN_SECRET` must be set in `/etc/torus/token.environment`.
  See v0.17.0 for the matching rpm change.
- Teach `worklog` about missing user keypairs.
- Teach `worklog` about approving invites.
- Unhide `worklog resolve`, as it can now be used to generate missing keypairs
  for an org, or approve an invite.

**Fixes**

- Fixed "unauthorized" error which occurred while updating email and password at
  the same time.
- Improve message for `machines list` when no machines are found.
- When encrypting or signing, do not use revoked keypairs.
- `torus ls` now returns *all* secrets that match the given path, if a `*` was
  provided or the path contained an alternation it wouldn't have been returned.

## v0.17.0

_2016-11-15_

**Notable Changes**

- Introduced `--format, -f` to `torus view` for specifying the format of out the output (env, json, verbose).
- Updated the `--verbose, -v` option for `torus view` to be a shortcut to `--format verbose`.
- Include a systemd service unit with the rpm packaging, to run the torus daemon in a system wide machine mode. When the unit is running, users in the `torus` group can access it. To run the unit, both `TORUS_TOKEN_ID` and `TORUS_TOKEN_SECRE` must be set in `/etc/torus/token.environment`.
- Introduced `torus profile update` for changing the current users name, email, or password.
- Introduced `torus profile view` for displaying current identity, removing such information from `torus status`
- Began publishing deb, rpm, brew, and binary releases at [get.torus.sh](https://get.torus.sh) increasing the number of ways you can download and install `torus`.

## v0.16.0

_2016-11-09_

**Notable Changes**

- Introduced `--user, -u` and `--machine, -m` flags to `torus set`, `torus
  unset`, `torus view`, `torus run`, and `torus ls` for specifying machine or
  user identity
- Introduce `machines roles list` and `machines roles create` commands for
  viewing and creating machine roles.
- Machine teams no longer appear under `teams list` nor can you view machine
  teams through `teams members`.
- The `machines` command now appears under the `ORGANIZATIONS` category when
  listing commands with `torus help`.
- Introduce more release formats: npm, binary/zip, rpm/yum, & homebrew
- Provide more detailed error messages.

**Fixes**

- Listing teams no longer results in a panic when an unknown org is specified.
- `torus status` properly displays the identity segment for a machine in the credential path.
- Various typo fixes.

## v0.15.0

_2016-11-01_

**Notable Changes**

- Added Contributor Guide, CLA and Code of Conduct as a part of our open sourcing effort
- Introducing the ability to create, list, view, and destroy machines to support secret access in automated environments (e.g. continuous integration or production).

**Fixes**

- Errors encountered during an interactive prompt are no longer hidden, they are surfaced to the user.

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
