# Toolbox

Toolbox bootstraps infrastructure state storage, enrolls k3s servers, and exports
cluster credentials. Run it from the repository root so inventory paths resolve
correctly.

```sh
toolbox --help
toolbox cluster enroll --help
toolbox cluster enroll --environment staging
toolbox cluster kubeconfig --environment staging --output infra/staging/kubeconfig.yaml
toolbox infra state ensure --account-id "$CLOUDFLARE_ACCOUNT_ID" --bucket tfstate-production
```

Cluster commands use the explicit `--environment` inventory, not `KUBECONFIG`.
Kubeconfig export defaults to `infra/kubeconfig.yaml`. Use `--output` to select
another destination or `--output -` to send credentials to stdout.
State storage uses `CLOUDFLARE_TFSTATE_API_TOKEN` for authentication.
Terragrunt invokes these operations through the infrastructure hooks.

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
    └── state/
```

`internal/cli/` owns Cobra commands, flag validation, and output routing.
`root.go` lists the command tree. The CLI imports `cluster/` and `state/`.
These packages accept contexts and explicit inputs, not Cobra commands.
They do not import each other.

`internal/state/` owns R2 bucket provisioning through the Cloudflare SDK.
It leaves existing buckets unchanged and handles concurrent creation.

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
make -C toolbox lint
```

The test target runs vet and race-enabled tests. Tests use temporary files,
subprocess fixtures, and local HTTP, TLS, SSH, and SFTP servers. They do not
enroll real nodes or create real Cloudflare buckets. Tests live beside their
owner and cover enrollment safety, kubeconfig publication, and state bootstrap.

The lint target uses the repository's shared `.golangci.yml` and includes test
files. `nix develop` provides golangci-lint. Checks cover formatting, selected Go
style conventions, correctness, and cognitive complexity above 20.
That complexity threshold is a project policy, not a Google style requirement.
Lint exits unsuccessfully when findings exist and does not rewrite code.
The configuration does not enforce the entire Google Go style guide or impose
a line-length limit. Existing findings are not suppressed by a baseline.

Put `terragrunt` on `PATH` to run the local state-bootstrap contract test. It uses
fake toolbox and OpenTofu executables and skips if Terragrunt is absent.
Set `SFTP_SERVER` to OpenSSH's `sftp-server`
executable for remote-file publication tests. Without it, tests find OpenSSH
through `sshd` on `PATH` and skip if `sshd` is absent.
