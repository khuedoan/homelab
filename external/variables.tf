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

variable "ntfy" {
  type = object({
    url   = string
    topic = string
  })

  sensitive = true
}

variable "extra_secrets" {
  type        = map(string)
  description = "Key-value pairs of extra secrets that cannot be randomly generated (e.g. third party API tokens)"
  sensitive   = true
  default     = {}
}
