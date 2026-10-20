# Testing

Run these commands from the repository root.

## Unit tests

```sh
make test
```

## Integration tests

Run tests against staging:

```sh
make -C tests e2e env=staging
```

| Check | What it tests |
| --- | --- |
| API | API readiness, node identities, and membership through the VIP |
| Networking | DNS, pod traffic, and Service traffic across nodes when available |
| NetworkPolicy | Selective egress denial and connectivity from unselected pods |
| LoadBalancer | HTTP traffic through Istio's advertised gateway addresses |
| Storage | Data persistence with RBD and CephFS, including shared mounts |
| Snapshots | CSI snapshot creation and restore |
| Backups | Velero backup and restore through object storage |
| Secrets | OpenBao Secret injection and namespace permissions |
| Registry | HTTPS registry API, not image push or pull |
| GitOps | Argo CD deployment and self-healing |
| Apps | HTTPS responses from application routes |

You can also run individual checks:

```sh
make -C tests e2e env=staging filter='^Networking$'
```

## Benchmarks

```sh
make -C tests benchmarks env=staging
```

| Check | What it tests |
| --- | --- |
| Storage | Storage performance |
| Security | [K3s CIS checks](https://docs.k3s.io/security/hardening-guide) |
