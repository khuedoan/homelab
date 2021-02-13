resource "lxd_container" "k8s" {
  count     = 1
  name      = "k8s${count.index}"
  image     = "ubuntu:20.04"
  ephemeral = false

  config = {
    "security.privileged"  = true
    "security.nesting"     = true
    "linux.kernel_modules" = "ip_tables,ip6_tables,nf_nat,overlay,br_netfilter"
    "raw.lxc"       = <<-EOT
      lxc.apparmor.profile=unconfined
      lxc.cap.drop=
      lxc.cgroup.devices.allow=a
      lxc.mount.auto=proc:rw sys:rw
    EOT
    "user.user-data"       = <<-EOT
      #cloud-config
      ssh_authorized_keys:
        - ${file("~/.ssh/id_rsa.pub")}
      disable_root: false
      packages:
        - apt-transport-https
        - ca-certificates
        - curl
        - gnupg-agent
        - software-properties-common
      runcmd:
        - curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
        - add-apt-repository "deb [arch=$(dpkg --print-architecture)] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
        - apt-get update -y
        - apt-get install -y docker-ce docker-ce-cli containerd.io
        - mkdir -p /etc/systemd/system/docker.service.d/
        - printf "[Service]\nMountFlags=shared" > /etc/systemd/system/docker.service.d/mount_flags.conf
        - systemctl start docker
        - systemctl enable docker
    EOT
  }

  limits = {
    cpu = 2
  }
}

resource "time_sleep" "wait_cloud_init" {
  depends_on = [lxd_container.k8s]

  create_duration = "240s"
}

resource "rke_cluster" "cluster" {
  dynamic "nodes" {
    for_each = lxd_container.k8s

    content {
      address = nodes.value.ip_address
      user    = "root"
      role = [
        "controlplane",
        "etcd",
        "worker"
      ]
      ssh_key = file("~/.ssh/id_rsa")
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
