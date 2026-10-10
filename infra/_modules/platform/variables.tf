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

variable "ceph_replica_count" {
  type    = number
  default = 2
  validation {
    condition     = var.ceph_replica_count >= 1 && floor(var.ceph_replica_count) == var.ceph_replica_count
    error_message = "Ceph replica count must be a positive integer."
  }
}
