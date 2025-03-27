resource "kubernetes_secret" "external" {
  metadata {
    name      = var.name
    namespace = var.namespace

    annotations = {
      "app.kubernetes.io/managed-by" = "Terraform"
    }
  }

  data = var.data
}
