# Development environment

## Caveats compare to production environment

The development cluster doesn't have the following features:

- There is no valid domain name, hence no SSL certificates (some services require valid SSL certificates)
- Only accessible on the host machine
- No backup
<!-- TODO more caveats here -->

Please keep in mind that the development cluster may be unstable and things may break (it's for development after all).

## Prerequisites

Host machine:

- Recommended hardware specifications:
    - CPU: 4 cores
    - RAM: 16 GiB
- OS: Linux (Windows and macOS are untested, please let me know if it works)
- Available ports: `80` and `443`

Install the following packages:

- `docker`
- `make`

Clone the repository and checkout the development branch:

```sh
git clone https://github.com/khuedoan/homelab
git checkout dev
```

## Build

Open the tools container, which includes all the tools needed:

```sh
make tools
```

Build a development cluster and bootstrap it:

```
make
```

!!! note

    It will take about 15 to 30 minutes to build depending on your internet connection

## Explore

The homepage should be available at <https://home.127-0-0-1.nip.io> (ignore the security warning because we don't have valid certificates).

## Clean up

Delete the cluster:

```sh
k3d cluster delete homelab-dev
```
