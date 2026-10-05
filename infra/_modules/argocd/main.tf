terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.3.0"
    }
  }
}

variable "kubeconfig" {
  type = string
}

provider "helm" {
  kubernetes = {
    config_path = var.kubeconfig
  }
}

resource "helm_release" "release" {
  name       = "argocd"
  namespace  = "argocd"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  version    = "10.9.6"
  values = [yamlencode({
    global = { domain = "argocd.khuedoan.com" }
    configs = {
      params = {
        "server.insecure"             = true
        "controller.diff.server.side" = true
      }
      cm = {
        "resource.ignoreResourceUpdatesEnabled"             = true
        "resource.customizations.ignoreResourceUpdates.all" = <<-EOT
          jsonPointers:
            - /status
        EOT
      }
    }
    server = {
      ingress = {
        enabled          = true
        ingressClassName = "nginx"
        annotations      = { "cert-manager.io/cluster-issuer" = "letsencrypt-prod" }
        tls              = true
      }
      metrics = { enabled = true, serviceMonitor = { enabled = false } }
    }
    dex = { enabled = false }
    # TODO: Revisit resource sizing after measuring usage.
    controller = {
      replicas = 1
      resources = {
        requests = { cpu = "100m", memory = "256Mi" }
        limits   = { memory = "1Gi" }
      }
      metrics = { enabled = true, serviceMonitor = { enabled = false } }
    }
    repoServer = {
      replicas = 1
      resources = {
        requests = { cpu = "100m", memory = "128Mi" }
        limits   = { memory = "512Mi" }
      }
      metrics = { enabled = true, serviceMonitor = { enabled = false } }
    }
    redis = { metrics = { enabled = true, serviceMonitor = { enabled = false } } }
  })]
  create_namespace = true
  atomic           = true
  wait             = true
  wait_for_jobs    = true
  timeout          = 600

  lifecycle {
    prevent_destroy = true
  }
}
