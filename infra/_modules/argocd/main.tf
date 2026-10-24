terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.3.0"
    }
  }
}

variable "domain" {
  type = string
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
  version    = "10.10.2"
  values = [yamlencode({
    global = { domain = "argocd.${var.domain}" }
    configs = {
      rbac = { "policy.default" = "role:readonly" }
      params = {
        "server.insecure"              = true
        "controller.diff.server.side"  = true
        "reposerver.parallelism.limit" = 1
      }
      cm = {
        "admin.enabled" = false
        "oidc.config" = yamlencode({
          name                     = "SSO"
          issuer                   = "https://auth.${var.domain}/oauth2/openid/argocd"
          clientID                 = "argocd"
          clientSecret             = "$argocd-sso:client_secret"
          requestedScopes          = ["openid", "profile", "email", "groups"]
          enablePKCEAuthentication = true
        })
        "resource.ignoreResourceUpdatesEnabled"             = true
        "resource.customizations.ignoreResourceUpdates.all" = <<-EOT
          jsonPointers:
            - /status
        EOT
      }
    }
    server = {
      httproute = {
        enabled = true
        parentRefs = [{
          name        = "gateway"
          namespace   = "istio-system"
          sectionName = "https"
        }]
        hostnames = ["argocd.${var.domain}"]
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
