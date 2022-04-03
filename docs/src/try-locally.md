# Try locally

## Caveats compare to production environment

The development cluster doesn't have the following features:

- There is no valid domain name, hence no SSL certificates (some services require valid SSL certificates)
- Only accessible on the host machine
- No backup
<!-- TODO more caveats here -->

Please keep in mind that the development cluster may be unstable and things may break (it's for development after all).

## Prerequisites

Host machine:

- OS: Linux (Windows and macOS are untested)
- Recommended hardware specifications:
  - CPU: 4 cores
  - RAM: 16 GiB

Install the following packages:

- `docker`
- `make`

Clone the repository (follow the [configuration guide](./deployment/configuration.md) if you want to customize it):

```sh
git clone https://github.com/khuedoan/homelab
git checkout dev
```

## Build

Open the tools container:

```sh
make tools
```

Build a development cluster and bootstrap it:

```
make
```

Look for the dashboard URL in the command output.

## Clean up

Delete the cluster:

```sh
k3d cluster delete homelab-dev
```
