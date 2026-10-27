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

Run the following command:

```sh
lab user create johndoe \
  --kubeconfig infra/production/kubeconfig.yaml \
  --url https://auth.example.com \
  --display-name "John Doe" \
  --email johndoe@example.com
```

Share the one-time link privately for the user to setup passkeys or password + TOTP.

Add `--group` to add more access:

| Group | Access |
| --- | --- |
| `users` (default) | Use applications and view dashboards |
| `developers` | Edit dashboards and push to the GitOps repository |
| `admins` | Admin access to applications |
