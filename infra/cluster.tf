resource "tls_private_key" "ssh" {
  algorithm   = "ECDSA"
  ecdsa_curve = "P256"
}

resource "local_file" "ssh_private_key" {
  content         = tls_private_key.ssh.private_key_pem
  filename        = "${path.module}/private.pem"
  file_permission = "0600"
}

resource "lxd_profile" "kubenode" {
  name = "kubenode"

  config = {
    "limits.cpu"         = 2
    "limits.memory"      = "4GiB"
    "limits.memory.swap" = false
    # "security.nesting"     = true
    # "security.privileged"  = true
    # "linux.kernel_modules" = "ip_tables,ip6_tables,nf_nat,overlay,br_netfilter"
    # "raw.lxc"              = <<-EOT
    #   lxc.apparmor.profile=unconfined
    #   lxc.cap.drop=
    #   lxc.cgroup.devices.allow=a
    #   lxc.mount.auto=proc:rw sys:rw cgroup:rw
    # EOT
    "user.user-data" = <<-EOT
      #cloud-config
      ssh_authorized_keys:
        - ${tls_private_key.ssh.public_key_openssh}
      disable_root: false
      runcmd:
        - curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
        - add-apt-repository "deb [arch=$(dpkg --print-architecture)] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
        - apt-get update -y
        - apt-get install -y docker-ce docker-ce-cli containerd.io open-iscsi
        - mkdir -p /etc/systemd/system/docker.service.d/
        - printf "[Service]\nMountFlags=shared" > /etc/systemd/system/docker.service.d/mount_flags.conf
        - mount --make-rshared /
        - systemctl enable --now docker
        - systemctl enable --now open-iscsi
    EOT
  }

  # # echo "262144" > /sys/module/nf_conntrack/parameters/hashsize
  # device {
  #   type = "disk"
  #   name = "hashsize"

  #   properties = {
  #     source = "/sys/module/nf_conntrack/parameters/hashsize"
  #     path   = "/sys/module/nf_conntrack/parameters/hashsize"
  #   }
  # }

  # device {
  #   type = "unix-char"
  #   name = "kmsg"

  #   properties = {
  #     source = "/dev/kmsg"
  #     path   = "/dev/kmsg"
  #   }
  # }

  device {
    name = "eth0"
    type = "nic"

    properties = {
      nictype = "macvlan"
      parent  = "eno1"
    }
  }

  device {
    type = "disk"
    name = "root"

    properties = {
      pool = "default"
      path = "/"
      size = "32GiB"
    }
  }
}

# TODO (optimize) DRY master and worker definition
resource "lxd_container" "masters" {
  count     = 3
  name      = "master-${count.index}"
  image     = "ubuntu:20.04"
  # TODO (bug) Use containers instead of virtual machines for Kubernetes nodes https://bugs.launchpad.net/ubuntu/+source/lxc/+bug/1226855
  type      = "virtual-machine"
  ephemeral = false

  profiles = [lxd_profile.kubenode.name]

  config = {
    # TODO (bug) Should be posible to put it in the profile instead lxd_profile.kubenode.config, and make it a variable
    # https://github.com/terraform-lxd/terraform-provider-lxd/blob/master/lxd/resource_lxd_container.go#L473
    "user.access_interface" = "enp5s0"
  }

  provisioner "local-exec" {
    command = "ansible all -u root --private-key ${local_file.ssh_private_key.filename} -i ${self.ip_address}, -m 'wait_for_connection' -a '' && ansible all -u root -i ${self.ip_address}, -m 'wait_for' -a 'path=/var/run/docker.sock'"
    environment = {
      ANSIBLE_HOST_KEY_CHECKING = "False"
    }
  }
}

resource "lxd_container" "workers" {
  count     = 3
  name      = "worker-${count.index}"
  image     = "ubuntu:20.04"
  type      = "virtual-machine"
  ephemeral = false

  profiles = [lxd_profile.kubenode.name]

  config = {
    "user.access_interface" = "enp5s0"
  }

  provisioner "local-exec" {
    command = "ansible all -u root --private-key ${local_file.ssh_private_key.filename} -i ${self.ip_address}, -m 'wait_for_connection' -a '' && ansible all -u root -i ${self.ip_address}, -m 'wait_for' -a 'path=/var/run/docker.sock'"
    environment = {
      ANSIBLE_HOST_KEY_CHECKING = "False"
    }
  }
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
      ssh_key = tls_private_key.ssh.private_key_pem
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
      ssh_key = tls_private_key.ssh.private_key_pem
    }
  }

  ingress {
    provider = "none"
  }

  ignore_docker_version = true
}

resource "local_file" "kube_config_yaml" {
  filename          = "${path.root}/kube_config.yaml"
  sensitive_content = rke_cluster.cluster.kube_config_yaml
  file_permission   = "0600"
}

module "rke_cluster_bootstrap" {
  source      = "./modules/kubernetes-cluster-bootstrap"
  kube_config = rke_cluster.cluster.kube_config_yaml
}
