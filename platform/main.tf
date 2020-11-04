provider "rke" {
    debug = true
  }

resource rke_cluster "cluster" {
  dynamic "nodes" {
    for_each = [
      "192.168.1.21",
      "192.168.1.22",
      "192.168.1.23"
    ]
    content {
      address = nodes.value
      user    = "root"
      role    = [
        "controlplane",
        "etcd",
        "worker"
      ]
      ssh_key = file("~/.ssh/id_rsa")
    }
  }

  dynamic "nodes" {
    for_each = [
      "192.168.1.17"
    ]
    content {
      address = nodes.value
      user    = "root"
      role    = [
        "worker"
      ]
      ssh_key = file("~/.ssh/id_rsa")
    }
  }
}

resource "local_file" "kube_config_yaml" {
  filename = "${path.root}/kube_config.yaml"
  content  = rke_cluster.cluster.kube_config_yaml
}
