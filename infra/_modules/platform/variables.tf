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
  type = map(object({
    path      = string
    namespace = string
    values    = string
  }))
  default = {}
}
