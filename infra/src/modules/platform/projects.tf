resource "kubernetes_manifest" "system_project" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "AppProject"
    metadata = {
      name      = "system"
      namespace = "argocd"
    }
    spec = {
      sourceRepos                = ["*"]
      destinations               = [{ server = "*", namespace = "*" }]
      clusterResourceWhitelist   = [{ group = "*", kind = "*" }]
      namespaceResourceWhitelist = [{ group = "*", kind = "*" }]
    }
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "kubernetes_manifest" "apps_project" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "AppProject"
    metadata = {
      name      = "apps"
      namespace = "argocd"
    }
    spec = {
      sourceRepos              = ["*"]
      destinations             = [{ name = "in-cluster", namespace = "*" }]
      clusterResourceWhitelist = [{ group = "", kind = "Namespace" }]
      namespaceResourceWhitelist = [
        { group = "", kind = "ConfigMap" },
        { group = "", kind = "PersistentVolumeClaim" },
        { group = "", kind = "Secret" },
        { group = "", kind = "Service" },
        { group = "", kind = "ServiceAccount" },
        { group = "apps", kind = "DaemonSet" },
        { group = "apps", kind = "Deployment" },
        { group = "gateway.networking.k8s.io", kind = "HTTPRoute" },
        { group = "networking.k8s.io", kind = "NetworkPolicy" },
      ]
    }
  }

  lifecycle {
    prevent_destroy = true
  }
}
