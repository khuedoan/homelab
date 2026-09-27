resource "kubernetes_secret" "ntfy_auth" {
  metadata {
    name      = "webhook-transformer"
    namespace = "monitoring-system"

    annotations = {
      "app.kubernetes.io/managed-by" = "Terraform"
    }
  }

  data = {
    NTFY_URL   = var.ntfy.url
    NTFY_TOPIC = var.ntfy.topic
  }
}
