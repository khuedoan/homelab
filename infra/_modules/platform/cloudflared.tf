resource "kubectl_manifest" "cloudflared" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "cloudflared"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://bjw-s-labs.github.io/helm-charts"
        chart          = "app-template"
        targetRevision = "3.1.0"
        helm = {
          releaseName = "cloudflared"
          values = yamlencode({
            configMaps = {
              config = {
                data = {
                  "config.yaml" = <<-EOT
tunnel: homelab
credentials-file: /etc/cloudflared/credentials.json
metrics: 0.0.0.0:2000
no-autoupdate: true
ingress:
  - hostname: www.khuedoan.com
    service: https://public-istio.istio-system
  - hostname: draw.khuedoan.com
    service: https://public-istio.istio-system
  - hostname: chat.khuedoan.com
    service: https://public-istio.istio-system
  - hostname: '*.khuedoan.com'
    service: https://gateway-istio.istio-system
  - service: http_status:404
originRequest:
  originServerName: tunnel.khuedoan.com
EOT

                }
                enabled = true
              }
            }
            controllers = {
              cloudflared = {
                containers = {
                  app = {
                    args = ["tunnel", "--config", "/etc/cloudflared/config.yaml", "run"]
                    image = {
                      repository = "docker.io/cloudflare/cloudflared"
                      tag        = "2024.4.0"
                    }
                    # TODO: Revisit resource sizing after measuring usage.
                    resources = { requests = { cpu = "10m", memory = "32Mi" } }
                  }
                }
              }
            }
            persistence = {
              config = {
                enabled = true
                globalMounts = [{
                  path    = "/etc/cloudflared/config.yaml"
                  subPath = "config.yaml"
                }]
                name = "cloudflared-config"
                type = "configMap"
              }
              credentials = {
                enabled = true
                globalMounts = [{
                  path    = "/etc/cloudflared/credentials.json"
                  subPath = "credentials.json"
                }]
                name = "cloudflared-credentials"
                type = "secret"
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "cloudflared" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
