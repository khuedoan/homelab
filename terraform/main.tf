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
  count     = 1
  name      = "controller-${count.index}"
  image     = "images:ubuntu/18.04"
  ephemeral = false
  profiles  = [ lxd_profile.kubernetes.name ]

  provisioner "local-exec" {
  command = <<EXEC
    lxc exec ${self.name} -- bash -xe -c '
      apt-get install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common
      curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
      add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
      apt-get update
      apt-get install -y docker-ce docker-ce-cli containerd.io
      systemctl enable --now docker
      apt-get update && apt-get install -y apt-transport-https curl
      curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
      echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list
      apt-get update
      apt-get install -y kubelet kubeadm kubectl
      apt-mark hold kubelet kubeadm kubectl
      systemctl daemon-reload
      systemctl enable --now kubelet
      apt-get install -y linux-image-$(uname -r)
      kubeadm init
    '
    EXEC
  }
}

resource "lxd_container" "kubernetes_workers" {
  count     = 1
  name      = "worker-${count.index}"
  image     = "images:ubuntu/18.04"
  ephemeral = false
  profiles  = [ lxd_profile.kubernetes.name ]
}
