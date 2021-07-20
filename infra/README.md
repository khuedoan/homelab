# Infrastructure

Bare metal servers life cycle management:

- Automatically discover bare metal servers
- Automatically wipe the disk
- Install Linux on empty machines via the network (using iPXE)

Kubernetes cluster creation:

- Create ephemeral management cluster (using Kubernetes in Docker)
- Create target cluster
- Pivot the management plane over to the new cluster
- Remove the ephemeral management cluster
