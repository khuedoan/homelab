# Add or remove nodes

Or how to scale vertically. To replace the same node with a clean OS, remove it and add it again.

## Add new nodes

> You can add multiple nodes at the same time

Ensure that it meets the requirements in [prerequisites](../deployment/prerequisites.md), then add its details to the inventory **at the end of the group** (kube_control_plane or kube_node):

```diff
diff --git a/metal/inventories/master/inventory.ini b/metal/inventories/master/inventory.ini
index fe0ca8b..ddbca8d 100644
--- a/metal/inventories/master/inventory.ini
+++ b/metal/inventories/master/inventory.ini
@@ -18,6 +18,7 @@ odroid-hc4
 k8s-odroid-hc4-1
 k8s-odroid-hc4-2
 k8s-odroid-hc4-3
+k8s-odroid-hc4-4

 [amd64]
 k8s-amd64-1
```

Setup OS and network: [manual Setup](../deployment/manual-setup.md)

Join the cluster:

```bash
make metal
```

That's it!

## Remove a node

> It is recommended to remove nodes one at a time

Remove it from the inventory:

```diff
diff --git a/metal/inventories/master/inventory.ini b/metal/inventories/master/inventory.ini
index fe0ca8b..19891bf 100644
--- a/metal/inventories/master/inventory.ini
+++ b/metal/inventories/master/inventory.ini
@@ -17,7 +17,6 @@ odroid-hc4
 [odroid-hc4]
 k8s-odroid-hc4-1
 k8s-odroid-hc4-2
-k8s-odroid-hc4-3

 [amd64]
 k8s-amd64-1
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
