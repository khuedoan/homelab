terraform {
  # TODO (feature) Generate endpoint automatically (terragrunt for variable)
  backend "etcdv3" {
    endpoints = [
      "192.168.1.110:23799"
    ]
    lock = true
  }

  required_providers {
    rke = {
      source  = "rancher/rke"
      version = "1.2.2"
    }
  }
}
