# Add or remove nodes

Or how to scale horizontally. To replace the same node with a clean OS, remove it and add it again.

## Add new nodes

!!! tip

    You can add multiple nodes at the same time

Add its details to the inventory:

```diff title="infra/production/metal/hosts.json"
 {
   "homelab-1": {"ip": "192.168.1.110", "mac_address": "02:00:00:00:00:01"},
   "homelab-2": {"ip": "192.168.1.111", "mac_address": "02:00:00:00:00:02"},
   "homelab-3": {"ip": "192.168.1.112", "mac_address": "02:00:00:00:00:03"},
+  "homelab-4": {"ip": "192.168.1.113", "mac_address": "02:00:00:00:00:04"},
 }
```

Install the OS and join the cluster:

```
make infra env=production
```

Then verify the new node:

```sh
export KUBECONFIG="$PWD/infra/production/kubeconfig.yaml"
kubectl get nodes
```

That's it!

## Remove a node

!!! danger

    It is recommended to remove nodes one at a time

Use the environment's kubeconfig:

```sh
export KUBECONFIG="$PWD/infra/production/kubeconfig.yaml"
```

Drain the node:

```sh
kubectl drain ${NODE_NAME} --delete-emptydir-data --ignore-daemonsets --force
```

Remove the node from the cluster

```sh
kubectl delete node ${NODE_NAME}
```

Shutdown the node:

```
ssh root@${NODE_IP} poweroff
```

Remove it from the inventory:

```diff title="infra/production/metal/hosts.json"
 {
   "homelab-1": {"ip": "192.168.1.110", "mac_address": "02:00:00:00:00:01"},
   "homelab-2": {"ip": "192.168.1.111", "mac_address": "02:00:00:00:00:02"},
   "homelab-3": {"ip": "192.168.1.112", "mac_address": "02:00:00:00:00:03"},
-  "homelab-4": {"ip": "192.168.1.113", "mac_address": "02:00:00:00:00:04"},
 }
```

Reapply the environment with `make infra env=production`.
Don't replace the initializer with a fresh server in an existing cluster.

Optionally, after Ceph finishes recovering, clean up the removed node's OSD records using the [Rook guide](https://rook.io/docs/rook/latest/Storage-Configuration/Advanced/ceph-osd-mgmt/#remove-an-osd).
