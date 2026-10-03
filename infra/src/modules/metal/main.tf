resource "local_file" "hosts" {
  content         = jsonencode(var.hosts)
  filename        = "${path.module}/flake/hosts.json"
  file_permission = "600"
}

resource "local_file" "cluster" {
  content         = jsonencode(var.cluster)
  filename        = "${path.module}/flake/cluster.json"
  file_permission = "600"
}
