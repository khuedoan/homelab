provider "lxd" {
  generate_client_certificates = true
  accept_remote_certificate    = true
}

resource "lxd_network" "kubernetes_network" {
  name = "kubernetes"

  config = {
    "ipv4.address" = "10.150.19.1/24"
    "ipv4.nat"     = "true"
  }
}

resource "lxd_container" "kubernetes_controllers" {
  count     = 3
  name      = "controller-${count.index}"
  image     = "images:ubuntu/18.04"
  ephemeral = false

  config = {
    "boot.autostart" = true
  }

  limits = {
    cpu = 2
  }

  device {
    name = "eth0"
    type = "nic"

    properties = {
      nictype = "bridged"
      parent  = "${lxd_network.kubernetes_network.name}"
    }
  }
}

resource "lxd_container" "kubernetes_workers" {
  count     = 3
  name      = "worker-${count.index}"
  image     = "images:ubuntu/18.04"
  ephemeral = false

  config = {
    "boot.autostart" = true
  }

  limits = {
    cpu = 2
  }

  device {
    name = "eth0"
    type = "nic"

    properties = {
      nictype = "bridged"
      parent  = "${lxd_network.kubernetes_network.name}"
    }
  }
}
