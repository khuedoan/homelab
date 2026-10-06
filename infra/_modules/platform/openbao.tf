locals {
  openbao_readers = merge({
    forgejo    = ["forgejo.admin"]
    grafana    = ["kanidm.grafana"]
    renovate   = ["forgejo.renovate"]
    woodpecker = ["forgejo.woodpecker", "woodpecker.agent"]
    zot        = ["registry.admin"]
    }, {
    for name in var.apps : name => [name == "paperless" ? "paperless.admin" : "${name}.auth"]
    if contains(["paperless", "tailscale", "wireguard"], name)
  })
}

resource "kubectl_manifest" "openbao" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "openbao"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "ghcr.io/bank-vaults/helm-charts"
        targetRevision = "1.24.1"
        chart          = "vault-operator"
        helm = {
          releaseName = "openbao"
          values      = yamlencode({ watchNamespace = "openbao" })
        }
        }, {
        repoURL        = "ghcr.io/bank-vaults/helm-charts"
        targetRevision = "0.4.1"
        chart          = "secrets-webhook"
        helm = {
          releaseName = "openbao"
          values = yamlencode({
            autoscaling = { hpa = { enabled = false } }
            resources   = { requests = { cpu = "10m", memory = "32Mi" } }
            env = {
              BAO_ADDR                     = "http://openbao.openbao.svc.cluster.local:8200"
              BAO_PATH                     = "kubernetes"
              BAO_ADDR_ALLOWLIST           = "http://openbao.openbao.svc.cluster.local:8200"
              BAO_ALLOW_OBJECT_SKIP_VERIFY = "false"
            }
            secretsFailurePolicy = "Fail"
            secrets              = { objectSelector = { matchLabels = { "homelab.khuedoan.com/bao-secret" = "true" } } }
            pods                 = { objectSelector = { matchLabels = { "homelab.khuedoan.com/bao-pod" = "true" } } }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "openbao" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}

resource "kubectl_manifest" "openbao_resources" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "openbao-resources"
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
          releaseName = "openbao-resources"
          values = yamlencode({
            global         = { createDefaultServiceAccount = false }
            serviceAccount = { main = { forceRename = "openbao" } }
            rbac = {
              roles = {
                main = {
                  forceRename = "openbao"
                  type        = "Role"
                  rules = [
                    { apiGroups = [""], resources = ["secrets"], verbs = ["get", "create", "update", "patch"] },
                    { apiGroups = [""], resources = ["pods"], verbs = ["get", "update", "patch"] },
                  ]
                }
              }
              bindings = {
                main = {
                  forceRename = "openbao"
                  type        = "RoleBinding"
                  roleRef     = { apiGroup = "rbac.authorization.k8s.io", kind = "Role", name = "openbao" }
                  subjects    = [{ kind = "ServiceAccount", name = "openbao", namespace = "openbao" }]
                }
                auth-delegator = {
                  forceRename = "openbao-auth-delegator"
                  type        = "ClusterRoleBinding"
                  roleRef     = { apiGroup = "rbac.authorization.k8s.io", kind = "ClusterRole", name = "system:auth-delegator" }
                  subjects    = [{ kind = "ServiceAccount", name = "openbao", namespace = "openbao" }]
                }
              }
            }
            persistence = {
              data = {
                forceRename = "openbao-data"
                type        = "persistentVolumeClaim"
                accessMode  = "ReadWriteOnce"
                size        = "2Gi"
                # TODO: Move secrets storage to replicated storage with tested backups.
                storageClass = "local-path"
                annotations = {
                  "argocd.argoproj.io/sync-wave"    = "1"
                  "argocd.argoproj.io/sync-options" = "Delete=false,Prune=false"
                }
              }
            }
            rawResources = {
              vault = {
                forceRename = "openbao"
                manifest = {
                  apiVersion = "vault.banzaicloud.com/v1alpha1"
                  kind       = "Vault"
                  metadata = {
                    namespace   = "openbao"
                    annotations = { "argocd.argoproj.io/sync-wave" = "1" }
                  }
                  spec = {
                    size            = 1
                    image           = "quay.io/openbao/openbao:2.7.1"
                    bankVaultsImage = "ghcr.io/bank-vaults/bank-vaults:v1.33.2"
                    configPath      = "/openbao/config"
                    serviceAccount  = "openbao"
                    statsdDisabled  = true
                    unsealConfig    = { kubernetes = { secretNamespace = "openbao", secretName = "openbao-unseal" } }
                    config = {
                      api_addr      = "http://openbao.openbao.svc.cluster.local:8200"
                      cluster_addr  = "https://openbao.openbao.svc.cluster.local:8201"
                      disable_mlock = true
                      # TODO: Use verified TLS for in-cluster clients.
                      listener = { tcp = { address = "[::]:8200", tls_disable = true } }
                      storage  = { raft = { path = "/openbao/data" } }
                      ui       = true
                    }
                    volumeMounts = [{ name = "data", mountPath = "/openbao/data" }]
                    volumes      = [{ name = "data", persistentVolumeClaim = { claimName = "openbao-data" } }]
                    externalConfig = {
                      auth = [{
                        type   = "kubernetes"
                        path   = "kubernetes"
                        config = { kubernetes_host = "https://kubernetes.default.svc.cluster.local:443" }
                        roles = [for namespace, paths in local.openbao_readers : {
                          name                             = namespace
                          bound_service_account_names      = ["default"]
                          bound_service_account_namespaces = [namespace]
                          audience                         = "https://kubernetes.default.svc"
                          policies                         = [namespace]
                          ttl                              = "5m"
                        }]
                      }]
                      policies = [for namespace, paths in local.openbao_readers : {
                        name = namespace
                        rules = join("\n", [for path in paths : <<-EOT
                          path "secret/data/${path}" {
                            capabilities = ["read"]
                          }
                        EOT
                        ])
                      }]
                      secrets = [{
                        path    = "secret"
                        type    = "kv"
                        options = { version = "2" }
                        configuration = {
                          data = [for record, key in {
                            "forgejo.admin"    = "password"
                            "registry.admin"   = "password"
                            "woodpecker.agent" = "secret"
                            "paperless.admin"  = "PAPERLESS_ADMIN_PASSWORD"
                            } : {
                            name        = record
                            create_only = true
                            options     = { cas = 0 }
                            data        = { (key) = "$${randAlphaNum 32}" }
                          }]
                        }
                      }]
                    }
                  }
                }
              }
              route = {
                forceRename = "openbao"
                manifest = {
                  apiVersion = "gateway.networking.k8s.io/v1"
                  kind       = "HTTPRoute"
                  metadata   = { namespace = "openbao" }
                  spec = {
                    parentRefs = [{ name = "gateway", namespace = "istio-system", sectionName = "https" }]
                    hostnames  = ["openbao.khuedoan.com"]
                    rules      = [{ backendRefs = [{ name = "openbao", port = 8200 }] }]
                  }
                }
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "openbao" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
