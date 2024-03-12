# Single node cluster adjustments

!!! danger

    This is not officially supported and I don't regularly test it,
    I highly recommend using multiple nodes.

    Using a single node could lead to data loss unless your backup strategy is rock solid,
    make sure you are **ABSOLUTELY CERTAIN** this is what you want.

Update the following changes, then commit and push.

## Remove storage redundancy

Set pod counts and number of data copies to `1`:

```yaml title="system/rook-ceph/values.yaml" hl_lines="4 6 11 12 18 22 25"
rook-ceph-cluster:
  cephClusterSpec:
    mon:
      count: 1
    mgr:
      count: 1
  cephBlockPools:
    - name: standard-rwo
      spec:
        replicated:
          size: 1
          requireSafeReplicaSize: false
  cephFileSystems:
    - name: standard-rwx
      spec:
        metadataPool:
          replicated:
            size: 1
        dataPools:
          - name: data0
            replicated:
              size: 1
        metadataServer:
          activeCount: 1
          activeStandby: false
```

## Disable automatic upgrade

Because they will try to drain the only node, the pods will have no place to go.
Remove them entirely:

```sh
rm -rf system/kured
```

Commit and push the change.
You can revert it later when you add more nodes.
