# System

## prefs
The Torus CLI has a number of preferences that can be stored on your system. These preferences change the experience of the tool.

No preferences are required to be set in order to interact with the hosted Torus service.

There are two categories of preferences: Core and Defaults. Core contains preferences related to the internal operations of the tool. Defaults contains values that will be used when executing commands in absence of specified flags.

The following are the available preferences:

Preference | Description
---- | ----
`core.registry_uri` | The hostname (including protocol) of the Torus Registry
`core.ca_bundle_file` | Certificate bundle used to communicate with the Torus Registry.
`core.public_key_file` | Location of the ed25519 public key file.
`core.context` | Boolean determining if the `.torus.json` and defaults should be considered during command execution
`core.auto_confirm` | Boolean determining if confirmation prompts should be automatically skipped (equivalent of always using `-y` command option)
`core.vim` | Boolean determining if CLI input should use Vim bindings
`defaults.org` | Organization name to be used with context
`defaults.project` | Project name to be used with context
`defaults.environment` | Environment name to be used with context
`defaults.service` | Service name to be used with context

### set
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus prefs set <key> <value>` sets a preference value by key, where key is the dot-delimited path of the category and preference name.

If value is omitted it defaults to boolean true, except for the Vim setting, which is false by default.

### list
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus prefs list` displays all currently set preferences by category in ini format.

## daemon
Torus CLI uses a daemon to manage your active session and to perform cryptographic operations. By default your Torus daemon operates out of `~/.torus`.

### status
###### Added [v0.5.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus daemon status` displays the current state of the daemon process as well as its PID.

### start
###### Added [v0.5.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus daemon start` initiates the daemon process if it is not already running.

### stop
###### Added [v0.5.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus daemon stop` halts the daemon process if it is running.

## version
###### Added [v0.1.0](https://github.com/manifoldco/torus-cli/blob/master/CHANGELOG.md)

`torus version` displays the current version of the Torus CLI, Daemon and Registry.
