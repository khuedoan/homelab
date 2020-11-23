resource "kubernetes_ingress" "grafana_ingress" {
  metadata {
    name      = "grafana-ingress"
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

resource "kubernetes_ingress" "vault_ingress" {
  metadata {
    name      = "vault-ingress"
    namespace = helm_release.vault.namespace
  }

  spec {
    rule {
      host = "vault.khuedoan.com"
      http {
        path {
          backend {
            service_name = "vault"
            service_port = 8200
          }
        }
      }
    }
  }
}
