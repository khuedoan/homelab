resource "null_resource" "ansible_provisioner" {
  triggers = {
    ansible_hash = md5(join("", [for files in fileset("${var.directory}/", "**") : file("${var.directory}/${files}")]))
  }

  provisioner "local-exec" {
    command = "ansible-playbook --user ${var.user} --inventory ${join(",", var.inventory)}, --private-key ${var.private_key} ${var.directory}/${var.playbook}"

    environment = {
      ANSIBLE_HOST_KEY_CHECKING = "False"
    }
  }
}
