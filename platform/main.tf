provider "rke" {
  debug = true
}

provider "helm" {
  kubernetes {
    config_path = "${path.root}/kube_config.yaml"
  }
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
      role = [
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
      role = [
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

resource "helm_release" "metallb" {
  name       = "metallb"
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "metallb"
  version    = "1.0.1"

  namespace        = "metallb-system"
  create_namespace = true

  set {
    name  = "configInline"
    value = <<EOT
      address-pools:
      - name: default
        protocol: layer2
        addresses:
        - 192.168.1.150-192.168.1.180
    EOT
  }
}

resource "helm_release" "prometheus" {
  name       = "kube-prometheus-stack"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  version    = "11.1.2"

  namespace        = "monitoring-system"
  create_namespace = true
}
