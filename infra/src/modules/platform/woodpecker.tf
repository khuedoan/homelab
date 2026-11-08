resource "kubectl_manifest" "woodpecker" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "woodpecker"
      namespace = "argocd"
    }
    spec = {
      project = kubectl_manifest.system_project.name
      sources = [{
        repoURL        = "https://woodpecker-ci.org"
        targetRevision = "3.7.5"
        chart          = "woodpecker"
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
  })
  lifecycle {
    prevent_destroy = true
  }
}

resource "kubectl_manifest" "woodpecker_resources" {
  depends_on        = [kubectl_manifest.openbao]
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name        = "woodpecker-resources"
      namespace   = "argocd"
      annotations = { "argocd.argoproj.io/compare-options" = "IncludeMutationWebhook=true" }
    }
    spec = {
      project = kubectl_manifest.system_project.name
      sources = [{
        repoURL        = "https://bjw-s-labs.github.io/helm-charts"
        targetRevision = "5.2.1"
        chart          = "app-template"
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
  })
  lifecycle {
    prevent_destroy = true
  }
}
