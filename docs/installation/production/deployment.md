# Deployment

Open the development shell if you haven't already:

```sh
nix develop
```

Build the lab:

```sh
make infra env=production
```

Yes it's that simple!

It will take a while to download everything,
you can read the [architecture document](../../reference/architecture/overview.md) while waiting for the deployment to complete.
