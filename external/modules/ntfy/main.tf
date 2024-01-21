resource "kubernetes_secret" "ntfy_auth" {
  metadata {
    name      = "ntfy.auth"
    namespace = "global-secrets"

    annotations = {
      "app.kubernetes.io/managed-by" = "Terraform"
    }
  }

  data = {
    url      = var.auth.url
    username = var.auth.username
    password = var.auth.password
  }
}
