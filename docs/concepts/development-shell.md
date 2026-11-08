# Development shell

The development shell makes it easy to get all of the tools and credentials.

## Prerequisites

### NixOS

Add the following to your NixOS configuration:

```nix
{
  nix.settings.experimental-features = [
    "nix-command"
    "flakes"
  ];
  programs.direnv = {
    enable = true;
    silent = true;
  };
}
```

### Other distros

1. Install Nix using the [official installer](https://nixos.org/download).
2. Enable [Flakes](https://nixos.wiki/wiki/Flakes) in `~/.config/nix/nix.conf`:

   ```conf
   experimental-features = nix-command flakes
   ```

3. Install [direnv](https://direnv.net) and [enable its shell hook](https://direnv.net/man/direnv.1.html#setup).

## Load the environment

From the repo root:

```sh
direnv allow .
direnv allow infra/production
direnv allow infra/staging
```

From now on, `cd` to the project automatically loads the required tools and credentials.
