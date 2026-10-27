# Use both GitHub and Forgejo

Even though we self-host Forgejo, you may still want to use GitHub as a backup and for discovery.

Add both push URLs (replace my repositories with yours):

```sh
git remote set-url --add --push origin https://git.khuedoan.com/ops/homelab.git
git remote set-url --add --push origin git@github.com:khuedoan/homelab
```

Now you can just run `git push` like usual and it will push to both GitHub and Forgejo.
