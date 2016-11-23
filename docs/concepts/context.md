# Context

Each Torus command requires input to tell it which part of your organization structure you're interacting with.

Typically this information is passed through command options (also known as flags).

For example, when we want to set a secret we could use:

```
torus set -o manifold -p guides -s www -e staging port 80 -u * -i *
```

Typing each of these command options for every interaction becomes a pain, so the Torus CLI uses context to help infer which resource you're interacting with.

Context is a cascade of values which ultimately determine which resource you're acting upon. Starting with system defaults, then your linked context followed by any command options.

### System default

Each command has its own system default values, where applicable. A definition of the command options can be found either through `torus help` or in the command's documentation.

For example, with `torus set`:

Command Option | Default Value
---- | ----
Identity | Currently authenticated username or machine name
Instance | `*`

### Preference defaults

Your system preferences, found in your `.torusrc` file, are used to identify system-wide defaults. They are managed with the [prefs](./system.md#prefs) command.

Supported system default preferences:
- org
- project
- environment
- service

### Linked directory

Your project's `.torus.json` file, which can be created through [torus link](../project-structure.md#link), is then used to source Organization and Project (if present).

Any time Torus is executed within this directory or one of its child directories these values will be sourced.

### Command options

Command options take presedence during execution of a Torus command, overwriting any values sourced from context.
