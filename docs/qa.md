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

If you have `ag` installed, start fresh `npm uninstall -g ag`

- [ ]   `npm install -g ag` installs `ag`
- [ ]   `ag help` displays the help prompt after an `npm install -g ag`
- [ ]   `ag prefs list` displays the path to your `public_key_file`
- [ ]   `ag signup` prompts you for an access code, username, name and email, before verifying and authenticating you
- [ ]  A user cannot sign-up without a valid access code
- [ ]   `ag status` displays your current working context
- [ ]   `ag logout` logs you out
- [ ]   `ag login` prompts you for an email and password, before authenticating you

### Teams

- [ ]   `ag teams list --org [username]` displays `owner` `admin` and `member` teams, and you are a member of each
- [ ]   `ag teams create [name] --org [org-name]` creates an org.
- [ ]   `ag invites send [email] —org [org-name]` generates an access code and sends it to the user
- [ ]   `ag invites accept —org [org] [email] [code]` prompts a user to sign-up or login
- [ ]  The user cannot manage resources in this org
- [ ]   `ag invites list —org [org]` lists the outstanding invite
- [ ]   `ag invites approve [email] —org [org]` approves the users invite
- [ ]  The user is now a member of the members team
- [ ]   `ag teams add [name] [username] —org [org-name]` adds the user to a team
- [ ]   `ag teams remove [name] [username] —org [org-name]` removes the user from a team
- [ ]  Users cannot remove themselves from system-teams

### Context

- [ ]   `ag status` will not display a context outside of a linked directory
- [ ]   `ag link` prompts you to select or create an org, and project
- [ ]   `ag link -f` will over-write a previous `ag link`
- [ ]   `ag status` now displays a valid: Identity, username, org, project, environment, service, instance and path
- [ ]  Commands can now be executed without the `—org` and `--project` flags

### Access Controls

- [ ]   `ag policies list --org [org]` displays three `system-default` policies attched to the `owner` , `admin` , and `member` teams
- [ ]   `ag allow crudl [path] [team]` generates a policy, displays an `allow` effect, appropriate actions, the correct resource. and attaches it `[team]`
- [ ]  Members of `[team]` have appropriate access
- [ ]   `ag deny crudl [path] [team]` generates a policy, with a `deny` effect, appropriate actions, the correct resource, and attaches it to `[team]`
- [ ]  Members of `[team]` are denied access to the appropriate resource
- [ ]   `ag policies list —org [org]` displays the generated policies attached to the appripriate `[team]`
- [ ]   `ag policies detach [name] [team] —org [org]` detaches the policy from the team
- [ ]  Members of `[team]` have appropriate access
- [ ]  System policies cannot be detached
- [ ]  `ag policies view [name]` displays statements contained within the policy

### Critical Path

- [ ]   `ag orgs create [name]` prompts you to confirm the org name, and creates it
- [ ]   `ag projects create [name] —org [org]` prompts you to confirm the project name, and creates it
- [ ]   `ag services create [name] —project [project] —org [org]` creates a service.
- [ ]   `ag set [key] [value] —service [service]` will set the variable
- [ ]  You can not set variables in services that do not exist
- [ ]   `ag set [path] [value]` will set the variable
- [ ]   `ag view —service [service]` will list secrets
- [ ]  A secret can be set across services
- [ ]  A secret can be set across environments
- [ ]  More specific paths win
- [ ]  Latest set wins
- [ ]   `ag run <command> —service [service]` will inject secrets and execute the command
- [ ]   `ag run <command>` without a service defaults to "default" service
- [ ]   `ag unset [key] —service [service]` will unset the variable
