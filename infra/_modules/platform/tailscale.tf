resource "kubectl_manifest" "tailscale" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "tailscale"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://bjw-s-labs.github.io/helm-charts"
        chart          = "app-template"
        targetRevision = "5.3.0"
        helm = {
          releaseName = "tailscale"
          values = yamlencode({
            controllers = {
              main = {
                strategy       = "Recreate"
                serviceAccount = { identifier = "main" }
                containers = {
                  main = {
                    image = { repository = "ghcr.io/tailscale/tailscale", tag = "v1.104.1" }
                    env = {
                      TS_HOSTNAME            = "${var.resource_prefix}-router"
                      TS_KUBE_SECRET         = "tailscale-state"
                      TS_ROUTES              = var.load_balancer_cidr
                      TS_USERSPACE           = "false"
                      TS_DEBUG_FIREWALL_MODE = "auto"
                      TS_AUTH_ONCE           = "true"
                      TS_BOOT_TIMEOUT        = "24h"
                      TS_ACCEPT_DNS          = "false"
                      TS_ENABLE_HEALTH_CHECK = "true"
                    }
                    securityContext = { privileged = true }
                    resources = {
                      requests = { cpu = "10m", memory = "64Mi" }
                      limits   = { memory = "256Mi" }
                    }
                    probes = {
                      readiness = {
                        enabled = true
                        custom  = true
                        spec = {
                          httpGet       = { path = "/healthz", port = 9002 }
                          periodSeconds = 10
                        }
                      }
                    }
                  }
                }
              }
            }
            defaultPodOptions = {
              automountServiceAccountToken = true
              labels                       = { "istio.io/dataplane-mode" = "none" }
            }
            serviceAccount = { main = { enabled = true } }
            rbac = {
              roles = {
                state = {
                  type = "Role"
                  rules = [{
                    apiGroups = [""]
                    resources = ["secrets"]
                    verbs     = ["create"]
                    }, {
                    apiGroups     = [""]
                    resources     = ["secrets"]
                    resourceNames = ["tailscale-state"]
                    verbs         = ["get", "update", "patch"]
                  }]
                }
              }
              bindings = {
                state = {
                  type     = "RoleBinding"
                  roleRef  = { identifier = "state" }
                  subjects = [{ identifier = "main" }]
                }
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "tailscale" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
