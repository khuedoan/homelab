variable "cloudflare_email" {
  type = string
}

variable "cloudflare_api_key" {
  type      = string
  sensitive = true
}

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
