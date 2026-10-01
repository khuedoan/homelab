# Toolbox

`toolbox` is a Go CLI packaged in the Nix development shell. Commands use a resource followed by an action.

```sh
nix develop
toolbox --help
toolbox users --help
toolbox backup restore --help
```

## Commands

| Command | Purpose |
| --- | --- |
| `toolbox status` | List Argo CD applications and cluster ingresses. |
| `toolbox apps create NAME` | Create a Helm chart skeleton under `apps/NAME`. Existing directories are not overwritten. |
| `toolbox argocd admin-password` | Print the initial Argo CD admin password. |
| `toolbox dns list` | List ingress IP addresses and DNS names. |
| `toolbox users create USERNAME FULL_NAME EMAIL` | Create a Kanidm account, add it to `editor`, and issue a credential reset token. |
| `toolbox users reset-password ACCOUNT` | Recover a Kanidm account password through the server. |
| `toolbox wireguard config PEER` | Print the peer QR code and WireGuard configuration. |
| `toolbox backup setup --namespace NAMESPACE --pvc PVC` | Apply an ExternalSecret and scheduled VolSync ReplicationSource. |
| `toolbox backup restore --namespace NAMESPACE --pvc PVC` | Apply an ExternalSecret and one-shot VolSync ReplicationDestination. This does not stop workloads or wait for the restore to finish. |
| `toolbox integrations setup` | Configure the legacy Gitea, Dex, Woodpecker, and Kanidm integrations. This recovers admin passwords and is intended for initial setup. |
| `toolbox helm diff --repository URL --source REF --target REF --subpath PATH` | Compare rendered Helm charts between Git revisions. |
| `toolbox screenshots capture --output DIRECTORY` | Capture the five configured application pages using installed Firefox at 1920 × 1080. `--profile` selects a Firefox profile for authenticated pages. |
| `toolbox completion SHELL` | Generate shell completion for Bash, Zsh, Fish, or PowerShell. |

Account creation and integration setup require an installed `kanidm` CLI compatible with the deployed Kanidm server. Screenshot capture requires Firefox.

## Kubeconfig

Kubeconfig selection has the following precedence:

1. `--kubeconfig PATH`
2. `KUBECONFIG`
3. `infra/kubeconfig.yaml`, relative to the current directory

## Development

Without the packaged binary, commands run through Go:

```sh
go -C toolbox run . --help
make -C toolbox test
```
