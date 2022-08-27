variable "zerotier_central_url" {
  description = "ZeroTier Central API endpoint"
  type        = string
  default     = "https://my.zerotier.com/api" # https://github.com/zerotier/go-ztcentral/blob/4d397d1e82c043a5376789177ad55536044d69ce/client.go#L44
}

variable "zerotier_central_token" {
  description = "ZeroTier Central API Token"
  type        = string
  sensitive   = true
}

variable "name" {
  description = "Network name"
  type        = string
  default     = "homelab"
}

variable "description" {
  description = "Network description"
  type        = string
  default     = "Homelab network"
}

variable "managed_route" {
  description = "ZeroTier managed route"
  type        = string
  default     = "10.147.17.0/24"
}

variable "bridged_routes" {
  description = "List of bridged routes" # TODO
  type        = list(string)
}
