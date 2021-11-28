# Configuration

## SSH keys

> Skip this step if you already have an Ed25519 key pair

Generate SSH private key and public key on the controller (your laptop or desktop):

```sh
ssh-keygen -t ed25519
```

## Update variables and parameters

This project follows GitOps practice, the git repository is the source of truth for you homelab setup, so you will need to [fork this repository](https://github.com/khuedoan/homelab) and update these config files based on those information:

- [ ] `metal/inventories/prod.ini`
- [ ] `metal/group_vars/all.yml`
- [ ] TODO git repository config in `bootstrap/...`
- [ ] TODO single place for Ingress domain
- [ ] TODO single place for docs link config
