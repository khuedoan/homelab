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
  ephemeral = false

  config = {
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
      parent  = "eno1" # TODO make parent interface a variable
    }
  }
}

resource "null_resource" "ansible" {
  triggers = {
    ansible_hash = md5(join("", [for f in fileset("${path.module}/ansible/", "**") : file("${path.module}/ansible/${f}")]))
  }

  provisioner "local-exec" {
    command = "ansible-playbook -u ubuntu -i ${lxd_container.vpn.ip_address}, --private-key ${local_file.ssh_private_key.filename} ${path.module}/ansible/main.yml"

    environment = {
      ANSIBLE_HOST_KEY_CHECKING = "False"
    }
  }
}
