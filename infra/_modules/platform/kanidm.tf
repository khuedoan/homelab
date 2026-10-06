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
        repoURL        = "https://bjw-s-labs.github.io/helm-charts"
        targetRevision = "5.2.1"
        chart          = "app-template"
        helm = {
          releaseName = "kanidm"
          values = yamlencode({
            configMaps = {
              config = {
                suffix = "config"
                data = {
                  "server.toml" = <<-EOT
bindaddress = "[::]:443"
ldapbindaddress = "[::]:636"
trust_x_forward_for = true
db_path = "/data/kanidm.db"
tls_chain = "/tls/tls.crt"
tls_key = "/tls/tls.key"
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
                  path     = "/tls"
                  readOnly = true
                }]
                # TODO: Automate Kanidm's certificate reload after renewal.
                name = "kanidm-backend-tls"
                type = "secret"
              }
            }
            service = {
              main = {
                controller = "main"
                ports = {
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
            route = {
              main = {
                enabled = true
                kind    = "HTTPRoute"
                parentRefs = [{
                  name        = "gateway"
                  namespace   = "istio-system"
                  sectionName = "https"
                }]
                hostnames = ["auth.khuedoan.com"]
                rules = [{
                  backendRefs = [{
                    identifier = "main"
                    port       = "https"
                  }]
                }]
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

resource "kubectl_manifest" "kanidm_resources" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "kanidm-resources"
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
          releaseName = "kanidm-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            rawResources = {
              certificate = {
                forceRename = "kanidm-backend"
                manifest = {
                  apiVersion = "cert-manager.io/v1"
                  kind       = "Certificate"
                  metadata   = { namespace = "kanidm" }
                  spec = {
                    secretName = "kanidm-backend-tls"
                    issuerRef  = { kind = "ClusterIssuer", name = "letsencrypt-prod" }
                    dnsNames   = ["auth.khuedoan.com"]
                  }
                }
              }
              backend-tls-policy = {
                forceRename = "kanidm"
                manifest = {
                  apiVersion = "gateway.networking.k8s.io/v1"
                  kind       = "BackendTLSPolicy"
                  metadata   = { namespace = "kanidm" }
                  spec = {
                    targetRefs = [{
                      group       = ""
                      kind        = "Service"
                      name        = "kanidm"
                      sectionName = "https"
                    }]
                    validation = {
                      hostname                = "auth.khuedoan.com"
                      wellKnownCACertificates = "System"
                      subjectAltNames         = [{ type = "Hostname", hostname = "auth.khuedoan.com" }]
                    }
                  }
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
