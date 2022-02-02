# Provisioning flow

![Provisioning flow](../images/provisioning_flow.png)

Everything is automated, after you edit the configuration files, you just need to run a single `make` command and it will:

- Build the `./metal` layer:
  - Create an ephemeral, stateless PXE server
  - Install Linux on all servers in parallel
  - Build a Kubernetes cluster (based on k3s)
- Build the `./bootstrap` layer:
  - Install ArgoCD
  - Install ApplicationSet to manage other layers (and also manage itself)

From now on, ArgoCD will do the rest:

- Build the `./system` layer (storage, networking, monitoring, etc)
- Build the `./platform` layer (Gitea, Vault, SSO, etc)
- Build the `./apps` layer: (Syncthing, Jellyfin, etc)
