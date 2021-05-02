variable "directory" {
  description = "Path to Ansible directory"
  type        = string
}

variable "playbook" {
  description = "Path to Ansible playbook, relative to Ansible directory"
  type        = string
  default     = "main.yml"
}

variable "user" {
  description = "User to connect as"
  type        = string
  default     = "ubuntu"
}

variable "inventory" {
  description = "List of hosts for Ansible to run against"
  type        = list(string)
}

variable "private_key" {
  description = "Private key file to authenticate the connection"
  type        = string
}
