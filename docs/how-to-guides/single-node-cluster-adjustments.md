# Single node cluster adjustments

Update the following changes, then commit and push.

## Reduce Longhorn replica count

Set the `defaultClassReplicaCount` to 1:

```yaml title="system/longhorn-system/values.yaml" hl_lines="6"
--8<--
system/longhorn-system/values.yaml
--8<--
```

## Disable automatic upgrade

Because they will try to drain the only node, the pods will have no place to go.
Remove them entirely:

```sh
rm -rf system/kured
```

Commit and push the change.
You can revert it later when you add more nodes.
