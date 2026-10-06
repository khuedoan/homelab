resource "kubectl_manifest" "forgejo" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "forgejo"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "code.forgejo.org/forgejo-helm"
        targetRevision = "17.1.7"
        chart          = "forgejo"
        helm = {
          releaseName = "forgejo"
          values = yamlencode({
            fullnameOverride = "forgejo"
            image            = { tag = "16.0.5" }
            # TODO: Revisit resource sizing after measuring usage.
            resources = { requests = { cpu = "50m", memory = "128Mi" } }
            gitea = {
              admin = { existingSecret = "forgejo-admin" }
              config = {
                database = { DB_TYPE = "sqlite3" }
                actions  = { ENABLED = false }
                repository = {
                  DEFAULT_BRANCH      = "master"
                  DISABLED_REPO_UNITS = "repo.wiki,repo.projects,repo.packages"
                  DISABLE_STARS       = true
                }
                "service.explore" = { DISABLE_USERS_PAGE = true }
                server = {
                  DOMAIN       = "git.khuedoan.com"
                  ROOT_URL     = "https://git.khuedoan.com"
                  LANDING_PAGE = "explore"
                  OFFLINE_MODE = true
                }
                webhook = { ALLOWED_HOST_LIST = "private" }
              }
            }
            persistence = {
              claimName = "forgejo-data"
              size      = "2Gi"
              # TODO: Replace temporary local-path storage with replicated storage.
              storageClass = "local-path"
            }
            httpRoute = {
              enabled = true
              parentRefs = [{
                name        = "gateway"
                namespace   = "istio-system"
                sectionName = "https"
              }]
              hostnames = ["git.khuedoan.com"]
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "forgejo" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}

resource "kubectl_manifest" "forgejo_resources" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "forgejo-resources"
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
          releaseName = "forgejo-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            rawResources = {
              admin = {
                forceRename = "forgejo-admin"
                manifest = {
                  apiVersion = "v1"
                  kind       = "Secret"
                  metadata = {
                    labels = { "homelab.khuedoan.com/bao-secret" = "true" }
                    annotations = {
                      "argocd.argoproj.io/sync-wave"                               = "-1"
                      "secrets-webhook.security.bank-vaults.io/provider"           = "bao"
                      "secrets-webhook.security.bank-vaults.io/bao-role"           = "forgejo"
                      "secrets-webhook.security.bank-vaults.io/bao-path"           = "kubernetes"
                      "secrets-webhook.security.bank-vaults.io/bao-serviceaccount" = "default"
                    }
                  }
                  data = {
                    username = base64encode("forgejo_admin")
                    password = base64encode("bao:secret/data/forgejo.admin#password")
                  }
                }
              }
              source = {
                forceRename = "forgejo-config-source"
                manifest = {
                  apiVersion = "v1"
                  kind       = "ConfigMap"
                  data = merge(
                    { for filename in fileset("${path.module}/forgejo/files/config", "*") : filename => file("${path.module}/forgejo/files/config/${filename}") },
                    {
                      "config.yaml" = yamlencode({
                        organizations = [{
                          name        = "ops"
                          description = "Operations"
                        }]
                        repositories = [
                          {
                            name    = "homelab"
                            owner   = "ops"
                            private = false
                            migrate = { source = "https://github.com/khuedoan/homelab", mirror = false }
                          },
                          {
                            name    = "blog"
                            owner   = "khuedoan"
                            migrate = { source = "https://github.com/khuedoan/blog", mirror = true }
                          },
                          {
                            name    = "backstage"
                            owner   = "khuedoan"
                            migrate = { source = "https://github.com/khuedoan/backstage", mirror = true }
                          },
                        ]
                      })
                    }
                  )
                }
              }
              configuration = {
                forceRename = "forgejo-config"
                manifest = {
                  apiVersion = "batch/v1"
                  kind       = "Job"
                  metadata = {
                    annotations = {
                      "argocd.argoproj.io/hook"               = "PostSync"
                      "argocd.argoproj.io/hook-delete-policy" = "BeforeHookCreation,HookSucceeded"
                    }
                  }
                  spec = {
                    backoffLimit = 10
                    template = {
                      spec = {
                        restartPolicy = "Never"
                        containers = [{
                          name  = "apply"
                          image = "golang:1.26-alpine"
                          env = [
                            { name = "FORGEJO_HOST", value = "http://forgejo-http:3000" },
                            { name = "FORGEJO_USER", valueFrom = { secretKeyRef = { name = "forgejo-admin", key = "username" } } },
                            { name = "FORGEJO_PASSWORD", valueFrom = { secretKeyRef = { name = "forgejo-admin", key = "password" } } },
                          ]
                          workingDir   = "/go/src/forgejo-config"
                          command      = ["sh", "-c"]
                          args         = ["go run ."]
                          volumeMounts = [{ name = "source", mountPath = "/go/src/forgejo-config" }]
                        }]
                        volumes = [{ name = "source", configMap = { name = "forgejo-config-source" } }]
                      }
                    }
                  }
                }
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "forgejo" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
