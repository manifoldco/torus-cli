# Secrets
The Torus CLI is used to set and access secrets and config. Each value is encrypted client-side according to our [cryptography](../internals/crypto.md) architecture and stored in the Torus Registry.

A secret is a single piece of configuration which should be encrypted.

Torus exposes your decrypted secrets to your process through environment variables. This means that anything you can store in an environment variable, you can set in Torus. Support for file storage (such as certificates) is planned, but not currently supported.

### Command Options

  Option | Description
  ---- | ----
  --org, ORG, -o ORG | Executing the command for the specified org.
  --project PROJECT, -p PROJECT | Executing the command for the specified project.
  --environment ENV, -e ENV | Executing the command for the specified environment.
  --service SERVICE, -s SERVICE | Execute the command for the specified environment. (default: default)
  --user USER, -u USER | Execute the command for the specified user identity. (default: *)
  --machine MACHINE, -m MACHINE | Execute the command for the specified machine identity. (default: *)
  --instance INSTANCE, -i INSTANCE | Execute the command for the specified instance identity. (default: *)

The `--environment`, `--service`, `--user`, `--machine`, and `--instance` flags can be specified many times when setting, unsetting, or importing secrets.

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

Credentials retrieved
Keypairs retrieved
Encrypting key retrieved
Credential encrypted
Completed Operation

Credential PORT has been set at /myorg/api/production/auth/*/*/PORT
```

**Using flags inside a linked directory**

You can link a directory to an org and project using the [`torus link`](./project-structure.md) command. Once a directory is linked, you only need to specify the flags you want to override.

```bash
# Setting the port for the auth service in staging
$ torus set -e staging -s auth PORT 8080

Credentials retrieved
Keypairs retrieved
Encrypting key retrieved
Credential encrypted
Completed Operation

Credential PORT has been set at /myorg/api/staging/auth/*/*/PORT
```

**Using multiple flags to share a secret between environments**

You can specify the environment, service, user, machine, and instance flags many times to share a secret.

```bash
# Setting the same port for production and staging for the auth service
$ torus set -e production -e staging -s auth PORT 8080

Credentials retrieved
Keypairs retrieved
Encrypting key retrieved
Credential encrypted
Completed Operation

Credential PORT has been set at /myorg/api/[production|staging]/auth/*/*/PORT
```

**Setting a secret with a `*` value**

You can set a secret to be shared across all environments, or services by specifying a value of `*`. For example, if you set an environment to be `*` then any environment (production, staging, dev, etc) will have access to the value.

```bash
# Setting the same port across all environments for the auth service
$ torus set -e * -s auth PORT 8080

Credentials retrieved
Keypairs retrieved
Encrypting key retrieved
Credential encrypted
Completed Operation

Credential PORT has been set at /myorg/api/[production|staging]/auth/*/*/PORT
```

## unset
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus unset <name|path>` unsets the value for the specified name (or [path](../concepts/path.md)).

## import
###### Added [v0.25.0](https://github.com/manifoldco/torus-cli/blob/v0.25.0/CHANGELOG.md)

`torus import <file>` or using stdin redirection (e.g. `torus import -e production <prod.env`) imports the contents of an `.env` file to the specified path.

**Example**

```bash
$ cat prod.env
PORT=4000
DOMAIN=mydomain.co
MYSQL_URL=mysql://user:pass@host.com:4321/mydb

$ torus import -e production test.env

Credentials retrieved
Keypairs retrieved
Encrypting key retrieved
Credential encrypted
Credential encrypted
Completed Operation

Credential port has been set at /myorg/myproject/production/default/*/*/PORT
Credential mysql_url has been set at /myorg/myproject/production/default/*/*/MYSQL_URL
```

## export
###### Added [v0.28.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus export [file-path]` or using stdout redirection (e.g. `torus export -e production > config.env`) exports the secrets for a specific project, environment, and service to a file or stdout. The output format can be specified using the `--format, -f` flag supporting `env`, `bash`, `powershell`, `cmd` (windows command prompt), `fish`, and `json`.

**Examples**

All examples assumed the command is ran inside a [linked](./project-structure.md#link) directory.

```bash
# Exporting secrets from a specific environment to stdin
$ torus export -e prod
MYSQLDB_URL="mysql://user:password@host.com:3306/db"

# Exporting secrets from a specific environment to stdin for powershell
$ torus export -e prod
$Env:MYSQLDB_URL = "mysql://user:password@host.com:3306/db"

# Exporting secrets from a specific environment piped to a file
$ torus export -e prod > config.env

# Exporting secrets from a specific environment to a file
$ torus export -e prod config.env

# Exporting secrets from a specific environment into a bash shell
$ eval "$(torus export -e prod -f bash)"
$ echo $MYSQLDB_URL
mysql://user:password@host.com:3306/db

# Exporting secrets from a specific environment and service into a powershell
$ torus export -e prod -s api -f powershell
```

## view
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus view` displays secrets in the current [context](./project-structure.md#link).

By default items are displayed in environment variable format.

### Command Options

  Option | Description
  ---- | ----
  --verbose, -v | List the sources of the secrets (shortcut for --format verbose)
  --format FORMAT, -f FORMAT | Format used to display data (json, env, verbose) (default: env)

## run
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus run [--] <command> [<arguments>...]` executes the supplied command and injects your secrets into its environment.

By prefixing your process execution with `torus run` we are able to fetch, decrypt and inject your secrets into the process environment based on the [context](./project-structure.md#link) of the Torus client.

To ensure that your command’s arguments and options are passed correctly you may need to separate your `torus run` definition from your command definition with `--`, for example:

```
# Inject secrets into the process from the production environment and www service.
$ torus run -e production -s www -- node ./bin/www --app api
```

## list
###### Added [v0.28.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus list [name]...` allows a user to explore the objects and values set within their organization. Users can find the location of a particular secret by passing its name as an argument, or view all secrets within different branches of their organization.

Each level within the organization can be inspected by changing the options supplied to the command.

### Command Options

  Option | Description
  ---- | ----
  --org, -o | Required flag to specify org.
  --project, -p | Required flag to specify project.
  --env, -e | Specify environment filter(s) for displayed secrets. To specify multiple environments, pass multiple flags (ie. -e env1 -e env2). This flag is optional.
  --service, -s | Specify service filter(s) for displayed secrets. To specify multiple services, pass multiple flags (ie. -s ser1 -s ser2). This flag is optional.
  --verbose, -v | Show which type of path is being displayed, shortcut for --format=verbose
  --format FORMAT, -f FORMAT | Format used to display data (simple, verbose) (default: simple)

### Examples

List all secrets in a project:
```
$ torus list
/org/project/
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

Find the location of a particular secret:
```
$ torus list secret1
/org/project/
    env1/
        ser1/
            secret1      
        ser2/          
    env2/
        ser1/
            secret1
```

List all secrets in a specific environment:
```
$ torus list -e env2
/org/project/
    env2/
        ser1/
            secret1
```

List all secrets in a specific service:
```
$ torus list -s ser1
/org/project/
    env1/
        ser1/
            secret1               
    env2/
        ser1/
            secret1
```
