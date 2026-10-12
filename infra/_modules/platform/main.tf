terraform {
  required_providers {
    kubectl = {
      source  = "alekc/kubectl"
      version = "~> 2.0"
    }
  }
}

provider "kubectl" {
  config_path      = var.kubeconfig
  load_config_file = true
}

locals {
  ignore_differences = concat([for kind in ["Deployment", "StatefulSet", "DaemonSet"] : {
    group = "apps"
    kind  = kind
    jqPathExpressions = [
      ".spec.template.metadata.annotations[\"homelab.khuedoan.com/bao-secrets-revision\"]",
    ]
    }], [{
    group = "apps"
    kind  = "StatefulSet"
    jqPathExpressions = [
      ".spec.volumeClaimTemplates[]?.apiVersion",
      ".spec.volumeClaimTemplates[]?.kind",
    ]
    }, {
    kind = "Secret"
    jqPathExpressions = [
      "select(.metadata.labels[\"homelab.khuedoan.com/bao-secret\"] == \"true\") | .data",
      "select(.metadata.labels[\"homelab.khuedoan.com/bao-secret\"] == \"true\") | .stringData",
    ]
  }])
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
    syncOptions = ["CreateNamespace=true", "ServerSideApply=true", "RespectIgnoreDifferences=true"]
    managedNamespaceMetadata = {
      annotations = { "volsync.backube/privileged-movers" = "true" }
    }
  }
}
