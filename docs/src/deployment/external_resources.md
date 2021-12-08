# External resources

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

The code for these resources is located at `./external`.

## Setup

### Create Terraform workspace

TODO

### Get Cloudflare API key

TODO

### Get Backblaze API key

TODO

## Deploy

Export environment variables for API keys:

```sh
export CLOUDFLARE_API_KEY=xxx
export B2_APPLICATION_KEY_ID=xxx
export B2_APPLICATION_KEY=xxx
```

Apply Terraform:

```sh
make external
```
