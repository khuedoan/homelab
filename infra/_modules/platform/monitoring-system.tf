resource "kubectl_manifest" "monitoring_system" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "monitoring-system"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = var.repository.url
        targetRevision = var.repository.revision
        path           = "infra/_modules/platform/monitoring-system"
        helm = {
          releaseName = "monitoring-system"
          values = yamlencode({
            kube-prometheus-stack = {
              alertmanager = {
                alertmanagerSpec = {
                  # TODO: Revisit resource sizing after measuring usage.
                  resources = { requests = { cpu = "10m", memory = "64Mi" } }
                  containers = [{
                    args = ["--port=8081", "--config=/config/alertmanager-to-ntfy.jsonnet", "--upstream-host=https://ntfy.sh"]
                    envFrom = [{
                      secretRef = {
                        name = "webhook-transformer"
                      }
                    }]
                    image     = "ghcr.io/khuedoan/webhook-transformer:v0.0.3"
                    name      = "ntfy-relay"
                    resources = { requests = { cpu = "10m", memory = "32Mi" } }
                    volumeMounts = [{
                      mountPath = "/config"
                      name      = "config"
                    }]
                  }]
                  volumes = [{
                    configMap = {
                      name = "webhook-transformer"
                    }
                    name = "config"
                  }]
                }
                config = {
                  receivers = [{
                    name = "ntfy"
                    webhook_configs = [{
                      send_resolved = true
                      url           = "http://localhost:8081"
                    }]
                  }]
                  route = {
                    group_by        = ["namespace"]
                    group_interval  = "5m"
                    group_wait      = "30s"
                    receiver        = "ntfy"
                    repeat_interval = "12h"
                    routes = [{
                      matchers = ["alertname = \"Watchdog\""]
                      receiver = "ntfy"
                    }]
                  }
                }
              }
              grafana = {
                enabled                = false
                forceDeployDashboards  = true
                forceDeployDatasources = true
                additionalDataSources = [{
                  name = "Loki"
                  type = "loki"
                  url  = "http://loki.loki:3100"
                }]
              }
              prometheusOperator = {
                # TODO: Revisit resource sizing after measuring usage.
                resources                = { requests = { cpu = "10m", memory = "64Mi" } }
                prometheusConfigReloader = { resources = { requests = { cpu = "10m", memory = "32Mi" } } }
                admissionWebhooks = {
                  patch = { resources = { requests = { cpu = "10m", memory = "32Mi" } } }
                }
              }
              kube-state-metrics       = { resources = { requests = { cpu = "10m", memory = "32Mi" } } }
              prometheus-node-exporter = { resources = { requests = { cpu = "10m", memory = "32Mi" } } }
              prometheus = {
                prometheusSpec = {
                  ruleSelectorNilUsesHelmValues           = false
                  serviceMonitorSelectorNilUsesHelmValues = false
                  # TODO: Revisit resource sizing and retention after measuring usage.
                  resources = {
                    limits = {
                      memory = "1Gi"
                    }
                    requests = {
                      cpu    = "100m"
                      memory = "256Mi"
                    }
                  }
                  retention = "6h"
                  storageSpec = {
                    volumeClaimTemplate = {
                      spec = {
                        accessModes = ["ReadWriteOnce"]
                        resources = {
                          requests = {
                            storage = "2Gi"
                          }
                        }
                        # TODO: Replace temporary local-path storage with replicated storage.
                        storageClassName = "local-path"
                      }
                    }
                  }
                  podMonitorSelectorNilUsesHelmValues = false
                  probeSelectorNilUsesHelmValues      = false
                }
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "monitoring-system" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
