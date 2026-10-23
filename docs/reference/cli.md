# CLI

The `lab` CLI automates common tasks.

```sh
lab --help
```

| Command | Purpose |
| --- | --- |
| `lab infra state ensure --bucket NAME` | Ensure the Cloudflare R2 state bucket exists |
| `lab infra state credentials --account-id ID --bucket NAME` | Supply temporary S3 credentials to OpenTofu's credential process |
| `lab infra secrets sync --environment ENV --kubeconfig PATH` | Copy explicit OpenTofu secret outputs to OpenBao |
| `lab infra forgejo sync --kubeconfig PATH --domain DOMAIN` | Configure Forgejo SSO and Woodpecker integration |
| `lab infra forgejo push --kubeconfig PATH --domain DOMAIN` | Push committed `HEAD` to Forgejo's `ops/homelab` repository |
| `lab cluster enroll --environment ENV` | Join installed K3s servers and verify Ready membership |
| `lab cluster kubeconfig --environment ENV --output PATH` | Export kubeconfig after verifying access through the API VIP |
| `lab sso group ensure NAME` | Ensure a Kanidm group exists |
| `lab sso client ensure NAME --redirect-uri URI` | Ensure an SSO client exists |
| `lab sso secrets sync` | Copy SSO client credentials to OpenBao |
| `lab sso openbao configure --redirect-uri URI` | Configure OpenBao OIDC login |
| `lab user create NAME --display-name NAME --email EMAIL` | Create a user and print a one-use setup link |
| `lab completion SHELL` | Generate shell completion |

The `sso` and `user` commands also require `--kubeconfig PATH --url URL`.
