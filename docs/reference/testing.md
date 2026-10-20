# Testing

Run these commands from the repository root.

## Local checks

```sh
just test
```

List available tests:

```sh
just test list
```

## Integration tests

Run tests against staging:

```sh
just test e2e staging
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
just test e2e staging networking
```

## Benchmarks

```sh
just test benchmarks staging
```

| Check | What it tests |
| --- | --- |
| Storage | Storage performance |
| Security | [K3s CIS checks](https://docs.k3s.io/security/hardening-guide) |
