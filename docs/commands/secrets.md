# Secrets
The Torus CLI is used to set and access secrets and config. Each value is encrypted client-side according to our [cryptography](../internals/crypto.md) architecture and stored in the Torus Registry.

A secret is a single piece of configuration which should be encrypted.

Torus exposes your decrypted secrets to your process through environment variables. This means that anything you can store in an environment variable, you can set in Torus. Support for file storage (such as certificates) is planned, but not currently supported.  

### Command Options

  Option | Description
  ---- | ----
  --org, ORG, -o ORG | Executing the command for the specified org.
  --project PROJECT, -p PROJECT | Executing the command for the specified project.
  --environment ENV, -e ENV | Executing the command for the specified environment. Can be specified multiple times.
  --service SERVICE, -s SERVICE | Execute the command for the specified environment. (default: default)
  --user USER, -u USER | Execute the command for the specified user identity. (default: *)
  --machine MACHINE, -m MACHINE | Execute the command for the specified machine identity. (default: *)
  --instance INSTANCE, -i INSTANCE | Execute the command for the specified instance identity. (default: *)

## set
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus set <name|path> <value>` sets a value for the specified name (or [path](../concepts/path.md)).

This is how all secrets are stored in Torus.

## unset
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus unset <name|path>` unsets the value for the specified name (or [path](../concepts/path.md)).

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
torus run -o example -- node ./bin/www --app api
```

## ls
###### Added [v0.13.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus ls [path]` allows a user to explore the objects and values set within their organization.

Each level within the organization can be inspected by changing the segments supplied in the path. Wildcards cannot be used for the organization segment of the [path](../concepts/path.md).

Segments | Result | Example
---- | ---- | ----
0 | All organizations| `torus ls /`
1 | All projects within an org | `torus ls /org`
2 | All environments (for the defined project) within an org | `torus ls /org/*`
3 | All services (for the defined project) within an org | `torus ls /org/*/*`
4 | All secrets within the defined path (org and path required) | `torus ls /org/project/*/*`
