provider "rke" {
    debug = true
  }

resource rke_cluster "cluster" {
  nodes {
    address = "192.168.1.17"
    user    = "root"
    role    = [
      "controlplane",
      "etcd",
      "worker"
    ]
    ssh_key = file("~/.ssh/id_rsa")
  }

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
        "worker",
        "etcd"
      ]
      ssh_key = file("~/.ssh/id_rsa")
    }
  }
}

resource "local_file" "kube_cluster_yaml" {
  filename = "${path.root}/kube_config_cluster.yml"
  content  = rke_cluster.cluster.kube_config_yaml
}
