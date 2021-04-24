module "rke_cluster_bootstrap" {
  source      = "./modules/kubernetes-cluster-bootstrap"
  kube_config = rke_cluster.cluster.kube_config_yaml
}
