resource "kubernetes_manifest" "woodpecker" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "woodpecker"
      namespace = "argocd"
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "ghcr.io/woodpecker-ci/helm"
        chart          = "woodpecker"
        targetRevision = "3.7.6"
        helm = {
          releaseName = "woodpecker"
          values = yamlencode({
            agent = {
              mapAgentSecret             = false
              extraSecretNamesForEnvFrom = ["woodpecker-agent-secret"]
              # TODO: Revisit resource sizing after measuring usage.
              resources = { requests = { cpu = "10m", memory = "64Mi" } }
              env = {
                WOODPECKER_BACKEND_K8S_STORAGE_RWX = false
                WOODPECKER_MAX_WORKFLOWS           = 10
              }
              replicaCount = 2
            }
            server = {
              createAgentSecret          = false
              extraSecretNamesForEnvFrom = ["woodpecker-secret", "woodpecker-agent-secret"]
              resources                  = { requests = { cpu = "50m", memory = "128Mi" } }
              persistentVolume = {
                size = "2Gi"
              }
              env = {
                WOODPECKER_ADMIN                     = "forgejo_admin"
                WOODPECKER_GITEA                     = true
                WOODPECKER_GITEA_URL                 = "https://git.${var.domain}"
                WOODPECKER_HOST                      = "https://ci.${var.domain}"
                WOODPECKER_OPEN                      = true
                WOODPECKER_EXPERT_WEBHOOK_HOST       = "http://woodpecker-server.woodpecker"
                WOODPECKER_DEFAULT_PIPELINE_CONFIGS  = ".workflows/"
                WOODPECKER_AUTHENTICATE_PUBLIC_REPOS = true
                WOODPECKER_PLUGINS_TRUSTED_CLONE     = "docker.io/woodpeckerci/plugin-git,quay.io/woodpeckerci/plugin-git,docker.io/renovate/renovate"
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "woodpecker" }
      ignoreDifferences = local.ignore_differences
      syncPolicy = merge(local.sync_policy, {
        automated = {
          enabled  = !var.recovery_mode
          prune    = true
          selfHeal = true
        }
      })
    }
  }
  lifecycle {
    prevent_destroy = true
  }
}

resource "kubernetes_manifest" "woodpecker_resources" {
  depends_on = [kubernetes_manifest.openbao]
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name        = "woodpecker-resources"
      namespace   = "argocd"
      annotations = { "argocd.argoproj.io/compare-options" = "IncludeMutationWebhook=true" }
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "ghcr.io/bjw-s-labs/helm"
        chart          = "app-template"
        targetRevision = "5.3.0"
        helm = {
          releaseName = "woodpecker-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            secrets = {
              secret = {
                forceRename = "woodpecker-secret"
                stringData = {
                  WOODPECKER_GITEA_CLIENT = "bao:secret/data/woodpecker/forgejo#client_id"
                  WOODPECKER_GITEA_SECRET = "bao:secret/data/woodpecker/forgejo#client_secret"
                }
              }
              agent = {
                forceRename = "woodpecker-agent-secret"
                stringData = {
                  WOODPECKER_AGENT_SECRET = "bao:secret/data/woodpecker/agent#secret"
                }
              }
            }
            rawResources = {
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
      destination       = { name = "in-cluster", namespace = "woodpecker" }
      ignoreDifferences = local.ignore_differences
      syncPolicy = merge(local.sync_policy, {
        automated = {
          enabled  = !var.recovery_mode
          prune    = true
          selfHeal = true
        }
      })
    }
  }
  lifecycle {
    prevent_destroy = true
  }
}
