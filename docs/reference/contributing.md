# Contributing

## How to contribute

### Bug report

You can [create a new GitHub issue](https://github.com/khuedoan/homelab/issues/new/choose) with the bug report template.

### Submitting patches

Because you may have a lot of changes in your fork, you can't create a pull request directly from your `master` branch.
Instead, create a branch from the upstream repository and commit your changes there:

```sh
git remote add upstream https://github.com/khuedoan/homelab
git fetch upstream
git checkout upstream/master
git checkout -b contrib-fix-something

# Make your changes here
#
# nvim README.md
# git cherry-pick a1b2c3
#
# commit, push, etc. as usual
```

Then you can send the patch using [GitHub pull request](https://github.com/khuedoan/homelab/pulls) or `git send-email` to <mail@khuedoan.com>.
