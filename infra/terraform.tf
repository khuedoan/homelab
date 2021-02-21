terraform {
  backend "etcdv3" {
    endpoints = ["localhost:2379"]
    lock      = true
  }

  required_providers {
    lxd = {
      source  = "terraform-lxd/lxd"
      version = "1.5.0"
    }

    rke = {
      source  = "rancher/rke"
      version = "1.1.7"
    }
  }
}

provider "lxd" {
  generate_client_certificates = true
  accept_remote_certificate    = true
}

provider "rke" {
  debug = true
}
