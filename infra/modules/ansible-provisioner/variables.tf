var "ansible_directory" {
  description = "Path to Ansible directory"
  type        = string
}

var "ansible_playbook" {
  description = "Path to Ansible playbook, relative to Ansible directory"
  type        = string
  default     = "main.yml"
}

var "ansible_user" {
  description = "User to connect as"
  type        = string
  default     = "ubuntu"
}

var "ansible_inventory" {
  description = "List of hosts for Ansible to run against"
  type        = list(string)
}

var "ansible_private_key" {
  description = "Private key file to authenticate the connection"
  type        = list(string)
}
