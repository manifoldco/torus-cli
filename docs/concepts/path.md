# Path
Inside the Torus data structure your secrets are leaf-nodes stored in a hierarchy (tree). Each node within this hierarchy is named, and the absolute address of a leaf node (string representation of all segments) is called the path.

The path begins with a forward slash and has seven slash-delimited segments, most of which are standard Objects inside Torus:

- Organization
- Project
- Environment
- Service
- Identity
- Instance
- Secret  

A complete path:

```
/org/project/environment/service/identity/instance/secret
```

Some segments are unique to the path:

Name | Description | Default value
---- | ---- | ----
Identity | A globally unique username or machine name | Current authenticated username or machine name
Instance | Unique ID of the accessing process | `*`
**For example:**
`/manifoldco/torus-cli/production/docs/cdn/1/token`

Commands that take advantage of the path:

- [set](../commands/secrets.md#set)
- [unset](../commands/secrets.md#unset)
- [ls](../commands/secrets.md#ls)
- [allow](../commands/access-control.md#allow)
- [deny](../commands/access-control.md#deny)

## Wildcards and Sharing
Path segments may contain wildcards (with the exception of Organization and Project). Through this secrets can be shared across multiple nodes.

Wildcards support prefixes, so that you can namespace objects:

```
/org/project/env-*/service/identity/instance/secret
```

So the value of "secret" is available to all applicable environments that match the wildcard such as: "env-1", "env-development", "env-ironment".

Paths support `**` to be expanded to fill absent segments.

#### Examples

The following paths are equivalent when supplied to a command:

```
/org/project/**/name
/org/project/*/*/*/*/name
```

## Alternations
Path segments may also contain alternations (otherwise known as "OR" statements). Alternations enable you to allow two or more values to match at a particular depth of the hierarchy.

Each alternation is wrapped in square brackets with each value (that adheres to the path segment naming scheme) delimited by the pipe character.

#### Examples

An alternation for the names "one" and "two" is:

```
[one|two]
```

The following would make "secret" available to the "development" and "staging" environments: 

```
/org/project/[development|staging]/service/identity/instance/secret
```

We can also use wildcards in a segment making "secret" available to the "development" as well as "dev-\*” environments (such as "dev-jane" and "dev-john"):

```
/org/project/[dev-*|development]/service/identity/instance/secret
```
