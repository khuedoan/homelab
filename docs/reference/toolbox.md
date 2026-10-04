# Toolbox

`toolbox` is a Go CLI packaged in the Nix development shell. It bootstraps
infrastructure state storage, enrolls k3s servers, and exports kubeconfig.
Commands use a resource followed by an action.

```sh
nix develop
toolbox --help
toolbox cluster enroll --help
toolbox infra state ensure --help
```

## Commands

| Command | Purpose |
| --- | --- |
| `toolbox infra state ensure --account-id ID --bucket NAME` | Ensure the Cloudflare R2 state bucket exists without changing existing buckets. |
| `toolbox cluster enroll --environment ENV` | Join installed servers to the configured k3s cluster and verify Ready membership. |
| `toolbox cluster kubeconfig --environment ENV --output PATH` | Export kubeconfig after verifying authenticated access through the control-plane VIP. |
| `toolbox completion SHELL` | Generate shell completion for Bash, Zsh, Fish, or PowerShell. |

`ENV` is `staging` or `production`. Cluster commands run from the repository root
and read `infra/ENV/metal/hosts.json` and `infra/ENV/cluster/config.json`.
SSH requires an authorized key and trusted host keys in `known_hosts`.

## State storage

State storage authenticates through `CLOUDFLARE_TFSTATE_API_TOKEN`.
`--account-id` defaults to `CLOUDFLARE_ACCOUNT_ID`. `--bucket` is required.
The command does not require an existing Kubernetes cluster or Terraform state.

## Cluster access

Enrollment uses verified root SSH connections. It does not install nodes,
reset datastores, or overwrite conflicting credentials. `--timeout` defaults
to `10m` for enrollment and `2m` for kubeconfig export.

Kubeconfig export defaults to `infra/kubeconfig.yaml`. `--output PATH` selects
another destination. Files are published atomically with mode `0600`.
`--output -` sends credentials to stdout.
Cluster commands use the explicit environment inventory, not `KUBECONFIG`.

## Development

Without the packaged binary, commands run through Go:

```sh
go -C toolbox run . --help
make -C toolbox test
```
