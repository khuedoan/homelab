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

    http = {
      source = "hashicorp/http"
      version = "~> 2.1.0"
    }
  }
}

provider "cloudflare" {
  # Environment variables
  # CLOUDFLARE_API_TOKEN
}

provider "b2" {
  # Environment variables
  # B2_APPLICATION_KEY
  # B2_APPLICATION_KEY_ID
}

provider "kubernetes" {
  config_path = "${path.root}/../metal/kubeconfig.yaml"
}
