# Access Control
The Torus CLI can be used to restrict access to secrets stored within your org. To control access to specific objects we will be using [Path](../concepts/path.md) expressions.

## teams
Users in Torus can be grouped into Teams inside an Organization. These teams have segmented access to the secrets which have been saved.

Each command within this group must be supplied an Organization flag using `--org <name>`, or `-o <name>` for short. The organization can also be supplied by executing these commands within a [linked directory](./project-structure.md#link).

Organizations have three default teams:

- Member
- Admin
- Owner

The creator of the organization is automatically added to all three teams. Anyone invited to the org is automatically added to the member team (and cannot be removed from that team without removing them from t#he organization).

Only users who are a member of the "admin" team can manage resources within an organization.

### Command Options

All teams commands accept the following flags:

  Option | Environment Variable | Description
  ---- | ---- | ----
  --org, ORG, -o ORG | TORUS_ORG | The org the team or teams belong to

### create
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus teams create [name]` will create a new team for the specified organization.

A team is given a unique name within the organization that adheres to the naming scheme. If no name argument is supplied, the user will be prompted to enter the new team’s name.

### list
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus teams list` displays all available teams for the specified organization. Each team that the authenticated user is currently a member of will be noted with a `*`.

### members
`torus teams members <name>` displays all members for the specified team name. Should the authenticated user be a member of the team they will be noted with a `*`.

To display all members of your organization display the members of the "member" team using `torus teams members member`. This is useful when using the add and remove commands.

**Example**

```
$ torus teams members member --org matt

    USERNAME   NAME
*   matt       Matt Wright
    barnaby    barnaby

team member has (2) members.
```

### add
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus teams add <username> <name>` adds the user specified (by username) to the specified team.

To display all members in your organization see the [`members`](./#members) command.

### remove
###### Added [v0.4.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus teams remove <username> <name>` removes the user specified (by username) from the specified team.

Users cannot be removed from the "member" team. Owners cannot remove themselves from the "owner" team.

## policies
Access to resources is controlled using documents that define access called Policies.

A policy contains rules about which actions can be taken on specific resources. Policies are then attached to teams, which enables you to control access in a very specific manner.

Each default team has its own default policy, and can have additional policies attached to the team. You can view the [complete default policies here](../concepts/policies.md), in short:

**Member team:** cannot manage resources, can set credentials in their own "dev-username" environment and share credentials through the "dev-\*"
environment.

**Admin team:** has mostly full access. Can manage all resources: teams, policies, projects, environments, etc. Cannot add members to "owner" team.

**Owner team:** can do everything found in admin, but can add members to the "owner" team.

Each command within this group must be supplied an Organization flag using `--org <name>`, or `-o <name>` for short. The organization can also be supplied by executing these commands within a [linked directory](./project-structure.md#link).

### Command Options

All policies commands accept the following flags:

  Option | Environment Variable | Description
  ---- | ---- | ----
  --org, ORG, -o ORG | TORUS_ORG | The org the policy or policies belong to

### list
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus policies list` displays all available policies for the specified organization.

Each row has a name, type (system or member), and list of teams and machine roles the policy is attached to.

### view
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus policies view <name>` displays all of the rules within the named policy.

Each row has the effect (allow or deny), the list of actions (crudl - create, read, update, delete, list), and the resource path.

### attach
###### Added [v0.26.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus policies attach <name> <team|machine-role>` attaches the policy (identified by the poliy name) to the given team (or machine role).

This enables you to re-use policies by attaching one to multiple teams and machine roles.

### detach
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus policies detach <name> <team|machine-role>` detaches the policy (identified by name) from the given team (or machine role).

This enables you to lift restrictions (or grants) from a team.

### delete
###### Added [v0.26.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus policies delete <name>` deletes a policy and all of it's attachments (identified by name) from the given org.

This enables you to permanently delete a policy from an organization.

#### Command Options

The delete command accepts the following additional flags:

  Option | Description
  ----   | -----
  --yes, -y | Automatically accept the confirm dialog

## allow
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus allow <crudl> <path> <team|machine-role>` generates a new policy and attaches it to the given team (or machine role). The policy created is given a generated name.

CRUDL (create, read, update, delete, list) represents the actions that are being granted. The supplied Path represents the resource that you are enabling the aforementioned actions on.

If a name (`--name` flag) is not provided, one will be automatically generated.

#### Command Options

The `allow` command accepts the following flags:

  Option | Environment Variable | Description
  ----   | ----- | ----
  --org ORG, -o ORG | TORUS_ORG | The org to generate the policy for
  --name NAME, -n NAME | TORUS_NAME | The name to give the generated policy (e.g. allow-prod-env)
  --description DESCRIPTION, -d DESCRIPTION | TORUS_DESCRIPTION | A sentence or two explaining the purpose of the policy

**Example**

```bash
# Create a policy allowing it's subjects to read secrets from prod environment
# in the api project which belongs to the myorg organization and attach it to
# the api-prod-machines machine role.
$ torus allow -n read-api-prod-env rl /myorg/api/prod/** api-prod-machines
```

## deny
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus deny <crudl> <path> <team|machine-role>` generates a new policy and attaches it to the given team (or machine role). The policy created is given a generated name.

CRUDL (create, read, update, delete, list) represents the actions that are being denied (or restricted). The supplied Path represents the resource that you are disabling the aforementioned actions on.

If a name (`--name` flag) is not provided, one will be automatically generated.

#### Command Options

The `deny` command accepts the following flags:

  Option | Environment Variable | Description
  ----   | ----- | ----
  --org ORG, -o ORG | TORUS_ORG | The org to generate the policy for
  --name NAME, -n NAME | TORUS_NAME | The name to give the generated policy (e.g. allow-prod-env)
  --description DESCRIPTION, -d DESCRIPTION | TORUS_DESCRIPTION | A sentence or two explaining the purpose of the policy
