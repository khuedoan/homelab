# Toolbox

Toolbox is the administrative CLI for the homelab. Run it from the repository
root so inventory and chart paths resolve correctly.

```sh
toolbox --help
toolbox cluster enroll --help
toolbox cluster kubeconfig --environment staging
```

`--kubeconfig` takes precedence over `KUBECONFIG`. An empty value selects
`infra/kubeconfig.yaml`. Each subprocess receives that selection without changing
the parent environment. Cluster enrollment and kubeconfig export use their explicit
`--environment` inventory rather than the ambient Kubernetes context.

## Code ownership

Toolbox has one Go module and one executable entry point. Its packages are internal
to that executable:

```text
toolbox/
├── go.mod
├── main.go
└── internal/
    ├── cli/
    ├── cluster/
    ├── identity/
    ├── charts/
    ├── backup/
    ├── state/
    └── process/
```

`internal/cli/` owns Cobra commands, flag validation, warnings, and output routing.
`root.go` lists the command tree. `process.go` converts parsed flags and streams
into command-local subprocess configuration. Each command family has a named
file. Short status, DNS, WireGuard, Argo CD password, and screenshot operations stay direct.

The operation packages accept contexts and explicit inputs, not Cobra commands:

- `identity/` owns Gitea and Kanidm integration setup, Kanidm account creation,
  and account recovery. Ingress lookup, Secret publication, and PTY login are private.
- `charts/` owns chart scaffolding and revision comparison, including temporary
  files, rendering, and cleanup.
- `backup/` owns PVC backup and restore policy and ordered resource application.
- `state/` owns R2 bucket provisioning through the Cloudflare SDK.
- `process/` owns child environment and stream handling. `Run` forwards output,
  `Output` captures stdout and forwards diagnostics, and `PrivateOutput` captures
  credentials without forwarding or retaining stderr. `Command` leaves streams
  unset for private Secret writes and PTY use.

The CLI imports operation packages. Identity, charts, and backup import process.
Operation packages do not import the CLI or each other. Kubernetes subprocess
helpers stay private to their workflow owner. Inventory-driven cluster access
uses native SSH, SFTP, and Kubernetes clients.

`internal/cluster/` owns enrollment and authenticated cluster access:

- `inventory.go` parses and validates initializer, joiner, and VIP addresses.
- `enroll.go` coordinates preflight, sequential joins, and readiness checks.
- `state.go` classifies observed node state without remote I/O.
- `join.go` inspects and joins nodes under a remote lock.
- `files.go` validates private remote files and publishes tokens without overwrite.
- `ssh.go` owns trusted SSH authentication, connection lifetimes, and locking.
- `nodes.go` verifies API identities and Ready control-plane membership.
- `kubeconfig.go` verifies initializer and VIP access before returning credentials.

Enrollment inspects all joiners before writing credentials. It rechecks initializer
credentials before each join and preserves completed joins after interruption.
It does not install machines, reset datastores, or overwrite conflicting credentials.

## Verification

```sh
make -C toolbox test
```

The target runs vet and race-enabled tests. Tests use temporary files, subprocess
fixtures, and local HTTP, TLS, SSH, and SFTP servers. They do not enroll real nodes
or create real Cloudflare buckets. Workflow tests live beside their owner. CLI
tests cover flags, routing, streams, and credential handling across package boundaries.

Put `terragrunt` on `PATH` to run the local state-bootstrap contract test. It uses
fake toolbox and OpenTofu executables and skips if Terragrunt is absent.
Set `SFTP_SERVER` to OpenSSH's `sftp-server`
executable for remote-file publication tests. Without it, tests find OpenSSH
through `sshd` on `PATH` and skip if `sshd` is absent.
