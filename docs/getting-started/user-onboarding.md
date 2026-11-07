# User onboarding

## For user

### Create user

Ask an admin to create your account, provide the following information:

- [ ] Full name (John Doe)
- [ ] Select a username (`johndoe`)
- [ ] Email address (`johndoe@example.com`)

### Install companion apps

For all users:

- [ ] A password manager (I personally recommend [Bitwarden](https://bitwarden.com/download))
- [ ] A [Matrix chat client](https://matrix.org/clients) (optional, you can use the web version)

For technical users:

- [ ] [Docker](https://docs.docker.com/engine/install)
- [ ] [Nix](https://nixos.org/download) and [direnv](https://direnv.net) (optional, but highly recommended)
- [ ] [Lens](https://k8slens.dev) (optional, you can use `kubectl` in the development shell)

## For admin

Create the account:

```sh
just user create production johndoe "John Doe" johndoe@example.com
```

Share the one-time link privately for the user to set up passkeys or password + TOTP.

To include additional access groups when creating the account, pass a comma-separated list:

```sh
just user create production johndoe "John Doe" johndoe@example.com admins,developers
```

To generate a credential reset link:

```sh
just user reset production johndoe
```

Set access groups:

```sh
just user groups production johndoe admins,developers
```

| Group | Access |
| --- | --- |
| `users` | Use applications and view dashboards (added by default) |
| `developers` | Edit dashboards and push to the GitOps repository |
| `admins` | Admin access to applications |
