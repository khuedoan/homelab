terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 3.2.1"
    }
  }
}

provider "kubernetes" {
  config_path = var.kubeconfig
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
    retry = {
      limit = -1
      backoff = {
        duration    = "10s"
        factor      = 2
        maxDuration = "3m"
      }
    }
    syncOptions = [
      "CreateNamespace=true",
      "ServerSideApply=true",
      "ApplyOutOfSyncOnly=true",
    ]
  }
}
