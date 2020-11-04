terraform {
  required_version = ">= 0.13"

  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "khuedoan"

    workspaces {
      name = "homelab"
    }
  }

  required_providers {
    rke = {
      source = "rancher/rke"
      version = "1.1.3"
    }
  }
}
