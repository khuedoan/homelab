resource "lxd_container" "test1" {
  name      = "test1"
  image     = "ubuntu:20.04"
  ephemeral = false

  config = {
    "boot.autostart" = true
  }

  limits = {
    cpu = 2
  }
}
