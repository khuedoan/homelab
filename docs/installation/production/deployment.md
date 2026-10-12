# Deployment

Open the development shell if you haven't already:

```sh
nix develop
```

Before applying, configure the [R2 state credentials](../../how-to-guides/infrastructure-state.md).
If upgrading from the HTTP backend, complete the coordinated state transition
in that guide before any plan or apply.

Build the lab:

```sh
make infra env=production
```

Yes it's that simple!

!!! example

    <script id="asciicast-xkBRkwC6e9RAzVuMDXH3nGHp7" src="https://asciinema.org/a/xkBRkwC6e9RAzVuMDXH3nGHp7.js" async></script>

It will take a while to download everything,
you can read the [architecture document](../../reference/architecture/overview.md) while waiting for the deployment to complete.
