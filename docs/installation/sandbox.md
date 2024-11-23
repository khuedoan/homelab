# Development sandbox

The sandbox is intended for trying out the homelab without any hardware or testing changes before applying them to the production environment.

## Prerequisites

Host machine:

- Recommended hardware specifications:
    - CPU: 4 cores
    - RAM: 16 GiB
- OS: Linux (Windows and macOS are untested, please let me know if it works)
- Available ports: `80` and `443`

Install the following packages:

- `docker`
- `nix` (see [development shell](../concepts/development-shell.md) for the installation guide)

Clone the repository and checkout the development branch:

```sh
git clone https://github.com/khuedoan/homelab
git checkout dev
```

## Build

Open the development shell, which includes all the tools needed:

```sh
nix develop
```

Build a development cluster and bootstrap it:

```
make
```

!!! note

    It will take about 15 to 30 minutes to build depending on your internet connection

## Explore

The homepage should be available at <https://home.127-0-0-1.nip.io> (ignore the security warning because we don't have valid certificates).

See [admin credentials](../post-installation/#admin-credentials) for default passwords.

If you want to make some changes, simply commit to the local `dev` branch and push it to Gitea in the sandbox:

```sh
git remote add sandbox https://git.127-0-0-1.nip.io/ops/homelab
git config http.https://git.127-0-0-1.nip.io.sslVerify false

git add foobar.txt
git commit -m "feat: harness the power of the sun"
git push sandbox # you can use the gitea_admin account
```

## Clean up

Delete the cluster:

```sh
k3d cluster delete homelab-dev
```

## Caveats compare to production environment

The development cluster doesn't have the following features:

- There is no valid domain name, hence no SSL certificates (some services require valid SSL certificates)
- Only accessible on the host machine
- No backup

Please keep in mind that the development cluster may be unstable and things may break (it's for development after all).
