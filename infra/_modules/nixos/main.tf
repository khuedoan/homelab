locals {
  # nixie annotates the tracked hosts.json with observed runtime state (ip,
  # machine_id_hash) after installing. That state is not part of the flake
  # input, so drop it before writing the flake's copy to avoid tofu seeing
  # drift every time nixie records a host.
  nixie_state_fields = ["ip", "machine_id_hash"]
  inventory = {
    for name, host in var.hosts : name => {
      for key, value in host : key => value
      if !contains(local.nixie_state_fields, key)
    }
  }
}

resource "local_file" "hosts" {
  content         = jsonencode(local.inventory)
  filename        = "${var.flake}/hosts.json"
  file_permission = "600"
}
