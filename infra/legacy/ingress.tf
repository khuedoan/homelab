resource "kubernetes_ingress" "grafana_ingress" {
  metadata {
    name      = "grafana-ingress"
    namespace = helm_release.prometheus.namespace
  }

  spec {
    rule {
      host = "grafana.${local.domain}"
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

resource "kubernetes_ingress" "longhorn" {
  metadata {
    name      = "longhorn-ingress"
    namespace = helm_release.longhorn.namespace
  }

  spec {
    rule {
      host = "longhorn.${local.domain}"
      http {
        path {
          backend {
            service_name = "longhorn-frontend"
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
      host = "vault.${local.domain}"
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
