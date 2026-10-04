# Backup and restore

## Prerequisites

Create an S3 bucket to store backups. You can use AWS S3, Minio, or
any other S3-compatible provider.

- For AWS S3, your bucket URL might look something like this:
  `https://s3.amazonaws.com/my-homelab-backup`.
- For Minio, your bucket URL might look something like this:
  `https://my-s3-host.example.com/homelab-backup`.

Follow your provider's documentation to create a service account with the
following policy (replace `my-homelab-backup` with your actual bucket name):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::my-homelab-backup",
        "arn:aws:s3:::my-homelab-backup/*"
      ]
    }
  ]
}
```

Save the access key and secret key to a secure location, such as a password
manager. While you're at it, generate a new password for Restic encryption and
save it there as well.

!!! example

    I use Minio for my homelab backups. Here's how I set it up:

    - Create a bucket named `homelab-backup`.
    - Create a service account under Identity -> Service Accounts -> Create
      Service Account:
        - Enable Restrict beyond user policy.
        - Paste the policy above.
        - Click Create and copy the access key and secret key
    - I also set up Minio replication to store backups in two locations: one in
      my house and one remotely.

## Add backup credentials to global secrets

Add the following to `external/terraform.tfvars`:

```hcl
extra_secrets = {
  restic-password = "xxxxxxxxxxxxxxxxxxxxxxxx"
  restic-s3-bucket = "https://s3.amazonaws.com/my-homelab-backup-xxxxxxxxxx"
  restic-s3-access-key = "xxxxxxxxxxxxxxxx"
  restic-s3-secret-key = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}
```

Then apply the changes:

```sh
make external
```

You may want to back up the `external/terraform.tfvars` file to a secure location as well.

## Add backup configuration for volumes

!!! warning

    Restore existing backup data before enabling scheduled backups on a new
    cluster. A backup of an empty volume may overwrite data you need to recover.

Configure each volume with a VolSync `ReplicationSource` that specifies the
source PVC, backup schedule, retention policy, and Restic repository Secret.
The repository Secret needs `RESTIC_REPOSITORY`, `RESTIC_PASSWORD`,
`AWS_ACCESS_KEY_ID`, and `AWS_SECRET_ACCESS_KEY`.
These resources require a running VolSync controller.

## Restore from backup

Stop workloads that write to the destination volume before restoring data.
Configure a VolSync `ReplicationDestination` with the destination PVC, a manual
trigger, and the repository Secret. Wait for the restore to complete and verify
the restored data before restarting workloads or enabling scheduled backups.
