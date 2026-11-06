variable "cloudflare_account_id" {
  type = string
}

variable "resource_prefix" {
  type = string
}

variable "domain" {
  type = string
}

variable "zone_name" {
  type = string
}

variable "extra_records" {
  description = "Static DNS records keyed by name relative to domain."
  type = map(object({
    type    = string
    content = string
    proxied = optional(bool, false)
  }))
  default = {}
}
