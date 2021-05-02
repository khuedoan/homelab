resource "helm_release" "metallb" {
  name       = "metallb"
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "metallb"
  version    = "2.3.5"

  namespace        = "metallb-system"
  create_namespace = true

  values = [
    file("${path.module}/values/metallb.yaml")
  ]
}

resource "helm_release" "nginx" {
  name       = "ingress-nginx"
  repository = "https://kubernetes.github.io/ingress-nginx"
  chart      = "ingress-nginx"
  version    = "3.29.0"

  namespace        = "ingress-nginx"
  create_namespace = true
}

resource "helm_release" "cert_manager" {
  name       = "cert-manager"
  repository = "https://charts.jetstack.io"
  chart      = "cert-manager"
  version    = "1.3.1"

  namespace        = "cert-manager"
  create_namespace = true

  values = [
    file("${path.module}/values/cert-manager.yaml")
  ]
}

resource "helm_release" "prometheus" {
  name       = "kube-prometheus-stack"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  version    = "15.1.1"

  namespace        = "monitoring-system"
  create_namespace = true
}

resource "helm_release" "longhorn" {
  name       = "longhorn"
  repository = "https://charts.longhorn.io"
  chart      = "longhorn"
  version    = "1.1.0"

  namespace        = "longhorn-system"
  create_namespace = true
}

resource "helm_release" "argocd" {
  name       = "argocd"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  version    = "3.1.2"

  namespace        = "argocd"
  create_namespace = true
}

resource "helm_release" "vault" {
  name       = "vault"
  repository = "https://helm.releases.hashicorp.com"
  chart      = "vault"
  version    = "0.11.0"

  namespace        = "vault"
  create_namespace = true

  values = [
    file("${path.module}/values/vault.yaml")
  ]

  # TODO (optimize) HA Vault and auto unseal Vault
}

# TODO (feature) Automatic ingress and tunnel for all services
