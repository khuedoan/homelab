provider "rke" {
    debug = true
  }

resource rke_cluster "cluster" {
  dynamic "nodes" {
    for_each = [
      "192.168.1.110",
      "192.168.1.111",
      "192.168.1.112"
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
      "192.168.1.113"
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
