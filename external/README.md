# External resources

**WIP documents**

> These resources are optional, the homelab still works without them but will lack some features like trusted certificates and offsite backup

Although I try to keep the amount of external resources to the minimum, there's still need for a few of them.
Below is a list of external resources and why we need them (also see some [alternatives](#alternatives) below).

- Terraform Cloud:
  - Workspace to store the state for external resources
- Cloudflare:
  - DNS
  - DNS-01 challenge for Let's Encrypt
  - Tunnel to public services to the internet without port-forwarding
- Backblaze:
  - B2 storage with S3 compatible API for offsite backup

This layer will:

- Create external resources
- Add external secrets to namespaces

## Prerequisites

### Create Terraform workspace

Terraform is stateful, which means it needs somewhere to store it's state. The Terraform Cloud is one option for a state backend with a generous free tier perfect for a homelab.

1. Sign up for a [Terraform Cloud](https://cloud.hashicorp.com/products/terraform)
2. Create a workspace named `homelab-external`, this is the workspace where your homelab state will be stored.
3. Change the "Execution Mode" from "Remote" to "Local". This will ensure your local machine, which can access your lab, is the one executing the terraform plan rather than the cloud runners.

If you decide to use a different terraform backend, you'll need to edit the [external/versions.tf](./versions.tf) file as required.

### Create Cloudflare API token

<https://dash.cloudflare.com/profile/api-tokens>

Terraform API token summary:

```
This API token will affect the below accounts and zones, along with their respective permissions

└── Khue Doan - Argo Tunnel:Edit, Account Settings:Read
    └── khuedoan.com - Zone:Read, DNS:Edit

Client IP Address Filtering

└── Is in - 117.xxx.xxx.xxx, 2402:xxx:xxx:xxx:xxx:xxx:xxx:xxx
```

### Create Backblaze API key

<https://secure.backblaze.com/app_keys.htm>

```
Name of Key: Homelab
Allow access to Bucket(s): All
Type of Access: Read and Write
```

## Deploy

Apply Terraform (you will be prompted to login to Terraform Cloud and enter API keys from the previous steps):

```sh
make
```

## Alternatives

- Terraform Cloud: any other [Terraform backends](https://www.terraform.io/language/settings/backends)
- Cloudflare Tunnel: you can build a small VPS in the cloud and route traffic via it using Wireguard and HAProxy. 
- Backblaze B2: any S3 compatible object storage, such as S3 Glacier, Minio...
