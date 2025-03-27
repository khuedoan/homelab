variable "name" {
  type    = string
  default = "external"
}

variable "namespace" {
  type    = string
  default = "global-secrets"
}

variable "data" {
  type = map(string)
}
