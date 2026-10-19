variable "kubeconfig" {
  type = string
}

variable "control_plane_ips" {
  type = list(string)
}

variable "repository" {
  type = object({
    url      = string
    revision = string
  })
}

variable "backup_location" {
  type = object({
    bucket   = string
    endpoint = string
  })
}

variable "backup_read_only" {
  type        = bool
  default     = false
  description = "Prevent backup creation and deletion during recovery and disable the backup schedule."
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

variable "load_balancer_cidr" {
  type        = string
  description = "Application LoadBalancer subnet advertised by kube-vip and Tailscale."
}
