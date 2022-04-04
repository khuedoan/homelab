# Architecture

TODO

## Components

```
+--------------+
|    ./apps    |
|--------------|
|  ./platform  |
|--------------|       +------------+
|   ./system   |- - - -| ./external |
|--------------|       +------------+
| ./bootstrap  |
|--------------|
|   ./metal    |
|--------------|
|   HARDWARE   |
+--------------+
```

- `./metal`: bare metal management, install Linux and Kubernetes
- `./bootstrap`: GitOps bootstrap with ArgoCD
- `./system`: critical system components for the cluster (load balancer, storage, ingress, operation tools...)
- `./platform`: essential components for service hosting platform (vault, git...)
- `./apps`: user facing applications
- `./external` (optional): externally managed services
- `./tools`: tools container, includes all the tools you'll need
- `./docs`: all documentation go here, this will generate a searchable web UI
