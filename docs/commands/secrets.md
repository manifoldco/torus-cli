# Secrets
The Torus CLI is used to set and access secrets and config. Each value is encrypted client-side according to our [cryptography](../internals/crypto.md) architecture and stored in the Torus Registry.

A secret is a single piece of configuration which should be encrypted.

Torus exposes your decrypted secrets to your process through environment variables. This means that anything you can store in an environment variable, you can set in Torus. Support for file storage (such as certificates) is planned, but not currently supported.

### Command Options

All secrets commands accept the following flags:

  Option | Environmant Variable | Description
  ---- | ---- | ----
  --org, ORG, -o ORG | TORUS_ORG | Execute the command for the specified org.
  --project PROJECT, -p PROJECT | TORUS_PROJECT | Execute the command for the specified project.
  --environment ENV, -e ENV | TORUS_ENVIRONMENT | Execute the command for the specified environment.
  --service SERVICE, -s SERVICE | TORUS_SERVICE | Execute the command for the specified environment. (default: default)

The `--environment`, and `--service` flags can be specified many times when setting, unsetting, or importing secrets.

By default, the environment will be set to the user's development environment for the specified (or linked) project (e.g. if the user's username is `joe` the environment will be `dev-joe`). An environment must always be supplied for a machine.

## set
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus set <name|path> <value>` or `torus set <name|path>=<value>` sets a value
for the specified name (or [path](../concepts/path.md)).

This is how all secrets are stored in Torus.

#### Examples

**Using flags**

```bash
# Setting the port for the production auth service inside myorg's api project.
$ torus set -o myorg -p api -e production -s auth PORT 3000

Credential port has been set at /myorg/api/production/auth/port
```

**Using flags inside a linked directory**

You can link a directory to an org and project using the [`torus link`](./project-structure.md) command. Once a directory is linked, you only need to specify the flags you want to override.

```bash
# Setting the port for the auth service in staging
$ torus set -e staging -s auth PORT 8080

Credential port has been set at /myorg/api/staging/auth/port
```

**Using multiple flags to share a secret between environments**

You can specify the environment, service, user, machine, and instance flags many times to share a secret.

```bash
# Setting the same port for production and staging for the auth service
$ torus set -e production -e staging -s auth PORT 8080

Credential port has been set at /myorg/api/[production|staging]/auth/port
```

**Setting a secret with a `*` value**

You can set a secret to be shared across all environments, or services by specifying a value of `*`. For example, if you set an environment to be `*` then any environment (production, staging, dev, etc) will have access to the value.

```bash
# Setting the same port across all environments for the auth service
$ torus set -e * -s auth port 8080

