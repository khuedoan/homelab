# Add or remove nodes

Or how to scale vertically. To replace the same node with a clean OS, remove it and add it again.

## Add new nodes

!!! tip

    You can add multiple nodes at the same time

Add its details to the inventory **at the end of the group** (masters or workers):

```diff title="metal/inventories/prod.yml"
diff --git a/metal/inventories/prod.yml b/metal/inventories/prod.yml
index 7f6474a..1bb2cbc 100644
--- a/metal/inventories/prod.yml
+++ b/metal/inventories/prod.yml
@@ -8,3 +8,4 @@ metal:
     workers:
       hosts:
         metal3: {ansible_host: 192.168.1.113, mac: '00:23:24:d1:f5:69', disk: sda, network_interface: eno1}
+        metal4: {ansible_host: 192.168.1.114, mac: '00:11:22:33:44:55', disk: sda, network_interface: eno1}
```

Install the OS and join the cluster:

```
make metal
```

That's it!

## Remove a node

!!! danger

    It is recommended to remove nodes one at a time

Remove it from the inventory:

```diff title="metal/inventories/prod.yml"
diff --git a/metal/inventories/prod.yml b/metal/inventories/prod.yml
index 7f6474a..d12b50a 100644
--- a/metal/inventories/prod.yml
+++ b/metal/inventories/prod.yml
@@ -4,7 +4,6 @@ metal:
       hosts:
         metal0: {ansible_host: 192.168.1.110, mac: '00:23:24:d1:f3:f0', disk: sda, network_interface: eno1}
         metal1: {ansible_host: 192.168.1.111, mac: '00:23:24:d1:f4:d6', disk: sda, network_interface: eno1}
-        metal2: {ansible_host: 192.168.1.112, mac: '00:23:24:e7:04:60', disk: sda, network_interface: eno1}
     workers:
       hosts:
         metal3: {ansible_host: 192.168.1.113, mac: '00:23:24:d1:f5:69', disk: sda, network_interface: eno1}
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
