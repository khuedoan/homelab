variable "kubeconfig" {
  type = string
}

variable "repository" {
  type = object({
    url      = string
    revision = string
  })
}

variable "apps" {
  type    = set(string)
  default = []
}

variable "domain" {
  type = string
}

variable "resource_prefix" {
  type = string
}
