# Tests

Run these commands from the repository root.

## Run offline checks

```sh
make test
```

This runs toolbox unit tests, config-loader tests, vet, and race checks. It does
not contact a cluster.

```sh
make -C tests lint
```

The lint target uses the shared `.golangci.yml` and includes test files.
`nix develop` provides golangci-lint. Checks cover formatting, selected Go style
conventions, correctness, and cognitive complexity above 20. That
threshold is a project policy, not a Google style requirement. Findings fail
the command without rewriting code. Lint does not contact a cluster.
Run `make lint` from the repository root to check both Go modules.

## Validate a cluster

Run live tests against staging. Tests never use your ambient Kubernetes context.

```sh
make -C tests e2e config=config/staging.json
make -C tests e2e config=config/staging.json filter='^Networking$'
```

`TestCluster` contains the following capability checks:

- `API` authenticates through the control-plane VIP, checks readiness and cluster
  membership, and compares node and namespace identities with each node's local API.
- `Networking` checks DNS, direct pod HTTP, and Service HTTP. With multiple nodes,
  the server and probe run on different nodes.
- `LoadBalancer` reaches a temporary HTTP backend through every advertised
  external address and checks the exact response body.
- `Storage` provisions each configured class, writes data, deletes the writer pod,
  and reads the same data in another pod. RWO stays on the writer's node.
  RWX also mounts the volume in concurrent pods, on different nodes when available.
- `Registry` checks HTTPS `/v2/` and the Distribution v2 response header. It does
  not test image push or pull, and does not write registry data.
- `GitOps` creates an Argo CD Application at a pinned Git revision, checks its
  HTTP response, and verifies that self-heal repairs a changed deployment image.
- `Apps` checks HTTPS 200 for each configured application ingress.

The fixture verifies inventory identities over SSH before exporting a private
mode-0600 kubeconfig through toolbox. Filtered runs still perform this preflight.
Tests create isolated namespaces and delete their resources on success or failure.
Cleanup waits for namespace deletion and volume reclamation. Storage classes must
use the `Delete` reclaim policy. Tests do not install or reset nodes, run Terraform,
or change enrollment.

Prerequisites are Go, root SSH access, trusted host keys, and access to the VIP on
port 6443. Each selected node needs a management IP and MAC in the inventory.
SSH uses `SSH_AUTH_SOCK`, or `SSH_KEY`, defaulting to `~/.ssh/id_ed25519`.
Toolbox builds from the checkout unless `TOOLBOX_BIN` specifies an absolute path.
Workload tests need image-pull access and schedulable nodes.

## Configure capabilities

Config paths are absolute or relative to `tests/`. Inventories and VIP settings
come from `infra/<environment>/metal/hosts.json` and `cluster/config.json`.
Known addresses and MACs must not overlap excluded environments.
`dns_domain` specifies the expected Kubernetes service DNS suffix.

```json
{
	"environment": "staging",
	"exclude_environments": ["production"],
	"dns_domain": "cluster.local",
	"apps": [],
	"storage": [
		{"class": "standard-rwo", "access_mode": "ReadWriteOnce"},
		{"class": "standard-rwx", "access_mode": "ReadWriteMany"}
	],
	"registry": {"namespace": "zot", "ingress": "zot"},
	"gitops_namespace": "argocd",
	"load_balancer": {
		"namespace": "ingress-nginx",
		"service": "ingress-nginx-controller",
		"ingress_class": "nginx"
	}
}
```

Only configure storage and ingresses that the target is expected to provide.
Empty lists and omitted registry or GitOps settings skip those capabilities.
A missing configured capability fails. HTTPS checks require valid certificates
and connect directly without using proxy environment variables.

An omitted `load_balancer` creates a temporary LoadBalancer Service on port 80.
That port must be free on nodes when using K3s ServiceLB. A configured
`load_balancer` tests an existing ingress controller's LoadBalancer Service on
port 80 through a temporary ingress and backend. This avoids competing with
the controller for the same host port. The test runner must reach the advertised
external addresses directly.

Target names and capability expectations come from the config and inventory
files. The test code does not branch on environment names.
`config/staging.json` selects API, networking, the existing ingress controller's
LoadBalancer, local-path storage, and Argo CD. Registry and application HTTPS
checks are not configured. Excluded inventories are read only to reject
overlapping addresses and MACs.

The GitOps check needs access to GitHub and Docker Hub. A single-node target
cannot prove cross-node traffic, shared storage across nodes, or API failover.

## Run benchmarks

Benchmarks are opt-in Go tests that use client-go to create Jobs and PVCs.
They share the cluster fixture and namespace cleanup with the E2E checks.

```sh
make -C tests benchmarks config=config/staging.json filter='^Security$'
make -C tests benchmarks config=config/staging.json filter='^Storage$'
```

Storage benchmarks skip when the selected config has no storage classes.

`Storage` runs fio for 30 seconds on a 1 GiB PVC for each configured class. It
checks for successful read and write I/O and reports IOPS and KiB per second.
It does not impose a performance threshold. The fio container uses the dbench
image pinned by digest. `Security` runs kube-bench's K3s CIS 1.7 checks on every
selected node.
It fails for CIS failures and reports manual warnings separately.

Security Jobs use host PID access and read-only mounts of K3s data, kubelet data,
systemd units, and the Nix store. The target must permit these workloads. Storage
benchmarks generate disk load. Both checks can fail without changing cluster
configuration. Go cleanup cannot run after a forced process termination, so
inspect namespaces prefixed `homelab-test-` after an interrupted run.

## Layout

```text
tests/
├── config/             Staging target and expected capabilities
├── e2e/                Capability assertions and Pod helpers
├── benchmarks/         fio, CIS checks, and Job helpers
├── internal/
│   ├── fixture/        Safe connection, SSH, namespaces, and PVCs
│   └── testenv/        Config loading and target isolation checks
├── Makefile
├── go.mod
└── go.sum
```

Each suite owns its assertions and workload definitions. Neither suite imports
the other. Shared fixture code handles connection and resource lifecycle only.
`fixture.Cluster` holds the validated target and typed and dynamic client-go clients. Helpers use
the current subtest's context, and cleanup has a separate bounded context.
