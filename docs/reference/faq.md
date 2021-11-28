# FAQ

## Do I need to install Linux on my servers before provisioning the homelab?

No, and it's the beauty of this set up. You start from scratch (empty hard drive), type a single command on your laptop/PC and it will install the OS for you automatically, in parallel via the network.

## Do I need to keep the PXE server running?

No, the ephemeral PXE server is stateless, after Linux is installed on your servers you can shut it down (or not, ideally you don't even need to care about its existence).
The Ansible set up in `./metal` is idempotent and will start the PXE server if needed.
