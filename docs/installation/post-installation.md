# Post-installation

## Backup secrets

Save the following files to a safe location like a password manager (if you're using the sandbox, you can skip this step):

- `~/.ssh/id_ed25519`
- `~/.ssh/id_ed25519.pub`
- `./infra/kubeconfig.yaml`
- `~/.terraform.d/credentials.tfrc.json`
- `./external/terraform.tfvars`

## Admin credentials

- ArgoCD:
    - Username: `admin`
    - Password: run `toolbox argocd admin-password`
- Gitea:
    - Username: `gitea_admin`
    - Password: get from `global-secrets` namespace
- Kanidm:
    - Usernames: `admin` and `idm_admin`
    - Password: run `toolbox users reset-password admin` and `toolbox users reset-password idm_admin`
- Jellyfin and other applications in the \*arr stack: see the [dedicated guide for media management](../how-to-guides/media-management.md)
- Other apps:
    - Username: `admin`
    - Password: get from `global-secrets` namespace

`toolbox` is available in `nix develop`. See the [command reference](../reference/toolbox.md) for all commands and kubeconfig options.

## Backup

Now is a good time to set up backups for your homelab.
Follow the [backup and restore guide](../how-to-guides/backup-and-restore.md) to get started.

## Run the full test suite

After the homelab has been stabilized, you can run the full test suite to ensure that everything is working properly:

```sh
make test
```

!!! info

    The "full" test suit is still in its early stages, so any contribution is greatly appreciated.
