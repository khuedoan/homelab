variable "flake" {
  type        = string
  description = "Path to the NixOS flake directory"
}

variable "hosts" {
  type        = any
  description = "Host inventory keyed by host name"
}
