terraform {
  required_providers {
    lxd = {
      source  = "terraform-lxd/lxd"
      version = "1.5.0"
    }

    rke = {
      source  = "rancher/rke"
      version = "1.1.7"
    }

    helm = {
      source = "hashicorp/helm"
      version = "2.1.1"
    }
  }
}
