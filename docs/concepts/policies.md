# Policies
Torus CLI has resource-based policies for access control. In order to facilitate the basics of secret sharing we include three default policies.

Policies are attached to teams and machine roles to construct what the members are entitled to access.

## Default Policies
The three default teams each have their own default policy.

### Member
Every user added to an organization is automatically made a member of the "member" team and will be given access according to the policy below.

```
allow -r--l /${org}/*/[dev-${username}|dev-@]
allow crudl /${org}/*/[dev-${username}|dev-@]/*
allow crudl /${org}/*/[dev-${username}|dev-@]/*/*
allow crudl /${org}/*/[dev-${username}|dev-@]/*/*/*
allow crudl /${org}/*/[dev-${username}|dev-@]/*/*/*/*
```

### Admin
Through the default admin policy users are given full access to the Torus organization except for adding users to the Owner team.

```
deny  crudl teams:owner
allow crudl teams:*
allow -ru-l /${org}
allow crudl /${org}/*
allow crudl /${org}/*/*
allow crudl /${org}/*/*/*
allow crudl /${org}/*/*/*/*
allow crudl /${org}/*/*/*/*/*
allow crudl /${org}/*/*/*/*/*/*
```

### Owner
Based on the admin policy, owners are given full access to their Torus organization; however, by default they are the only ones permitted to add users to the "owner" team.

```
allow crudl teams:*
allow -rudl /${org}
allow crudl /${org}/*
allow crudl /${org}/*/*
allow crudl /${org}/*/*/*
allow crudl /${org}/*/*/*/*
allow crudl /${org}/*/*/*/*/*
allow crudl /${org}/*/*/*/*/*/*
```

## Machines
Unlike users, machines are not automatically given access through default policies. A machine starts out with zero permissions and must have policies attached to its machine role to open its access.
