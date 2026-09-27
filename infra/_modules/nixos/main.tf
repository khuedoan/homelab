resource "local_file" "hosts" {
  content         = jsonencode(var.hosts)
  filename        = "${var.flake}/hosts.json"
  file_permission = "600"
}
