# Configuration

## SSH keys

> Skip this step if you already have an Ed25519 key pair

Generate SSH private key and public key on the controller (your laptop or desktop):

```sh
ssh-keygen -t ed25519
```

## Fork this repository

Because this repository applies GitOps practices, this git repository is the source of truth for my homelab, so you'll need to folk it to make it yours.

After forking it, clone and replace the reference to my repository with yours (without `https://`):

```sh
./scripts/replace-gitops-repo "gitservicelikegithub.com/yourname/homelab"
```

Then commit and push the changes.

## Server list

Edit the following file and replace the MAC addresses with the one on your servers.
The IP addresses are the desired ones, since your servers have no operating system installed yet.

> If you're trying it out on the dev VM, use `dev.ini` instead of `prod.ini`

```ini
; metal/inventories/prod.ini
{{#include ../../../metal/inventories/prod.ini:5:}}
```

## Server hardware info

> Skip this step if you're trying out the dev VM

Change the following parameters based on your hardware.

```yaml
# metal/group_vars/all.yaml
{{#include ../../../metal/group_vars/all.yml:6:}}
```

- Disk: based on `/dev/$DISK`, in my case it's `sda`, but yours can be `sdb`, `nvme0n1`...
- Network interface: usually it's `eth0`, mine is `eno1`

## Update Ingresses to point to your domain

> Use [`home.arpa`](https://datatracker.ietf.org/doc/html/rfc8375) if you don't have a domain

My domain is `khuedoan.com`, run the following command to replace it with yours:

```sh
./scripts/replace-domain "yourdomain.com"
```

Commit and push the changes.
