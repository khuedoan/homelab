terraform {
  required_providers {
    helm = {
      source = "hashicorp/helm"
      version = "2.1.1"
    }
  }
}

provider "helm" {
  kubernetes {
    host                   = yamldecode(var.kube_config)["clusters"][0]["cluster"]["server"]
    cluster_ca_certificate = base64decode(yamldecode(var.kube_config)["clusters"][0]["cluster"]["certificate-authority-data"])
    client_certificate     = base64decode(yamldecode(var.kube_config)["users"][0]["user"]["client-certificate-data"])
    client_key             = base64decode(yamldecode(var.kube_config)["users"][0]["user"]["client-key-data"])
  }
}
