terraform {
  required_providers {
    kubectl = {
      source  = "alekc/kubectl"
      version = "~> 3.0"
    }
  }
}

provider "kubectl" {
  config_path      = var.kubeconfig
  load_config_file = true
}

locals {
  ignore_differences = [{
    group = "apps"
    kind  = "StatefulSet"
    jqPathExpressions = [
      ".spec.volumeClaimTemplates[]?.apiVersion",
      ".spec.volumeClaimTemplates[]?.kind",
    ]
  }]
  sync_policy = {
    automated = {
      prune    = true
      selfHeal = true
    }
    retry = {
      limit = -1
      backoff = {
        duration    = "10s"
        factor      = 2
        maxDuration = "3m"
      }
    }
    syncOptions = ["CreateNamespace=true", "ServerSideApply=true"]
  }
}
