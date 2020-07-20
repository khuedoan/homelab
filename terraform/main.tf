provider "lxd" {
  generate_client_certificates = true
  accept_remote_certificate    = true
}

resource "lxd_network" "kubernetes_subnetwork" {
  name = "kubernetes"

  config = {
    "ipv4.address" = "10.150.19.1/24"
    "ipv4.nat"     = "true"
  }
}

resource "lxd_container" "test1" {
  count     = 3
  name      = "test-${count.index}"
  image     = "images:ubuntu/18.04"
  ephemeral = false

  config = {
    "boot.autostart" = true
  }

  limits = {
    cpu = 2
  }
}
