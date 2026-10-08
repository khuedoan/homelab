resource "kubectl_manifest" "woodpecker" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "woodpecker"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://woodpecker-ci.org"
        targetRevision = "3.7.4"
        chart          = "woodpecker"
        helm = {
          releaseName = "woodpecker"
          values = yamlencode({
            agent = {
              # TODO: Revisit resource sizing after measuring usage.
              resources = { requests = { cpu = "10m", memory = "64Mi" } }
              env = {
                WOODPECKER_BACKEND_K8S_STORAGE_RWX = false
                WOODPECKER_MAX_WORKFLOWS           = 10
              }
              persistence = {
                # TODO: Replace temporary local-path storage with replicated storage.
                storageClass = "local-path"
              }
              replicaCount = 2
            }
            server = {
              resources = { requests = { cpu = "50m", memory = "128Mi" } }
              persistentVolume = {
                size = "2Gi"
                # TODO: Replace temporary local-path storage with replicated storage.
                storageClass = "local-path"
              }
              env = {
                WOODPECKER_ADMIN               = "forgejo_admin"
                WOODPECKER_GITEA               = true
                WOODPECKER_GITEA_URL           = "https://git.${var.domain}"
                WOODPECKER_HOST                = "https://ci.${var.domain}"
                WOODPECKER_OPEN                = true
                WOODPECKER_EXPERT_WEBHOOK_HOST = "http://woodpecker-server.woodpecker"
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "woodpecker" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}

resource "kubectl_manifest" "woodpecker_resources" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "woodpecker-resources"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://bjw-s-labs.github.io/helm-charts"
        targetRevision = "5.2.1"
        chart          = "app-template"
        helm = {
          releaseName = "woodpecker-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            rawResources = {
              secret = {
                forceRename = "woodpecker-secret"
                manifest = {
                  apiVersion = "v1"
                  kind       = "Secret"
                  metadata = {
                    namespace = "woodpecker"
                    labels    = { "homelab.khuedoan.com/bao-secret" = "true" }
                    annotations = {
                      "secrets-webhook.security.bank-vaults.io/provider"           = "bao"
                      "secrets-webhook.security.bank-vaults.io/bao-role"           = "woodpecker"
                      "secrets-webhook.security.bank-vaults.io/bao-path"           = "kubernetes"
                      "secrets-webhook.security.bank-vaults.io/bao-serviceaccount" = "default"
                    }
                  }
                  data = {
                    WOODPECKER_GITEA_CLIENT = base64encode("bao:secret/data/forgejo.woodpecker#client_id")
                    WOODPECKER_GITEA_SECRET = base64encode("bao:secret/data/forgejo.woodpecker#client_secret")
                    WOODPECKER_AGENT_SECRET = base64encode("bao:secret/data/woodpecker.agent#secret")
                  }
                }
              }
              route = {
                forceRename = "woodpecker"
                manifest = {
                  apiVersion = "gateway.networking.k8s.io/v1"
                  kind       = "HTTPRoute"
                  metadata   = { namespace = "woodpecker" }
                  spec = {
                    parentRefs = [{
                      name        = "gateway"
                      namespace   = "istio-system"
                      sectionName = "https"
                    }]
                    hostnames = ["ci.${var.domain}"]
                    rules     = [{ backendRefs = [{ name = "woodpecker-server", port = 80 }] }]
                  }
                }
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "woodpecker" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
