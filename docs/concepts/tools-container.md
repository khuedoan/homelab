# Tools container

The tools container makes it easy to get all of the dependencies needed to interact with the homelab.

## How to open it

You can use the default Docker wrapper, or use Nix if you have Nix installed:

=== "Docker"

    ```sh
    make tools
    ```

=== "Nix"

    ```sh
    nix-shell
    ```

    !!! tip

        If you have [`direnv`](https://direnv.net) installed, you can run `direnv allow` once and it will automatically enter the Nix shell.

It will open a shell like this:

```
[nix-shell:/home/khuedoan/Documents/homelab]# echo hello
hello
```

## How it works

- All dependencies are defined in `./shell.nix`
- When you run `make tools`, it will run a thin Docker wrapper with the `nixos/nix` image (because not everyone has Nix installed) and mount some required volumes
- `nix-shell` will start an interactive shell based on the Nix expression in `./shell.nix` and install everything from there

## Known issues

- If your Docker engine is not running in rootless mode, all files created by the tools container will be owned by `root`
