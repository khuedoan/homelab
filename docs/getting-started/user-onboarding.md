# User onboarding

## For user

Ask an admin to create your account, provide the following information:

- [ ] Full name (John Doe)
- [ ] Select a username (`johndoe`)
- [ ] Email address (`johndoe@example.com`)

## For admin

Create the account and share the one-time link privately for the user to set up passkeys:

```sh
just user create production johndoe "John Doe" johndoe@example.com
```

To include additional access groups when creating the account:

```sh
just user create production johndoe "John Doe" johndoe@example.com admins,developers
```

To generate a credential reset link:

```sh
just user reset production johndoe
```

To update access groups:

```sh
just user groups production johndoe admins,developers
```

| Group | Access |
| --- | --- |
| `users` | Use applications and view dashboards (added by default) |
| `developers` | Edit dashboards and push to the GitOps repository |
| `admins` | Admin access to applications |
