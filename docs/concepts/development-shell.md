# Development shell

The development shell makes it easy to get all of the dependencies needed to interact with the homelab.

## Prerequisites

!!! info

    NixOS users can skip this step.

Install Nix using one of the following methods:

- [Official Nix installer](https://nixos.org/download)
- [Determinate Nix Installer](https://docs.determinate.systems/getting-started/#installer)

If you're using the official installer, add the following to your
`~/.config/nix/nix.conf` to enable [Flakes](https://nixos.wiki/wiki/Flakes):

```conf
experimental-features = nix-command flakes
```

## How to open it

Run the following command:

```sh
nix develop
```

It will open a shell with all the dependencies defined in `./flake.nix`:

```
[khuedoan@ryzentower:~/Documents/homelab]$ which kubectl
/nix/store/0558zzzqynzw7rx9dp2i7jymvznd1cqx-kubectl-1.30.1/bin/kubectl
```

!!! tip

    If you have [`direnv`](https://direnv.net) installed, you can run `direnv
    allow` once and it will automatically enter the Nix shell every time you
    `cd` into the project.
