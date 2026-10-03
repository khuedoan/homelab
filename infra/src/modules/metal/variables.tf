variable "hosts" {
  type        = any
  description = "Host inventory keyed by host name"
}

variable "cluster" {
  type = object({
    init_host          = string
    vip                = string
    load_balancer_cidr = string
    network = object({
      prefix_length = number
      gateway       = string
    })
  })
  description = "Cluster initializer and control-plane virtual IP"
}
