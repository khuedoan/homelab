# CLI

Use `just` to run common tasks.

```sh
just --list
```

| Command | Purpose |
| --- | --- |
| `just deploy ENV` | Deploy the lab and configure services |
| `just user create ENV NAME "FULL NAME" EMAIL [GROUPS]` | Create a user and print an enrollment link; optional groups are comma-separated |
| `just user reset ENV NAME` | Print a credential reset link for an existing user |
| `just infra tfstate ENV` | Create the state bucket and credentials |
| `just infra apply ENV` | Apply infrastructure |
| `just infra secrets ENV` | Copy OpenTofu secret outputs to OpenBao |
| `just infra sso ENV` | Configure SSO and sync OAuth credentials |
| `just infra forgejo ENV` | Configure Forgejo and push committed `HEAD` |
| `just test` | Run local checks |
| `just lint` | Check Go and infrastructure code |
| `just fmt` | Format code |
| `just docs` | Preview the documentation |

`just deploy` loads the environment's credentials. For individual apply or setup commands, load them explicitly:

```sh
just --dotenv-path infra/production/.env infra secrets production
```

See the [testing reference](testing.md) for integration tests and benchmarks.
