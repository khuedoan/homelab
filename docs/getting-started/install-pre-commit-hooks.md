# Install pre-commit hooks

Git hook scripts are useful for identifying simple issues before commiting changes.

Install [pre-commit](https://pre-commit.com/#install) first, one-liner for Arch users:

```sh
sudo pacman -S python-pre-commit
```

Then install git hook scripts:

```sh
make git-hooks
```

If you want to enable pre-commit on all repositories without enabling it manually, see [automatically enabling pre-commit on repositories](https://pre-commit.com/#automatically-enabling-pre-commit-on-repositories).
