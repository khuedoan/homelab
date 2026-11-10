# Post-installation

## Access the cluster

Export the environment's kubeconfig, e.g., production:

```sh
export KUBECONFIG="$PWD/infra/production/kubeconfig.yaml"
```

## Create users

Follow the [user onboarding guide](user-onboarding.md) to create accounts and assign access.
Use SSO to sign in to applications.

## Verify backups

Follow the [backup and restore guide](../how-to-guides/backup-and-restore.md) to check backups and restore data.

## Run tests

Run offline checks:

```sh
just test
```

Run live tests against staging:

```sh
just test e2e staging
```

See the [testing reference](../reference/testing.md) for individual checks.
