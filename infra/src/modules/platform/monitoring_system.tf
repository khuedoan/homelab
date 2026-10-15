locals {
  # K3s shares metrics across components; keep the copies used by the dashboards.
  control_plane_service_monitors = {
    for component, patterns in {
      apiserver = [
        "(container|kubelet|scheduler)_.+",
        "apiserver_(request_body_size_bytes|response_sizes|watch_events_sizes)_.+",
        "apiserver_watch_(list_duration_seconds|cache_read_wait_seconds)_.+",
      ]
      controller_manager = ["(apiserver|etcd|kubelet|container|scheduler)_.+"]
      scheduler          = ["(apiserver|etcd|kubelet|container|workqueue)_.+"]
      kubelet            = ["(apiserver|etcd|scheduler|workqueue|container)_.+"]
      } : component => {
      metricRelabelings = [{
        action       = "drop"
        sourceLabels = ["__name__"]
        regex        = join("|", patterns)
      }]
    }
  }
}

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
        repoURL        = "https://prometheus-community.github.io/helm-charts"
        targetRevision = "92.3.0"
        chart          = "kube-prometheus-stack"
        helm = {
          releaseName = "monitoring-system"
          values = yamlencode({
            alertmanager = {
              alertmanagerSpec = {
                # TODO: Revisit resource sizing after measuring usage.
                resources = { requests = { cpu = "10m", memory = "64Mi" } }
              }
              config = {
                receivers = [{ name = "null" }]
                route = {
                  group_by        = ["namespace"]
                  group_interval  = "5m"
                  group_wait      = "30s"
                  receiver        = "null"
                  repeat_interval = "12h"
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
            kubeEtcd                 = { endpoints = var.control_plane_ips }
            kubeApiServer            = { serviceMonitor = local.control_plane_service_monitors.apiserver }
            kubeControllerManager = {
              endpoints      = var.control_plane_ips
              serviceMonitor = local.control_plane_service_monitors.controller_manager
            }
            kubeScheduler = {
              endpoints      = var.control_plane_ips
              serviceMonitor = local.control_plane_service_monitors.scheduler
            }
            kubelet = {
              serviceMonitor = local.control_plane_service_monitors.kubelet
            }
            prometheus = {
              additionalPodMonitors = [{
                name              = "ztunnel"
                namespaceSelector = { matchNames = ["istio-system"] }
                selector          = { matchLabels = { app = "ztunnel" } }
                podTargetLabels   = ["app"]
                podMetricsEndpoints = [{
                  port = "ztunnel-stats"
                  path = "/stats/prometheus"
                }]
              }]
              additionalServiceMonitors = [{
                name              = "istiod"
                namespaceSelector = { matchNames = ["istio-system"] }
                selector          = { matchLabels = { app = "istiod" } }
                endpoints         = [{ port = "http-monitoring" }]
              }]
              prometheusSpec = {
                ruleSelectorNilUsesHelmValues           = false
                serviceMonitorSelectorNilUsesHelmValues = false
                resources = {
                  limits   = { memory = "1Gi" }
                  requests = { cpu = "100m", memory = "512Mi" }
                }
                scrapeInterval = "60s"
                containers = [{
                  name = "prometheus"
                  env  = [{ name = "GOMEMLIMIT", value = "768MiB" }]
                }]
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
                    }
                  }
                }
                podMonitorSelectorNilUsesHelmValues = false
                probeSelectorNilUsesHelmValues      = false
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
