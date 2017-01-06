# Project Structure
Torus secrets are part of a hierarchy, which is represented by a [path](../concepts/path.md). The segments that the path is comprised of are the project structure objects.

## projects
A project is a grouping of services, environments and secrets typically synonymous with a codebase. It is the first level of segmentation inside of an organization.

### create
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus projects create [name]` creates a new project for the specified organization.

A project is given a unique name within the organization that adheres to the system naming scheme. If no name argument is supplied, the user will be prompted to enter the new project’s name.

When a project is created, a `default` service is created as well.

### list
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus projects list` displays all projects for the specified organization.

## services
A service is an entity synonymous with an application process.

With the advent of micro-service architectures, projects are having more than one application process and in a lot of cases these need unique configuration.

### create
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus services create [name]` creates a new service for the specified organization.

A service is given a unique name within the organization that adheres to the system naming scheme. If no name argument is supplied, the user will be prompted to enter the new service’s name.

### list
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus services list` displays all services for the specified organization.

## envs
An environment is a grouping of services that live within a project which have their own configuration requirements.

By default every user in an organization is given a `dev-$username` environment (where $username is the user’s profile username). This is synonymous will a user’s local environment.

Users are also given access to share credentials across all users using the `dev-*` environment. This is a paradigm that exists through the default [access controls](./access-control.md) (and doesn’t actually exist as an environment object for the organization).

### create
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus envs create [name]` creates a new environment for the specified organization.

An environment is given a unique name within the organization that adheres to the system naming scheme. If no name argument is supplied, the user will be prompted to enter the new environment’s name.

### list
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus envs list` displays all services for the specified organization.

## link
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus link` enables a user to tie their current working directory to a selected organization and project. The selected values are output in a `.torus.json` file which can be committed to source control (so that other users in the organization can take advantage of it).

The Torus CLI then uses these values in determining which entities to interact with in subsequent commands.

If a user is operating the CLI in a directory which has not been linked, they will be required to supply any and all command options (such as org, and project) with each command.

In a linked directory the project and organization are inherited from the `.torus.json` file and do not need to be supplied; however, if they are present the command options will take precedence.

The context features provided as a result of `torus link` can be disabled using [preferences](./system.md#prefs).

## unlink
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus unlink` destroys the current working directory’s `.torus.json` file, thus ceasing any inferred context for subsequent commands performed in that directory.

## status
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus status` displays the current working directory’s context. The user is given each segment of the path which has been inferred (or supplied) as well as the completed path itself.
