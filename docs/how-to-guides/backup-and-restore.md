# Backup and restore

## Prerequisites

Keep access to your Cloudflare account and existing OpenTofu state in R2. The state contains the backup credentials and Kopia repository password.

From the development shell, select the environment and its kubeconfig:

```sh
env=production
export KUBECONFIG="$PWD/infra/$env/kubeconfig.yaml"
```

## Back up the cluster

Velero runs hourly backups. To create one now:

```sh
velero backup create --from-schedule cluster --wait
velero backup get
velero backup describe BACKUP_NAME --details
```

Check the result for warnings and failed uploads.

## Restore an application

Pause the affected Argo CD Applications and their ApplicationSet, then stop the workloads.

Start in read-only mode:

```sh
export TF_VAR_backup_read_only=true
(cd "infra/$env/platform" && terragrunt run apply)
```

Check that app reconciliation remains paused after applying the platform.

Select a backup and restore the app namespace:

```sh
velero backup get
velero restore create --from-backup BACKUP_NAME --include-namespaces APP_NAMESPACE --wait
velero restore get
velero restore describe RESTORE_NAME --details
```

Verify the restored data before resuming workloads and GitOps reconciliation.
Then resume scheduled backups:

```sh
unset TF_VAR_backup_read_only
(cd "infra/$env/platform" && terragrunt run apply)
```
