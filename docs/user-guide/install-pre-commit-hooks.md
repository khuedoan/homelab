# Install pre-commit hooks

> TODO: organize developer-focused documentation

Git hook scripts are useful for identifying simple issues before commiting changes.

Install [pre-commit](https://pre-commit.com/#install) first, one-liner for Arch users:

```sh
sudo pacman -S python-pre-commit
```

Then install git hook scripts:

```sh
make git-hooks
```
