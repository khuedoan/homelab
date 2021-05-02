resource "null_resource" "ansible_provisioner" {
  triggers = {
    ansible_hash = md5(join("", [for files in fileset("${var.ansible_directory}/", "**") : file("${var.ansible_directory}/${files}")]))
  }

  provisioner "local-exec" {
    command = "ansible-playbook --user ${var.ansible_user} --inventory ${join(",", var.ansible_inventory)}, --private-key ${var.ansible_private_key} ${path.ansible_directory}/${var.ansible_playbook}"

    environment = {
      ANSIBLE_HOST_KEY_CHECKING = "False"
    }
  }
}
