# Infrastructure state and locking

`infra/root.hcl` stores each unit's OpenTofu state in Cloudflare R2 using its
S3-compatible API. The bucket remains `homelab-<environment>-tfstate` and the key
remains `<unit-relative-path>/tfstate.json`, for example
`homelab-staging-tfstate` / `cloudflare/tfstate.json`.

OpenTofu's `use_lockfile = true` acquires an exclusive lock by conditionally
creating `<state-key>.tflock` with `If-None-Match: *`. R2 supports that conditional
write. Runs from different hosts against the same environment and unit therefore
share a lock; different units and environments use different lock objects.
This is separate from S3 Object Lock (retention), which is not required.

## Credentials and prerequisites

- Use OpenTofu **1.10 or newer**, preferably the current stable patch release,
  and the repository's Terragrunt development shell. Both Terragrunt and the
  generated OpenTofu configuration enforce the minimum OpenTofu version.
- Cloudflare browser OAuth still authenticates the Cloudflare provider and
  `toolbox infra state ensure`, which creates the R2 bucket if needed.
- Separately obtain R2 **S3 API credentials** from an authorized administrator.
  Scope them to the state buckets you operate, with Object Read & Write access:
  listing and reading state, writing state and lock objects, and deleting lock
  objects. Ordinary read-only object credentials cannot run locking plans.
  Follow [Cloudflare's R2 authentication guide](https://developers.cloudflare.com/r2/api/tokens/).
- Supply the R2 access key ID as `AWS_ACCESS_KEY_ID` and its secret as
  `AWS_SECRET_ACCESS_KEY` in your shell or secret manager. An AWS SDK credentials
  profile is also supported. Clear an unrelated `AWS_SESSION_TOKEN` when using
  non-session R2 credentials. Do not reuse an AWS account's credentials here.
- A Cloudflare OAuth bearer token is **not** an R2 S3 key pair. This repository
  does not create long-lived credentials or convert the OAuth token into them.
  Do not put either key in HCL, command-line backend arguments, committed files,
  or generated `backend.tf`; the SDK reads them at runtime.
- Use the default OpenTofu workspace. This repository isolates environments by
  buckets and units by keys, not by OpenTofu workspaces.

Backend generation deliberately does not use Terragrunt's AWS S3 bucket
bootstrap. The existing toolbox hook manages the R2 bucket through Cloudflare;
OpenTofu then initializes the generated S3 backend normally. No DynamoDB or AWS
account is needed.

## Upgrade an existing HTTP backend

This is a coordinated change for **every client** of the state, including CI,
other machines, old checkouts, and cached/generated configurations. Old HTTP
clients do not honor S3 locks and can still overwrite state while a new client
holds a lock. Do not run mixed clients.

1. Stop and drain all runs for the environment. Keep them paused until every
   client is upgraded. Discard old saved plans, which embed backend settings.
2. Before changing the checkout, use the old configuration to pull each unit's
   current state. Save backups outside the repository in a private directory
   (`umask 077`), and record the state lineage, serial, and resource addresses.
   Treat backups as secrets. Do not continue if a known existing unit has no
   readable state.
3. Install this configuration and provide the R2 S3 credentials above on every
   machine/runner. Verify the account, bucket, and key exactly match the previous
   HTTP object's account, bucket, and path. For an existing deployment, do not
   let a newly created empty bucket substitute for the original bucket.
4. Reinitialize **each unit individually** with an explicit command, for example:

   ```sh
   cd infra/staging/cloudflare
   terragrunt run -- init -reconfigure
   terragrunt run -- state pull > /private/backup/staging-cloudflare-after.json
   ```

   Replace the example backup path with an existing private directory. With the
   same account, bucket, key, and default workspace, both APIs refer to the same
   R2 object; no state copy is required. `-reconfigure` changes the client's
   backend metadata. Do not use `-migrate-state` to copy the object onto itself,
   delete the old object, or push an empty state. If the intended destination
   differs, stop and plan a separate backed-up migration instead.
5. Compare the new pull to the backup: lineage, serial, and resource addresses
   must match. If they do not, stop; do not apply. Repeat for all units and all
   runners. New checkout/cache directories must also initialize against the
   verified existing object.
6. Run a plan on staging and inspect it for unexpected recreation. Verify lock
   contention using a disposable test key before resuming production automation.
   Resume runs only after old clients and old saved plans are retired.

The previous unconditional `init -reconfigure` has been removed. Later backend
changes must be reviewed explicitly, rather than silently adopting a different
state location. Never use `-lock=false` to work around contention.

## Lock contention and recovery

A concurrent run fails with the holder's lock information by default. To wait
for an ordinary busy lock, use a bounded timeout, for example
`terragrunt run -- plan -lock-timeout=5m`.

A crashed or interrupted client can leave a lock behind. OpenTofu also has a
[reported S3 lock retry issue](https://github.com/opentofu/opentofu/issues/4405):
a lost response can leave the caller's own lock behind. This is an availability
failure; bypassing it with unlocked operations reintroduces data loss.

Inspect the lock ID, owner, operation, and creation time. Confirm with all
operators and runners that the holder is no longer running before using
`terragrunt run -- force-unlock <LOCK_ID>` in the matching unit. Keep the
interactive confirmation. A lock's age alone is not proof that it is abandoned.
Never automatically expire/delete locks or unlock another active operation.

Keep private state backups. R2 does not implement S3 bucket versioning, so native
locking is not a backup or protection against authorized clients bypassing it.

## Offline regression test

From the repository root, with Python 3, OpenTofu, Terragrunt, and Git on `PATH`:

```sh
make test-state-backend
```

The test renders the actual root backend with a fake Cloudflare identity, then
runs two real OpenTofu processes against a loopback-only in-memory S3 mock. It
checks the conditional lock request, same-key contention, owner-only unlock,
release, independent unit/environment keys, and reuse of an existing state object. It uses only the built-in
`terraform_data` provider and does not log in, contact R2, or touch real state.
It verifies client wiring and lock protocol behavior, not R2 service behavior or
real credentials; validate those separately on a disposable R2 key.

References: [OpenTofu S3 backend](https://opentofu.org/docs/language/settings/backends/s3/),
[Cloudflare S3 API compatibility](https://developers.cloudflare.com/r2/api/s3/api/).
