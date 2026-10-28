# Single node cluster adjustments

> [!WARNING]
>
> This is not officially supported and I don't regularly test it,
> I highly recommend using multiple nodes.
>
> Using a single node could lead to data loss unless your backup strategy is rock solid,
> make sure you are **ABSOLUTELY CERTAIN** this is what you want.

Make the following changes, then commit and push.

## Configure the node

- Keep one host in `infra/production/metal/hosts.json`.
- Set it as `init_host` in `infra/production/cluster/config.json`.

## Remove storage redundancy

Set `ceph_replica_count` to `1` in the `inputs` block of `infra/production/platform/terragrunt.hcl`:

```hcl title="infra/production/platform/terragrunt.hcl"
ceph_replica_count = 1
```
