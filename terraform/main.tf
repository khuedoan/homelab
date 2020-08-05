provider "lxd" {
  generate_client_certificates = true
  accept_remote_certificate    = true
}

resource "lxd_network" "kubernetes" {
  name = "kubernetes"

  config = {
    "ipv4.address" = "10.150.19.1/24"
    "ipv4.nat"     = "true"
  }
}

resource "lxd_profile" "kubernetes" {
  name = "kubernetes"

  config = {
    "limits.cpu"           = 2
    "linux.kernel_modules" = "ip_tables,ip6_tables,netlink_diag,nf_nat,overlay"
    "raw.lxc"              = "lxc.apparmor.profile=unconfined\nlxc.cap.drop= \nlxc.cgroup.devices.allow=a\nlxc.mount.auto=proc:rw sys:rw"
    "security.privileged"  = true
    "security.nesting"     = true
  }

  device {
    name = "eth0"
    type = "nic"

    properties = {
      nictype = "routed"
      parent  = "enp0s3"
    }
  }

  device {
    name = "root"
    type = "disk"

    properties = {
      path = "/"
      pool = "default"
    }
  }
}

resource "lxd_container" "kubernetes_controllers" {
  count     = 1
  name      = "controller-${count.index}"
  image     = "images:ubuntu/18.04"
  ephemeral = false
  profiles  = [ lxd_profile.kubernetes.name ]

  provisioner "local-exec" {
  command = <<EXEC
    lxc exec ${self.name} -- bash -xe -c '
      apt-get install -y curl
      curl -sfL https://get.k3s.io | sh -
      apt-get install -y linux-image-$(uname -r)
      mknod /dev/kmsg c 1 11
    '
    EXEC
  }
}

resource "lxd_container" "kubernetes_workers" {
  count     = 0
  name      = "worker-${count.index}"
  image     = "images:ubuntu/18.04"
  ephemeral = false
  profiles  = [ lxd_profile.kubernetes.name ]
}
