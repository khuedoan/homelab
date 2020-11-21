provider "rke" {
  debug = true
}

provider "kubernetes" {
  config_path = "${path.root}/kube_config.yaml"
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

  ingress {
    provider = "none"
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

resource "helm_release" "nginx" {
  name       = "ingress-nginx"
  repository = "https://kubernetes.github.io/ingress-nginx"
  chart      = "ingress-nginx"
  version    = "3.10.1"

  namespace        = "ingress-nginx"
  create_namespace = true
}

resource "helm_release" "prometheus" {
  name       = "kube-prometheus-stack"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  version    = "11.1.2"

  namespace        = "monitoring-system"
  create_namespace = true
}

resource "helm_release" "longhorn" {
  name       = "longhorn"
  repository = "https://charts.longhorn.io"
  chart      = "longhorn"
  version    = "1.0.2"

  namespace        = "longhorn-system"
  create_namespace = true
}

resource "kubernetes_ingress" "grafana_ingress" {
  metadata {
    name = "grafana-ingress"
    namespace = helm_release.prometheus.namespace
  }

  spec {
    rule {
      host = "grafana.khuedoan.com"
      http {
        path {
          backend {
            service_name = "kube-prometheus-stack-grafana"
            service_port = 80
          }
         }
       }
     }
   }
 }

resource "helm_release" "vault" {
  name       = "vault"
  repository = "https://helm.releases.hashicorp.com"
  chart      = "vault"
  version    = "0.8.0"

  namespace        = "vault"
  create_namespace = true
}
