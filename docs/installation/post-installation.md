# Post-installation

## Backup secrets

Save the following files to a safe location like a password manager (if you're using the sandbox, you can skip this step):

- `~/.ssh/id_ed25519`
- `~/.ssh/id_ed25519.pub`
- `./metal/kubeconfig.yaml`
- `~/.terraform.d/credentials.tfrc.json`
- `./external/terraform.tfvars`

## Admin credentials

- ArgoCD:
    - Username: `admin`
    - Password: run `./scripts/argocd-admin-password`
- Vault:
    - Root token: run `./scripts/vault-root-token`
- Grafana:
    - Username: `admin`
    - Password: `prom-operator` (TODO: use random password)
- Gitea:
    - Username: `gitea_admin`
    - Password: get from Vault
