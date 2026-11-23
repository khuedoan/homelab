resource "kubernetes_manifest" "forgejo_runners" {
  depends_on = [kubernetes_manifest.openbao_resources]

  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"

    metadata = {
      name        = "forgejo-runners"
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
          values = yamlencode({
            global = { createDefaultServiceAccount = false }

            controllers = {
              main = {
                strategy = "Recreate"

                containers = {
                  main = {
                    image = {
                      # TODO switch to nix flake when we have nix cache
                      repository = "nixery.dev/shell/git/go/gcc/just/golangci-lint/terragrunt/opentofu/python3/nodejs/forgejo-runner/dumb-init/nix"
                      tag        = "latest@sha256:caca8c757c0a5246d28987350f1142592166d6d4d4395d8565fb11cc49aaf05c"
                    }
                    command = ["/bin/dumb-init", "--"]
                    args = [
                      "/bin/forgejo-runner", "daemon",
                      "--url=http://forgejo-http.forgejo:3000",
                      "--uuid=$(RUNNER_UUID)",
                      "--token-url=file:/credentials/token",
                      "--label=linux:host",
                    ]
                    env = {
                      HOME = "/workspace"
                      RUNNER_UUID = {
                        valueFrom = {
                          secretKeyRef = { name = "forgejo-runners", key = "uuid" }
                        }
                      }
                    }
                    securityContext = {
                      allowPrivilegeEscalation = false
                      capabilities             = { drop = ["ALL"] }
                      seccompProfile           = { type = "RuntimeDefault" }
                    }
                    resources = {
                      requests = {
                        cpu               = "100m"
                        memory            = "512Mi"
                        ephemeral-storage = "512Mi"
                      }
                      limits = {
                        cpu               = "2"
                        memory            = "2Gi"
                        ephemeral-storage = "8Gi"
                      }
                    }
                  }
                }
              }
            }

            configMaps = {
              nix = {
                # TODO re-enable sandbox?
                data = {
                  "nix.conf" = <<-EOF
                    experimental-features = nix-command flakes
                    sandbox = false
                    build-users-group =
                  EOF
                }
              }
            }

            secrets = {
              runner = {
                stringData = {
                  uuid  = "bao:secret/data/forgejo/linux-runner#uuid"
                  token = "bao:secret/data/forgejo/linux-runner#secret"
                }
              }
            }

            persistence = {
              nix = {
                type         = "configMap"
                identifier   = "nix"
                globalMounts = [{ path = "/etc/nix" }]
              }
              credentials = {
                type = "secret"
                name = "forgejo-runners"
              }
              workspace = {
                type      = "emptyDir"
                sizeLimit = "6Gi"
              }
              tmp = {
                type = "emptyDir"
              }
            }
          })
        }
      }]

      destination = { name = "in-cluster", namespace = "forgejo" }
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
