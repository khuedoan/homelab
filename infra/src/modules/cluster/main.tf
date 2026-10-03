terraform {
  required_providers {
    external = {
      source  = "hashicorp/external"
      version = "~> 2.3"
    }
  }
}

variable "hosts" {
  type = map(object({ ip = string }))
}

variable "initializer" {
  type = string
}

variable "vip" {
  type = string
}

variable "kubeconfig" {
  type = string
}

locals {
  config = {
    hosts       = var.hosts
    initializer = var.initializer
    vip         = var.vip
    kubeconfig  = abspath(var.kubeconfig)
  }
}

data "external" "nodes" {
  program = ["python3", "${path.module}/cluster.py", "identities"]
  query   = { hosts = jsonencode(var.hosts) }
}

resource "terraform_data" "enroll" {
  triggers_replace = {
    config = local.config
    nodes  = data.external.nodes.result
  }
  provisioner "local-exec" {
    command     = "python3 \"${path.module}/cluster.py\" enroll"
    environment = { CLUSTER_CONFIG = jsonencode(local.config) }
  }
}

output "kubeconfig_path" {
  value      = local.config.kubeconfig
  depends_on = [terraform_data.enroll]
}
