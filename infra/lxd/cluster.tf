resource "lxd_profile" "kubenode" {
  name = "kubenode"

  config = {
    "limits.cpu" = 2
    "limits.memory.swap" = false
    "user.access_interface" = "eth0"
    "security.nesting"     = true
    "security.privileged"  = true
    "linux.kernel_modules" = "ip_tables,ip6_tables,nf_nat,overlay,br_netfilter"
    "raw.lxc"       = <<-EOT
      lxc.apparmor.profile=unconfined
      lxc.cap.drop=
      lxc.cgroup.devices.allow=a
      lxc.mount.auto=proc:rw sys:rw cgroup:rw
    EOT
    "user.user-data"       = <<-EOT
      #cloud-config
      ssh_authorized_keys:
        - ${file(var.ssh_public_key)}
      disable_root: false
      runcmd:
        - curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
        - add-apt-repository "deb [arch=$(dpkg --print-architecture)] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
        - apt-get update -y
        - apt-get install -y docker-ce docker-ce-cli containerd.io
        - mkdir -p /etc/systemd/system/docker.service.d/
        - printf "[Service]\nMountFlags=shared" > /etc/systemd/system/docker.service.d/mount_flags.conf
        - mount --make-rshared /
        - systemctl start docker
        - systemctl enable docker
    EOT
  }

  # echo "262144" > /sys/module/nf_conntrack/parameters/hashsize
  device {
    type = "disk"
    name = "hashsize"

    properties = {
      source = "/sys/module/nf_conntrack/parameters/hashsize"
      path = "/sys/module/nf_conntrack/parameters/hashsize"
    }
  }

  device {
    type = "unix-char"
    name = "kmsg"

    properties = {
      source = "/dev/kmsg"
      path = "/dev/kmsg"
    }
  }

  device {
    name = "eth0"
    type = "nic"

    properties = {
      network = "lxdfan0"
    }
  }

  device {
    type = "disk"
    name = "root"

    properties = {
      pool = "local"
      path = "/"
    }
  }
}

resource "lxd_container" "masters" {
  count     = 1
  name      = "master-${count.index}"
  image     = "ubuntu:20.04"
  ephemeral = false

  profiles = [lxd_profile.kubenode.name]
}

resource "lxd_container" "workers" {
  count     = 1
  name      = "worker-${count.index}"
  image     = "ubuntu:20.04"
  ephemeral = false

  profiles = [lxd_profile.kubenode.name]
}

resource "time_sleep" "wait_cloud_init" {
  depends_on = [
    lxd_container.masters,
    lxd_container.workers
  ]

  create_duration = "5m"
}

resource "rke_cluster" "cluster" {
  dynamic "nodes" {
    for_each = lxd_container.masters

    content {
      address = nodes.value.ip_address
      user    = "root"
      role = [
        "controlplane",
        "etcd"
      ]
      ssh_key = file(var.ssh_private_key)
    }
  }

  dynamic "nodes" {
    for_each = lxd_container.workers

    content {
      address = nodes.value.ip_address
      user    = "root"
      role = [
        "worker"
      ]
      ssh_key = file(var.ssh_private_key)
    }
  }

  ingress {
    provider = "none"
  }

  ignore_docker_version = true

  depends_on = [time_sleep.wait_cloud_init]
}

resource "local_file" "kube_config_yaml" {
  filename = "${path.root}/kube_config.yaml"
  content  = rke_cluster.cluster.kube_config_yaml
}
