variable "hosts" {
  type        = any
  description = "Host inventory keyed by host name"
}

variable "cluster" {
  type = object({
    init_host = string
    vip       = string
  })
  description = "Cluster initializer and control-plane virtual IP"
}
