locals {
  # nixie records the installation identity after installing. Exclude it from
  # the flake input to avoid drift. K3s autodetects its IP on the first install;
  # a discovered IP is retained here for subsequent rebuilds.
  nixie_state_fields = ["machine_id_hash"]
  inventory = {
    for name, host in var.hosts : name => {
      for key, value in host : key => value
      if !contains(local.nixie_state_fields, key)
    }
  }
}

resource "local_file" "hosts" {
  content         = jsonencode(local.inventory)
  filename        = "${path.module}/hosts.json"
  file_permission = "600"
}

resource "local_file" "cluster" {
  content         = jsonencode(var.cluster)
  filename        = "${path.module}/cluster.json"
  file_permission = "600"
}
