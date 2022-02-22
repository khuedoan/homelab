# Single node cluster adjustments

Update the following changes, then commit and push.

## Reduce Longhorn replica count

Set the `defaultClassReplicaCount` to 1:

```yaml
# system/longhorn-system/values.yaml
{{#include ../../../system/longhorn-system/values.yaml}}
```

## Disable automated system upgrade

Because the system upgrade controller will try to drain the node,
the pods will have no place to go on a single node set up.

Remove the system upgrade controller entirely:

```sh
rm -rf system/system-upgrade
```
