include "root" {
  path = find_in_parent_folders("root.hcl")
}

dependency "dns" {
  config_path = "../dns"
}

inputs = {
  domain  = "khuedoan.com"
  zone_id = dependency.dns.outputs.zone_id
}
