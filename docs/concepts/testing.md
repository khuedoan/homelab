# Testing the system

The Go test module lives in `tests/`. It uses the standard `testing` package,
Testify assertions, and [Terratest](https://terratest.gruntwork.io) helpers.

E2E tests exercise the Kubernetes API, networking, external load balancing,
persistent storage, registry API, and application ingresses. Capabilities come
from an explicit target configuration. Tests never use the ambient Kubernetes
context. Workloads and opt-in benchmark Jobs use client-go and clean up their
resources.

Run offline checks with `make test`. Run cluster checks with
`make -C tests e2e config=config/staging.json`.

See `tests/README.md` for commands, configuration, coverage, and prerequisites.
