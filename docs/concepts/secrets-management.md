# Secrets management

Infrastructure exports, Kanidm SSO clients, and Forgejo OAuth credentials are
published to OpenBao's `secret` KV v2 mount by the toolbox. Applications use
Kubernetes Secrets labeled `homelab.khuedoan.com/bao-secret: "true"` whose initial
`stringData` contains references such as
`bao:secret/data/woodpecker/forgejo#client_secret`.

The secrets webhook resolves these references when a Secret is created or
updated. It is an admission webhook, not a controller watching OpenBao. Changing
KV alone does not refresh an existing Secret or an environment variable in a
running container. Argo CD ignores the resolved Secret data to avoid fighting
with the webhook.

## Synchronization and rotation

The existing commands now reconcile both storage layers and request workload
reloads:

```sh
make sync-secrets env=staging
make sso env=staging
make sync-forgejo env=staging
```

`make sync env=staging` runs these in order. Use the corresponding `production`
environment only when the rotation is intended there. Commands use an explicit
cluster-admin kubeconfig, not an ambient `KUBECONFIG`.

Each toolbox secret sync:

1. Publishes changed KV records, preserving unchanged KV versions.
2. Lists only labeled Secrets in the namespaces of the records being synced.
3. Updates only mapped keys for those records, preserving other data and metadata.
   The command never deletes or recreates a Secret. Missing Secrets are left for
   GitOps and the admission webhook to create, which also supports first bootstrap.
4. Finds Deployments, StatefulSets, and DaemonSets that refer to those Secrets
   through `env`, `envFrom`, init containers, Secret volumes, or projected volumes.
   It patches their pod-template annotation to request a normal controller rollout.

The rollout marker is a deterministic hash of the Kubernetes UIDs and resource
versions of **all** managed Secrets consumed by that workload. No secret values
or hashes of credential values are placed in workload annotations. Including all
dependencies keeps separate infra, SSO, and Forgejo syncs from oscillating the
marker. An unchanged repeat sync makes no Secret or workload writes. The first
sync after adopting this mechanism requests one baseline rollout; a later
metadata-only Secret update can also request a rollout because resource versions
represent the whole Secret.

Reconciliation runs even when KV values are unchanged. If a Secret update or
workload patch fails after KV was written, fix the reported problem and rerun the
same sync command. Already-successful updates are preserved and the failed step
is retried. Kubernetes conflicts are handled without overwriting unrelated Secret
changes; workload-patch conflicts return an error for a fresh sync retry. API
errors that could contain credential values are not printed.

Argo CD's shared application policy ignores only the toolbox-owned pod-template
annotation and uses `RespectIgnoreDifferences=true` so later syncs preserve it.
Apply that policy before using the updated toolbox on an existing installation.
Without it, GitOps could remove the marker and undo a requested rollout.
The existing labeled-Secret data ignores are also respected during apply, so
changes to unmapped static data keys (for example a literal username) require an
explicit operator update or migration; changing those keys in Git alone is not
a rotation path.

## Reference mappings

The webhook replaces reference strings with secret values. Toolbox therefore
cannot recover references by examining live Secret data. Existing repository
Secrets are covered by the explicit namespace/name/key registry in
`toolbox/internal/secrets/reconcile.go`: Actual Budget, Argo CD, cert-manager,
cloudflared, external-dns, Forgejo admin and SSO, Grafana, Woodpecker OAuth and agent,
and the registry admin Secret. This fallback works on already-deployed Secrets
without deleting or recreating them. Keep this registry aligned with manifest
renames or changed bindings.

For new or customized Secrets, retain the mappings in an annotation in GitOps:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: example-sso
  namespace: example
  labels:
    homelab.khuedoan.com/bao-secret: "true"
  annotations:
    homelab.khuedoan.com/bao-references: >-
      {"CLIENT_SECRET":"bao:secret/data/example/sso#client_secret"}
stringData:
  CLIENT_SECRET: bao:secret/data/example/sso#client_secret
```

The annotation replaces the fallback mapping entirely. It maps Kubernetes data
keys to unversioned references in `secret/data/<same namespace>/...#<field>`.
Do not put resolved values in the annotation. Cross-namespace, templated,
version-pinned, and non-KV references are intentionally unsupported. Missing
fields, invalid mappings, and unknown labeled Secrets fail explicitly rather
than guessing a binding. Unlabeled Secrets are never changed. Immutable Secrets
cannot be rotated in place and require an operator-managed migration.

## Reload boundaries and operational checks

A successful sync means the Secret and rollout-template updates were accepted;
it does not wait for new pods to become Ready. Check workload health afterward,
for example:

```sh
kubectl --kubeconfig infra/staging/kubeconfig.yaml -n woodpecker get deployments,statefulsets,pods
kubectl --kubeconfig infra/staging/kubeconfig.yaml -n woodpecker rollout status deployment/woodpecker-server
```

Use the workload kind/name actually rendered by the installed chart. Paused
Deployments, OnDelete controllers, and partitioned StatefulSets return an error
instead of claiming to have reloaded all replicas. Change their rollout strategy
or handle their upgrade explicitly before rerunning sync. Normal rollout failure
and unavailable capacity still require the usual Kubernetes diagnosis.

This is explicit reconciliation, not a continuously running secret controller.
Only records supplied by the calling command are reconciled; this does not scan
all live Bao records. In particular, the operator-generated `woodpecker/agent`,
`forgejo/admin`, and `zot/admin` records have legacy bindings but are not supplied
by the current `make sync` commands. Manually changing those records in Bao needs
an explicit operator-managed Secret update and consumer reload. Other manual Bao
edits may be overwritten by a command's authoritative input values. Run the
relevant sync after changing its source of truth or recovering a failed sync.
Consumers
that read Secrets through the Kubernetes API, such as Argo CD and cert-manager,
use their own watches/reloads; there is no pod-template Secret reference to follow.
Direct Pods, already-running Jobs, custom controllers, CSI/external references,
and application-specific provisioning beyond a pod restart are not automatically
restarted. New Jobs read updated Secrets when they start. Applications that copy
credentials into persistent state must provide their own reload/provisioning flow.
Rotating a provider credential may temporarily interrupt service before rollout
finishes; where supported, use overlapping credentials and an appropriate
maintenance window.
