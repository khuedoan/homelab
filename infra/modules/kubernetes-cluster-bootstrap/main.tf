resource "helm_release" "metallb" {
  name       = "metallb"
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "metallb"
  # TODO upgrade metallb helm version
  version    = "1.0.1"

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
  # TODO upgrade nginx helm version
  version    = "3.10.1"

  namespace        = "ingress-nginx"
  create_namespace = true
}

resource "helm_release" "cert_manager" {
  name       = "cert-manager"
  repository = "https://charts.jetstack.io"
  chart      = "cert-manager"
  # TODO upgrade cert-manager helm version
  version    = "1.0.4"

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
  # TODO upgrade prometheus stack helm version
  version    = "11.1.2"

  namespace        = "monitoring-system"
  create_namespace = true
}

resource "helm_release" "longhorn" {
  name       = "longhorn"
  repository = "https://charts.longhorn.io"
  chart      = "longhorn"
  # TODO upgrade longhorn helm version
  version    = "1.0.2"

  namespace        = "longhorn-system"
  create_namespace = true
}

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
