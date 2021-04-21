resource "helm_release" "metallb" {
  name       = "metallb"
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "metallb"
  version    = "2.3.5"

  namespace        = "metallb-system"
  create_namespace = true

  set {
    name  = "configInline"
    # TODO use ./values/metallb.yaml for this
    value = <<EOT
      address-pools:
      - name: default
        protocol: layer2
        addresses:
        - 192.168.1.150-192.168.1.180
    EOT
  }
}

resource "helm_release" "nginx" {
  name       = "ingress-nginx"
  repository = "https://kubernetes.github.io/ingress-nginx"
  chart      = "ingress-nginx"
  version    = "3.29.0"

  namespace        = "ingress-nginx"
  create_namespace = true
}

# TODO (optimize) should cert manager be part of bootstrap?
resource "helm_release" "cert_manager" {
  name       = "cert-manager"
  repository = "https://charts.jetstack.io"
  chart      = "cert-manager"
  version    = "1.3.1"

  namespace        = "cert-manager"
  create_namespace = true

  # TODO use ./values/cert-manager.yaml for this
  set {
    name  = "installCRDs"
    value = "true"
  }
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
  chart      = "argocd"
  version    = "3.1.2"

  namespace        = "argocd"
  create_namespace = true
}

# TODO move Vault out of bootstrap
resource "helm_release" "vault" {
  name       = "vault"
  repository = "https://helm.releases.hashicorp.com"
  chart      = "vault"
  # TODO upgrade vault helm version
  version    = "0.8.0"

  namespace        = "vault"
  create_namespace = true

  # TODO HA Vault
  # TODO Auto unseal Vault
}

# TODO automatic ingress and tunnel for all services
