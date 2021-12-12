# External resources

**WIP documents**

> These resources are optional, the homelab still works without them but will lack some features like trusted certificates and offsite backup

Although I try to keep the amount of external resources to the minimum, there's still need for a few of them.
Below is a list of external resources and why we need them.

- Terraform Cloud:
  - Workspace to store the state for external resources
- Cloudflare:
  - DNS
  - DNS-01 challenge for Let's Encrypt
  - Tunnel to public services to the internet without port-forwarding
- Backblaze:
  - B2 storage with S3 compatible API for offsite backup

This layer will deploy resources that require external dependencies using the following provisioners:

- Terraform:
  - Create external resources
  - Add external secrets to namespaces
  - Create an ApplicationSet
- ArgoCD (via the ApplicationSet created by Terraform):
  - Deploy Helm charts in the subdirectories

## Prerequisites

### Create Terraform workspace

TODO

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

Export environment variables for API keys:

```sh
export CLOUDFLARE_API_TOKEN=xxx
export B2_APPLICATION_KEY_ID=xxx
export B2_APPLICATION_KEY=xxx
```

Apply Terraform:

```sh
make
```
