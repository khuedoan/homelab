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
  }
}

provider "lxd" {
  generate_client_certificates = true
  accept_remote_certificate    = true

  lxd_remote {
    name     = "local"
    scheme   = "https"
    address  = "192.168.64.4"
    password = "1"
    default  = true
  }
}
