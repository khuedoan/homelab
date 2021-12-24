variable "cloudflare_email" {
  type = string
}

variable "cloudflare_api_key" {
  type = string
  sensitive = true
}

variable "cloudflare_account_id" {
  type = string
}

variable "b2_application_key_id" {
  type = string
}

variable "b2_application_key" {
  type = string
  sensitive = true
}
