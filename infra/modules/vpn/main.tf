resource "tls_private_key" "ssh" {
  algorithm   = "ECDSA"
  ecdsa_curve = "P256"
}

resource "local_file" "ssh_private_key" {
  content         = tls_private_key.ssh.private_key_pem
  filename        = "${path.module}/private.pem"
  file_permission = "0600"
}

resource "lxd_container" "vpn" {
  name      = "vpn"
  image     = "ubuntu:20.04"
  type      = "virtual-machine" # TODO (feature) Upgrade hosts kernel to use Wireguard in container
  ephemeral = false

  config = {
    "user.access_interface" = "enp5s0"
    "security.secureboot"   = false
    "user.user-data" = templatefile(
      "${path.module}/cloud-init.yaml.tpl",
      {
        ssh_public_key = tls_private_key.ssh.public_key_openssh
      }
    )
  }

  limits = {
    cpu    = "1"
    memory = "256MiB"
  }

  device {
    name = "eth0"
    type = "nic"

    properties = {
      nictype = "macvlan"
      parent  = "eno1" # TODO (optimize) Make parent interface a variable
    }
  }

  device {
    type = "disk"
    name = "root"

    properties = {
      pool = "default"
      path = "/"
      size = "8GiB"
    }
  }
}

module "ansible_provisioner" {
  source      = "../ansible-provisioner"
  directory   = "${path.module}/ansible"
  private_key = local_file.ssh_private_key.filename
  inventory = [
    lxd_container.vpn.ip_address
  ]
}
