terraform {
  backend "etcdv3" {
    lock = true
  }

  required_providers {
    rke = {
      source  = "rancher/rke"
      version = "1.2.2"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "2.2.0"
    }
  }
}
