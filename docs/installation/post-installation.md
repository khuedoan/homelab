# Post-installation

## Create users

Follow the [user onboarding guide](../getting-started/user-onboarding.md) to create accounts and assign access.
Use SSO to sign in to applications.

## Verify backups

Follow the [backup and restore guide](../how-to-guides/backup-and-restore.md) to check backups and restore data.

## Run tests

Run offline checks:

```sh
make test
```

Run live tests against staging:

```sh
make -C tests e2e env=staging
```

See the [testing reference](../reference/testing.md) for individual checks.
