include "root" {
  path = find_in_parent_folders("root.hcl")
}

include "cloudflare" {
  path = find_in_parent_folders("cloudflare.hcl")
}

inputs = {
  resource_prefix = "homelab-production"
  domain          = "khuedoan.com"
  zone_name       = "khuedoan.com"
}
