provider "rke" {
  debug = true
}

locals {
  hosts   = yamldecode(file("../metal/hosts.yml"))
  user    = local.hosts.metal.vars.ansible_user
  ssh_key = file(local.hosts.metal.vars.ansible_ssh_private_key_file)
}

resource "rke_cluster" "cluster" {
  dynamic "nodes" {
    for_each = [
      local.hosts.metal.hosts.metal0.ansible_host,
      local.hosts.metal.hosts.metal1.ansible_host,
      local.hosts.metal.hosts.metal2.ansible_host
    ]

    content {
      address = nodes.value
      user    = local.user
      role = [
        "controlplane",
        "etcd",
        "worker"
      ]
      ssh_key = local.ssh_key
    }
  }

  dynamic "nodes" {
    for_each = [
      local.hosts.metal.hosts.metal3.ansible_host
    ]

    content {
      address = nodes.value
      user    = local.user
      role = [
        "worker"
      ]
      ssh_key = local.ssh_key
    }
  }

  ingress {
    provider = "none"
  }

  # For CoreOS
  network {
    plugin = "canal"
    options = {
      canal_flex_volume_plugin_dir = "/opt/kubernetes/kubelet-plugins/volume/exec/nodeagent~uds"
      flannel_backend_type = "vxlan"
    }
  }

  services {
    kube_controller {
      extra_args = {
        flex-volume-plugin-dir = "/opt/kubernetes/kubelet-plugins/volume/exec/"
      }
    }
  }
}

resource "local_file" "kube_config_yaml" {
  filename          = "${path.root}/kube_config.yaml"
  sensitive_content = rke_cluster.cluster.kube_config_yaml
  file_permission   = "0600"
}
