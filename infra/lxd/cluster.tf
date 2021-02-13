resource "lxd_container" "k8s" {
  count     = 1
  name      = "k8s${count.index}"
  image     = "ubuntu:20.04"
  ephemeral = false

  config = {
    security.privileged  = true
    security.nesting     = true
    limits.memory.swap   = false
    limits.cpus          = 1
    linux.kernel_modules = "ip_tables,ip6_tables,nf_nat,overlay,br_netfilter"
    raw.lxc       = <<-EOT
      lxc.apparmor.profile=unconfined
      lxc.cap.drop=
      lxc.cgroup.devices.allow=a
      lxc.mount.auto=proc:rw sys:rw
    EOT
    user.user-data       = <<-EOT
      #cloud-config
      ssh_authorized_keys:
        - ${file("~/.ssh/id_rsa.pub")}
      disable_root: false
      runcmd:
        - apt-get update
        - apt-get install -y iptables git telnet vim software-properties-common resolvconf linux-headers-$(uname -r)
        - echo "nameserver 1.1.1.1" > /etc/resolvconf/resolv.conf.d/tail
        - echo "RateLimitIntervalSec=0" >> /etc/systemd/journald.conf
        - echo "RateLimitBurst=0" >> /etc/systemd/journald.conf
        - systemctl restart systemd-journald.service
        - systemctl start resolvconf
        - /opt/rke2/run_rke2.sh
      write_files:
      - path: /opt/rke2/run_rke2.sh
        permissions: "0755"
        owner: root:root
        content: |
          #!/bin/bash
          curl -fsSL https://raw.githubusercontent.com/rancher/rke2/master/install.sh --output install.sh
          chmod u+x install.sh
          INSTALL_RKE2_METHOD='tar' INSTALL_RKE2_TYPE=server INSTALL_RKE2_VERSION=v1.19.7+rke2r1 ./install.sh
          systemctl enable rke2-server
          systemctl start rke2-server
    EOT
  }

  limits = {
    cpu = 2
  }
}

# resource "rke_cluster" "cluster" {
#   dynamic "nodes" {
#     for_each = lxd_container.k8s

#     content {
#       address = nodes.value.ip_address
#       user    = "root"
#       role = [
#         "controlplane",
#         "etcd",
#         "worker"
#       ]
#       ssh_key = file("~/.ssh/id_rsa")
#     }
#   }

#   ingress {
#     provider = "none"
#   }
# }

# resource "local_file" "kube_config_yaml" {
#   filename = "${path.root}/kube_config.yaml"
#   content  = rke_cluster.cluster.kube_config_yaml
# }
