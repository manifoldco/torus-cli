# QA Checklist Template

This template contains the manual qa checklist used for testing a release of
the CLI or Registry. Depending on the changelist a component of this checklist
may be skipped.

Copy it into the body of a comment in the release issue. Check the boxes and
note any failures by linking the appropriate bug.

### Installation and Signup

If you have `torus` installed, start fresh `npm uninstall -g torus-cli`

- [ ]   `npm install -g torus-cli` installs `torus`
- [ ]   `torus help` displays the help prompt after an
        `npm install -g torus-cli`
- [ ]   `torus prefs list` displays the path to your `public_key_file`
- [ ]   `torus signup` prompts you for an verification code username, name and
        email, before verifying and authenticating you
- [ ]   A user cannot perform any writes without verifying their account
- [ ]   `torus status` displays your current working context
- [ ]   `torus logout` logs you out
- [ ]   `torus login` prompts you for an email and password, before
        authenticating you
- [ ]   You can login using environment variables (`TORUS_EMAIL` and
        `TORUS_PASSWORD`)

### Account

- [ ]   `torus profile view` displays your current identity
- [ ]   `torus profile update` allows you to change your name, email and
        password.
- [ ]   Changing email generates a new verification code
- [ ]   A user cannot perform any writes without verifying their account

### Teams

- [ ]   `torus teams list --org [username]` displays `owner` `admin` and
        `member` teams, and you are a member of each
- [ ]   `torus teams create [name] --org [org-name]` creates an org.
- [ ]   `torus invites send [email] —org [org-name]` generates an access code
        and sends it to the user
- [ ]   `torus invites accept —org [org] [email] [code]` prompts a user to
        sign-up or login
- [ ]   The user cannot manage resources in this org
- [ ]   `torus invites list —org [org]` lists the outstanding invite
- [ ]   `torus invites approve [email] —org [org]` approves the users invite
- [ ]   The user is now a member of the members team
- [ ]   `torus teams add [name] [username] —org [org-name]` adds the user to a
        team
- [ ]   `torus teams remove [name] [username] —org [org-name]` removes the user
        from a team
- [ ]   Users cannot remove themselves from system-teams

### Context

- [ ]   `torus status` will not display a context outside of a linked directory
- [ ]   `torus link` prompts you to select or create an org, and project
- [ ]   `torus link -f` will over-write a previous `torus link`
- [ ]   `torus status` now displays a valid: org, project, environment, service,
        instance and path
- [ ]   Commands can now be executed without the `—org` and `--project` flags

### Access Controls

- [ ]   `torus policies list --org [org]` displays three `system-default`
        policies attched to the `owner` , `admin` , and `member` teams
- [ ]   `torus allow crudl [path] [team]` generates a policy, displays an
        `allow` effect, appropriate actions, the correct resource. and attaches
        it to `[team]`
- [ ]   Members of `[team]` have appropriate access
- [ ]   `torus deny crudl [path] [team]` generates a policy, with a `deny`
        effect, appropriate actions, the correct resource, and attaches it to
        `[team]`
- [ ]   Members of `[team]` are denied access to the appropriate resource
- [ ]   `torus policies list —org [org]` displays the generated policies
        attached to the appripriate `[team]`
- [ ]   `torus policies detach [name] [team] —org [org]` detaches the policy
        from the team
- [ ]   Members of `[team]` have appropriate access
- [ ]   System policies cannot be detached
- [ ]   `torus policies view [name]` displays statements contained within the
        policy

### Worklog/Keyring Versioning

- [ ]   *prerequisite:* Create an org, and invite another user.
- [ ]   *prerequisite:* Create a team, add the other user to it, and give it
        access to a project (`projA`)
- [ ]   *prerequisite:* Create a second project (`projB`) that the other user
        does not have access to.
- [ ]   *prerequiresite:* Set two secrets in both projects
- [ ]   `torus orgs remove [username] —org [org-name]` removes the user from the
        org.
- [ ]   `torus worklog list` shows the secrets from `projA` as needing to be
        changed, but does not show the secrets from `projB`.
- [ ]   `torus worklog view <id>` shows a worklog entry by id.
- [ ]   After each secret from `projA` is changed, it no longer appears in the
        worklog.
- [ ]   `torus view` can display a mixture of old and new secrets (ie after a
        single new set in `projA`).

### List

- [ ]   `torus ls`, in an unlinked directory, list all orgs
- [ ]   `torus ls` works in a linked directory, showing all paths for the
        context
- [ ]   `torus ls /` lists all orgs
- [ ]   `torus ls "/*"` lists all projects
- [ ]   `torus ls "/*/*"` lists all environments
- [ ]   `torus ls "/*/*/*"` lists all services
- [ ]   `torus ls "/$org/$proj/*/*"` lists all secrets for `/$org/$proj/*/*/*`

### Machines

- [ ]   Only admins and owners can create machines
- [ ]   `torus machines create` works with context
- [ ]   `torus machines create` supports flags (e.g. `-o, -t` etc)
- [ ]   `torus machines create` allows you to select an existing org and team
- [ ]   `torus machines create` allows you to create a new machine team
- [ ]   You can login using `TORUS_TOKEN_ID` and `TORUS_TOKEN_SECRET`
        environment variables
- [ ]   A machine can read but not write (e.g. `view, run, ls, envs list,
        services list, projects list, orgs list, status`)
- [ ]   `torus machines list` displays all machines
- [ ]   `torus machines list --role [role]` shows machines belonging to that
        team
- [ ]   `torus machines list --destroyed` shows destroyed machines
- [ ]   `torus machines view [identity]` shows a single machine's details by id
- [ ]   `torus machines view [name]` shows a single machine's details by name
- [ ]   `torus machines destroy [identity]` destroys a machine by id after
        confirming
- [ ]   `torus machines destroy [name]` destroys a machine by name after
        confirming
- [ ]   `torus machines list --destroyed` shows only destroyed machines
- [ ]   `torus machines roles create` creates a machine role
- [ ]   `torus machines roles lists` lists all machine roles (including the
        machine team)

### Critical Path

- [ ]   `torus orgs create [name]` creates an org.
- [ ]   `torus projects create [name] —org [org]` creates a project.
- [ ]   `torus services create [name] —project [project] —org [org]` creates a
        service.
- [ ]   `torus set [key] [value] —service [service]` will set the variable
- [ ]   You can set using environment variables (e.g. `TORUS_ORG`)
- [ ]   You can not set variables in services that do not exist
- [ ]   `torus set [path] [value]` will set the variable
- [ ]   `torus view —service [service]` will list secrets
- [ ]   A secret can be set across services
- [ ]   A secret can be set across environments
- [ ]   More specific paths win
- [ ]   Latest set wins
- [ ]   `torus run <command> —service [service]` will inject secrets and execute
        the command
- [ ]   `torus run <command>` without a service defaults to "default" service
- [ ]   `torus unset [key] —service [service]` will unset the variable
