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
    "raw.lxc"              = "lxc.apparmor.profile=unconfined\nlxc.cap.drop= \nlxc.cgroup.devices.allow=a\nlxc.mount.auto=proc:rw\nsys:rw"
    "security.privileged"  = "true"
    "security.nesting"     = "true"
  }

  device {
    name = "eth0"
    type = "nic"

    properties = {
      nictype = "bridged"
      parent  = "lxdbr0"
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
  count     = 3
  name      = "controller-${count.index}"
  image     = "images:ubuntu/18.04"
  ephemeral = false
  profiles  = [ lxd_profile.kubernetes.name ]
}

resource "lxd_container" "kubernetes_workers" {
  count     = 3
  name      = "worker-${count.index}"
  image     = "images:ubuntu/18.04"
  ephemeral = false
  profiles  = [ lxd_profile.kubernetes.name ]
}
