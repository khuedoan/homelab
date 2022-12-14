# Restore backup

- scale down replicas and delete pvc
- restore latest backup from [UI](https://longhorn.k8s.grigri/#/backup):
  - pv with the same name
  - pvc with the same name
- scale up and recreate pvc. Warning: change storageClass to `longhorn-static`
