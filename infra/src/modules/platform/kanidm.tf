resource "kubernetes_manifest" "kanidm" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "kanidm"
      namespace = "argocd"
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "ghcr.io/bjw-s-labs/helm"
        chart          = "app-template"
        targetRevision = "5.3.0"
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
domain = "${var.domain}"
origin = "https://auth.${var.domain}"
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
                      tag        = "1.11.2"
                    }
                    probes = {
                      readiness = {
                        enabled = true
                        type    = "HTTPS"
                        path    = "/status"
                        port    = 443
                      }
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
                hostnames = ["auth.${var.domain}"]
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
      destination       = { name = "in-cluster", namespace = "kanidm" }
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

resource "kubernetes_manifest" "kanidm_resources" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "kanidm-resources"
      namespace = "argocd"
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "ghcr.io/bjw-s-labs/helm"
        chart          = "app-template"
        targetRevision = "5.3.0"
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
                    issuerRef  = { kind = "ClusterIssuer", name = "acme" }
                    dnsNames   = ["auth.${var.domain}"]
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
                      hostname                = "auth.${var.domain}"
                      wellKnownCACertificates = "System"
                      subjectAltNames         = [{ type = "Hostname", hostname = "auth.${var.domain}" }]
                    }
                  }
                }
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "kanidm" }
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
