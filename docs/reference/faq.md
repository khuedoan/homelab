# FAQ

## Do I need to install Linux on my servers before provisioning the homelab?

No, and it's the beauty of this set up. You start from scratch (empty hard drive), type a single command on your laptop/PC and it will install the OS for you automatically, in parallel via the network.

## Do I need to keep the PXE server running?

No, the ephemeral PXE server is stateless, after Linux is installed on your servers you can shut it down (or not, ideally you don't even need to care about its existence).
The Ansible set up in `./metal` is idempotent and will start the PXE server if needed.

## Why use Fedora CoreOS instead of a traditional Linux distro?

There are several benefits:

- Automatic update
- Atomic upgrade
- Immutable
- Minimal
- Faster install time (3 minutes compare to 5 minutes on Fedora or CentOS)
- Faster provisioning (Docker already installed, save 5 minutes)

However this is a fairly new distro, so it may not be really stable yet.

## Where Terraform state is stored?

In a Docker container on the first node, which was created by the `./metal` layer (it's not HA _yet_).

However I'm experimenting with Cluster API, remove the needs for a Terraform state storage.
