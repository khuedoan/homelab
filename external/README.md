# External resources

> WIP documents

This layer will deploy resources that require external dependencies using the following provisioners:

- Terraform:
  - Create external resources
  - Add external secrets to namespaces
  - Create an ApplicationSet
- ArgoCD (via the ApplicationSet created by Terraform):
  - Deploy Helm charts in the subdirectories

## Setup

- Create Terraform Cloud workspace
- Run `terraform login`
- Create Cloudflare key at <https://dash.cloudflare.com/profile/api-tokens>
- Create Backblaze key at <https://secure.backblaze.com/app_keys.htm>

```sh
export CLOUDFLARE_API_TOKEN='xxx'
export B2_APPLICATION_KEY_ID='xxx'
export B2_APPLICATION_KEY='xxx'
export KUBE_CONFIG_PATH="$PWD/../metal/kubeconfig.yaml"
```

```sh
terraform apply
```
