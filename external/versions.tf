terraform {
  required_version = "~> 1.1.0"

  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "khuedoan"

    workspaces {
      name = "homelab-external"
    }
  }

  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 3.4.0"
    }

    b2 = {
      source  = "Backblaze/b2"
      version = "~> 0.7.0"
    }

    kubernetes = {
      source = "hashicorp/kubernetes"
      version = "~> 2.7.0"
    }
  }
}
