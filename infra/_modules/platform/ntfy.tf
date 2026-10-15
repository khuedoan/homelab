resource "kubectl_manifest" "ntfy" {
  depends_on        = [kubectl_manifest.openbao_resources]
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "ntfy"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://bjw-s-labs.github.io/helm-charts"
        targetRevision = "5.3.0"
        chart          = "app-template"
        helm = {
          releaseName = "ntfy"
          values = yamlencode({
            controllers = {
              main = {
                strategy = "Recreate"
                containers = {
                  main = {
                    image = { repository = "docker.io/binwiederhier/ntfy", tag = "v2.30.0" }
                    args  = ["serve"]
                    env = {
                      NTFY_BASE_URL            = "https://ntfy.${var.domain}"
                      NTFY_UPSTREAM_BASE_URL   = "https://ntfy.sh"
                      NTFY_BEHIND_PROXY        = "true"
                      NTFY_CACHE_FILE          = "/var/lib/ntfy/cache.db"
                      NTFY_AUTH_FILE           = "/var/lib/ntfy/auth.db"
                      NTFY_AUTH_DEFAULT_ACCESS = "deny-all"
                      NTFY_AUTH_ACCESS         = "alerts:alerts:rw"
                    }
                    envFrom = [{ secretRef = { name = "ntfy-auth" } }]
                    resources = {
                      requests = { cpu = "10m", memory = "32Mi" }
                      limits   = { memory = "128Mi" }
                    }
                    probes = {
                      liveness  = { enabled = true, custom = true, spec = { httpGet = { path = "/v1/health", port = 80 } } }
                      readiness = { enabled = true, custom = true, spec = { httpGet = { path = "/v1/health", port = 80 } } }
                    }
                  }
                }
              }
            }
            service = { main = { controller = "main", ports = { http = { port = 80 } } } }
            route = {
              main = {
                parentRefs = [{ name = "gateway", namespace = "istio-system", sectionName = "https" }]
                hostnames  = ["ntfy.${var.domain}"]
                rules      = [{ backendRefs = [{ identifier = "main", port = 80 }] }]
              }
            }
            persistence = {
              data = {
                accessMode   = "ReadWriteOnce"
                size         = "1Gi"
                globalMounts = [{ path = "/var/lib/ntfy" }]
              }
            }
            secrets = {
              auth = {
                suffix      = "auth"
                annotations = { "argocd.argoproj.io/sync-wave" = "-1" }
                labels      = { "homelab.khuedoan.com/bao-secret" = "true" }
                stringData = {
                  NTFY_AUTH_USERS  = "alerts:$${bao:secret/data/monitoring-system/ntfy#password_hash}:user"
                  NTFY_AUTH_TOKENS = "alerts:$${bao:secret/data/monitoring-system/ntfy#token}"
                  token            = "bao:secret/data/monitoring-system/ntfy#token"
                }
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "monitoring-system" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
