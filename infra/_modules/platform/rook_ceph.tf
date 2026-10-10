resource "kubectl_manifest" "rook_ceph" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "rook-ceph"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://charts.rook.io/release"
        chart          = "rook-ceph"
        targetRevision = "1.13.5"
        helm = {
          releaseName = "rook-ceph"
          values = yamlencode({
            # TODO: Revisit resource sizing after measuring usage.
            resources = { requests = { cpu = "10m", memory = "64Mi" } }
            monitoring = {
              enabled = true
            }
          })
        }
        }, {
        repoURL        = "https://charts.rook.io/release"
        chart          = "rook-ceph-cluster"
        targetRevision = "1.13.5"
        helm = {
          releaseName = "rook-ceph"
          values = yamlencode({
            cephBlockPools = [{
              name = "standard-rwo"
              spec = {
                requireSafeReplicaSize = var.ceph_replica_count > 1
                replicated = {
                  size = var.ceph_replica_count
                }
              }
              storageClass = {
                allowVolumeExpansion = true
                enabled              = true
                isDefault            = false
                name                 = "standard-rwo"
                parameters = {
                  "csi.storage.k8s.io/controller-expand-secret-name"      = "rook-csi-rbd-provisioner"
                  "csi.storage.k8s.io/controller-expand-secret-namespace" = "{{ .Release.Namespace }}"
                  "csi.storage.k8s.io/node-stage-secret-name"             = "rook-csi-rbd-node"
                  "csi.storage.k8s.io/node-stage-secret-namespace"        = "{{ .Release.Namespace }}"
                  "csi.storage.k8s.io/provisioner-secret-name"            = "rook-csi-rbd-provisioner"
                  "csi.storage.k8s.io/provisioner-secret-namespace"       = "{{ .Release.Namespace }}"
                  imageFeatures                                           = "layering,fast-diff,object-map,deep-flatten,exclusive-lock"
                }
              }
            }]
            cephBlockPoolsVolumeSnapshotClass = {
              enabled   = true
              isDefault = true
            }
            cephClusterSpec = {
              dashboard = {
                ssl = false
              }
              logCollector = {
                enabled = false
              }
              mgr = {
                count = min(2, var.ceph_replica_count)
              }
              mon = {
                count = var.ceph_replica_count == 1 ? 1 : 3
              }
              removeOSDsIfOutAndSafeToRemove = true
              # TODO: Revisit resource sizing after measuring usage.
              resources = {
                mgr = {
                  limits = {
                    memory = "1Gi"
                  }
                  requests = {
                    cpu    = "50m"
                    memory = "128Mi"
                  }
                }
                mon = {
                  limits = {
                    memory = "2Gi"
                  }
                  requests = {
                    cpu    = "50m"
                    memory = "100Mi"
                  }
                }
                osd = {
                  limits = {
                    memory = "4Gi"
                  }
                  requests = {
                    cpu    = "50m"
                    memory = "128Mi"
                  }
                }
              }
            }
            cephFileSystemVolumeSnapshotClass = {
              enabled   = true
              isDefault = false
            }
            cephFileSystems = [{
              name = "standard-rwx"
              spec = {
                dataPools = [{
                  name                   = "data0"
                  requireSafeReplicaSize = var.ceph_replica_count > 1
                  replicated = {
                    size = var.ceph_replica_count
                  }
                }]
                metadataPool = {
                  requireSafeReplicaSize = var.ceph_replica_count > 1
                  replicated = {
                    size = var.ceph_replica_count
                  }
                }
                metadataServer = {
                  activeCount       = 1
                  activeStandby     = var.ceph_replica_count > 1
                  priorityClassName = "system-cluster-critical"
                  resources = {
                    limits = {
                      memory = "4Gi"
                    }
                    requests = {
                      cpu    = "50m"
                      memory = "100Mi"
                    }
                  }
                }
              }
              storageClass = {
                allowVolumeExpansion = true
                enabled              = true
                isDefault            = false
                name                 = "standard-rwx"
                parameters = {
                  "csi.storage.k8s.io/controller-expand-secret-name"      = "rook-csi-cephfs-provisioner"
                  "csi.storage.k8s.io/controller-expand-secret-namespace" = "{{ .Release.Namespace }}"
                  "csi.storage.k8s.io/node-stage-secret-name"             = "rook-csi-cephfs-node"
                  "csi.storage.k8s.io/node-stage-secret-namespace"        = "{{ .Release.Namespace }}"
                  "csi.storage.k8s.io/provisioner-secret-name"            = "rook-csi-cephfs-provisioner"
                  "csi.storage.k8s.io/provisioner-secret-namespace"       = "{{ .Release.Namespace }}"
                }
                pool = "data0"
              }
            }]
            cephObjectStores = []
            monitoring = {
              createPrometheusRules = true
              enabled               = true
            }
          })
        }
        }, {
        repoURL        = "https://piraeus.io/helm-charts"
        chart          = "snapshot-controller"
        targetRevision = "2.2.1"
        helm = {
          releaseName = "rook-ceph"
          values = yamlencode({
            # TODO: Revisit resource sizing after measuring usage.
            resources = { requests = { cpu = "10m", memory = "32Mi" } }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "rook-ceph" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
