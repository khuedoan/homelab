resource "kubernetes_secret" "ntfy_auth" {
  metadata {
    name      = "webhook-transformer"
    namespace = "monitoring-system"

    annotations = {
      "app.kubernetes.io/managed-by" = "Terraform"
    }
  }

  data = {
    NTFY_URL   = var.auth.url
    NTFY_TOPIC = var.auth.topic
  }
}
