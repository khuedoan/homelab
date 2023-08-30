terraform {
  required_version = "~> 1.3.0"

  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "homelab-external"

    workspaces {
      name = "homelab-external"
    }
  }

  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 3.11.0"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.9.0"
    }

    http = {
      source  = "hashicorp/http"
      version = "~> 3.2.0"
    }
  }
}

provider "cloudflare" {
  email   = var.cloudflare_email
  api_key = var.cloudflare_api_key
}

provider "kubernetes" {
  # Use KUBE_CONFIG_PATH environment variables
  # Or in cluster service account
}
