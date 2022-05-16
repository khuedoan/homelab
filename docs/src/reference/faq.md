# FAQ

## Is it necessary to install Linux on my servers before setting up the homelab?

No, and that's the beauty of this setup.
You start with empty hard drives, type a single command on your laptop/PC,
and it will install the operating system for you automatically and in parallel over the network.

## Is it necessary to keep the PXE server running?

No, the ephemeral PXE server is stateless, once Linux is installed on your servers,
you can shut it down (or not, ideally, you don't even need to be aware of its existence).
The Ansible setup in `./metal` is idempotent and will start the PXE server if needed.
