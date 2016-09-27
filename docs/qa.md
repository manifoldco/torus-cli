# QA Checklist Template

This template contains the manual qa checklist used for testing a release of
the CLI or Registry. Depending on the changelist a component of this checklist
may be skipped.

Copy it into the body of a comment in the release issue. Check the boxes and
note any failures by linking the appropriate bug.

### Waitlist QA Flow

- [ ]   [https://arigato-www-staging.herokuapp.com](https://arigato-www-staging.herokuapp.com) loads
- [ ]  The disclaimer expands and collapses
- [ ]   **Every ** link on the page works as you would expect it to
- [ ]  If you enter an (email@arigato.sh) email address and click `Add me to the alpha` , you are added to the waitlist
- [ ]  By default, `video-1` is selected
- [ ]  All three videos play
- [ ]  You can leave emoji feedback
- [ ]  You can initiate an intercom session using the intercom 'bubble'
- [ ]  You can initiate an intercom session by click `Send us your feedback`

### Installation and Signup

If you have `torus` installed, start fresh `npm uninstall -g torus-cli`

- [ ]   `npm install -g torus-cli` installs `torus`
- [ ]   `torus help` displays the help prompt after an `npm install -g torus-cli`
- [ ]   `torus prefs list` displays the path to your `public_key_file`
- [ ]   `torus signup` prompts you for an access code, username, name and email, before verifying and authenticating you
- [ ]  A user cannot sign-up without a valid access code
- [ ]   `torus status` displays your current working context
- [ ]   `torus logout` logs you out
- [ ]   `torus login` prompts you for an email and password, before authenticating you
- [ ]  You can login using environment variables (`TORUS_EMAIL` and `TORUS_PASSWORD`)
### Teams

- [ ]   `torus teams list --org [username]` displays `owner` `admin` and `member` teams, and you are a member of each
- [ ]   `torus teams create [name] --org [org-name]` creates an org.
- [ ]   `torus invites send [email] —org [org-name]` generates an access code and sends it to the user
- [ ]   `torus invites accept —org [org] [email] [code]` prompts a user to sign-up or login
- [ ]  The user cannot manage resources in this org
- [ ]   `torus invites list —org [org]` lists the outstanding invite
- [ ]   `torus invites approve [email] —org [org]` approves the users invite
- [ ]  The user is now a member of the members team
- [ ]   `torus teams add [name] [username] —org [org-name]` adds the user to a team
- [ ]   `torus teams remove [name] [username] —org [org-name]` removes the user from a team
- [ ]  Users cannot remove themselves from system-teams

### Context

- [ ]   `torus status` will not display a context outside of a linked directory
- [ ]   `torus link` prompts you to select or create an org, and project
- [ ]   `torus link -f` will over-write a previous `torus link`
- [ ]   `torus status` now displays a valid: Identity, username, org, project, environment, service, instance and path
- [ ]  Commands can now be executed without the `—org` and `--project` flags

### Access Controls

- [ ]   `torus policies list --org [org]` displays three `system-default` policies attched to the `owner` , `admin` , and `member` teams
- [ ]   `torus allow crudl [path] [team]` generates a policy, displays an `allow` effect, appropriate actions, the correct resource. and attaches it `[team]`
- [ ]  Members of `[team]` have appropriate access
- [ ]   `torus deny crudl [path] [team]` generates a policy, with a `deny` effect, appropriate actions, the correct resource, and attaches it to `[team]`
- [ ]  Members of `[team]` are denied access to the appropriate resource
- [ ]   `torus policies list —org [org]` displays the generated policies attached to the appripriate `[team]`
- [ ]   `torus policies detach [name] [team] —org [org]` detaches the policy from the team
- [ ]  Members of `[team]` have appropriate access
- [ ]  System policies cannot be detached
- [ ]  `torus policies view [name]` displays statements contained within the policy

### Critical Path

- [ ]   `torus orgs create [name]` prompts you to confirm the org name, and creates it
- [ ]   `torus projects create [name] —org [org]` prompts you to confirm the project name, and creates it
- [ ]   `torus services create [name] —project [project] —org [org]` creates a service.
- [ ]   `torus set [key] [value] —service [service]` will set the variable
- [ ]  You can set using environment variables (e.g. `TORUS_ORG`)
- [ ]  You can not set variables in services that do not exist
- [ ]   `torus set [path] [value]` will set the variable
- [ ]   `torus view —service [service]` will list secrets
- [ ]  A secret can be set across services
- [ ]  A secret can be set across environments
- [ ]  More specific paths win
- [ ]  Latest set wins
- [ ]   `torus run <command> —service [service]` will inject secrets and execute the command
- [ ]   `torus run <command>` without a service defaults to "default" service
- [ ]   `torus unset [key] —service [service]` will unset the variable
