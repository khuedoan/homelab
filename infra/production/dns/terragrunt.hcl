include "root" {
  path = find_in_parent_folders("root.hcl")
}

inputs = {
  domain    = "khuedoan.com"
  zone_name = "khuedoan.com"
  extra_records = {
    # Static site hosting is a commodity at this point,
    # so this isn't really vendor lock-in.
    # I also need the docs outside the lab in case it goes down.
    homelab = {
      type    = "CNAME"
      content = "khuedoan.github.io"
    }
  }
}
