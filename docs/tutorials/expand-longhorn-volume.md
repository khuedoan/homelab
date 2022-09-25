# Expand Longhorn volume

Longhorn requires volumes to be detached in order to expand them.

An example of volume expansion:

```bash
kubectl -n gitea delete --cascade=orphan sts gitea-postgres
kubectl -n gitea delete pod gitea-postgres-0
kubectl -n gitea patch pvc pgdata-gitea-postgres-0 -p '{ "spec": { "resources": { "requests": { "storage": "1.5Gi" }}}}'
```
