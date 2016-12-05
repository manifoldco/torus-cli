# Organizations
## orgs
Every user who signs up for Torus is given a personal organization. Here they can collaborate on all of their personal projects. Each default org is given the same name as the user’s username automatically.

Users can have as many organizations as their account allows (currently unlimited).

### create
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus orgs create [name]` creates a new organization.

Each organization name is globally unique and must adhere to the system naming scheme. If no name argument is supplied, the user will be prompted to enter the new org’s name.

### list
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus orgs list` displays all organizations the current session has access to.

### remove
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus orgs remove [username]` removes the specified user from the specified organization.

## keypairs
Every user/machine in the Torus ecosystem has both a signing and an encryption key per-organization. These key pairs are generated when an entity joins an organization.

The keys are used to sign and encrypt the objects you interact with inside Torus and facilitate sharing and auditing. If you do not have the requisite keys they will need to be generated before most operations can take place.

### list
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus keypairs list` displays the available key pairs for the specified organization.

### generate
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus keypairs generate` creates the requisite key pairs (that are missing) for the specified organization.

## worklog
Torus worklog facilitates maintenance tasks which are generated as a result of actions taken throughout your organization (for example: a secret needs to be rotated due to a user being removed from the org).

### list
###### Added [v0.12.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus worklog list` displays all pending work items for the specified organization.

### view
###### Added [v0.12.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus worklog view <identity>` displays the details of an individual worklog item inside the specified organization.

### resolve
###### Added [v0.18.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus worklog resolve [identity...]` attempts to automatically resolve the
given worklog items, or all worklog items within the org if no identies are
specified.

Not all worklog items can be automatically resolved. For instance, secret
rotation; Torus doesn't know the new value you've chosen for a secret!

## invites
Users want to share their secrets with other users. To do this we allow users to invite others to join an organization and collaborate on that project structure according to pre-established and user-defined [access controls](./access-control.md).

Inviting a user to an organization is a multi-step process:
1.  Send an invite to the user’s email
2.  User must accept the invite
3.  Admin must approve the invite and complete their induction to the org

Only organization administrators (including owners) may send invitations.

### send
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus invites send [email]` sends an invite code to the supplied email address for the specified organization.

By default the user is invited to join the `member` team. This can be changed/augmented using command options.

### list
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus invites list` displays all outstanding invitations to the specified organization.

By default only invites which have not yet been approved will be shown.

### Command Options

Option | Description
---- | ----
--approved | Display approved invites instead of pending

### approve
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus invites approve <email>` finalizes the end-user’s membership to the organization. To be approved it must already be accept by the individual it was sent to.

### accept
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus invites accept <email> <code>` accepts an invitation to join the specified organization.

A user will receive their invite code by email when sent an invitation to join the organization. When executed the user will be prompted if they would like to log in to an existing Torus account or sign up for a new account.

The resulting authenticated account will be the one added to the org.

## machines
Machines are a method of authenticating systems which are not owned by an individual (i.e. server instances).

Each machine is given an ID and Token value which are synonymous with a user’s email and password. These are used to authenticate the CLI on a per-system basis.

Roles are assigned to each machine for [access control](./access-control.md).

### create
###### Added [v0.15.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus machines create [name]` creates a new machine for the specified organization.

A machine is given a unique name within the organization that adheres to the system naming scheme.

### list
###### Added [v0.15.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus machines list` displays all available machines for the specified organization.

The Machine listing can be filtered using the role or destroyed command options.

### view
###### Added [v0.15.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus machines view <id|name>` displays a machine’s details by id or name for the specified organization.

### destroy
###### Added [v0.15.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus machines destroy <id|name>` destroys a machine by id or name for the specified organization.

### roles
Machines are given roles (similar to how users are added to teams) which enable you to finely control what a machine has access to when deployed.

Once created a role can have any number of policies attached to it.

#### create
###### Added [v0.16.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus machines roles create <name>` creates a new role for the specified organization.

A role is given a unique name within the organization’s roles that adheres to the system naming scheme.

#### list
###### Added [v0.16.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus machines roles list` displays all available roles for the specified organization.
