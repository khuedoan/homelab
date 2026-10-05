resource "kubectl_manifest" "kanidm" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "kanidm"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = var.repository.url
        targetRevision = var.repository.revision
        path           = "infra/_modules/platform/kanidm"
        helm = {
          releaseName = "kanidm"
          values = yamlencode({
            app-template = {
              configMaps = {
                config = {
                  data = {
                    "server.toml" = <<-EOT
bindaddress = "[::]:443"
ldapbindaddress = "[::]:636"
trust_x_forward_for = true
db_path = "/data/kanidm.db"
tls_chain = "/data/ca.crt"
tls_key = "/data/tls.key"
domain = "auth.khuedoan.com"
origin = "https://auth.khuedoan.com"
EOT

                  }
                  enabled = true
                }
              }
              controllers = {
                main = {
                  containers = {
                    main = {
                      image = {
                        repository = "docker.io/kanidm/server"
                        tag        = "1.3.3"
                      }
                      # TODO: Revisit resource sizing after measuring usage.
                      resources = { requests = { cpu = "50m", memory = "128Mi" } }
                    }
                  }
                  statefulset = {
                    volumeClaimTemplates = [{
                      accessMode = "ReadWriteOnce"
                      globalMounts = [{
                        path = "/data"
                      }]
                      name = "data"
                      size = "1Gi"
                      # TODO: Replace temporary local-path storage with replicated storage.
                      storageClass = "local-path"
                    }]
                  }
                  type = "statefulset"
                }
              }
              persistence = {
                config = {
                  enabled = true
                  globalMounts = [{
                    path    = "/data/server.toml"
                    subPath = "server.toml"
                  }]
                  name = "kanidm-config"
                  type = "configMap"
                }
                tls = {
                  enabled = true
                  globalMounts = [{
                    path    = "/data/ca.crt"
                    subPath = "ca.crt"
                    }, {
                    path    = "/data/tls.key"
                    subPath = "tls.key"
                  }]
                  name = "kanidm-selfsigned-certificate"
                  type = "secret"
                }
              }
              service = {
                main = {
                  ports = {
                    http = {
                      enabled = false
                    }
                    https = {
                      port     = 443
                      protocol = "HTTPS"
                    }
                    ldap = {
                      port     = 636
                      protocol = "TCP"
                    }
                  }
                }
              }
              ingress = {
                main = {
                  annotations = {
                    "cert-manager.io/cluster-issuer"               = "letsencrypt-prod"
                    "nginx.ingress.kubernetes.io/backend-protocol" = "HTTPS"
                  }
                  className = "nginx"
                  enabled   = true
                  hosts = [{
                    host = "auth.khuedoan.com"
                    paths = [{
                      path     = "/"
                      pathType = "Prefix"
                      service = {
                        name = "main"
                        port = "https"
                      }
                    }]
                  }]
                  tls = [{
                    hosts      = ["auth.khuedoan.com"]
                    secretName = "kanidm-tls-certificate"
                  }]
                }
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "kanidm" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
