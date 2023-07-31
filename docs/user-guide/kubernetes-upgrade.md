# Kubernetes upgrade

Update `k3s_version` to desired version and then run:

```bash
cd metal
ANSIBLE_EXTRA_ARGS="-t k3s-upgrade -e serial=1" make cluster
```

*Note*: it worked perfectly with serial=100% or running it by default.