Credential port has been set at /myorg/api/*/auth/port
```

## unset
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus unset <name|path>` unsets the value for the specified name (or [path](../concepts/path.md)).

The explicit values used to set the secret must be provided to unset it.

#### Examples

**Unsetting a secret set using default values**

```bash
$ torus unset port
You are about to unset "/myorg/myproject/dev-matt/default/port". This cannot be undone.
✔ Do you wish to continue? [y/N] y

Credential port has been unset at /myorg/myproject/dev-matt/default/port.
```

**Unsetting a secret by providing its full path**

It's easiest to unset a secret by providing its full path. You can get a secrets full path by passing the `-v, --verbose` flag to [`torus view`](#view) or [`torus list`](#list).

```bash
$ torus unset /myorg/myproject/dev-matt/default/port
You are about to unset "/myorg/myproject/dev-matt/default/port". This cannot be undone.
✔ Do you wish to continue? [y/N] y

Credential port has been unset at /myorg/myproject/dev-matt/default/port.
```

## import
###### Added [v0.25.0](https://github.com/manifoldco/torus-cli/blob/v0.25.0/CHANGELOG.md)

`torus import <file>` or using stdin redirection (e.g. `torus import -e production <prod.env`) imports the contents of an `.env` file to the specified path.

**Example**

```bash
$ cat prod.env
DOMAIN="mydomain.co"
PORT="4000"

$ torus import -e production prod.env

Credential domain has been set at /myorg/myproject/production/default/domain
Credential port has been set at /myorg/myproject/production/default/port
```

## export
###### Added [v0.28.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus export [file-path]` or using stdout redirection (e.g. `torus export -e production > config.env`) exports the secrets for a specific project, environment, and service to a file or stdout. The output format can be specified using the `--format, -f` flag supporting `env`, `bash`, `powershell`, `cmd` (windows command prompt), `fish`, `json`, `tfvars` (for exporting to [terraform](https://terraform.io) variable files).

#### Examples

All examples assumed the command is ran inside a [linked](./project-structure.md#link) directory.

**Exporting secrets from a specific environment to stdin**

```bash
$ torus export -e prod
MYSQLDB_URL="mysql://user:password@host.com:3306/db"
```

**Exporting secrets from a specific environment to stdout for powershell**

```bash
$ torus export -e prod
$Env:MYSQLDB_URL = "mysql://user:password@host.com:3306/db"
```

**Exporting secrets from a specific environment piped to a file**

```bash
$ torus export -e prod > config.env
$ cat config.env
MYSQLDB_URL="mysql://user:password@host.com:3306/db"
```

**Exporting secrets from a specific environment to a file**

```bash
$ torus export -e prod config.env
$ cat config.env
MYSQLDB_URL="mysql://user:password@host.com:3306/db"
```

**Exporting secrets from a specific environment into a bash shell**

```bash
$ eval "$(torus export -e prod -f bash)"
$ echo $MYSQLDB_URL
mysql://user:password@host.com:3306/db
```

**Exporting secrets to a tfvars file**

```bash
$ torus export -e prod -s api -f tfvars secrets.tfvars
$ terraform plan -var-file=secrets.tfvars
```

## view
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus view` displays secrets in the current [context](./project-structure.md#link).

Items are displayed in environment variable format.

### Command Options

  Option | Description
  ---- | ----
  --verbose, -v | List the sources of the secrets (shortcut for --format verbose)

## run
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus run [--] <command> [<arguments>...]` executes the supplied command and injects your secrets into its environment.

By prefixing your process execution with `torus run` we are able to fetch, decrypt and inject your secrets into the process environment based on the [context](./project-structure.md#link) of the Torus client.

To ensure that your command’s arguments and options are passed correctly you may need to separate your `torus run` definition from your command definition with `--`.

Torus will inject the current org, project, environment, and service into the processes through the `TORUS_ORG`, `TORUS_PROJECT`, `TORUS_ENVIRONMENT`, and `TORUS_SERVICE` environment variables.

#### Examples

**Injecting secrets into a process using flags**

```bash
$ torus run -e production -s www -- node ./bin/www --log_level=error
```

**Injecting secrets into a process using environment variables**

```bash
$ TORUS_ENVIRONMENT=production TORUS_SERVICE=www torus run -- node ./bin/www --log_level=error
```

**Using the project structure environment variables**

```bash
$ cat example.sh
#!/usr/bin/bash

echo "org: $TORUS_ORG"
echo "project: $TORUS_PROJECT"
echo "env: $TORUS_ENV"
echo "service: $TORUS_SERVICE"
$ torus run -- bash example.sh
org: test
project: api
env: dev-test
service: default
```

## list
###### Added [v0.28.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus list [name]...` allows a user to explore the secrets stored within a project. Users can search for specific secrets by supplying a name as an argument or filter down to specific environments and services using the `--environment, -e` and `--service, -s` flags.

### Command Options

The list command accepts the following flags in addition to flags supported by all secret commands.

  Option | Environment Variable | Description
  ---- | ---- | ----
  --team, -t | TORUS_TEAM | Only show secrets that the specified team(s) can access. To specify multiple teams, pass multiple flags (eg. `torus list -t team1 -t team2`). This flag is optional.
  --verbose, -v | TORUS_VERBOSE | Show which type of path is being displayed, shortcut for

### Examples

**List all secrets inside a project from a linked directory**

```
$ torus list

env1/
    ser1/
        secret1
        secret2
    ser2/
        secret3
env2/
    ser1/
        secret1
```

**List all secrets inside a project using flags**

```bash
$ torus list -o myorg -p api

env1/
    ser1/
        secret1
        secret2
    ser2/
        secret3
env2/
    ser1/
        secret1
```

**List all instances of a secret from a linked directory**

```bash
$ torus list secret1

env1/
    ser1/
        secret1
    ser2/
env2/
    ser1/
        secret1
```

**List all secrets in a specific environment from a linked directory**

```bash
$ torus list -e env2

env2/
    ser1/
        secret1
```

**List all secrets in a specific service from a linked directory**

```bash
$ torus list -s ser1
env1/
    ser1/
        secret1
env2/
    ser1/
        secret1
```

**List all secrets in a specific environment for a specific service in verbose mode from a linked directory**

```bash
$ torus list -e env1 -s ser1 -v

env1/
    ser1/
        secret1   (/org/project/env1/ser1/secret1)
```

**Use `--team` flag to limit the secrets displayed by list.**

In this example, the team "team1" only has access to the secrets in "staging"

```bash
$ torus list

prod/
    default/
        secret1
	secret2
	secret3
staging/
    default/
        secret4

(4) secrets found
$ torus list -t team1

staging/
    default/
        secret4

(1) secrets found
```
