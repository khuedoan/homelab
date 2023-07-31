# Kubernetes upgrade

Update `k3s_version` to desired version and then run:

```bash
cd metal
ANSIBLE_EXTRA_ARGS="-t k3s-upgrade -e serial=1" make cluster
```
