# External resources

!!! info

    These resources are optional, the homelab still works without them but will lack some features like trusted certificates and offsite backup

Although I try to keep the amount of external resources to the minimum, there's still need for a few of them.
Below is a list of external resources and why we need them (also see some [alternatives](#alternatives) below).

| Provider        | Resource  | Purpose                                                                                                     |
| --------        | --------  | -------                                                                                                     |
| Terraform Cloud | Workspace | Terraform state backend                                                                                     |
| Cloudflare      | DNS       | DNS and [DNS-01 challenge](https://letsencrypt.org/docs/challenge-types/#dns-01-challenge) for certificates |
| Cloudflare      | Tunnel    | Public services to the internet without port forwarding                                                     |
| ntfy            | Topic     | External notification service to receive alerts                                                             |

## Create credentials

You'll be asked to provide these credentials on first build.

### Create Terraform workspace

Terraform is stateful, which means it needs somewhere to store its state. Terraform Cloud is one option for a state backend with a generous free tier, perfect for a homelab.

1. Sign up for a [Terraform Cloud](https://cloud.hashicorp.com/products/terraform) account
2. Create a workspace named `homelab-external`, this is the workspace where your homelab state will be stored.
3. Change the "Execution Mode" from "Remote" to "Local". This will ensure your local machine, which can access your lab, is the one executing the Terraform plan rather than the cloud runners.

If you decide to use a [different Terraform backend](https://www.terraform.io/language/settings/backends#available-backends), you'll need to edit the `external/versions.tf` file as required.

### Cloudflare

- Buy a domain and [transfer it to Cloudflare](https://developers.cloudflare.com/registrar/get-started/transfer-domain-to-cloudflare) if you haven't already
- Get Cloudflare email and account ID
- Global API key: <https://dash.cloudflare.com/profile/api-tokens>

<!-- TODO switch to API token instead of API key? -->
<!-- Terraform API token summary: -->

<!-- ``` -->
<!-- This API token will affect the below accounts and zones, along with their respective permissions -->

<!-- └── Khue Doan - Argo Tunnel:Edit, Account Settings:Read -->
<!--     └── khuedoan.com - Zone:Read, DNS:Edit -->

<!-- Client IP Address Filtering -->

<!-- └── Is in - 117.xxx.xxx.xxx, 2402:xxx:xxx:xxx:xxx:xxx:xxx:xxx -->
<!-- ``` -->

<!-- ### Create Minio keys -->

<!-- TODO: skip this for now -->

<!-- ### Create AWS API key -->

<!-- TODO: skip this for now -->

### ntfy

- Choose a topic name like <https://ntfy.sh/random_topic_name_here_a8sd7fkjxlkcjasdw33813> (treat it like you password)

## Alternatives

To avoid vendor lock-in, each external provider must have an equivalent alternative that is easy to replace:

- Terraform Cloud:
    - Any other [Terraform backends](https://www.terraform.io/language/settings/backends)
- Cloudflare DNS:
    - Update cert-manager and external-dns to use a different provider
    - [Alternate DNS setup](../../how-to-guides/alternate-dns-setup.md)
- Cloudflare Tunnel:
    - Use port forwarding if it's available
    - Create a small VPS in the cloud and utilize Wireguard to route traffic via it
    - Access everything via VPN
    - See also [awesome tunneling](https://github.com/anderspitman/awesome-tunneling)
- ntfy:
    - [Self-host your own ntfy server](https://docs.ntfy.sh/install)
    - Any other [integration supported by Grafana Alerting](https://grafana.com/docs/grafana/latest/alerting/alerting-rules/manage-contact-points/integrations/#list-of-supported-integrations)
