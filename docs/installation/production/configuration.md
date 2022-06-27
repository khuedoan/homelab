# Configuration

Open the tools container if you haven't already:

```sh
make tools
```

Run the following script to configure the homelab:

```sh
make configure
```

Example input:

<!-- TODO update example input -->

```
Text editor (nvim):
Enter seed repo (github.com/khuedoan/homelab): github.com/example/homelab
Enter your domain (khuedoan.com): example.com
```

It will prompt you to edit the inventory, for example:

> - The IP addresses are the desired ones, not the current one, since your servers have no operating system installed yet.
> - Disk: based on `/dev/$DISK`, in my case it's `sda`, but yours can be `sdb`, `nvme0n1`...
> - Network interface: usually it's `eth0`, mine is `eno1`
> - MAC address: the **lowercase, colon separated** MAC address of the above network interface

```yaml title="metal/inventories/prod.yml"
--8<--
metal/inventories/prod.yml
--8<--
```

At the end it will show what has changed. After examining the diff, commit and push the changes.
