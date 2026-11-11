resource "kubernetes_manifest" "rook_ceph" {
  depends_on = [kubernetes_manifest.cert_manager_self_signed]
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "rook-ceph"
      namespace = "argocd"
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "ghcr.io/rook"
        chart          = "rook-ceph"
        targetRevision = "1.21.2"
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
        repoURL        = "https://ceph.github.io/ceph-csi-operator"
        chart          = "ceph-csi-drivers"
        targetRevision = "1.1.0"
        helm = {
          releaseName = "ceph-csi-drivers"
          values = yamlencode({
            operatorConfig = {
              driverSpecDefaults = {
                nodePlugin = {
                  priorityClassName = "system-node-critical"
                  resources = {
                    plugin = {
                      requests = { cpu = "50m", memory = "128Mi" }
                      limits   = { memory = "512Mi" }
                    }
                    registrar = {
                      requests = { cpu = "10m", memory = "32Mi" }
                      limits   = { memory = "128Mi" }
                    }
                    logRotator = {
                      requests = { cpu = "5m", memory = "8Mi" }
                      limits   = { memory = "64Mi" }
                    }
                  }
                }
                controllerPlugin = {
                  priorityClassName = "system-cluster-critical"
                  resources = {
                    plugin = {
                      requests = { cpu = "50m", memory = "64Mi" }
                      limits   = { memory = "256Mi" }
                    }
                    provisioner = {
                      requests = { cpu = "10m", memory = "128Mi" }
                      limits   = { memory = "256Mi" }
                    }
                    attacher = {
                      requests = { cpu = "10m", memory = "64Mi" }
                      limits   = { memory = "256Mi" }
                    }
                    resizer = {
                      requests = { cpu = "10m", memory = "64Mi" }
                      limits   = { memory = "256Mi" }
                    }
                    snapshotter = {
                      requests = { cpu = "10m", memory = "64Mi" }
                      limits   = { memory = "256Mi" }
                    }
                    logRotator = {
                      requests = { cpu = "5m", memory = "8Mi" }
                      limits   = { memory = "64Mi" }
                    }
                  }
                }
              }
            }
            drivers = {
              rbd = {
                name           = "rook-ceph.rbd.csi.ceph.com"
                snapshotPolicy = "volumeSnapshot"
                controllerPlugin = {
                  replicas = 2
                }
              }
              cephfs = {
                name = "rook-ceph.cephfs.csi.ceph.com"
                controllerPlugin = {
                  replicas = 2
                }
              }
              nfs    = { enabled = false }
              nvmeof = { enabled = false }
            }
          })
        }
        }, {
        repoURL        = "ghcr.io/rook"
        chart          = "rook-ceph-cluster"
        targetRevision = "1.21.2"
        helm = {
          releaseName = "rook-ceph"
          values = yamlencode({
            cephBlockPools = [{
              name = "standard-rwo"
              spec = {
                replicated = {
                  requireSafeReplicaSize = var.ceph_replica_count > 1
                  size                   = var.ceph_replica_count
                }
              }
              storageClass = {
                allowVolumeExpansion = true
                enabled              = true
                isDefault            = true
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
                    memory = "1Gi"
                  }
                }
              }
            }
            cephFileSystemVolumeSnapshotClass = {
              enabled   = true
              isDefault = true
            }
            cephFileSystems = [{
              name = "standard-rwx"
              spec = {
                dataPools = [{
                  name = "data0"
                  replicated = {
                    requireSafeReplicaSize = var.ceph_replica_count > 1
                    size                   = var.ceph_replica_count
                  }
                }]
                metadataPool = {
                  replicated = {
                    requireSafeReplicaSize = var.ceph_replica_count > 1
                    size                   = var.ceph_replica_count
                  }
                }
                metadataServer = {
                  activeCount       = 1
                  activeStandby     = var.ceph_replica_count > 1
                  priorityClassName = "system-cluster-critical"
                  placement = {
                    podAntiAffinity = {
                      requiredDuringSchedulingIgnoredDuringExecution = [{
                        labelSelector = { matchLabels = { app = "rook-ceph-mds", rook_file_system = "standard-rwx" } }
                        topologyKey   = "kubernetes.io/hostname"
                      }]
                    }
                  }
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
        repoURL        = "ghcr.io/piraeusdatastore/helm-charts"
        chart          = "snapshot-controller"
        targetRevision = "5.4.0"
        helm = {
          releaseName = "rook-ceph"
          values = yamlencode({
            controller = {
              resources = { requests = { cpu = "10m", memory = "64Mi" } }
            }
            webhook = {
              tls = { certManagerIssuerRef = { kind = "ClusterIssuer", name = "self-signed" } }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "rook-ceph" }
      ignoreDifferences = local.ignore_differences
      syncPolicy = merge(local.sync_policy, {
        automated = {
          enabled  = true
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
