# Configuration

## Fork this repository

Because this repository applies GitOps practices, this git repository is the source of truth for my homelab, so you'll need to folk it to make it yours.

After forking it, clone and replace the reference to my repository with yours (without `https://`):

```sh
./scripts/replace-gitops-repo "github.com/yourname/homelab"
```

Then commit and push the changes.

## Server list and hardware info

Edit the following file and replace the hardware information with the one on your servers:

> Use `dev.yml` instead of `prod.yml` if you're trying out the dev VM

```yaml
# metal/inventories/prod.yml
{{#include ../../../metal/inventories/prod.yml:3:}}
```

- The IP addresses are the desired ones, not the current one, since your servers have no operating system installed yet.
- Disk: based on `/dev/$DISK`, in my case it's `sda`, but yours can be `sdb`, `nvme0n1`...
- Network interface: usually it's `eth0`, mine is `eno1`

## Update Ingresses to point to your domain

> Use [`home.arpa`](https://datatracker.ietf.org/doc/html/rfc8375) if you don't have a domain

My domain is `khuedoan.com`, run the following command to replace it with yours:

```sh
./scripts/replace-domain "yourdomain.com"
```

Commit and push the changes.
