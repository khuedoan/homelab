terraform {
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
      version = "~> 2.1.0"
    }
  }
}

provider "cloudflare" {
  email   = "josephkleinsorge@gmail.com"
  api_key = "a865618d79734cb2bb61200d22d6ce764d6bb"
}

provider "kubernetes" {
  # Use KUBE_CONFIG_PATH environment variables
  # Or in cluster service account
}
